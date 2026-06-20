import time

import zenoh

from .tagged_string import TaggedString


def main():
    with zenoh.open(zenoh.Config()) as session:
        publisher = session.declare_publisher("demo/out/py")
        subscriber = session.declare_subscriber("demo/out/*")

        time.sleep(0.5)

        msg = TaggedString(id=67, s="hello from python!")
        publisher.put(msg.to_msgpack())
        print(f"Python Sent: {msg}")

        deadline = time.time() + 6
        received_any = False

        while time.time() < deadline:
            sample = subscriber.try_recv()

            if sample is not None:
                received = TaggedString.from_msgpack(bytes(sample.payload))
                print(f"Python Received: {received}")
                received_any = True
                continue

            time.sleep(0.05)

        if not received_any:
            print("Timeout: no message received")


if __name__ == "__main__":
    main()
