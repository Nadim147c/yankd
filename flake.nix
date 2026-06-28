{

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      inherit (nixpkgs) lib;

      perSystem =
        f:
        lib.genAttrs systems (
          system:
          let
            pkgs = import nixpkgs { inherit system; };
          in
          f { inherit lib system pkgs; }
        );
    in
    {
      packages = perSystem (
        { pkgs, system, ... }: {
          duckdb-rapidfuzz-extension = pkgs.callPackage ./nix/duckdb-rapidfuzz-extension.nix { };
          yankd = pkgs.callPackage ./nix/package.nix {
            inherit (self.packages.${system}) duckdb-rapidfuzz-extension;
          };
          default = self.packages.${system}.yankd;

          ci-go-vet = pkgs.writeShellApplication {
            name = "go-vet";
            runtimeInputs = with pkgs; [ go ];
            text = ''
              go vet -v ./...
            '';
          };

          ci-go-test = pkgs.writeShellApplication {
            name = "go-test";
            runtimeInputs = with pkgs; [ go ];
            text = ''
              go test -v ./...
            '';
          };

          ci-go-lint = pkgs.writeShellApplication {
            name = "go-test";
            runtimeInputs = with pkgs; [ golangci-lint ];
            text = ''
              golangci-lint run
            '';
          };

          ci-format = pkgs.writeShellApplication {
            name = "format";
            runtimeInputs = with pkgs; [ gofumpt ];
            text = ''
              gofumpt -d -e .
            '';
          };

          ci-go-mod-tidy = pkgs.writeShellApplication {
            name = "go-mod-tidy";
            runtimeInputs = with pkgs; [
              go
              git
            ];
            text = ''
              go mod tidy
              env PAGER= git diff --exit-code
            '';
          };
        }
      );
      devShells = perSystem (
        { pkgs, ... }: {
          default = pkgs.callPackage ./nix/shell.nix { };
        }
      );
    };
}
