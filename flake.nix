{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    systems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];

    forAllSystems = f:
      nixpkgs.lib.genAttrs systems (
        system: f (import nixpkgs {inherit system;})
      );
  in {
    packages = forAllSystems (pkgs: {
      default = pkgs.callPackage ./nix/package.nix {};
    });

    devShells = forAllSystems (pkgs: {
      default = pkgs.callPackage ./nix/shell.nix {};
    });

    homeModules.yankd = import ./nix/home-module.nix self;
    homeModules.default = self.homeModules.yankd;
  };
}
