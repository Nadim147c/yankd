{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };
  outputs = inputs @ {
    self,
    flake-parts,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = {pkgs, ...}: {
        packages.default = pkgs.callPackage ./nix/package.nix {};
        devShells.default = pkgs.callPackage ./nix/shell.nix {};
      };

      flake = {
        homeModules.yankd = import ./nix/home-module.nix self;
        homeModules.default = self.homeModules.yankd;
      };
    };
}
