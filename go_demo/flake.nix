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
        {
          packages.default = pkgs.buildGoModule {
            pname = "go_demo";
            version = "0.1.0";
            src = ./.;
            vendorHash = null;
            # vendorHash = "sha256-qaRFf+OrzBHZYBQT0u3RiTJ9XtS6/s9P0UGjhmnLIso=";

            env.CGO_ENABLED = "1";
            proxyVendor = true;

            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.zenoh-c ];
          };

        };
    };
}
