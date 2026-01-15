import sys
from time import sleep

import zenoh

has_rx = False


# a callback to run by the subscriber
def listen(sample: zenoh.Sample):
    global has_rx
    print("Python ← Received:", sample.payload.to_string())
    has_rx = True


def main():
    global has_rx
    with zenoh.open(zenoh.Config().from_env()) as session:
        pub = session.declare_publisher("python/helloworld")

        sub = session.declare_subscriber("rust/helloworld", listen)

        # Wait for subscribers to be ready
        sleep(0.5)

        # Now publish
        pub.put("Hello, from Python!")
        print("Python → Published")

        print("Python → Waiting for Rust message...")
        while not has_rx:
            pass
        print("Python done!")
        session.close()


if __name__ == "__main__":
    main()
