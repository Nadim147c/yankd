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

> [!CAUTION]
> 🚧 **Highly Experimental & Unstable**. This project is in active development
> and may break at any time. Expect bugs, missing features, unexpected behavior,
> and frequent changes.

## Features

- [x] Implements Wayland native protocol.
- [x] Saves raw and parsed image metadata.
- [x] Advanced search queries.
- [x] Restore clipboard content with metadata.
- [x] Get clipboard as `json`.
- [x] Delete clipboard item.
- [x] Delete all history.
- [x] Does not depend on `wl-clipboard`.
- [ ] Built-in way for interactive search.
  - Use quickshell example for search UI.

## Why?

When you copy an image, most web browsers include `image/html` with metadata
like `url` and `alt`. Yankd allows you to search your clipboard history using
this metadata.

Existing tools like `wl-paste` + `cliphist` lack the ability to pass this
metadata from the daemon to the database, or offer advanced querying to search
tools. Yankd solves this by directly implementing the
`wlr-data-control-unstable-v1` protocol and storing clipboard events in a
relational database (DuckDB), enabling advanced searches (fuzzy, exact, regex,
time, and type filters).

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

> [!IMPORTANT]
> All `yankd` command requires the daemon to be running on the background.

Check if yankd is installed or not:

```bash
yankd --help
```

### Daemon

Start yankd daemon by running `yankd daemon`.

#### Hyprland

Add this to hyprland.conf:

```lua
hl.dispatch(hl.dsp.exec_cmd("yankd daemon"))
```

#### Systemd User Service

Copy the following configuration to `~/.config/systemd/user/yankd.service`:

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

### Basic commands

List last 10 clipboard events[^1].

```bash
yankd list --format=json --limit=10 | jq
```

Get a clipboard event[^1] as JSON by ID[^2].

```bash
yankd get 019f0dea-40d5-773a-ad21-aff1feba4d34
```

Get a content of a primary[^3] clipboard event[^1]

```bash
yankd get --primary 019f0dea-40d5-773a-ad21-aff1feba4d34
```

Get a content of a primary mime-type[^3] clipboard event[^1] but encode to
base64 if the content is binary data.

```bash
yankd get --primary --base64 019f0dea-40d5-773a-ad21-aff1feba4d34
```

Set/Restore a clipboard event[^1] by ID[^2].

```bash
yankd set 019f0dea-40d5-773a-ad21-aff1feba4d34
```

### Search

Since yankd owns the database, it provides powerful controls over database
queries. Here are the query parameters:

- `after:"<time>"`: Filter items after this time. Can be human readable like
  `last friday`
- `before:"<time>"`: Filter items before this time.
- `type:"<mime-type>"`: Filter by mime-type. It's fuzzy match so don't need to
  exact.
- `[<query>`: Prefix match.
- `<query>]`: Suffix match.
- `/<pattern>/`: Regex match.
- `"<pattern>"`: Exact Keyword match. Can add as many as you want

#### Examples

Get events[^1] before first sunday of this month.

```bash
yankd search --format=json 'hello world before:"first sunday"'
```

Get events[^1] that contains the exact match **world**.

```bash
yankd search --format=json 'hello "world"'
```

Get events[^1] images where metadata has exact match "reddit.com" and also match
the regex `r/(golang|hyprland)`

```bash
yankd search --format=json 'type:image "reddit.com" /r\/(golang|hyprland)/'
```

## Hacking

Contributions are welcome. I don't care if your PR is AI generated. Review the
PR with as much time as you expect me to spend reviewing it.

## Hacking++

Yankd does not have a GUI and likely never will. However, you can easily create
a GUI wrapper without rewriting the search logic and Wayland protocol. Yankd
commands depend on the background daemon, sending requests to it and printing
the results. The daemon uses HTTP over Unix sockets. More documentation is
available in [ipc](./internal/ipc/README.md).

## License

This repository is licensed under the [GPL-v3](./LICENSE).

[^1]: A clipboard event refers to content collection of all mime-type each time
    a wayland client makes a offer.

[^2]: Yankd use UUID as clipboard event[^1] ID.

[^3]: Yankd tries to assign a primary mime type to every clipboard event[^1]
    from the mime-type list.
