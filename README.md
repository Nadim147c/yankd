# Yankd

A wayland native clipboard manager that implement `wlr-data-control-unstable-v1`.

### Install

#### NixOS

##### Try out

```bash
nix run github:Nadim147c/yankd --help

```

##### Use flake

```nix
yankd = {
    url = "github:Nadim147c/yankd";
    inputs.nixpkgs.follows = "nixpkgs";
    inputs.flake-parts.follows = "flake-parts"; # Optional
};
```

#### Manual

> Requires `git`, `make`, `install` and `go` installed.

```bash
git clone https://github.com/Nadim147c/yankd.git
cd yankd
make build
make install PREFIX=$HOME/.local/
```

#### Go Install

```
go install https://github.com/Nadim147c/yankd@latest
```
