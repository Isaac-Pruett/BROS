from dataclasses import dataclass

import msgpack


@dataclass
class TaggedString:
    id: int
    s: str

    def to_msgpack(self) -> bytes | None:
        v = msgpack.packb([self.id, self.s])
        if v is not None:
            return v

    @classmethod
    def from_msgpack(cls, data: bytes) -> "TaggedString":
        vals = msgpack.unpackb(data)
        return cls(id=vals[0], s=vals[1])
