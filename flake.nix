{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-sub.url = "./rust_sub";
    python-sub.url = "./python_sub";




  };

  outputs = { self, nixpkgs, flake-utils, rust-sub, python-sub }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

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
        rustApp = rust-sub.packages.${system}.default;
        pythonApp = python-sub.packages.${system}.default;

        default = self.packages.${system}.demo;
        # Launcher: Spins up all with shared config
        demo = pkgs.writeShellApplication {
          name = "launch-zenoh-apps";
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

      devShells.default = pkgs.mkShell {
        packages = [
          # pkgs.zenoh
          self.packages.${system}.default
          self.packages.${system}.rustApp
          self.packages.${system}.pythonApp
          pkgs.just

        ];
        shellHook = ''
          export ZENOH_CONFIG=${sharedConfig}
          export j="just"
          echo "Master dev shell ready. Run 'launch-zenoh-apps' to start all."
        '';
      };
    });
}
