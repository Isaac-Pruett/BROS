from time import sleep

import zenoh


def main():
    with zenoh.open(zenoh.Config()) as session:
        pub = session.declare_publisher("python/helloworld")
        sub = session.declare_subscriber("rust/helloworld")

        # Wait for subscribers to be ready
        sleep(0.5)

        # Now publish
        pub.put("Hello, from Python!")
        print("Python → Published")

        print("Python → Waiting for Rust message...")
        try:
            sample = sub.recv()
            print("Python ← Received:", sample.payload.to_string())
        except TimeoutError:
            print("Python ← Timeout waiting for Rust")

        print("Python done!")


if __name__ == "__main__":
    main()
