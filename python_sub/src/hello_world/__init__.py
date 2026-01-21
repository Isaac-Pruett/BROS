import os
import sys
from time import sleep

import zenoh


def main():
    with zenoh.open(zenoh.Config().from_env()) as session:
        pub = session.declare_publisher("python/helloworld")

        sub = session.declare_subscriber("rust/helloworld")

        # Wait for subscribers to be ready
        sleep(0.5)

        # Now publish
        pub.put("Hello, from Python!")
        print("Python → Published")

        sample = sub.recv()

        print("Python ← Received:", sample.payload.to_string())

        print("Python → Waiting for Rust message...")

        print("Python done!")
        session.close()


if __name__ == "__main__":
    main()
