{
  description = "mavlink_bridge — Rust Zenoh Node";
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
        packages.default = naersk-lib.buildPackage {
          src = ./.;
          pname = "mavlink_bridge";

          nativeBuildInputs = [
            pkgs.git
          ];

          # Optional: customize if needed
          # doCheck = true;
          # release = false;  # already default in naersk
        };
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
        apps.default = utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };
      }
    );
}
