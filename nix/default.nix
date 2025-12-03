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
    rev = "2423d1abfb3dcc3001a3872b5d2ef4b4d6691b8c";
    hash = "sha256-a9KOYrXcwEa/YAAVgYcvildoDzNgfb2JcJccHhbk0DI=";
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
