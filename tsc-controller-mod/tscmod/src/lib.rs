use futures_util::{SinkExt, StreamExt};
use libc::{c_char, c_float, c_int};
use libloading::Library;
use once_cell::sync::Lazy;
use std::collections::HashMap;
use std::ffi::{CStr};
use std::path::PathBuf;
use std::sync::{Arc, Mutex, RwLock};
use std::time::{Duration, Instant};
use tokio::runtime::Runtime;
use tokio::sync::broadcast;
use tokio::sync::mpsc::{self};
use tokio_tungstenite::connect_async;
use tungstenite::{protocol::Message, Utf8Bytes};
use windows::Win32::Foundation::HMODULE;
use windows::Win32::System::LibraryLoader::{GetModuleFileNameW};

static WS_PORT_OPTIONS: &[u16] = &[63241, 63242, 63243];

struct DLLLocoStateControlTarget {
    value: c_float,
    max_change_rate: c_float,
    hold: bool
}

struct DLLLocoStateDrivableActor {
    name: String,
    lastsent: Instant
}

struct DLLLocoState {
    drivable: DLLLocoStateDrivableActor,
    controls: HashMap<String, usize>,
    controlvalues: HashMap<String, c_float>,
    controltargetvalues: HashMap<String, DLLLocoStateControlTarget>
}

struct DLLState {
    rt: Option<Runtime>,
    stop_tx: Option<Arc<broadcast::Sender<()>>>,
    outgoing_tx: Option<mpsc::Sender<String>>,
    loco: Option<DLLLocoState>,
    current_port_index: usize,
}

static STATE: Lazy<Arc<RwLock<DLLState>>> = Lazy::new(|| {
    Arc::new(RwLock::new(DLLState {
        rt: None,
        stop_tx: None,
        outgoing_tx: None,
        loco: None,
        current_port_index: 0,
    }))
});

fn module_path_from_hmodule(hmodule: HMODULE) -> Option<PathBuf> {
    let mut buffer = vec![0u16; 260];

    let len = unsafe { GetModuleFileNameW(Some(hmodule), &mut buffer) };

    if len == 0 {
        return None;
    }

    buffer.truncate(len as usize);
    Some(PathBuf::from(String::from_utf16_lossy(&buffer)))
}

unsafe fn get_loco_name(lib: &Library) -> String {
    let loconame_raw = lib.get::<unsafe extern "C" fn() -> *const c_char>(b"GetLocoName").unwrap()();
    let cstr = CStr::from_ptr(loconame_raw);
    match cstr.to_str() {
        Ok(v) => v.to_string(),
        _ => String::new(),
    }
}

unsafe fn get_controller_list(lib: &Library) -> HashMap<String, usize> {
    let controllerlist_raw = lib.get::<unsafe extern "C" fn() -> *const c_char>(b"GetControllerList").unwrap()();
    let cstr = CStr::from_ptr(controllerlist_raw);

    match cstr.to_str() {
        Ok(v) => {
            if v.is_empty() {
                return HashMap::new();
            }

            let mut map = HashMap::<String, usize>::new();
            for (index, control_name) in v.split("::").enumerate() {
                map.insert(control_name.to_string(), index);
            }
            map
        }
        _ => HashMap::new(),
    }
}

unsafe fn get_controller_value(lib: &Library, index: c_int, value_type: c_int) -> c_float {
    lib.get::<unsafe extern "C" fn(c_int, c_int) -> c_float>(b"GetControllerValue").unwrap()(index, value_type)
}

unsafe fn set_controller_value(lib: &Library, index: c_int, value: c_float) {
    lib.get::<unsafe extern "C" fn(c_int, c_float) -> c_float>(b"SetControllerValue").unwrap()(index, value);
}

