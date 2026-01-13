{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-demo-sub.url = "./rust_sub";
    python-demo-sub.url = "./python_sub";
    simpledemo.url = "./simpledemo";
    camera-node.url = "./camera";

  };

  outputs = { self, nixpkgs, flake-utils, ...} @ inputs:
    flake-utils.lib.eachDefaultSystem (system: let
      syspkgs = nixpkgs.legacyPackages.${system};

      # Generate shared Zenoh config (customize as needed; could derive from template)
      sharedConfig = syspkgs.writeText "zenoh-config.json" ''
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
      nodes = {
        rust_demo = inputs.rust-demo-sub.packages.${system}.default;
        python_demo = inputs.python-demo-sub.packages.${system}.default;

        my_node = inputs.simpledemo.packages.${system}.default;

        default = self.packages.${system}.demo;

        camera = inputs.camera-node.packages.${system}.default;


        # Launcher: Spins up all with shared config
        demo = syspkgs.writeShellApplication {
          name = "demo-ping-pong-zenoh";
          runtimeInputs = [
            self.nodes.${system}.rust_demo
            self.nodes.${system}.python_demo

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
      };

      devShells.default = syspkgs.mkShell {
        packages = [
          # pkgs.zenoh
          self.nodes.${system}.demo
          self.nodes.${system}.rust_demo
          self.nodes.${system}.python_demo
          syspkgs.just

          self.nodes.${system}.my_node
          self.nodes.${system}.camera

        ];

        env.ZENOH_CONFIG = sharedConfig;

        shellHook = ''
          export ZENOH_CONFIG=${sharedConfig}
          alias j="just"
          echo "Run 'demo-ping-pong-zenoh' to start demo."
          echo "just is aliased to j"
        '';
      };
    });
}
