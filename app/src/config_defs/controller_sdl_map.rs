use serde::{Deserialize, Serialize};

use super::serde_sdl_guid::SDLGuid;

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
    pub id: SDLGuid,
    pub name: String,
    pub data: Vec<ControllerSdlMapControl>,
}

impl ControllerSdlMap {
    /**
     * Finds the mapping information by an SDL index and kind
     */
    pub fn find_by_sdl_index(
        &self,
        kind: &SDLControlKind,
        index: u8,
    ) -> Option<ControllerSdlMapControl> {
        self.data
            .iter()
            .find(|c| &c.kind == kind && c.index == index)
            .cloned()
    }

    /**
     * Finds the mapping information by it's friendly name
     */
    pub fn find_by_name<T: AsRef<str>>(&self, name: T) -> Option<ControllerSdlMapControl> {
        self.data.iter().find(|c| c.name == name.as_ref()).cloned()
    }
}
