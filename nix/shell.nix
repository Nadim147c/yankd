{ pkgs }:
pkgs.mkShell {
  name = "yankd";
  env.CGO_ENABLED = "1";
  nativeBuildInputs = with pkgs; [
    pkg-config
    stdenv.cc
  ];
  # Additional tooling
  buildInputs = with pkgs; [
    nix-fast-build
    go
    gofumpt
    golines
    gopls
    golangci-lint-langserver
    sql-formatter
    watchexec
    duckdb
    just
    gotestsum
  ];
}
