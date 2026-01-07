use std::time::Duration;
use zenoh;

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let session = zenoh::open(zenoh::Config::default()).await?;
    let publisher = session.declare_publisher("rust/helloworld").await?;
    let subscriber = session.declare_subscriber("python/helloworld").await?;

    // Wait for subscribers to be ready
    tokio::time::sleep(Duration::from_millis(500)).await;

    // Now publish
    publisher.put("Hello, from Rust!").await?;
    println!("Rust → Published");

    println!("Rust → Waiting for Python message...");
    match tokio::time::timeout(Duration::from_secs(8), subscriber.recv_async()).await {
        Ok(Ok(sample)) => {
            let msg = sample.payload().try_to_string().unwrap_or_default();
            println!("Rust ← Received: {msg:?}");
        }
        Ok(Err(e)) => println!("Rust ← Error receiving: {e}"),
        Err(_) => println!("Rust ← Timeout waiting for Python"),
    }

    println!("Rust done!");
    session.close().await?;
    Ok(())
}