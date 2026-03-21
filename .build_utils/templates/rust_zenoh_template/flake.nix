{
  description = "rust_zenoh_template — Rust Zenoh Node";

  inputs = {
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; # This should point to whichever nixpkgs rev you want.
    naersk.url = "github:nix-community/naersk/master";

    rust-overlay.url = "github:oxalica/rust-overlay";

  };

  outputs = { flake-parts, ... } @ inputs: flake-parts.lib.mkFlake { inherit inputs; } {
    imports = [
      # ./module.nix
      # inputs.foo.flakeModule
    ];

    perSystem = { config, self', inputs', pkgs, system, ... }: {



      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          git
          cargo
          rustc
          rustfmt
          pre-commit
          rustPackages.clippy
        ];
        RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
      };

      packages.default =
        let
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ inputs.rust-overlay.overlays.default ];
          };

          naersk-lib = pkgs.callPackage inputs.naersk {

            cargo = pkgs.rust-bin.stable.latest.default;
            rustc = pkgs.rust-bin.stable.latest.default;
          };
        in
        naersk-lib.buildPackage {
          src = ./.;
          pname = "rust_zenoh_template";
          nativeBuildInputs = [ pkgs.git ];
        };
    };

    flake = {
      # The usual flake attributes can be defined here, including
      # system-agnostic and/or arbitrary outputs.
    };

    # Declared systems that your flake supports. These will be enumerated in perSystem
    systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
  };
}
