{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-sub.url = "./rust_sub";
    python-sub.url = "./python_sub";
    mynode.url = "./mynode";

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
        rustApp = inputs.rust-sub.packages.${system}.default;
        pythonApp = inputs.python-sub.packages.${system}.default;
        myNodeApp = inputs.mynode.packages.${system}.default;

        default = self.packages.${system}.demo;
        # Launcher: Spins up all with shared config
        demo = syspkgs.writeShellApplication {
          name = "demo-ping-pong-zenoh";
          runtimeInputs = [
            self.packages.${system}.rustApp
            self.packages.${system}.pythonApp

          ];
          text = ''
            export ZENOH_CONFIG=${sharedConfig}
            echo "Launching with shared config: $ZENOH_CONFIG"

            hello &
            PYTHON_PID=$!
            rust-zenoh-app &
            RUST_PID=$!

            # trap 'kill $PYTHON_PID $RUST_PID 2>/dev/null' EXIT

            wait $PYTHON_PID $RUST_PID
            # wait  # Or use trap for signals/cleanup
          '';
        };
      };

      devShells.default = syspkgs.mkShell {
        packages = [
          # pkgs.zenoh
          self.packages.${system}.default
          self.packages.${system}.rustApp
          self.packages.${system}.pythonApp
          syspkgs.just

          self.packages.${system}.myNodeApp


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