pub fn mod_init(hmod: HMODULE) {
    println!("[tscmod][info] initializing tscmod");

    let dllpath = module_path_from_hmodule(hmod).unwrap();
    let dlldir = dllpath.parent().unwrap();
    let raildriverpath = dlldir.join("RailDriver64.dll");

    let mut st = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
    if st.rt.is_some() {
        return; // already running
    }

    // load raildriver lib
    let lib = unsafe {
        let lib = libloading::Library::new(raildriverpath).unwrap();
        lib.get::<unsafe extern "C" fn(bool)>(b"SetRailDriverConnected").unwrap()(true);
        Arc::new(Mutex::new(lib))
    };

    // create tokio runtime
    let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().expect("Failed to create runtime");

    // create channels
    let (stop_tx, _) = broadcast::channel::<()>(1);
    let (out_tx, mut out_rx) = mpsc::channel::<String>(64);
    let stop_tx_arc = Arc::new(stop_tx);

    let socket_thread_stop_tx = Arc::clone(&stop_tx_arc);
    rt.spawn(async move {
        loop {
            let current_port = STATE.read().unwrap_or_else(
                |poisoned| poisoned.into_inner()
            ).current_port_index;
            let ws_url = format!("ws://127.0.0.1:{}", WS_PORT_OPTIONS[current_port]);
            println!("[tscmod][info] attempting to connect to socket on port {}", WS_PORT_OPTIONS[current_port]);
            let mut sockst_stop_rx = socket_thread_stop_tx.subscribe();
            tokio::select! {
                _ = sockst_stop_rx.recv() => {
                    break;
                }
                conect_res = connect_async(ws_url.as_str()) => {
                    match conect_res {
                        Ok((ws_stream, response)) => {
                            let header = response.headers().get("X-TSW-Version");
                            if header.is_none() || header.unwrap().is_empty() {
                                println!("[socket_connection_lib][error] connected to unknown socket server - switching and retrying in 3s");
                                tokio::time::sleep(std::time::Duration::from_secs(3)).await;
                                /* update port index to next one */
                                let mut state_guard = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
                                state_guard.current_port_index = (state_guard.current_port_index + 1) % WS_PORT_OPTIONS.len();
                                continue;
                            }

                            let (mut ws_write, mut ws_read) = ws_stream.split();

                            let (reconnect_tx, mut reconnect_rx) = mpsc::channel::<()>(1);

                            // Forward incoming WS messages to callback
                            let client_read_loop_stop_tx = Arc::clone(&socket_thread_stop_tx);
                            tokio::spawn(async move {
                                loop {
                                    let mut client_read_loop_stop_rx = client_read_loop_stop_tx.subscribe();

                                    tokio::select! {
                                        _ = client_read_loop_stop_rx.recv() => {
                                            break;
                                        },
                                        Some(Ok(msg)) = ws_read.next() => {
                                            match msg {
                                                tungstenite::Message::Text(text) => {
                                                    let msg_split: Vec<&str> = text.split(",").collect();
                                                    if msg_split[0] == "direct_control" {
                                                        /* collect properties from direct control message */
                                                        let mut properties = HashMap::<&str, &str>::new();
                                                        for part in msg_split.iter().skip(1) {
                                                            let valuesplit: Vec<&str> = part.split("=").collect();
                                                            if valuesplit.len() == 2 {
                                                                properties.insert(valuesplit[0], valuesplit[1]);
                                                            }
                                                        }

                                                        /* now apply value */
                                                        if properties.contains_key("controls") && properties.contains_key("value") {
                                                            let mut guard = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
                                                            let st = &mut *guard;
                                                            if st.loco.is_some() {
                                                                let value: f32 = properties["value"].parse().unwrap();
                                                                let max_change_rate: f32 = match properties.contains_key("max_change_rate") {
                                                                    true => properties["max_change_rate"].parse().unwrap(),
                                                                    false => 999.0f32 /* 999 should be more than enough */
                                                                };
                                                                let hold: bool = match properties.contains_key("flags") {
                                                                    true => properties["flags"].split(',').any(|s| s.trim().contains("hold")),
                                                                    false => false
                                                                };
                                                                st.loco.as_mut().unwrap().controltargetvalues.insert(
                                                                    properties["controls"].to_string(),
                                                                    DLLLocoStateControlTarget { value, max_change_rate, hold }
                                                                );
                                                            }
                                                        }
                                                    }
                                                },
                                                tungstenite::Message::Close(_) => {
                                                    break;
                                                },
                                                _ => {},
                                            }
                                        }
                                    }
                                }
                                /* if this while ends - the read resulted in an error - try send reconnect_tx */
                                println!("[socket_connection_lib][info] closing connection and reconnecting due to error or close message");
                                let _ = reconnect_tx.try_send(());
                            });

                            // Outgoing loop
                            let mut client_write_loop_stop_rx = socket_thread_stop_tx.subscribe();
                            loop {
                                tokio::select! {
                                    Some(msg) = out_rx.recv() => {
                                      println!("[socket_connection_lib][info] sending message | {}", msg);
                                        if let Err(e) = ws_write.send(Message::Text(Utf8Bytes::from(msg))).await {
                                           println!("[socket_connection_lib][info] failed to send message | {}", e);
                                            break; // reconnect
                                        }
                                    }
                                    _ = reconnect_rx.recv() => {
                                      break;
                                    },
                                    _ = client_write_loop_stop_rx.recv() => {
                                        let _ = ws_write.send(Message::Close(None)).await;
                                        return;
                                    }
                                }
                            }
                        }
                        Err(e) => {
                            println!("[socket_connection_lib][error] failed to connect to socket - retrying in 3s | {}", e);
                            tokio::time::sleep(std::time::Duration::from_secs(3)).await;
                            /* update port index to next one */
                            let mut state_guard = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
                            state_guard.current_port_index = (state_guard.current_port_index + 1) % WS_PORT_OPTIONS.len();
                            continue;
                        }
                    }
                }
            }
        }
    });

    let read_state_lib = Arc::clone(&lib);
    let read_state_thread_stop_tx = Arc::clone(&stop_tx_arc);
    rt.spawn(async move {
        unsafe {
            let mut read_state_stop_rx = read_state_thread_stop_tx.subscribe();
            loop {
                tokio::select! {
                    _ = read_state_stop_rx.recv() => {
                        break;
                    }
                    _ = tokio::time::sleep(Duration::from_millis(300)) => {
                        let lib = read_state_lib.lock().unwrap();
                        let mut guard = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
                        let st = &mut *guard;

                        let loconame = get_loco_name(&lib);
                        if st.loco.is_none() || (
                            /* send name if changed */
                            st.loco.as_ref().unwrap().drivable.name != loconame
                            /* or it's been 2 secs - just to make sure */
                            || st.loco.as_ref().unwrap().drivable.lastsent.elapsed() > Duration::from_secs(2)
                         ) {
                            let controls = get_controller_list(&lib);
                            let loco = DLLLocoState {
                                drivable:  DLLLocoStateDrivableActor{
                                    name: loconame.to_string(),
                                    lastsent: Instant::now(),
                                },
                                controls: controls,
                                /* this will reset the controlvalues and controltarget values */
                                controlvalues: HashMap::new(),
                                controltargetvalues: HashMap::new()
                            };
                            st.loco = Some(loco);

                            let drivable_msg = format!("current_drivable_actor,name={}", loconame);
                            if let Some(tx) = &st.outgoing_tx {
                                let drivable_send_result = tx.try_send(drivable_msg);
                                if let Err(e) = drivable_send_result {
                                    println!("[tscmod][error] failed to send message {}", e.to_string());
                                }
                            }
                        }

                        let loco = st.loco.as_mut().unwrap();
                        for (control_name, index) in loco.controls.iter() {
                            let controlvalue = get_controller_value(&lib, (*index) as c_int, libraildriver::Kind::Current as c_int);
                            if loco.controlvalues.contains_key(control_name) && loco.controlvalues[control_name] == controlvalue {
                                /* skip sending if value is unchanged */
                                continue;
                            }

                            loco.controlvalues.insert(control_name.to_string(), controlvalue);
                            let msg = format!("sync_control_value,name={},property={},value={},normal_value={}", control_name, control_name, controlvalue, controlvalue);
                            if let Some(tx) = st.outgoing_tx.as_ref() {
                                let send_result = tx.try_send(msg);
                                if let Err(e) = send_result {
                                    println!("[tscmod][error] failed to send message {}", e.to_string());
                                }
                            }
                        }
                    }
                }
            }
        }
    });

    let control_tick_lib = Arc::clone(&lib);
    let control_tick_thread_stop_tx = Arc::clone(&stop_tx_arc);
    rt.spawn(async move {
        let mut control_tick_thread_stop_rx = control_tick_thread_stop_tx.subscribe();
        loop {
            tokio::select! {
                _ = control_tick_thread_stop_rx.recv() => {
                    break;
                }
                _ = tokio::time::sleep(Duration::from_millis(33)) => {
                    /* check state */
                    let lib = control_tick_lib.lock().unwrap();
                    let mut guard = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
                    let st = &mut *guard;
                    if !st.loco.is_some() {
                        continue;
                    }

                    if let Some(loco) = guard.loco.as_mut() {
                        let controltargetvalues_keys = loco.controltargetvalues.keys().cloned().collect::<Vec<String>>();
                        for key in controltargetvalues_keys.iter() {
                            if !loco.controls.contains_key(key) {
                            /* skip and delete from targets if not available in loco */
                                loco.controltargetvalues.remove(key);
                                continue;
                            }

                            unsafe {
                                let target_state = loco.controltargetvalues.get_mut(key).unwrap();
                                let control_index = loco.controls[key];
                                let currentvalue = get_controller_value(&lib, control_index as c_int, libraildriver::Kind::Current as c_int);
                                let delta = target_state.value - currentvalue;
                                let next_value = match delta > 0.0 {
                                    true => currentvalue + delta.abs().min(target_state.max_change_rate),
                                    false => currentvalue - delta.abs().min(target_state.max_change_rate)
                                };
                                set_controller_value(&lib, control_index as c_int, next_value as c_float);
                                if !target_state.hold && (next_value - target_state.value).abs() < 0.05f32 {
                                    /* has reached target value within margin of error of 0.05 */
                                    loco.controltargetvalues.remove(key);
                                }
                            }
                        }
                    }
                }
            }
        }
    });

    st.rt = Some(rt);
    st.stop_tx = Some(stop_tx_arc);
    st.outgoing_tx = Some(out_tx);
}

pub fn mod_destroy() {
    let mut st = STATE.write().unwrap_or_else(|poisoned| poisoned.into_inner());
    if let Some(stop_tx) = st.stop_tx.take() {
        let _ = stop_tx.send(());
    }
    st.rt.take().unwrap().shutdown_background(); // dropping runtime shuts it down
}

#[no_mangle]
#[cfg(target_os = "windows")]
pub extern "system" fn DllMain(hmod: HMODULE, fwd_reason: u32, _lp_reserved: *mut u8) -> i32 {
    /* DLL_PROCESS_ATTACH */
    if fwd_reason == 1 {
        mod_init(hmod);
    }

    /* DLL_PROCESS_DETACH */
    if fwd_reason == 0 {
        mod_destroy();
    }

    1
}
