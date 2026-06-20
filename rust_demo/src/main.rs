use std::time::{Duration, Instant};
use zenoh::Config;

use rmp_serde::{Deserializer, Serializer};
use serde::{Deserialize, Serialize};

mod tagged_string;
use tagged_string::TaggedString;

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let session = zenoh::open(Config::default()).await?;
    let publisher = session.declare_publisher("demo/out/rust").await?;
    let subscriber = session.declare_subscriber("demo/out/*").await?;

    tokio::time::sleep(Duration::from_millis(500)).await;

    let packet = TaggedString {
        id: 42,
        s: "hello from rust!".into(),
    };

    let mut buf = Vec::new();
    packet
        .serialize(&mut Serializer::new(&mut buf))
        .expect("Failed to serialize packet");

    println!("Rust Sent: {:?}", packet);
    publisher.put(buf).await?;

    let deadline = Instant::now() + Duration::from_secs(6);
    let mut received_any = false;

    while Instant::now() < deadline {
        let remaining = deadline.saturating_duration_since(Instant::now());

        match tokio::time::timeout(remaining, subscriber.recv_async()).await {
            Ok(Ok(sample)) => {
                let bytes = sample.payload().to_bytes();

                match TaggedString::deserialize(&mut Deserializer::new(&*bytes)) {
                    Ok(msg) => {
                        println!("Rust Received: {:?}", msg);
                        received_any = true;
                    }
                    Err(e) => eprintln!("Deserialize error: {e}"),
                }
            }
            Ok(Err(e)) => {
                eprintln!("Receive error: {e}");
                break;
            }
            Err(_) => break,
        }
    }

    if !received_any {
        println!("Timeout: no message received");
    }

    Ok(())
}
