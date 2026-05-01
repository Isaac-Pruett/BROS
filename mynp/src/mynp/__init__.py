import asyncio
import time

import msgpack
import zenoh

from .tagged_string import TaggedString


def main():
    with zenoh.open(zenoh.Config()) as session:
        publisher = session.declare_publisher("demo/out_py")
        subscriber = session.declare_subscriber("demo/out_rs")

        time.sleep(2)  # wait for subs

        msg = TaggedString(id=42, s="hello from python!")
        publisher.put(msg.to_msgpack())
        print(f"Sent: {msg}")

        # Wait up to 6 seconds for a reply
        sample = subscriber.recv()
        if sample is not None:
            received = TaggedString.from_msgpack(bytes(sample.payload))
            print(f"Received: {received}")
        else:
            print("Timeout: no message received")


if __name__ == "__main__":
    main()
