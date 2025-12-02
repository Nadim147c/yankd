{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "yankd";
  version = "0.0.1-dev.1";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "yankd";
    rev = "v${version}";
    hash = "sha256-EjR+c9X01A3wRbP0ng20d120wiZ4xNfQ+HsfjlcoMZE=";
  };

  vendorHash = "sha256-iK/YLHHNFrLyHRuQsAbBQ8qB8PQ98uyYTPWDfsQ14m0=";

  ldflags = ["-s" "-w"];

  meta = {
    description = "A (WIP) wayland native clipboard manager";
    homepage = "https://github.com/Nadim147c/yankd";
    license = lib.licenses.gpl3Only;
    mainProgram = "yankd";
  };
}
