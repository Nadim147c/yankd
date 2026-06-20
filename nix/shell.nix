{
  pkgs ? import <nixpkgs> { },
}:
pkgs.mkShell {
  name = "yankd";
  # Get dependencies from the main package
  inputsFrom = [ (pkgs.callPackage ./package.nix { }) ];
  # Additional tooling
  buildInputs = with pkgs; [
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
