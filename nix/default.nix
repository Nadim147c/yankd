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
  version = "0.0.1-dev-unstable-2025-12-03";

  src = fetchFromGitHub {
    owner = "Nadim147c";
    repo = "yankd";
    rev = "df62d08303bcb8068e570b7c84b789ae39baf4cd";
    hash = "sha256-clROQrBVCLWTPp3JD5teXsTLV0pHFZdiX3GnI6nS3/s=";
  };

  vendorHash = "sha256-SHzl0X3EqkE0aBHbHFgyeHG6T7j9T3vWaqmZG8x6J2Q=";

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
