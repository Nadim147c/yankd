{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };
  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = {pkgs, ...}: {
        packages.default = pkgs.callPackage ./nix/default.nix {};
        devShells.default = pkgs.callPackage ./nix/shell.nix {};
      };

      # TODO: create a flake module
      # flake = {
      #   homeModules.yankd = import ./module.nix self;
      #   homeModules.default = self.homeModules.yankd;
      # };
    };
}
