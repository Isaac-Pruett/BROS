{
  description = "Rust Zenoh subproject using naersk";

  inputs = {
    naersk.url = "github:nix-community/naersk/master";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils, naersk }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        naersk-lib = pkgs.callPackage naersk { };
      in
      {
        # Modern way: define packages.default
        packages.default = naersk-lib.buildPackage {
          src = ./.;
          pname = "rust-zenoh-app";
          # Optional: customize if needed
          # doCheck = true;
          release = false;  # already default in naersk
        };

        # Modern way: devShells.default
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            cargo
            rustc
            rustfmt
            pre-commit
            rustPackages.clippy
          ];
          RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
        };

        # Optional: expose as an app (for nix run)
        apps.default = utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };
      }
    );
}
