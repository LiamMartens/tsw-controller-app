use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
#[serde(rename_all = "snake_case")]
pub enum SDLControlKind {
    Button,
    Hat,
    Axis,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerSdlMapControl {
    /** matches the SDL kinds like b, h and a - this mapping will be similar to the gamepad mapping ie: a:b0 -> map "a" to button index 0 */
    pub kind: SDLControlKind,
    pub index: u8,
    /** this is the friendly name of the controller; such as "throttle1" */
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ControllerSdlMap {
    pub name: String,
    /* {0xVENDOR_ID}:{0xPRODUCT_ID} */
    pub usb_id: String,
    pub data: Vec<ControllerSdlMapControl>,
}
