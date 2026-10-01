#!/usr/bin/env bash
#
# drop-box-cli installer: clones the repo, builds the binary locally and
# puts it on your PATH. Building locally matters on macOS — a binary you
# compile yourself carries no quarantine attribute, so Gatekeeper never
# complains about an unsigned download.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/vaxann/drop-box-cli/main/install.sh | bash
#
# Options (environment variables):
#   INSTALL_DIR  target directory (default: /usr/local/bin if writable,
#                otherwise ~/.local/bin)
#   REPO_URL     git repository to build from (default: upstream)
#   WITH_RAYCAST set to 1 to also install the Raycast script command into
#                RAYCAST_SCRIPTS_DIR and, if npm is available, build the
#                Raycast extension in RAYCAST_DIR (macOS)
#   RAYCAST_SCRIPTS_DIR  Raycast script commands folder
#                (default: ~/raycast-scripts)
#   RAYCAST_DIR  where the extension lives
#                (default: $RAYCAST_SCRIPTS_DIR/drop-box-cli)

set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/vaxann/drop-box-cli.git}"
BIN_NAME="drop-box-cli"

info() { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

command -v git >/dev/null 2>&1 || fail "git is required (macOS: xcode-select --install)"
command -v go >/dev/null 2>&1 || fail "Go is required (macOS: brew install go; see https://go.dev/dl/)"

# Pick the install directory.
if [ -n "${INSTALL_DIR:-}" ]; then
  mkdir -p "$INSTALL_DIR"
elif [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
  INSTALL_DIR=/usr/local/bin
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

workdir="$(mktemp -d)"
trap '[ -n "${dev_pid:-}" ] && kill "$dev_pid" 2>/dev/null; rm -rf "$workdir"' EXIT

info "Cloning $REPO_URL"
git clone --quiet --depth 1 "$REPO_URL" "$workdir/src"
cd "$workdir/src"
version="$(git rev-parse --short HEAD)"

info "Building $BIN_NAME ($version) with $(go version | awk '{print $3}')"
CGO_ENABLED=0 go build -trimpath \
  -ldflags "-s -w -X main.version=$version" \
  -o "$workdir/$BIN_NAME" .

info "Installing to $INSTALL_DIR/$BIN_NAME"
install -m 0755 "$workdir/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"

if [ "${WITH_RAYCAST:-0}" = "1" ]; then
  RAYCAST_SCRIPTS_DIR="${RAYCAST_SCRIPTS_DIR:-$HOME/raycast-scripts}"
  RAYCAST_DIR="${RAYCAST_DIR:-$RAYCAST_SCRIPTS_DIR/drop-box-cli}"

  info "Installing the Raycast script command into $RAYCAST_SCRIPTS_DIR"
  mkdir -p "$RAYCAST_SCRIPTS_DIR"
  install -m 0755 "$workdir/src/raycast/script-commands/drop-box-cli-terminal.sh" "$RAYCAST_SCRIPTS_DIR/"

  if command -v npm >/dev/null 2>&1; then
    info "Building the Raycast extension in $RAYCAST_DIR"
    mkdir -p "$RAYCAST_DIR"
    # Replace the sources but keep node_modules to make updates fast. The
    # script command already went to the scripts folder, so leave it out.
    find "$RAYCAST_DIR" -mindepth 1 -maxdepth 1 ! -name node_modules -exec rm -rf {} +
    (cd "$workdir/src/raycast" && tar -cf - --exclude node_modules --exclude script-commands .) |
      (cd "$RAYCAST_DIR" && tar -xf -)
    (cd "$RAYCAST_DIR" && npm ci --no-audit --no-fund --loglevel=error)

    # `ray build` alone leaves Raycast with "Missing executable": only
    # `ray develop` (what `npm run dev` runs) registers the extension with
    # the running Raycast app. Run it until the first build is done, then
    # stop it — the extension stays installed.
    info "Registering the extension with Raycast"
    log="$workdir/ray-develop.log"
    (cd "$RAYCAST_DIR" && exec ./node_modules/.bin/ray develop --non-interactive --exit-on-error) >"$log" 2>&1 &
    dev_pid=$!
    for _ in $(seq 1 180); do
      grep -q "built extension successfully" "$log" && break
      kill -0 "$dev_pid" 2>/dev/null || break
      sleep 1
    done
    built=0
    grep -q "built extension successfully" "$log" && built=1
    sleep 3 # let Raycast pick the build up
    kill "$dev_pid" 2>/dev/null || true
    wait "$dev_pid" 2>/dev/null || true

    if [ "$built" != "1" ]; then
      cat "$log" >&2
      fail "building the Raycast extension failed (log above)"
    elif grep -q "Raycast is not running" "$log"; then
      info "Raycast is not running — start it and re-run the installer to register the extension"
    else
      info "Raycast extension installed. Assign a hotkey: Raycast Settings → Extensions → Drop Box CLI."
    fi
  else
    info "npm not found — skipping the Raycast extension (brew install node, then re-run)"
  fi
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    info "Done. Run: $BIN_NAME"
    ;;
  *)
    info "Done, but $INSTALL_DIR is not on your PATH."
    printf '    Add it, e.g.:\n'
    printf '      echo '\''export PATH="%s:$PATH"'\'' >> ~/.zshrc && exec zsh\n' "$INSTALL_DIR"
    ;;
esac
