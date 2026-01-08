#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3
"""
Create a new Zenoh node with template files.
Supports both Rust and Python nodes.
Automatically updates the master flake.nix with the new node.
"""

import os
import re
import sys
from pathlib import Path
from typing import Literal

RUST_MAIN_TEMPLATE = """use std::time::Duration;
use zenoh::prelude::*;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {{
    // Initialize Zenoh session
    let config = zenoh::Config::default();
    let session = zenoh::open(config).await?;

    println!("[{node_name}] Session opened");

    // Declare publisher
    let key_pub = "rust/{node_name}/data";
    let publisher = session.declare_publisher(key_pub).await?;
    println!("[{node_name}] Publisher declared on '{{}}'", key_pub);

    // Declare subscriber
    let key_sub = "python/helloworld";
    let subscriber = session.declare_subscriber(key_sub).await?;
    println!("[{node_name}] Subscriber declared on '{{}}'", key_sub);

    // Wait for discovery
    tokio::time::sleep(Duration::from_millis(500)).await;

    // Publish message
    let message = format!("Hello from {{}}!", "{node_name}");
    publisher.put(&message).await?;
    println!("[{node_name}] → Published: '{{}}'", message);

    // Wait for incoming messages
    println!("[{node_name}] ← Waiting for messages...");
    match tokio::time::timeout(Duration::from_secs(10), subscriber.recv_async()).await {{
        Ok(Ok(sample)) => {{
            let msg = sample
                .payload()
                .try_to_string()
                .unwrap_or_default();
            println!("[{node_name}] ← Received: '{{}}'", msg);
        }}
        Ok(Err(e)) => eprintln!("[{node_name}] ✗ Error receiving: {{}}", e),
        Err(_) => println!("[{node_name}] ⏱ Timeout waiting for messages"),
    }}

    println!("[{node_name}] Done!");
    session.close().await?;
    Ok(())
}}
"""

CARGO_TOML_TEMPLATE = """[package]
name = "{node_name}"
version = "0.1.0"
edition = "2021"

[dependencies]
tokio = {{ version = "1", features = ["full"] }}
zenoh = "1.0.0"

[[bin]]
name = "{node_name}"
path = "src/main.rs"
"""

RUST_FLAKE_TEMPLATE = """{{
  description = "{node_name} — Rust Zenoh Node";

  inputs = {{
    naersk.url = "github:nix-community/naersk";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  }};

  outputs = {{ self, nixpkgs, flake-utils, naersk }}:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${{system}};
        naersk-lib = pkgs.callPackage naersk {{}};
      in
      {{
        packages.default = naersk-lib.buildPackage {{
          src = ./.;
          pname = "{node_name}";

          # Optional: add system dependencies
          buildInputs = with pkgs; [
            # Add any required system libraries here
          ];
        }};

        devShells.default = pkgs.mkShell {{
          buildInputs = with pkgs; [
            cargo
            rustc
            rustfmt
            clippy
            rust-analyzer
            cargo-watch
          ];

          RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;

          shellHook = ''
            echo "🦀 Rust development environment for {node_name}"
            echo "Run 'cargo build' to build the project"
          '';
        }};

        apps.default = {{
          type = "app";
          program = "${{self.packages.${{system}}.default}}/bin/{node_name}";
        }};
      }}
    );
}}
"""

