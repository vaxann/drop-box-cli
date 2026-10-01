#!/bin/bash

# Raycast Script Command: open drop-box-cli in Terminal for the files
# selected in Finder. A zero-setup alternative to the Raycast extension.

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Upload Finder Selection in Terminal
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 📤
# @raycast.packageName drop-box-cli
# @raycast.description Send the files selected in Finder to an SSH server via drop-box-cli in Terminal

set -euo pipefail

# Raycast runs scripts with a minimal PATH, so look in the usual install
# locations explicitly.
bin=""
for c in "$(command -v drop-box-cli || true)" /usr/local/bin/drop-box-cli \
  "$HOME/.local/bin/drop-box-cli" "$HOME/go/bin/drop-box-cli" /opt/homebrew/bin/drop-box-cli; do
  if [ -n "$c" ] && [ -x "$c" ]; then bin="$c"; break; fi
done
if [ -z "$bin" ]; then
  echo "drop-box-cli not found — install it first"
  exit 1
fi

files=()
while IFS= read -r line; do
  [ -n "$line" ] && files+=("$line")
done < <(osascript -e '
tell application "Finder" to set sel to selection as alias list
set out to ""
repeat with f in sel
  set out to out & POSIX path of f & linefeed
end repeat
return out')

if [ ${#files[@]} -eq 0 ]; then
  echo "Select files in Finder first"
  exit 1
fi

cmd="$(printf '%q ' "$bin" "${files[@]}")&& exit"
osascript - "$cmd" <<'APPLESCRIPT' >/dev/null
on run argv
  tell application "Terminal"
    activate
    do script (item 1 of argv)
  end tell
end run
APPLESCRIPT
