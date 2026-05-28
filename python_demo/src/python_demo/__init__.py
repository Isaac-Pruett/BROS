import asyncio
import time

import msgpack
import zenoh

from .tagged_string import TaggedString


def main():
    with zenoh.open(zenoh.Config()) as session:
        publisher = session.declare_publisher("demo/out_py")
        subscriber = session.declare_subscriber("demo/out_rs")

        time.sleep(0.5)  # wait for subs

        msg = TaggedString(id=67, s="hello from python!")
        publisher.put(msg.to_msgpack())
        print(f"Sent: {msg}")

        deadline = time.time() + 6

        while time.time() < deadline:
            sample = subscriber.try_recv()
            if sample is not None:
                received = TaggedString.from_msgpack(bytes(sample.payload))
                print(f"Received: {received}")
                break
            time.sleep(0.05)
        else:
            print("Timeout: no message received")


if __name__ == "__main__":
    main()
