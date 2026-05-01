use zenoh_ext::Serialize;

struct Packet {
    x: f32,
    y: f32,
    z: f32,
}

impl Serialize for Packet {
    fn serialize(&self, serializer: &mut zenoh_ext::ZSerializer) {
        serializer.serialize(vec![self.x, self.y, self.z]);
    }
}
