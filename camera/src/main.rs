use std::process::exit;
use zenoh;

use a8mini_camera_rs::{self, A8Mini, control::A8MiniSimpleCommand};

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let config = zenoh::Config::from_env().unwrap_or(zenoh::Config::default());
    let session = zenoh::open(config).await?;

    let cam = match A8Mini::connect().await {
        Err(e) => {
            eprintln!("Could not connect to camera: {e}");
            exit(1);
        }
        Ok(cam) => cam,
    };

    match cam.send_command(A8MiniSimpleCommand::RebootCamera).await {
        Ok(_) => println!("Camera rebooted successfully"),
        Err(e) => eprintln!("Could not reboot camera: {e}"),
    }

    session.close().await?;
    Ok(())
}
