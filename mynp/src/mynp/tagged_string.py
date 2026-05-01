from dataclasses import dataclass


@dataclass
class TaggedString:
    id: int
    s: str

    def to_msgpack(self) -> bytes:
        return msgpack.packb([self.id, self.s])

    @classmethod
    def from_msgpack(cls, data: bytes) -> "TaggedString":
        vals = msgpack.unpackb(data)
        return cls(id=vals[0], s=vals[1])
