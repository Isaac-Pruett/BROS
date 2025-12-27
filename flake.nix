{
  description = "Master Zenoh monorepo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    rust-sub.url = "./rust_sub";
    python-sub.url = "./python_sub";

  };

  outputs = { self, nixpkgs, flake-utils, rust-sub, python-sub}:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      # Generate shared Zenoh config (customize as needed; could derive from template)
      sharedConfig = pkgs.writeText "zenoh-config.json" ''
        {
          "mode": "peer",
          "connect": { "endpoints": ["tcp/127.0.0.1:7447"] },
          "listen": { "endpoints": ["tcp/0.0.0.0:7447"] }
        }
      '';
    in {
      # Expose subproject packages for composition
      packages = {
        rustApp = rust-sub.packages.${system}.default;
        pythonApp = python-sub.packages.${system}.default;

        # Launcher: Spins up all with shared config
        default = pkgs.writeShellApplication {
          name = "launch-zenoh-apps";
          runtimeInputs = [
            self.packages.${system}.rustApp
            self.packages.${system}.pythonApp

          ];
          text = ''
            export ZENOH_CONFIG=${sharedConfig}
            echo "Launching with shared config: $ZENOH_CONFIG"
            hello &
            rust-zenoh-app &  # Adjust bin name if needed

            wait  # Or use trap for signals/cleanup
          '';
        };
      };

      devShells.default = pkgs.mkShell {
        packages = [
          pkgs.zenoh
          self.packages.${system}.rustApp
          self.packages.${system}.pythonApp

        ];
        shellHook = ''
          export ZENOH_CONFIG=${sharedConfig}
          echo "Master dev shell ready. Run 'launch-zenoh-apps' to start all."
        '';
      };
    });
}
