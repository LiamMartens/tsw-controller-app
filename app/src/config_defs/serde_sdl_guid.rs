use core::fmt;
use std::fmt::Formatter;

use sdl2::joystick::Guid;
use serde::{Deserialize, Deserializer, Serialize, Serializer};

#[derive(Clone, PartialEq)]
pub struct SDLGuid {
    data: Guid,
}

impl SDLGuid {
    pub fn new<T: AsRef<str>>(data: T) -> Self {
        SDLGuid {
            data: Guid::from_string(data.as_ref()).unwrap(),
        }
    }

    pub fn string(&self) -> String {
        self.data.string()
    }

    pub fn guid(&self) -> Guid {
        self.data
    }
}

impl fmt::Display for SDLGuid {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(f, "{:?}", self.data.string())
    }
}

impl fmt::Debug for SDLGuid {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(f, "{:?}", self.data.string())
    }
}

impl PartialEq<Guid> for SDLGuid {
    fn eq(&self, other: &Guid) -> bool {
        self.data == *other
    }
}

impl<'de> Deserialize<'de> for SDLGuid {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        Ok(SDLGuid {
            data: Guid::from_string(s.as_ref()).unwrap(),
        })
    }
}

impl Serialize for SDLGuid {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_str(self.data.string().as_ref())
    }
}
