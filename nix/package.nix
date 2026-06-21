# Do not call this package with regular arguments
{
  src,
  lib,
  stdenv,
  installShellFiles,
  buildGoApplication,
  writableTmpDirAsHomeHook,
}:
let
  inherit (lib) optionalString licenses;
in
buildGoApplication rec {
  pname = "yankd";
  version = "0-unstable-2025-03-06";

  inherit src;
  modules = ../gomod2nix.toml;

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
  ];

  nativeBuildInputs = [
    installShellFiles
    writableTmpDirAsHomeHook
  ];
  postInstall = optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) /* bash */ ''
    installShellCompletion --cmd yankd \
      --bash <($out/bin/yankd _carapace bash) \
      --fish <($out/bin/yankd _carapace fish) \
      --zsh <($out/bin/yankd _carapace zsh)
  '';

  meta = {
    description = "A wayland native clipboard manager";
    homepage = "https://github.com/Nadim147c/yankd";
    license = licenses.gpl3Only;
    mainProgram = "yankd";
  };
}
