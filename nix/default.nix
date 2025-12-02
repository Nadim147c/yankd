{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "yankd";
  version = "0-unstable-2025-12-02";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "yankd";
    rev = "0d54c956d4f612b84defd108c7eaebd479136777";
    hash = "sha256-mF0jGRyFcX1nvA/0ABsE+ImnwYzlPGDB+HcSPVdvNs0=";
  };

  vendorHash = "sha256-SHzl0X3EqkE0aBHbHFgyeHG6T7j9T3vWaqmZG8x6J2Q=";

  ldflags = ["-s" "-w" "-X" "main.version=${version}"];

  meta = {
    description = "A (WIP) wayland native clipboard manager";
    homepage = "https://github.com/Nadim147c/yankd";
    license = lib.licenses.gpl3Only;
    mainProgram = "yankd";
  };
}
