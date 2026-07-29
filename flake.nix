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

          ci-test = pkgs.writeShellApplication {
            name = "test";
            runtimeInputs = with pkgs; [ go ];
            text = ''
              go test -v ./...
            '';
          };

          ci-lint = pkgs.writeShellApplication {
            name = "lint";
            runtimeInputs = with pkgs; [
              coreutils
              deadnix
              fd
              go
              golangci-lint
              gotools
              ripgrep
              codespell
            ];
            text = ''
              golangci-lint run
              go fix -diff ./...
              go vet -v ./...
              codespell \
                -xinternal/wlr-data-control-unstable-v1/wlr-data-control-unstable-v1.xml \
                .

              deadcode_result=$(mktemp)
              deadcode ./... |
                rg -v "(ipc)" | # Allowed dead codes
                tee "$deadcode_result" && [ ! -s "$deadcode_result" ] # rest is error

              fd --type file '\.nix$' --exec-batch deadnix -f {}
            '';
          };

          ci-format = pkgs.writeShellApplication {
            name = "format";
            runtimeInputs = with pkgs; [
              gofumpt
              fd
              nixfmt
            ];
            text = ''
              gofumpt -d -e .
              fd --type file '\.nix$' --exec-batch nixfmt -c {}
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
