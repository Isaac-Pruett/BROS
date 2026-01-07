from time import sleep

import zenoh


def hello() -> None:
    # print("INIT")
    with zenoh.open(zenoh.Config()) as session:
        # print("python opened the zenoh session")
        publ = session.declare_publisher("python/helloworld")

        subs = session.declare_subscriber("rust/helloworld")

        publ.put("Hello, from Python!")
        print("python put the msg")
        sleep(3)

        packet = subs.recv()
        print("python recienved a massage")
        # print(packet)

        payload = packet.payload

        print(payload.to_string())

    print("Hello from hello-world!")
    sleep(4)
    print("haha, did async work?")