PYTHON_MAIN_TEMPLATE = """#!/usr/bin/env python3
\"\"\"
{node_name} - Python Zenoh Node
\"\"\"
import time
import zenoh

def main():
    # Initialize Zenoh session
    config = zenoh.Config()
    session = zenoh.open(config)
    print(f"[{node_name}] Session opened")

    # Declare publisher
    key_pub = "python/{node_name}/data"
    pub = session.declare_publisher(key_pub)
    print(f"[{node_name}] Publisher declared on '{{key_pub}}'")

    # Declare subscriber
    key_sub = "rust/helloworld"

    def listener(sample):
        payload = sample.payload.to_string()
        print(f"[{node_name}] ← Received: '{{payload}}'")

    sub = session.declare_subscriber(key_sub, listener)
    print(f"[{node_name}] Subscriber declared on '{{key_sub}}'")

    # Wait for discovery
    time.sleep(0.5)

    # Publish message
    message = f"Hello from {node_name}!"
    pub.put(message)
    print(f"[{node_name}] → Published: '{{message}}'")

    # Keep running to receive messages
    print(f"[{node_name}] ← Waiting for messages...")
    try:
        time.sleep(10)
    except KeyboardInterrupt:
        pass

    print(f"[{node_name}] Done!")
    session.close()

if __name__ == "__main__":
    main()
"""

PYTHON_FLAKE_TEMPLATE = """{{
  description = "{node_name} — Python Zenoh Node";

  inputs = {{
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  }};

  outputs = {{ self, nixpkgs, flake-utils }}:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${{system}};

        pythonEnv = pkgs.python3.withPackages (ps: [
          ps.eclipse-zenoh
        ]);

      in
      {{
        packages.default = pkgs.stdenv.mkDerivation {{
          pname = "{node_name}";
          version = "0.1.0";
          src = ./.;

          buildInputs = [ pythonEnv ];

          installPhase = ''
            mkdir -p $out/bin
            cp src/main.py $out/bin/{node_name}
            chmod +x $out/bin/{node_name}

            # Wrap with Python environment
            wrapProgram $out/bin/{node_name} \\
              --prefix PATH : ${{pythonEnv}}/bin
          '';

          nativeBuildInputs = [ pkgs.makeWrapper ];
        }};

        devShells.default = pkgs.mkShell {{
          buildInputs = [
            pythonEnv
            pkgs.python3Packages.ipython
            pkgs.ruff
          ];

          shellHook = ''
            echo "🐍 Python development environment for {node_name}"
            echo "Run 'python src/main.py' to start the node"
          '';
        }};

        apps.default = {{
          type = "app";
          program = "${{self.packages.${{system}}.default}}/bin/{node_name}";
        }};
      }}
    );
}}
"""

GITIGNORE_TEMPLATE = """.zed
.vscode
.venv
.direnv
*/__pycache__
*.pyc
result
result-*
target/
debug/
.DS_Store
Cargo.lock
"""

README_TEMPLATE = """# {node_name}

A Zenoh {node_type} node.

## Building

```bash
nix build
```

## Development

```bash
nix develop
```

## Running

```bash
nix run
```

Or directly:
```bash
./{run_command}
```
"""


def sanitize_node_name(name: str) -> str:
    """Convert node name to valid Nix identifier (alphanumeric + underscore/dash)"""
    return re.sub(r"[^a-zA-Z0-9_-]", "_", name)


