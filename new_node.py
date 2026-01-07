#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3
"""
Create a new Zenoh node with template files.
"""

import os
import sys
from pathlib import Path

RUST_TEMPLATE = """
use std::time::Duration;
use zenoh;

#[tokio::main]
async fn main() -> zenoh::Result<()> {{
    let session = zenoh::open(zenoh::Config::default()).await?;
    let publisher = session.declare_publisher("rust/helloworld").await?;
    let subscriber = session.declare_subscriber("python/helloworld").await?;

    // Wait for subscribers to be ready
    tokio::time::sleep(Duration::from_millis(500)).await;

    // Now publish
    publisher.put("Hello, from Rust!").await?;
    println!("Rust → Published");

    println!("Rust → Waiting for Python message...");
    match tokio::time::timeout(Duration::from_secs(8), subscriber.recv_async()).await {{
        Ok(Ok(sample)) => {{
            let msg = sample.payload().try_to_string().unwrap_or_default();
            println!("Rust ← Received: {{msg:?}}");
        }}
        Ok(Err(e)) => println!("Rust ← Error receiving: {{e}}"),
        Err(_) => println!("Rust ← Timeout waiting for Python"),
    }}

    println!("Rust done!");
    session.close().await?;
    Ok(())
}}
""".strip()

CARGO_TOML = """[package]
name = "{node_name}"
version = "0.1.0"
edition = "2024"

[dependencies]
tokio = "1.49.0"
zenoh = "1.7.1"

# this defines the binaries to build in the nix build schema
# the name of this package is the name of the binary that will be built.
[[bin]]
name = "{node_name}"
path = "src/main.rs"
""".strip()

FLAKE_NIX = """
{{
  description = "{node_name} -- Rust Zenoh Node";

  inputs = {{
    naersk.url = "github:nix-community/naersk/master";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    utils.url = "github:numtide/flake-utils";
  }};

  outputs = {{ self, nixpkgs, utils, naersk }}:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${{system}};
        naersk-lib = pkgs.callPackage naersk {{ }};
      in
      {{
        # Modern way: define packages.default
        packages.default = naersk-lib.buildPackage {{
          src = ./.;
          pname = "{node_name}";
          # doCheck = true;
          # release = false;  # already default in naersk
        }};

        # Modern way: devShells.default
        devShells.default = pkgs.mkShell {{
          buildInputs = with pkgs; [
            cargo
            rustc
            rustfmt
            pre-commit
            rustPackages.clippy
          ];
          RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
        }};

        # Optional: expose as an app (for nix run)
        apps.default = utils.lib.mkApp {{
          drv = self.packages.${{system}}.default;
        }};
      }}
    );
}}
""".strip()


def create_rust_node(node_name: str):
    """Create a new Rust Zenoh node."""
    node_dir = Path(f"{node_name}")

    if node_dir.exists():
        print(f"Error: {node_dir} already exists!")
        sys.exit(1)

    # Create directory structure
    node_dir.mkdir()
    src_dir = node_dir / "src"
    src_dir.mkdir()

    # Write files
    (src_dir / "main.rs").write_text(RUST_TEMPLATE.format(node_name=node_name))
    (node_dir / "Cargo.toml").write_text(CARGO_TOML.format(node_name=node_name))
    (node_dir / "flake.nix").write_text(FLAKE_NIX.format(node_name=node_name))

    print(f"✓ Created Rust node: {node_dir}")
    print(f"\nTo get started:")
    print(f"  cd {node_dir}")
    print(f"  nix develop")
    print(f"  cargo build")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: ./new_node.py <node_name>")
        sys.exit(1)

    create_rust_node(sys.argv[1])
