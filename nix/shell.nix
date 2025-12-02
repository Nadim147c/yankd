{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  name = "yankd";
  # Get dependencies from the main package
  inputsFrom = [(pkgs.callPackage ./default.nix {})];
  # Additional tooling
  buildInputs = with pkgs; [
    gnumake
    go
    gofumpt
    golines
    gopls
    revive
    sql-formatter
  ];
}