def add_node_to_master_flake(node_name: str, node_type: str):
    """
    Automatically add the new node to the master flake.nix
    """
    flake_path = Path("flake.nix")

    if not flake_path.exists():
        print(f"⚠️  Warning: Master flake.nix not found at {flake_path}")
        print("    Skipping automatic flake update.")
        return False

    try:
        content = flake_path.read_text()
        original_content = content

        # Sanitize the node name for use as Nix identifier
        safe_name = sanitize_node_name(node_name)

        # 1. Add to inputs section
        inputs_pattern = r"(inputs\s*=\s*\{[^}]*?)((\s*};)|$)"

        new_input = f"""
        {safe_name}.url = "./{node_name}";
        """

        def add_input(match):
            existing = match.group(1)
            # Check if already exists
            if safe_name in existing:
                return match.group(0)
            closing = match.group(2)
            return existing + new_input + "\n  " + closing

        content = re.sub(inputs_pattern, add_input, content, flags=re.DOTALL)

        # 2. Add to outputs function parameters
        outputs_pattern = r"outputs\s*=\s*\{\s*([^}]+?)\s*\}:"

        def add_output_param(match):
            params = match.group(1)
            # Check if already exists
            if safe_name in params:
                return match.group(0)
            # Add before the closing brace
            params_list = [p.strip() for p in params.split(",") if p.strip()]
            if safe_name not in params_list:
                params_list.append(safe_name)
            return f"outputs = {{ {', '.join(params_list)} }}:"

        content = re.sub(outputs_pattern, add_output_param, content)

        # 3. Add helper to get package
        helper_section = (
            r"(# Helper to safely get packages from subflakes\s*)(.*?)(\n\s*in)"
        )

        cap_name = "".join(
            word.capitalize() for word in safe_name.replace("-", "_").split("_")
        )
        new_helper = (
            f"\n      get{cap_name} = {safe_name}.packages.${{system}}.default or null;"
        )

        def add_helper(match):
            prefix = match.group(1)
            existing = match.group(2)
            # Check if already exists
            if f"get{cap_name}" in existing:
                return match.group(0)
            suffix = match.group(3)
            return prefix + existing + new_helper + suffix

        content = re.sub(helper_section, add_helper, content, flags=re.DOTALL)

        # 4. Add to packages section
        packages_pattern = r"(packages\s*=\s*\{[^}]*?# Re-export subproject packages\s*)(.*?)(# Combined launcher)"

        new_package = f"\n        {safe_name}App = get{cap_name};"

        def add_package(match):
            prefix = match.group(1)
            existing = match.group(2)
            # Check if already exists
            if f"{safe_name}App" in existing:
                return match.group(0)
            suffix = match.group(3)
            return prefix + existing + new_package + "\n        \n        " + suffix

        content = re.sub(packages_pattern, add_package, content, flags=re.DOTALL)

        # 5. Add to runtimeInputs in demo launcher
        runtime_pattern = (
            r"(runtimeInputs = pkgs\.lib\.filter \(x: x != null\) \[)(.*?)(\];)"
        )

        new_runtime = f"\n            get{cap_name}"

        def add_runtime(match):
            prefix = match.group(1)
            existing = match.group(2)
            # Check if already exists
            if f"get{cap_name}" in existing:
                return match.group(0)
            suffix = match.group(3)
            return prefix + existing + new_runtime + "\n          " + suffix

        content = re.sub(runtime_pattern, add_runtime, content, flags=re.DOTALL)

        # 6. Add launch command in demo script
        launch_pattern = r'(# Launch applications\s*)(.*?)(echo ""\s*echo "✓ All applications running")'

        new_launch = f"""
            ${{pkgs.lib.optionalString (get{cap_name} != null) ''
              echo "Starting {node_name}..."
              ${{get{cap_name}}}/bin/{node_name} &
              {safe_name.upper()}_PID=$!
            ''}}
            """

        def add_launch(match):
            prefix = match.group(1)
            existing = match.group(2)
            # Check if already exists
            if safe_name.upper() in existing:
                return match.group(0)
            suffix = match.group(3)
            return prefix + existing + new_launch + "\n            " + suffix

        content = re.sub(launch_pattern, add_launch, content, flags=re.DOTALL)

        # 7. Add to devShell packages
        devshell_pattern = r"(devShells\.default = pkgs\.mkShell \{[^}]*?packages = \[[^\]]*?)\] \+\+ pkgs\.lib\.filter"

        def add_devshell(match):
            existing = match.group(1)
            # Check if already exists
            if f"get{cap_name}" in existing:
                return match.group(0)
            return (
                existing
                + f"\n          get{cap_name}"
                + "\n        ] ++ pkgs.lib.filter"
            )

        content = re.sub(devshell_pattern, add_devshell, content, flags=re.DOTALL)

        # Only write if changes were made
        if content != original_content:
            flake_path.write_text(content)
            print(f"✓ Updated master flake.nix with {node_name}")
            return True
        else:
            print(f"ℹ️  Node {node_name} already exists in master flake.nix")
            return False

    except Exception as e:
        print(f"✗ Error updating master flake.nix: {e}")
        print(f"  You may need to manually add the node to the flake.")
        return False


