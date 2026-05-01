use std::time::Duration;
use zenoh::Config;

use rmp_serde::{Deserializer, Serializer};
use serde::{Deserialize, Serialize};

mod tagged_string;
use tagged_string::TaggedString;

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let session = zenoh::open(Config::default()).await?;
    let publisher = session.declare_publisher("demo/out_rs").await?;
    let subscriber = session.declare_subscriber("demo/out_py").await?;

    tokio::time::sleep(Duration::from_secs(2)).await; //wait on subs

    let packet = TaggedString {
        id: 42,
        s: "hello from rust!".into(),
    };

    // Serialize and publish
    let mut buf = Vec::new();
    packet
        .serialize(&mut Serializer::new(&mut buf))
        .expect("Failed to serialize packet");

    println!("Sent: {:?}", packet);

    publisher.put(buf).await?;

    // Wait up to 6 seconds for a reply
    tokio::time::timeout(Duration::from_secs(6), async {
        match subscriber.recv_async().await {
            Ok(pack) => {
                let bytes = pack.payload().to_bytes();
                match TaggedString::deserialize(&mut Deserializer::new(&*bytes)) {
                    Ok(msg) => println!("Received: {:?}", msg),
                    Err(e) => eprintln!("Deserialize error: {e}"),
                }
            }
            Err(e) => eprintln!("Receive error: {e}"),
        }
    })
    .await
    .ok(); // timeout Result can be ignored

    Ok(())
}
