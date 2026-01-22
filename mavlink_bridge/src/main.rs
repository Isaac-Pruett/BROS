use mavlink::error::MessageReadError;
use std::time::Duration;
use zenoh;

use mavlink::ardupilotmega::{HIL_STATE_DATA, MavMessage};
use mavlink::{self, MavConnection, MavlinkVersion};

mod state_generated;

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

                    attitude_pub
                        .put(
                            data.attitude_quaternion
                                .iter()
                                .flat_map(|&x| x.to_ne_bytes())
                                .collect::<Vec<u8>>(),
                        )
                        .await?;

                    let pos = vec![data.lon, data.lat, data.alt];
                    pos_pub
                        .put(
                            pos.iter()
                                .flat_map(|x| x.to_ne_bytes())
                                .collect::<Vec<u8>>(),
                        )
                        .await?;

                    let velo = vec![data.vx, data.vy, data.vz];
                    velo_pub
                        .put(
                            velo.iter()
                                .flat_map(|x| x.to_ne_bytes())
                                .collect::<Vec<u8>>(),
                        )
                        .await?;

                    let acc = vec![data.xacc, data.yacc, data.zacc];
                    acc_pub
                        .put(
                            acc.iter()
                                .flat_map(|x| x.to_ne_bytes())
                                .collect::<Vec<u8>>(),
                        )
                        .await?;
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
