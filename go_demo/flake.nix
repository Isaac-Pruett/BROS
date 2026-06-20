{
  description = "go_demo — Go Zenoh Node";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-parts.inputs.nixpkgs-lib.follows = "nixpkgs";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        let
          zenoh-c = pkgs.callPackage ../.build_utils/zenoh-c.nix { };
        in
        {
          packages.default = pkgs.buildGoModule {
            pname = "go_demo";
            version = "0.1.0";
            src = ./.;

            vendorHash = "sha256-Z2kguOCLppJNOownC6+ojtkcHpdwptHM5sUAo310els=";
            proxyVendor = true;

            env.CGO_ENABLED = "1";
            env.PKG_CONFIG_ALL_STATIC = "1";

            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ zenoh-c ];
          };
        };
    };
}
