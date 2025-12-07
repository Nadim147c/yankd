{
  buildGoModule,
  fetchFromGitHub,
  installShellFiles,
  lib,
  makeBinaryWrapper,
  stdenv,
  versionCheckHook,
  # TODO: remove wl-clipboard dependencies
  wl-clipboard,
}:
buildGoModule rec {
  pname = "yankd";
  version = "0.0.1-dev.3-unstable-2025-12-07";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "yankd";
    rev = "1a9b3e0c3d9a6fa93d9327216c4f487b568ea675";
    hash = "sha256-slVMMuqyYaF50rKkGTDPtpjySKkGZcuQb7FtgZopNcY=";
  };

  vendorHash = "sha256-qmKm1Y4q43hWRdF1leT+2UujX9VlBJmpP51rxhpnBc4=";

  nativeBuildInputs = [installShellFiles makeBinaryWrapper];
  propagatedBuildInputs = [wl-clipboard];

  nativeInstallCheckInputs = [versionCheckHook];
  versionCheckProgramArg = "--version";

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
