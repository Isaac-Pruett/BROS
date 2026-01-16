{
  description = "camera — Rust Zenoh Node";
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

        gstPluginPath = pkgs.lib.makeSearchPath "lib/gstreamer-1.0" (with pkgs.gst_all_1; [
          gstreamer
          gst-plugins-base
          gst-plugins-good
          gst-plugins-bad
          gst-plugins-ugly
          gst-libav
        ]);

        unwrappedCamera = naersk-lib.buildPackage {
          src = ./.;
          pname = "camera";
          nativeBuildInputs = with pkgs; [
            pkg-config
          ];
          buildInputs = with pkgs; [
            gst_all_1.gstreamer
            gst_all_1.gst-plugins-base
            gst_all_1.gst-plugins-good
            gst_all_1.gst-plugins-bad
            gst_all_1.gst-plugins-ugly
            gst_all_1.gst-libav
          ];
        };
      in
      {
        packages.default = pkgs.symlinkJoin {
          name = "camera";
          paths = [ unwrappedCamera ];
          buildInputs = [ pkgs.makeWrapper ];
          postBuild = ''
            wrapProgram $out/bin/camera \
              --set GST_PLUGIN_SYSTEM_PATH_1_0 "${gstPluginPath}"
          '';
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            cargo
            rustc
            rustfmt
            pre-commit
            rustPackages.clippy
            pkg-config
            gst_all_1.gstreamer
            gst_all_1.gst-plugins-base
            gst_all_1.gst-plugins-good
            gst_all_1.gst-plugins-bad
            gst_all_1.gst-plugins-ugly
            gst_all_1.gst-libav
          ];
          RUST_SRC_PATH = pkgs.rustPlatform.rustLibSrc;
          GST_PLUGIN_SYSTEM_PATH_1_0 = gstPluginPath;
        };

        apps.default = utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };
      }
    );
}
