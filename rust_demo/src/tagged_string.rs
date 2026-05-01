use rmp_serde::{Deserializer, Serializer};
use serde::{Deserialize, Serialize};

#[derive(Debug, PartialEq, Deserialize, Serialize)]
pub struct TaggedString {
    pub id: u32,
    pub s: String,
}
