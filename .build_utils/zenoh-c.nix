{
  lib,
  stdenv,
  fetchFromGitHub,
  cmake,
  cargo,
  rustPlatform,
  rustc,
  fixDarwinDylibNames,
}:

stdenv.mkDerivation rec {
  pname = "zenoh-c";
  version = "1.9.0";

  src = fetchFromGitHub {
    owner = "eclipse-zenoh";
    repo = "zenoh-c";
    tag = version;
    hash = "sha256-7Qe2Hefvp09DVSQuaRcoDBuvDSOO3YRdQFSXbR6t/DA=";
  };

  env.CARGO_HOME = "/tmp/cargo-home";

  cargoDeps = rustPlatform.fetchCargoVendor {
    inherit src pname version;
    hash = "sha256-DV2jt3E+dVhn9YeElzzCXQv7phqye+octHdGprKJEy4=";
  };

  outputs = [
    "out"
    "dev"
  ];

  nativeBuildInputs = [
    cmake
    cargo
    rustPlatform.cargoSetupHook
    rustc
  ]
  ++ lib.optionals stdenv.hostPlatform.isDarwin [
    fixDarwinDylibNames
  ];

  cmakeFlags = [
    "-DZENOHC_BUILD_WITH_UNSTABLE_API=ON"
    "-DBUILD_SHARED_LIBS=OFF"
  ];

  postInstall = ''
    substituteInPlace $out/lib/pkgconfig/zenohc.pc \
      --replace-fail "\''${prefix}/" ""
  '';

  postFixup = lib.optionalString stdenv.hostPlatform.isDarwin ''
    if [ -e "$out/lib/libzenohc.dylib" ]; then
      install_name_tool -id "$out/lib/libzenohc.dylib" "$out/lib/libzenohc.dylib"
    fi
  '';

  meta = {
    description = "C API for zenoh";
    homepage = "https://github.com/eclipse-zenoh/zenoh-c";
    license = with lib.licenses; [
      asl20
      epl20
    ];
    maintainers = with lib.maintainers; [ markuskowa ];
  };
}
