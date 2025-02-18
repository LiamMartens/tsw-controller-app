use mlua::{Lua, Result, Table};
use once_cell::sync::Lazy;

use tokio::runtime::Runtime;
static TOKIO_RUNTIME: Lazy<Runtime> = Lazy::new(|| tokio::runtime::Builder::new_multi_thread().enable_all().build().expect("Failed to create runtime"));

pub(crate) mod direct_controller_task;
pub(crate) mod sync_controller_task;

fn lib_init(lua: &Lua) -> Result<Table> {
    let table = lua.create_table().unwrap();
    let direct_controller_task = direct_controller_task::DirectControllerTask::connect(&TOKIO_RUNTIME, lua);
    let sync_controller_task = sync_controller_task::SyncControllerTask::connect(&TOKIO_RUNTIME, lua);
    table.set("direct_controller_task", direct_controller_task).unwrap();
    table.set("sync_controller_task", sync_controller_task).unwrap();
    Ok(table)
}

#[no_mangle]
pub unsafe extern "C" fn luaopen_tsw5_gamepad_lua_socket_connection(state: *mut mlua::lua_State) -> std::os::raw::c_int {
    mlua::Lua::entrypoint1(state, move |lua| lib_init(lua))
}
