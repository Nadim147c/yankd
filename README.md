<h1 align="center">Yankd</h1>
<h3 align="center">A wayland native clipboard manager.</h3>

<h1 align="center">
<a href="https://pkg.go.dev/github.com/Nadim147c/yankd">
<img src="https://img.shields.io/github/go-mod/go-version/Nadim147c/yankd?style=for-the-badge&logo=go&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/yankd">
<img src="https://img.shields.io/github/stars/Nadim147c/yankd?style=for-the-badge&logo=github&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/yankd/blob/main/LICENSE">
<img src="https://img.shields.io/github/license/Nadim147c/yankd?style=for-the-badge&logo=gplv3&labelColor=11140F&color=BBE9AA">
</a>
<a href="https://github.com/Nadim147c/yankd/commits">
<img src="https://img.shields.io/github/last-commit/Nadim147c/yankd?style=for-the-badge&logo=git&labelColor=11140F&color=BBE9AA">
</a>
</h1>

## Features

- [x] Implemets Wayland native protocal.
- [x] Saves raw and parsed image metadata.
- [x] DuckDB full-text search.
- [x] Restore clipboard content with metadata.
- [x] Get clipboard as `json`.
- [x] Delete clipboard item.
- [x] Delete all history.
- [x] Does not depends on wl-clipboard`.
- [ ] Builtin way for interective search.
  - Use quickshell example for search UI.

> [!CAUTION]
> 🚧 **Highly Experimental & Unstable**. This project is in active development
> and may break at any time. Expect bugs, missing features, unexpected behavior,
> and frequent changes.

## Install

#### NixOS

Use flake:

> Quick run:
>
> ```sh
> nix run github:Nadim147c/yankd -- --help
> ```

```nix
yankd = {
    url = "github:Nadim147c/yankd";
    inputs.nixpkgs.follows = "nixpkgs";
};
```

#### Install from source

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

## Usage

Check if yankd is installed or not:

```bash
yankd --help
```

#### Hyprland

Add this to hyprland.conf:

```ini
exec-once = yankd daemon
```

> This will start yankd daemon on startup. You can run
> `hyprctl dispatch exec yankd daemon` to start it manually.

#### Systemd User Service

Copy the following to `~/.config/systemd/user/yankd.service`:

```ini
[Unit]
Description=Yankd clipboard daemon
PartOf=graphical-session.target
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/yankd daemon
Restart=on-failure

[Install]
WantedBy=graphical-session.target
```

Then enable and start the service:

```bash
systemctl --user daemon-reload
systemctl --user enable --now yankd.service
```

## License

This repository is licensed under the [GPL-v3](./LICENSE).
