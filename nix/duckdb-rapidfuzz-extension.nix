{ stdenvNoCC, fetchurl }:
let
  meta = {
    "x86_64-linux" = {
      platform = "linux_amd64";
      hash = "sha256-/sh6bm26Y6O74s0GGSKW90urm9EzmH3E5qxSKQll0U8=";
    };
    "aarch64-linux" = {
      platform = "linux_arm64";
      hash = "";
    };
  };
  inherit (meta."${stdenvNoCC.hostPlatform.system}") platform hash;
in
stdenvNoCC.mkDerivation (finalAttrs: {
  inherit platform;
  pname = "duckdb-rapidfuzz-extension";
  version = "v1.5.3";
  src = fetchurl {
    inherit hash;
    url = "http://community-extensions.duckdb.org/${finalAttrs.version}/${finalAttrs.platform}/rapidfuzz.duckdb_extension.gz";
  };
  dontUnpack = true;
  installPhase = ''
    mkdir -p $out/share/duckdb
    gunzip $src -c > $out/share/duckdb/rapidfuzz.duckdb_extension
  '';
})
