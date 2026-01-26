use flatbuffers::FlatBufferBuilder;
use mavlink::common::HIL_STATE_QUATERNION_DATA;
use mavlink::error::MessageReadError;
use tokio::time::Sleep;
use tokio::{process, time};

use std::error::Error;
use std::sync::Arc;
use std::time::Duration;
use zenoh;

use mavlink::ardupilotmega::{MavMessage, MavModeFlag};
use mavlink::{self, MavConnection, MavlinkVersion, MessageData};

use crate::MAV_generated::mavlink_fb::HilStateQuaternion;

mod MAV_generated;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error + Send + Sync>> {
    let session =
        zenoh::open(zenoh::Config::from_env().unwrap_or(zenoh::Config::default())).await?;

    let mut mavconn = mavlink::connect::<MavMessage>("serial:/dev/ttyTHS1:921600")
        .expect("Failed to connect UART to MAVLink");

    let state_pub = session.declare_publisher("mavlink/state").await?;

    // Wait for subscribers to be ready
    tokio::time::sleep(Duration::from_millis(500)).await;

    mavconn.set_protocol_version(MavlinkVersion::V2);

    let vehicle = Arc::new(mavconn);

    tokio::spawn(async move {
        let vehicle = vehicle.clone();
        loop {
            let res = vehicle.send_default(&MavMessage::HEARTBEAT(
                mavlink::ardupilotmega::HEARTBEAT_DATA {
                    custom_mode: 0,
                    mavtype: mavlink::ardupilotmega::MavType::MAV_TYPE_QUADROTOR,
                    autopilot: mavlink::ardupilotmega::MavAutopilot::MAV_AUTOPILOT_ARDUPILOTMEGA,
                    base_mode: MavModeFlag::empty(),
                    system_status: mavlink::ardupilotmega::MavState::MAV_STATE_STANDBY,
                    mavlink_version: 0x3,
                },
            ));

            match res {
                Ok(_) => {
                    time::sleep(Duration::from_secs(1)).await;
                }
                Err(e) => {
                    eprintln!("Heartbeat send failed: {:?}", e);
                }
            }
            //time::sleep(Duration::from_millis(10)).await;
        }
    });

    tokio::spawn(async move {
        let vehicle = vehicle.clone();
        loop {
            match vehicle.try_recv() {
                Ok((_header, msg)) => match msg {
                    MavMessage::HIL_STATE_QUATERNION(data) => {
                        println!("{:?}", data);

                        let mut builder = FlatBufferBuilder::new();

                        // Create the struct directly
                        let state = HilStateQuaternion::new(
                            data.time_usec,
                            &data.attitude_quaternion,
                            data.rollspeed,
                            data.pitchspeed,
                            data.yawspeed,
                            data.lat,
                            data.lon,
                            data.alt,
                            data.vx,
                            data.vy,
                            data.vz,
                            data.ind_airspeed,
                            data.true_airspeed,
                            data.xacc,
                            data.yacc,
                            data.zacc,
                        );

                        builder.push(state);

                        // Get the serialized bytes
                        let payload = builder.finished_data();

                        state_pub.put(payload).await?
                    }
                    _ => {}
                },
                Err(e) => {
                    match e {
                        MessageReadError::Io(e) => {
                            if e.kind() == std::io::ErrorKind::WouldBlock {
                                // no messages to recv.
                                time::sleep(Duration::from_millis(5));
                            } else {
                                eprintln!("IO Message Read Err: {:?}", e);
                            }
                        }
                        _ => {} // block parser errors. skill issue.
                    }
                }
            }
        }
    });

    session.close().await?;
    Ok(())
}
