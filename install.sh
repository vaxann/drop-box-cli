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
trap 'rm -rf "$workdir"' EXIT

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
