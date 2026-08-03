#!/usr/bin/env sh

set -eu

REPO="${MONOLOGUE_REPO:-EveryInc/monologue-toolkit}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REQUESTED_VERSION="${MONOLOGUE_VERSION:-latest}"
BINARY_NAME="monologue"

usage() {
  cat <<'EOF'
Install the Monologue CLI from GitHub Releases.

Usage:
  install.sh
  install.sh --version v0.1.0
  install.sh --install-dir /custom/bin

Environment:
  MONOLOGUE_REPO      Override the GitHub repo. Default: EveryInc/monologue-toolkit
  MONOLOGUE_VERSION   Release tag to install. Default: latest
  INSTALL_DIR         Destination directory. Default: ~/.local/bin
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      REQUESTED_VERSION="$2"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd tar
need_cmd mktemp

uname_s="$(uname -s)"
uname_m="$(uname -m)"

case "$uname_s" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *)
    echo "Unsupported operating system: $uname_s" >&2
    exit 1
    ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $uname_m" >&2
    exit 1
    ;;
esac

archive_name="${BINARY_NAME}_${os}_${arch}.tar.gz"
checksums_name="checksums.txt"

if [ "$REQUESTED_VERSION" = "latest" ]; then
  release_path="latest/download"
else
  release_path="download/$REQUESTED_VERSION"
fi

archive_url="https://github.com/$REPO/releases/$release_path/$archive_name"
checksums_url="https://github.com/$REPO/releases/$release_path/$checksums_name"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

echo "Downloading $archive_url"
curl -fsSL "$archive_url" -o "$tmp_dir/$archive_name"

if curl -fsSL "$checksums_url" -o "$tmp_dir/$checksums_name"; then
  if command -v shasum >/dev/null 2>&1; then
    (
      cd "$tmp_dir"
      expected="$(grep "  $archive_name\$" "$checksums_name" | awk '{print $1}')"
      if [ -n "$expected" ]; then
        actual="$(shasum -a 256 "$archive_name" | awk '{print $1}')"
        if [ "$actual" != "$expected" ]; then
          echo "Checksum verification failed for $archive_name" >&2
          exit 1
        fi
      fi
    )
  elif command -v sha256sum >/dev/null 2>&1; then
    (
      cd "$tmp_dir"
      sha256sum -c "$checksums_name" --ignore-missing >/dev/null
    )
  fi
fi

mkdir -p "$INSTALL_DIR"
tar -xzf "$tmp_dir/$archive_name" -C "$tmp_dir"
install -m 0755 "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo
echo "Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"
echo
echo "Next steps:"
echo "  1. Make sure $INSTALL_DIR is on your PATH"
echo "  2. Run: $BINARY_NAME onboarding"
echo
echo "To update later, run: $BINARY_NAME update"
