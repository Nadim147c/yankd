{
  buildGoModule,
  installShellFiles,
  lib,
  stdenv,
  versionCheckHook,
}:
buildGoModule rec {
  pname = "yankd";
  version = "0.0.1-unstable-2025-03-06";

  src = ../.;

  vendorHash = "sha256-6YVIFvV0p7uZvVkuOx1tYOn+xRZ9AqKpNqm4xbWzqo8=";

  nativeBuildInputs = [ installShellFiles ];
  nativeInstallCheckInputs = [ versionCheckHook ];
  versionCheckProgramArg = "--version";

  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) /* bash */ ''
    installShellCompletion --cmd yankd \
      --bash <($out/bin/yankd _carapace bash) \
      --fish <($out/bin/yankd _carapace fish) \
      --zsh <($out/bin/yankd _carapace zsh)
  '';

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  meta = {
    description = "A wayland native clipboard manager";
    homepage = "https://github.com/Nadim147c/yankd";
    license = lib.licenses.gpl3Only;
    mainProgram = "yankd";
  };
}
