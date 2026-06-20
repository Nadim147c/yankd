{ stdenvNoCC, fetchurl }:
let
  platforms = {
    "x86_64-linux" = "linux_amd64";
    "aarch64-linux" = "linux_arm64";
  };
  hashs = {
    "x86_64-linux" = "sha256-4IhTXVvdlOKw9HaE9libfyJhL7yxLmrsH96TzT9yZb4=";
    "aarch64-linux" = "sha256-u3zEEdCiWjDmTZQKvrBgxk+Q6TrsryRcFqiJqcOEeAg=";
  };
  platform = platforms."${stdenvNoCC.hostPlatform.system}";
  hash = hashs."${stdenvNoCC.hostPlatform.system}";
in
stdenvNoCC.mkDerivation (finalAttrs: {
  inherit platform;
  pname = "duckdb-fts-extension";
  version = "v1.5.3";
  src = fetchurl {
    inherit hash;
    url = "http://extensions.duckdb.org/${finalAttrs.version}/${finalAttrs.platform}/fts.duckdb_extension.gz";
  };
  dontUnpack = true;
  installPhase = ''
    mkdir -p $out/$version/$platform
    gunzip $src -c > $out/$version/$platform/fts.duckdb_extension
  '';
})
