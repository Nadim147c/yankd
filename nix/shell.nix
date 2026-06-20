{
  gomod2nix,
  pkgs ? import <nixpkgs> { },
}:
pkgs.mkShell {
  name = "yankd";
  # Additional tooling
  buildInputs = with pkgs; [
    gomod2nix
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
