use mavlink::error::MessageReadError;
use std::time::Duration;
use zenoh;

use mavlink::ardupilotmega::{HIL_STATE_DATA, MavMessage};
use mavlink::{self, MavConnection, MavlinkVersion};

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let session =
        zenoh::open(zenoh::Config::from_env().unwrap_or(zenoh::Config::default())).await?;

    let mut mav = mavlink::connect::<MavMessage>("serial:/dev/ttyTHS1:921600")
        .expect("Failed to connect UART to MAVLink");

    let mut attitude_pub = session.declare_publisher("mavlink/attitude").await?;

    // Wait for subscribers to be ready
    tokio::time::sleep(Duration::from_millis(500)).await;

    mav.set_protocol_version(MavlinkVersion::V2);
    loop {
        match mav.try_recv() {
            Ok((_header, msg)) => match msg {
                MavMessage::HEARTBEAT(data) => {
                    println!("{:?}", data);
                }
                MavMessage::HIL_STATE_QUATERNION(data) => {
                    println!("{:?}", data);
                }
                MavMessage::GLOBAL_POSITION_INT(data) => {
                    println!("{:?}", data);
                }

                _ => {}
            },
            Err(MessageReadError::Io(e)) => {
                eprintln!("Error reading message from MAVLink (Io Err): {e}")
            }
            Err(MessageReadError::Parse(e)) => {
                eprintln!("Error reading message from MAVLink (Parse Err): {e}")
            }
        }
    }

    // session.close().await?;
}
