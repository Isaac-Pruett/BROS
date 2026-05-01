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

    let att = tagged_string {
        id: 42,
        s: "hello from rust!".into(),
    };

    // Serialize and publish
    let mut buf = Vec::new();
    att.serialize(&mut Serializer::new(&mut buf))
        .expect("Failed to serialize att");

    println!("Sent: {:?}", att);

    publisher.put(buf).await?;

    // Wait up to 6 seconds for a reply
    tokio::time::timeout(Duration::from_secs(6), async move {
        if let Some(pack) = subscriber.recv().ok() {
            let bytes = pack.payload().to_bytes();
            match tagged_string::deserialize(&mut Deserializer::new(&*bytes)) {
                Ok(msg) => println!("Received: {:?}", msg),
                Err(e) => eprintln!("Deserialize error: {e}"),
            }
        }
    })
    .await
    .ok(); // timeout Result can be ignored

    Ok(())
}
