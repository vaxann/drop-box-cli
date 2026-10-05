# Drop Box CLI for Raycast

Select files in Finder, press a hotkey, pick a server and a directory — the
files are uploaded and their remote paths are copied to the clipboard.

A thin UI over the [`drop-box-cli`](../README.md) binary's non-interactive
subcommands, so the server list, directory history and usage ordering are
shared with the terminal app.

## Commands

- **Upload Finder Selection** — servers (most used first) → directories
  (most used first; type a path to use a new one) → upload.
- **Upload to Last Destination** — no UI: uploads straight to the most
  recently used server and directory.

## Drop box and Finder Quick Action

On macOS the installer also builds **Drop Box CLI.app** in
`~/Applications` (override with `DROPLET_DIR=…`). Drag it to the Dock and
drop files on it — from Finder, a screenshot thumbnail, anywhere — and
Raycast opens **Upload Finder Selection** with exactly those files.

It also installs a Finder Quick Action: right-click files →
**Quick Actions → Upload with drop-box-cli** (or bind a key in System
Settings → Keyboard → Keyboard Shortcuts → Services).

Both open the command through a Raycast deeplink, so the first time
Raycast asks to confirm — tick "Always open" to skip it afterwards.

## Install

1. Install `drop-box-cli` together with the Raycast pieces:

   ```sh
   curl -fsSL https://raw.githubusercontent.com/vaxann/drop-box-cli/main/install.sh | WITH_RAYCAST=1 bash
   ```

   This puts the script command into your Raycast script commands folder
   (`~/raycast-scripts`, override with `RAYCAST_SCRIPTS_DIR=…`) and, if
   Node.js is installed (`brew install node`), builds the extension in
   `~/raycast-scripts/drop-box-cli` (override with `RAYCAST_DIR=…`) and
   registers it with Raycast — keep Raycast running while it installs.
   Re-running the installer updates both.
2. Manual alternative: `cd ~/raycast-scripts/drop-box-cli && npm run dev`,
   wait for "built extension successfully", then `Ctrl+C`. (`ray build`
   alone is not enough — Raycast then reports "Missing executable".)
3. Raycast Settings → Extensions → Drop Box CLI → record a hotkey for
   **Upload Finder Selection** (e.g. `⌥⌘U`) and, optionally, for
   **Upload to Last Destination**.

If the binary is not in `/usr/local/bin`, `~/.local/bin`, `~/go/bin` or
`/opt/homebrew/bin`, set its path in the extension preferences.

## Requirements

- **Finder access**: on first use macOS asks whether Raycast may control
  Finder — allow it (System Settings → Privacy & Security → Automation).
- **Key-based SSH**: uploads run without a terminal (`BatchMode=yes`), so a
  password or key passphrase cannot be typed. Keep the key in the agent and
  Keychain:

  ```
  # ~/.ssh/config
  Host *
    AddKeysToAgent yes
    UseKeychain yes
  ```

  and once: `ssh-add --apple-use-keychain ~/.ssh/id_ed25519`. Also connect
  to each server once from a terminal so its host key is known.
- Keep Raycast open until a large upload finishes: closing the window may
  stop it.

## Without the extension

`script-commands/drop-box-cli-terminal.sh` (**Upload Finder Selection in
Terminal**) is a Raycast Script Command that opens the interactive
`drop-box-cli` in Terminal for the Finder selection. The installer copies
it into `~/raycast-scripts`, so it shows up in Raycast as soon as that
folder is registered under Settings → Extensions → Script Commands.

## Development

```sh
npm install
npm run dev     # live-reloads into Raycast
npm run lint
```
