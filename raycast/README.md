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

## Install

1. Install `drop-box-cli` with the Raycast extension:

   ```sh
   curl -fsSL https://raw.githubusercontent.com/vaxann/drop-box-cli/main/install.sh | WITH_RAYCAST=1 bash
   ```

   Requires Node.js (`brew install node`). The extension sources land in
   `~/.local/share/drop-box-cli/raycast`.
2. First time only: run **Import Extension** in Raycast and pick that
   folder (alternatively run `npm run dev` there once and stop it).
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

`script-commands/drop-box-cli-terminal.sh` is a Raycast Script Command that
opens the interactive `drop-box-cli` in Terminal for the Finder selection.
Add the `script-commands` folder in Raycast Settings → Extensions → Script
Commands.

## Development

```sh
npm install
npm run dev     # live-reloads into Raycast
npm run lint
```
