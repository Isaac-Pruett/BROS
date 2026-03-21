use std::{thread::sleep, time::Duration};
use zenoh;
use zenoh::Wait;

fn main() -> zenoh::Result<()> {
    let session =
        zenoh::open(zenoh::Config::from_env().unwrap_or(zenoh::Config::default())).wait()?;
    let publisher = session.declare_publisher("rust/helloworld").wait()?;
    let subscriber = session.declare_subscriber("python/helloworld").wait()?;

    // Wait for subscribers to be ready
    sleep(Duration::from_millis(500));

    // Now publish
    publisher.put("Hello, from Rust!").wait()?;
    println!("Rust → Published");

    println!("Rust → Waiting for Python message...");
    match subscriber.recv() {
        Ok(sample) => {
            let msg = sample.payload().try_to_string().unwrap_or_default();
            println!("Rust ← Received: {msg:?}");
        }
        Err(e) => println!("Rust ← Error receiving: {e}"),
    }

    println!("Rust done!");
    session.close().wait()?;
    Ok(())
}
