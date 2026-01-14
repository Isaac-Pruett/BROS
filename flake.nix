{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    rust-demo-sub.url = "./rust_sub";
    python-demo-sub.url = "./python_sub";
    simpledemo.url = "./simpledemo";
  };

  outputs = inputs @ { flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      perSystem = { config, self', inputs', pkgs, system, ... }: let
        # Generate shared Zenoh config (customize as needed; could derive from template)
        sharedConfig = pkgs.writeText "zenoh-config.json" ''
          {
            "mode": "peer",
            "listen": { "endpoints": ["tcp/127.0.0.1:0"] },
            "scouting": {
              "multicast": {
                "enabled": true
              }
            }
          }
        '';
      in {
        # Expose subproject packages for composition
        packages = {
          rust_demo = inputs.rust-demo-sub.packages.${system}.default;
          python_demo = inputs.python-demo-sub.packages.${system}.default;
          my_node = inputs.simpledemo.packages.${system}.default;

          # Launcher: Spins up all with shared config
          demo = pkgs.writeShellApplication {
            name = "demo-ping-pong-zenoh";
            runtimeInputs = [
              self'.packages.rust_demo
              self'.packages.python_demo
            ];
            text = ''
              export ZENOH_CONFIG=${sharedConfig}
              echo "Launching with shared config: $ZENOH_CONFIG"
              # Start processes in background
              hello &
              PYTHON_PID=$!
              rust-zenoh-app &
              RUST_PID=$!
              # Trap EXIT and SIGINT (Ctrl+C)
              trap 'kill $PYTHON_PID $RUST_PID 2>/dev/null' EXIT SIGINT
              # Wait for both processes
              wait $PYTHON_PID $RUST_PID
            '';
          };

          default = self'.packages.demo;
        };

        devShells.default = pkgs.mkShell {
          packages = [
            self'.packages.demo
            self'.packages.rust_demo
            self'.packages.python_demo

            self'.packages.my_node
            pkgs.just
          ];
          env.ZENOH_CONFIG = sharedConfig;
          shellHook = ''
            export ZENOH_CONFIG=${sharedConfig}
            alias j="just"
            echo "Run 'demo-ping-pong-zenoh' to start demo."
            echo "just is aliased to j"
          '';
        };
      };
    };
}
