{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    rust-demo-sub.url = "./rust_sub";
    python-demo-sub.url = "./python_sub";
    simpledemo.url = "./simpledemo";
    camera-package.url = "./camera";
  };

  outputs = inputs @ { flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      imports = [
        # ./module.nix
        # inputs.foo.flakeModule
      ];

      perSystem = { config, self', inputs', pkgs, system, ... }: let
        # Allows definition of system-specific attributes
          # without needing to declare the system explicitly!
          #
          # Quick rundown of the provided arguments:
          # - config is a reference to the full configuration, lazily evaluated
          # - self' is the outputs as provided here, without system. (self'.packages.default)
          # - inputs' is the input without needing to specify system (inputs'.foo.packages.bar)
          # - pkgs is an instance of nixpkgs for your specific system
          # - system is the system this configuration is for

          # This is equivalent to packages.<system>.default

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
          camera = inputs.camera-package.packages.${system}.default;

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
            self'.packages.camera
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
      flake = {
        # The usual flake attributes can be defined here, including
        # system-agnostic and/or arbitrary outputs.
      };
    };
}
