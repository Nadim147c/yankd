{
  buildGoModule,
  fetchFromGitHub,
  installShellFiles,
  lib,
  makeBinaryWrapper,
  stdenv,
  # TODO: remove wl-clipboard dependencies
  wl-clipboard,
}:
buildGoModule rec {
  pname = "yankd";
  version = "0.0.1-dev.3-unstable-2025-12-04";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "yankd";
    rev = "1ff598e602b888b31975d115118f4de9cb7d3e5f";
    hash = "sha256-1bd4YepxuIL8q55IpbbyX5G2093apD+YPA0/e/QBMMo=";
  };

  vendorHash = "sha256-qmKm1Y4q43hWRdF1leT+2UujX9VlBJmpP51rxhpnBc4=";

  nativeBuildInputs = [installShellFiles makeBinaryWrapper];
  propagatedBuildInputs = [wl-clipboard];

  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
    installShellCompletion --cmd yankd \
      --bash <($out/bin/yankd _carapace bash) \
      --fish <($out/bin/yankd _carapace fish) \
      --zsh <($out/bin/yankd _carapace zsh)

    wrapProgram $out/bin/yankd \
      --prefix PATH : ${lib.makeBinPath [wl-clipboard]}
  '';

  ldflags = ["-s" "-w" "-X" "main.version=${version}"];

  meta = {
    description = "A wayland native clipboard manager";
    homepage = "https://github.com/Nadim147c/yankd";
    license = lib.licenses.gpl3Only;
    mainProgram = "yankd";
  };
}
