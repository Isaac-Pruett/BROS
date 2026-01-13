use std::process::exit;
use zenoh;
use zenoh::Wait;

use a8mini_camera_rs::{self, A8Mini, control::A8MiniSimpleCommand};

use gst::prelude::*;
use gstreamer as gst;
use gstreamer_app as gst_app;

#[tokio::main]
async fn main() -> zenoh::Result<()> {
    let config = zenoh::Config::from_env().unwrap_or(zenoh::Config::default());
    let session = zenoh::open(config).await?;

    let img_pub = session.declare_publisher("camera/video").await?;

    gst::init().expect("Failed to initialize GStreamer");

    let cam = match A8Mini::connect().await {
        Err(e) => {
            eprintln!("Could not connect to camera: {e}");
            exit(1);
        }
        Ok(cam) => cam,
    };

    // RTSP addr of the main (HQ) video stream off of the a8.
    // let rtsp_url = "rtsp://192.168.144.25:8554/video1";
    // println!("Connecting to RTSP stream: {}", rtsp_url);
    // let pipeline_str = format!(
    //     "rtspsrc location={} latency=0 ! rtph264depay ! h264parse ! avdec_h264 ! videoconvert ! appsink name=sink",
    //     rtsp_url
    // );

    // Isaac laptop cam BEGIN
    let video_device = "0"; // Camera index (0 is usually the default camera)
    println!("Connecting to video device: {}", video_device);

    let pipeline_str = format!(
        "avfvideosrc device-index={} ! videoconvert ! appsink name=sink",
        video_device
    );
    // Isaac loptop cam END

    let pipeline = gst::parse::launch(&pipeline_str)
        .expect("Failed to create pipeline.")
        .dynamic_cast::<gst::Pipeline>()
        .expect("Error in dynamically casting the return value of the pipeline builder to an actual gstreamer::Pipeline.");

    let appsink = pipeline
        .by_name("sink")
        .expect("Failed to get appsink")
        .dynamic_cast::<gst_app::AppSink>()
        .expect("Expected an appsink");

    // Configure the appsink
    appsink.set_caps(Some(
        &gst::Caps::builder("video/x-raw")
            .field("format", "RGB")
            .build(),
    ));
    appsink.set_property("emit-signals", true);
    appsink.set_property("sync", false);

    // Set up callback for new samples
    appsink.set_callbacks(
        gst_app::AppSinkCallbacks::builder()
            .new_sample(move |sink| {
                let sample = sink.pull_sample().map_err(|_| gst::FlowError::Error)?;
                let buffer = sample.buffer().ok_or(gst::FlowError::Error)?;
                let map = buffer.map_readable().map_err(|_| gst::FlowError::Error)?;

                // Get the actual frame data
                let frame_data = map.as_slice();
                println!("Received frame with {} bytes", frame_data.len());

                // Publish to Zenoh
                img_pub.put(frame_data).wait().map_err(|e| {
                    eprintln!("Failed to publish frame: {}", e);
                    gst::FlowError::Error
                })?;

                Ok(gst::FlowSuccess::Ok)
            })
            .build(),
    );

    // Start the pipeline
    pipeline
        .set_state(gst::State::Playing)
        .expect("Failed to set pipeline to Playing");

    println!("Pipeline started, receiving frames...");

    let bus = pipeline.bus().expect("Pipeline has no bus");
    for msg in bus.iter_timed(gst::ClockTime::NONE) {
        use gst::MessageView;

        match msg.view() {
            MessageView::Eos(..) => {
                println!("End of stream");
                break;
            }
            MessageView::Error(err) => {
                eprintln!(
                    "Error from {:?}: {} ({:?})",
                    err.src().map(|s| s.path_string()),
                    err.error(),
                    err.debug()
                );
                break;
            }
            MessageView::StateChanged(state_changed) => {
                if state_changed.src().map(|s| s == &pipeline).unwrap_or(false) {
                    println!(
                        "Pipeline state changed from {:?} to {:?}",
                        state_changed.old(),
                        state_changed.current()
                    );
                }
            }
            _ => (),
        }
    }

    // Cleanup
    pipeline
        .set_state(gst::State::Null)
        .expect("Failed to set pipeline to Null");

    session.close().await?;
    Ok(())
}
