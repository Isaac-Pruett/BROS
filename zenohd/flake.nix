{
  description = "zenohd (Eclipse Zenoh router daemon) built from github:eclipse-zenoh/zenoh";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };
  outputs = inputs @ { flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      imports = [];
      perSystem = { config, self', inputs', pkgs, system, ... }: let
        version = "1.9.0";  # ← only thing you need to change on upgrade
        zenohd = pkgs.rustPlatform.buildRustPackage rec {
          pname = "zenohd";
          inherit version;
          src = pkgs.fetchFromGitHub {
            owner = "eclipse-zenoh";
            repo = "zenoh";
            rev = "refs/tags/${version}";  # ← unambiguously targets a tag
            hash = "sha256-sFHUphFu5a+buSa3GQvSmGo8SFtn3V5ZqTOnWMPlvs8=";      # ← nix build will print the real one
          };
          cargoBuildFlags = [ "--bin" "zenohd" ];
          doCheck = false;
          cargoHash = "sha256-1PjtZ5/bAnLlMbkcKAA6DCKDafItGiATjct5Pv8muas=";   # ← same: nix build prints the real one
          nativeBuildInputs = with pkgs; [ pkg-config ];
          buildInputs = with pkgs; [ openssl ];
          meta = with pkgs.lib; {
            description = "Zenoh router daemon (zenohd)";
            homepage = "https://zenoh.io";
            license = licenses.epl20;
            maintainers = [];
            mainProgram = "zenohd";
            platforms = platforms.linux ++ platforms.darwin;
          };
        };
      in {
        packages = {
          zenohd = zenohd;
          default = self'.packages.zenohd;
        };
        devShells.default = pkgs.mkShell {
          packages = [ self'.packages.default ];
          shellHook = "";
        };
      };
      flake = {};
    };
}