def create_rust_node(node_name: str):
    """Create a new Rust Zenoh node."""
    node_dir = Path(node_name)
    if node_dir.exists():
        print(f"✗ Error: {node_dir} already exists!")
        sys.exit(1)

    # Create directory structure
    node_dir.mkdir()
    src_dir = node_dir / "src"
    src_dir.mkdir()

    # Write files
    (src_dir / "main.rs").write_text(RUST_MAIN_TEMPLATE.format(node_name=node_name))
    (node_dir / "Cargo.toml").write_text(
        CARGO_TOML_TEMPLATE.format(node_name=node_name)
    )
    (node_dir / "flake.nix").write_text(RUST_FLAKE_TEMPLATE.format(node_name=node_name))
    (node_dir / ".gitignore").write_text(GITIGNORE_TEMPLATE)
    (node_dir / "README.md").write_text(
        README_TEMPLATE.format(
            node_name=node_name, node_type="Rust", run_command=f"cargo run"
        )
    )

    print(f"✓ Created Rust node: {node_dir}/")

    # Update master flake
    if add_node_to_master_flake(node_name, "rust"):
        print(f"✓ Node integrated into monorepo")
        print(f"\n💡 Run 'nix flake lock' to update lockfile")

    print(f"\n📦 Next steps:")
    print(f"  cd {node_dir}")
    print(f"  nix develop        # Enter dev shell")
    print(f"  cargo build        # Build the project")
    print(f"  cargo run          # Run the node")


def create_python_node(node_name: str):
    """Create a new Python Zenoh node."""
    node_dir = Path(node_name)
    if node_dir.exists():
        print(f"✗ Error: {node_dir} already exists!")
        sys.exit(1)

    # Create directory structure
    node_dir.mkdir()
    src_dir = node_dir / "src"
    src_dir.mkdir()

    # Write files
    main_file = src_dir / "__init__.py"
    main_file.write_text(PYTHON_MAIN_TEMPLATE.format(node_name=node_name))
    main_file.chmod(0o755)

    (node_dir / "flake.nix").write_text(
        PYTHON_FLAKE_TEMPLATE.format(node_name=node_name)
    )
    (node_dir / ".gitignore").write_text(GITIGNORE_TEMPLATE)
    (node_dir / "README.md").write_text(
        README_TEMPLATE.format(
            node_name=node_name, node_type="Python", run_command=f"python src/main.py"
        )
    )

    print(f"✓ Created Python node: {node_dir}/")

    # Update master flake
    if add_node_to_master_flake(node_name, "python"):
        print(f"✓ Node integrated into monorepo")
        # print(f"\n💡 Run 'nix flake lock' to update lockfile")

    print(f"\n📦 Next steps:")
    print(f"  cd {node_dir}")
    print(f"  nix develop              # Enter dev shell")
    print(f"  python src/main.py       # Run the node")


def main():
    if len(sys.argv) not in [2, 3]:
        print("Usage: ./new_node.py <node_name> [rust|python]")
        print("\nExamples:")
        print("  ./new_node.py my_sensor rust    # Create Rust node")
        print("  ./new_node.py my_actuator python # Create Python node")
        print("  ./new_node.py my_node            # Create Rust node (default)")
        sys.exit(1)

    node_name = sys.argv[1]
    node_type = sys.argv[2] if len(sys.argv) == 3 else "rust"

    if node_type not in ["rust", "python"]:
        print(f"✗ Error: Invalid node type '{node_type}'. Must be 'rust' or 'python'")
        sys.exit(1)

    if node_type == "rust":
        create_rust_node(node_name)
    else:
        create_python_node(node_name)


if __name__ == "__main__":
    main()
