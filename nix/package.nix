# Do not call this package with regular arguments
{
  lib,
  stdenv,
  buildGoModule,
  installShellFiles,
  duckdb-fts-extension,
  writableTmpDirAsHomeHook,
}:
let
  inherit (lib) cleanSource optionalString licenses;
  inherit (lib.fileset) toSource unions;
in
buildGoModule rec {
  pname = "yankd";
  version = "0-unstable-2025-03-06";

  src = cleanSource (toSource {
    root = ../.;
    fileset = unions [
      ../cmd
      ../internal
      ../main.go
      ../go.mod
      ../go.sum
    ];
  });

  postPatch = ''
    substituteInPlace internal/db/configure.sql \
      --replace-fail "INSTALL fts;" "FORCE INSTALL fts FROM '${duckdb-fts-extension}';"
  '';

  vendorHash = "sha256-bNlCh31+Ws73i169TaOIBkTKYdX/r5t3k7gUX7XCEY4=";

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
