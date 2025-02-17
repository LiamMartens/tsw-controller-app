use mlua::{Lua, Result, Table};
use once_cell::sync::Lazy;
use std::sync::Arc;
use tokio::runtime::Runtime;
use tokio_util::sync::CancellationToken;
mod module;

static TOKIO_RUNTIME: Lazy<Runtime> = Lazy::new(|| tokio::runtime::Builder::new_multi_thread().enable_all().build().expect("Failed to create runtime"));

fn init(lua: Arc<&Lua>) -> Result<Table> {
    let exports: Table = lua.create_table()?;

    exports.set(
        "init",
        lua.create_function(move |lua, ()| {
            println!("[TSW5GamepadMod] Starting new instance");

            /* create new module instance */
            let lua = Arc::new(lua);
            let cancel_token = CancellationToken::new();
            let table = module::Module::init(lua, &TOKIO_RUNTIME, cancel_token).unwrap();

            Ok(table)
        })?,
    )?;

    Ok(exports)
}

#[no_mangle]
pub unsafe extern "C" fn luaopen_tsw5_gamepad_lua_socket_connection(state: *mut mlua::lua_State) -> std::os::raw::c_int {
    mlua::Lua::entrypoint1(state, move |lua| {
        let lua_arc = Arc::new(lua);
        let table = init(lua_arc).unwrap();
        Ok(table)
    })
}
