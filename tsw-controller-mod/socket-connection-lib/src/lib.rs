pub(crate) mod direct_controller_task;
pub(crate) mod sync_controller_task;

use once_cell::sync::Lazy;
use tokio::runtime::Runtime;
static TOKIO_RUNTIME: Lazy<Runtime> = Lazy::new(|| {
    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .expect("Failed to create runtime")
});

static DIRECT_CONTROLLER_TASK: Lazy<direct_controller_task::DirectControllerTask> =
    Lazy::new(|| direct_controller_task::DirectControllerTask::new(&TOKIO_RUNTIME));

static SYNC_CONTROLLER_TASK: Lazy<sync_controller_task::SyncControllerTask> =
    Lazy::new(|| sync_controller_task::SyncControllerTask::new(&TOKIO_RUNTIME));

#[repr(C)]
pub struct ControlValue {
    pub direct_controller: &'static direct_controller_task::DirectControllerTask,
    pub sync_controller: &'static sync_controller_task::SyncControllerTask,
}

#[no_mangle]
pub unsafe extern "C" fn start() {
    DIRECT_CONTROLLER_TASK.spawn_dc_listener_task();
    DIRECT_CONTROLLER_TASK.spawn_queue_propagation_task();
    SYNC_CONTROLLER_TASK.spawn_sc_forwarding_task();
}

#[no_mangle]
pub unsafe extern "C" fn set_direct_controller_callback(
    callback: extern "C" fn(*const std::ffi::c_char),
) {
    DIRECT_CONTROLLER_TASK.set_callback(callback);
}

#[no_mangle]
pub unsafe extern "C" fn send_sync_controller_message(message: *const std::ffi::c_char) {
    SYNC_CONTROLLER_TASK.send(message);
}
