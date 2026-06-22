{ pkgs }:
pkgs.mkShell {
  name = "yankd";
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
