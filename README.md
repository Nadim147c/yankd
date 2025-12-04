# Yankd

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Nadim147c/yankd?style=for-the-badge&logo=go&labelColor=11140F&color=BBE9AA)](https://pkg.go.dev/github.com/Nadim147c/yankd)
[![GitHub Repo stars](https://img.shields.io/github/stars/Nadim147c/yankd?style=for-the-badge&logo=github&labelColor=11140F&color=BBE9AA)](https://github.com/Nadim147c/yankd)
[![GitHub License](https://img.shields.io/github/license/Nadim147c/yankd?style=for-the-badge&logo=gplv3&labelColor=11140F&color=BBE9AA)](./LICENSE)
[![GitHub Tag](https://img.shields.io/github/v/tag/Nadim147c/yankd?include_prereleases&sort=semver&style=for-the-badge&logo=git&labelColor=11140F&color=BBE9AA)](https://github.com/Nadim147c/yankd/tags)
[![Git Commit](https://img.shields.io/github/last-commit/Nadim147c/yankd?style=for-the-badge&logo=git&labelColor=11140F&color=BBE9AA)](https://github.com/Nadim147c/yankd/tags)

> [!CAUTION]
> 🚧 **Highly Experimental & Unstable**. This project is in active development and
> may break at any time. Expect bugs, missing features, unexpected behavior, and
> frequent changes.

A wayland native clipboard manager that implement `wlr-data-control-unstable-v1`.

## Install

### NixOS

#### Try out

```bash
nix run github:Nadim147c/yankd --help

```

#### Use flake

```nix
yankd = {
    url = "github:Nadim147c/yankd";
    inputs.nixpkgs.follows = "nixpkgs";
    inputs.flake-parts.follows = "flake-parts"; # Optional
};
```

### Manual

> Requires `git`, `make`, `install` and `go` installed.

```bash
git clone https://github.com/Nadim147c/yankd.git
cd yankd
make build
make install PREFIX=$HOME/.local/
```

### Go Install

```
go install https://github.com/Nadim147c/yankd@latest
```
