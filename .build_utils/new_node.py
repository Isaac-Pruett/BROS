#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3 cargo uv
"""
Create a new Zenoh node with template files.
Supports both Rust and Python nodes.
Automatically updates the master flake.nix with the new node.
"""

import re
import subprocess
import sys
from pathlib import Path

RUST_MAIN_TEMPLATE = """use std::{{thread::sleep, time::Duration}};
use zenoh;
use zenoh::Wait;

fn main() -> zenoh::Result<()> {{
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
    match subscriber.recv() {{
        Ok(sample) => {{
            let msg = sample.payload().try_to_string().unwrap_or_default();
            println!("Rust ← Received: {{msg:?}}");
            }}
            Err(e) => println!("Rust ← Error receiving: {{e}}"),
    }

    println!("Rust done!");
    session.close().wait()?;
    Ok(())
}}


"""

CARGO_TOML_TEMPLATE = """[package]
name = "{node_name}"
version = "0.1.0"
edition = "2024"

[dependencies]
tokio = {{ version = "1", features = ["full"] }}
zenoh = {version = "1", features = ["shared-memory"]}

flatbuffers = "25.12.19"

[[bin]]
name = "{node_name}"
path = "src/main.rs"
"""

RUST_FLAKE_TEMPLATE = """{{
  description = "{node_name} — Rust Zenoh Node";
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
        packages.default = naersk-lib.buildPackage {{
          src = ./.;
          pname = "{node_name}";
          # Optional: customize if needed
          # doCheck = true;
          # release = false;  # already default in naersk
        }};
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
        apps.default = utils.lib.mkApp {{
          drv = self.packages.${{system}}.default;
        }};
      }}
    );
}}
"""


PYTHON_PYROJ_TEMPLATE = """[project]
name = "{node_name}"
version = "0.1.0"
description = "{node_name} - Python Zenoh Node"
requires-python = ">=3.12"
dependencies = [
 "eclipse-zenoh>=1.7.1",
 "flatbuffers>=25.12.19",

]

# this defines the entrypoint of the python program.
# we look in the package hello_world and find the function main (inside __init__.py)
# the function must be visible (i.e. imported or defined) in __init__.py
[project.scripts]
{node_name} = "{node_name}:main"

[build-system]
requires = ["uv_build>=0.9.0,<0.10.0"]
build-backend = "uv_build"
"""

PYTHON_MAIN_TEMPLATE = """
\"\"\"
{node_name} - Python Zenoh Node
\"\"\"
from time import sleep

import zenoh


# a callback to run by the subscriber
def listen(sample: zenoh.Sample):
    print("Python ← Received:", sample.payload.to_string())


def main():
    with zenoh.open(zenoh.Config().from_env()) as session:
        pub = session.declare_publisher("python/helloworld")
        sub = session.declare_subscriber("rust/helloworld", listen)

        # Wait for subscribers to be ready
        sleep(0.5)

        # Now publish
        pub.put("Hello, from Python!")
        print("Python → Published")

        print("Python → Waiting for Rust message...")

        print("Python done!")
        session.close()


if __name__ == "__main__":
    main()

"""

PYTHON_FLAKE_TEMPLATE = """{{
  description = "{node_name} — Python Zenoh Node";
  inputs = {{
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    pyproject-nix = {{
      url = "github:pyproject-nix/pyproject.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    }};
    uv2nix = {{
      url = "github:pyproject-nix/uv2nix";
      inputs.pyproject-nix.follows = "pyproject-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    }};
    pyproject-build-systems = {{
      url = "github:pyproject-nix/build-system-pkgs";
      inputs.pyproject-nix.follows = "pyproject-nix";
      inputs.uv2nix.follows = "uv2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    }};
  }};
  outputs = {{
    nixpkgs,
    pyproject-nix,
    uv2nix,
    pyproject-build-systems,
    ...
  }}:
    let
      inherit (nixpkgs) lib;
      forAllSystems = lib.genAttrs lib.systems.flakeExposed;
      workspace = uv2nix.lib.workspace.loadWorkspace {{ workspaceRoot = ./.; }};
      overlay = workspace.mkPyprojectOverlay {{
        sourcePreference = "wheel";
      }};
      editableOverlay = workspace.mkEditablePyprojectOverlay {{
        root = "$REPO_ROOT";
      }};
      pythonSets = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${{system}};
          python = pkgs.python3;
        in
        (pkgs.callPackage pyproject-nix.build.packages {{
          inherit python;
        }}).overrideScope
          (
            lib.composeManyExtensions [
              pyproject-build-systems.overlays.wheel
              overlay
            ]
          )
      );
    in
    {{
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${{system}};
          pythonSet = pythonSets.${{system}}.overrideScope editableOverlay;
          virtualenv = pythonSet.mkVirtualEnv "{node_name}-dev-env" workspace.deps.all;
        in
        {{
          default = pkgs.mkShell {{
            packages = [
              virtualenv
              pkgs.uv
            ];
            env = {{
              UV_NO_SYNC = "1";
              UV_PYTHON = pythonSet.python.interpreter;
              UV_PYTHON_DOWNLOADS = "never";
            }};
            shellHook = ''
              unset PYTHONPATH
              export REPO_ROOT=$(git rev-parse --show-toplevel)
            '';
          }};
        }}
      );
      packages = forAllSystems (system: {{
        default = pythonSets.${{system}}.mkVirtualEnv "{node_name}-env" workspace.deps.default;
      }});
    }};
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
            node_name=node_name, node_type="Rust", run_command="cargo run"
        )
    )

    print(f"✓ Created Rust node: {node_dir}/")

    # # Update master flake
    # if add_node_to_master_flake(node_name, "rust"):
    #     print(f"✓ Node integrated into monorepo")
    #     print(f"\n💡 Run 'nix flake lock' to update lockfile")

    print("\n📦 Next steps:")
    print(f"  cd {node_dir}")
    print("  nix develop        # Enter dev shell")
    print("  cargo build        # Build the project")
    print("  cargo run          # Run the node")


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

    pkg_dir = src_dir / node_name
    pkg_dir.mkdir()

    # Write files
    main_file = pkg_dir / "__init__.py"
    main_file.write_text(PYTHON_MAIN_TEMPLATE.format(node_name=node_name))
    main_file.chmod(0o755)

    (node_dir / "flake.nix").write_text(
        PYTHON_FLAKE_TEMPLATE.format(node_name=node_name)
    )
    (node_dir / ".gitignore").write_text(GITIGNORE_TEMPLATE)
    (node_dir / "README.md").write_text(
        README_TEMPLATE.format(
            node_name=node_name, node_type="Python", run_command="uv run main"
        )
    )

    (node_dir / "pyproject.toml").write_text(
        PYTHON_PYROJ_TEMPLATE.format(node_name=node_name)
    )

    print(f"✓ Created Python node: {node_dir}/")

    # # Update master flake
    # if add_node_to_master_flake(node_name, "python"):
    #     print(f"✓ Node integrated into monorepo")
    #     # print(f"\n💡 Run 'nix flake lock' to update lockfile")

    print("\n📦 Next steps:")
    print(f"  cd {node_dir}")
    print("  nix develop       # Enter dev shell")
    print("  uv sync           # install the .venv")
    print("  uv run main       # Run the node")


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
        subprocess.run(["cargo", "fetch"], cwd=node_name)
    elif node_type == "python":
        create_python_node(node_name)
        subprocess.run(["uv", "sync"], cwd=node_name)


if __name__ == "__main__":
    main()
