use mavlink::common::HIL_STATE_QUATERNION_DATA;
use mavlink::error::MessageReadError;

use std::time::Duration;
use zenoh;

use mavlink::ardupilotmega::MavMessage;
use mavlink::{self, MavConnection, MavlinkVersion, MessageData};

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let session =
        zenoh::open(zenoh::Config::from_env().unwrap_or(zenoh::Config::default())).await?;

    let mut mav = mavlink::connect::<MavMessage>("serial:/dev/ttyTHS1:921600")
        .expect("Failed to connect UART to MAVLink");

    let state_pub = session.declare_publisher("mavlink/state").await?;

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

                    let mut buf = [0u8; HIL_STATE_QUATERNION_DATA::ENCODED_LEN];
                    let written_ct = data.ser(MavlinkVersion::V2, &mut buf);

                    let payload = &buf[..written_ct];

                    state_pub.put(payload).await?;
                }
                MavMessage::TRAJECTORY_REPRESENTATION_WAYPOINTS(data) => {}

                MavMessage::ACTUATOR_OUTPUT_STATUS(data) => {
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
