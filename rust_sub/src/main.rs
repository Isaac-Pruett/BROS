use std::{thread::sleep, time::Duration};

use zenoh;

#[tokio::main]
async fn main() {
    let config = zenoh::Config::default();

    let session = zenoh::open(config).await.unwrap();

    let publ = session.declare_publisher("rust/helloworld").await.unwrap();

    let subs = session
        .declare_subscriber("python/helloworld")
        .await
        .unwrap();

    publ.put("Hello, from Rust!").await.unwrap();

    println!("Rust put the data");
    tokio::time::sleep(Duration::from_millis(3000)).await;

    let data = subs.recv().unwrap();

    println!("Rust recienved a message");

    // dbg!(&data);

    let msg = data.payload().try_to_string().expect("msg is invalid");

    println!("{:?}", msg);

    sleep(Duration::from_millis(1500));
    println!("Hello, world!");
}
