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
        # Allows definition of system-specific attributes
        # without needing to declare the system explicitly!

        version = "1.9.0";  # ← change to a newer tag when available

        zenohd = pkgs.rustPlatform.buildRustPackage rec {
          pname = "zenohd";
          inherit version;

          src = pkgs.fetchFromGitHub {
            owner = "eclipse-zenoh";
            repo = "zenoh";
            rev = version;                     # or "main" / commit hash if you prefer bleeding-edge
            hash = "sha256-0y9C7A42eYTfnt1mjYKvtqg78KZsH9jB6OdskSYDwnM="; # ← placeholder, see below
          };

          # We build only the zenohd binary from the workspace
          cargoBuildFlags = [ "--bin" "zenohd" ];

          # Important: build all workspace members so features & dependencies resolve correctly
          # (zenohd depends on many workspace crates)
          doCheck = false;   # tests can be slow/heavy; set true if you want them

          cargoHash = "sha256-AO83hWzpD7T4iEfGf6p+XPuIfuki5mSkUgR8A3PajH0="; # ← placeholder

          nativeBuildInputs = with pkgs; [
            pkg-config
          ];

          buildInputs = with pkgs; [
            openssl
            # add more if linking errors appear (e.g. libclang, protobuf, etc.)
          ];

          meta = with pkgs.lib; {
            description = "Zenoh router daemon (zenohd)";
            homepage = "https://zenoh.io";
            license = licenses.epl20;
            maintainers = [ ];
            mainProgram = "zenohd";
            platforms = platforms.linux ++ platforms.darwin;
          };
        };
      in {
        # Expose subproject packages for composition
        packages = {

          zenohd = zenohd;

          default = self'.packages.zenohd;

        };

        devShells.default = pkgs.mkShell {
          packages = [
            self'.packages.default
          ];

          shellHook = ''

          '';
        };
      };
      flake = {
        # The usual flake attributes can be defined here, including
        # system-agnostic and/or arbitrary outputs.
      };
    };
}
