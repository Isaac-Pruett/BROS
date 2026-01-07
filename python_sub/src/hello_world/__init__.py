from time import sleep

import zenoh


def main():
    with zenoh.open(zenoh.Config()) as session:
        pub = session.declare_publisher("python/helloworld")
        sub = session.declare_subscriber("rust/helloworld", listen)

        # Wait for subscribers to be ready
        sleep(0.5)

        # Now publish
        pub.put("Hello, from Python!")
        print("Python → Published")

        print("Python → Waiting for Rust message...")
        sleep(2)

        print("Python done!")
        session.close()


# a callback to run by the subscriber
def listen(sample: zenoh.Sample):
    print("Python ← Received:", sample.payload.to_string())


if __name__ == "__main__":
    main()
