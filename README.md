# drop-box-cli

[![build](https://github.com/vaxann/drop-box-cli/actions/workflows/build.yml/badge.svg)](https://github.com/vaxann/drop-box-cli/actions/workflows/build.yml)

Drag-and-drop a file into a terminal window and beam it to an SSH server —
get back the file's **absolute path on the server**, already copied to your
clipboard. Handy when you need to hand a screenshot or any local file to an
LLM/agent that lives on a remote machine.

```
drop-box-cli  — drag a file into this window
→ /home/me/Pictures/screenshot.png⏎

Send "screenshot.png" to which server?
▸ 1. gpu-box  me@gpu.example.com
  2. staging  deploy@staging.example.com

Target directory on gpu-box:
▸ 1. ~/inbox
  2. /srv/uploads
  ✎  enter a new path…

✓ gpu-box:/home/me/inbox/screenshot.png
  copied to clipboard via wl-copy
```

## How it works

1. Run `drop-box-cli` in a dedicated terminal window. It waits for input.
2. Drag a file (or several) onto the window — the terminal inserts the
   path(s); press **Enter**. Quoted, backslash-escaped and `file://` paths
   from any common terminal are understood.
3. Pick a server from your `~/.ssh/config` — arrow keys, digits `1-9`, or
   just type to filter. Your most frequently used servers float to the top,
   so the usual destination is one keypress away. `Include` directives are
   followed; wildcard entries (`Host *`) and code-hosting services
   (github.com, gitlab.com, bitbucket.org, …) are hidden.
4. Pick a target directory: preconfigured directories for that server plus
   every directory you have used before, most frequently used first. Or
   enter a new one — it is remembered.
5. The directory is created on the server if needed (`~` is expanded
   remotely), the file is copied with `scp`, and the remote absolute path is
   printed and copied to the clipboard.

Transfers use your system `ssh`/`scp`, so keys, `ssh-agent`, `ProxyJump`,
`ControlMaster` and everything else in your SSH config just work.

## Install

### One-liner (recommended, especially on macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/vaxann/drop-box-cli/main/install.sh | bash
```

The script clones the repo, builds the binary **locally** and installs it to
`/usr/local/bin` (or `~/.local/bin` if that is not writable; override with
`INSTALL_DIR=…`). A locally built binary carries no quarantine attribute, so
macOS Gatekeeper does not demand any approvals — unlike a downloaded
unsigned binary. Requires `git` and Go (`brew install go`).

Add `WITH_RAYCAST=1` before `bash` to also install the
[Raycast commands](raycast/README.md) into `~/raycast-scripts`.

### With Go

```sh
go install github.com/vaxann/drop-box-cli@latest
```

### Prebuilt binaries

Every push to `main` builds macOS (arm64/amd64) and Linux binaries — grab
them from the latest [Actions run](https://github.com/vaxann/drop-box-cli/actions).
On macOS you will have to de-quarantine a downloaded binary yourself:
`xattr -d com.apple.quarantine drop-box-cli` — the install script above
avoids this entirely.

### Runtime requirements

OpenSSH (`ssh`/`scp`). For the clipboard, one of `wl-copy`, `xclip`, `xsel`
or `pbcopy` (macOS) is used if present; without one the path is still
printed.

## Usage

```sh
drop-box-cli              # loop mode: waits for drag-and-dropped files
drop-box-cli file.png     # one-shot: send the given file(s) and exit
drop-box-cli --version
```

Keys: type to filter · `↑`/`↓` move · `1-9` quick-select · `Enter` confirm ·
`Esc` back (quit from the first screen) · `Ctrl+C` quit anywhere.

### Non-interactive mode

For launchers and scripts (this is what the Raycast extension uses). These
subcommands never prompt: ssh runs with `BatchMode=yes`, so the server must
be reachable with a key from `ssh-agent`/Keychain. Add `--json` for
machine-readable output; errors go to stderr with a non-zero exit code.

```sh
drop-box-cli hosts [--json]                     # servers, most used first
drop-box-cli dirs --host gpu-box [--json]       # target dirs, most used first
drop-box-cli send --host gpu-box --dir ~/inbox [--json] [--no-clipboard] file ...
drop-box-cli last [--json]                      # most recent server and dir
```

`send` records the upload in the history just like the interactive mode.
To send a file literally named like a subcommand, use `drop-box-cli -- hosts`.

## Raycast

Select files in Finder, press a hotkey, pick a server and a directory —
done, the remote paths are on your clipboard. Install with
`WITH_RAYCAST=1` and see [raycast/README.md](raycast/README.md).

## Configuration

Optional. Preconfigure target directories per server in
`~/.config/drop-box-cli/config.yaml` (macOS:
`~/Library/Application Support/drop-box-cli/config.yaml`):

```yaml
servers:
  gpu-box:            # alias from ~/.ssh/config
    dirs:
      - ~/inbox
      - /srv/uploads
  staging:
    dirs:
      - /var/www/uploads
```

Directory usage history is kept next to it in `history.json` (most recently
used first, 50 entries per server).

## License

MIT
