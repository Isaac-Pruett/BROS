{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-demo-sub.url = "./rust_sub";
    python-demo-sub.url = "./python_sub";
    simpledemo.url = "./simpledemo";

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
      packages = {
        rust_demo = inputs.rust-demo-sub.packages.${system}.default;
        python_demo = inputs.python-demo-sub.packages.${system}.default;

        my_node = inputs.simpledemo.packages.${system}.default;

        default = self.packages.${system}.demo;


        # Launcher: Spins up all with shared config
        demo = syspkgs.writeShellApplication {
          name = "demo-ping-pong-zenoh";
          runtimeInputs = [
            self.packages.${system}.rust_demo
            self.packages.${system}.python_demo

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
          self.packages.${system}.demo
          self.packages.${system}.rust_demo
          self.packages.${system}.python_demo
          syspkgs.just

          self.packages.${system}.my_node


        ];

        env.ZENOH_CONFIG = sharedConfig;

        shellHook = ''
          # export ZENOH_CONFIG=${sharedConfig}
          alias j="just"
          echo "Master dev shell ready. Run 'demo-ping-pong-zenoh' to start demo."
        '';
      };
    });
}
