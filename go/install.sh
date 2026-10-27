#!/bin/sh
set -eu

REPO="MateeDevs/sentiary-tools"
BINARY="sentiary-cli"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin|linux) ;;
  *)
    echo "Unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64)
    arch="amd64"
    ;;
  arm64|aarch64)
    arch="arm64"
    ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

asset="${BINARY}-${os}-${arch}"
archive="${asset}.tar.gz"
release_version="${SENTIARY_VERSION:-${1:-}}"
case "$release_version" in
  *[!a-zA-Z0-9._-]*) echo "Invalid version" >&2; exit 1 ;;
esac
base_url="https://github.com/${REPO}/releases/latest/download"
if [ -n "$release_version" ]; then
  base_url="https://github.com/${REPO}/releases/download/${release_version}"
fi
url="${base_url}/${archive}"
install_dir="${INSTALL_DIR:-}"

if [ -z "$install_dir" ]; then
  if [ -w "/usr/local/bin" ]; then
    install_dir="/usr/local/bin"
  else
    install_dir="$HOME/.local/bin"
  fi
fi

mkdir -p "$install_dir"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp_dir/$archive"
  curl -fsSL "$base_url/SHA256SUMS" -o "$tmp_dir/SHA256SUMS"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$tmp_dir/$archive"
  wget -q "$base_url/SHA256SUMS" -O "$tmp_dir/SHA256SUMS"
else
  echo "curl or wget is required." >&2
  exit 1
fi

expected="$(awk -v asset="$archive" '$2 == asset { print $1 }' "$tmp_dir/SHA256SUMS")"
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp_dir/$archive" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  actual="$(shasum -a 256 "$tmp_dir/$archive" | awk '{print $1}')"
else
  echo "sha256sum or shasum is required" >&2; exit 1
fi
if [ -z "$expected" ] || [ "$actual" != "$expected" ]; then
  echo "Archive checksum mismatch" >&2; exit 1
fi
tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"

if [ -f "$tmp_dir/$asset" ]; then
  source_binary="$tmp_dir/$asset"
elif [ -f "$tmp_dir/$BINARY" ]; then
  source_binary="$tmp_dir/$BINARY"
else
  echo "The archive does not contain $asset or $BINARY." >&2
  exit 1
fi

install_path="$install_dir/$BINARY"
cp "$source_binary" "$install_path"
chmod +x "$install_path"

if command -v "$install_path" >/dev/null 2>&1; then
  version="$($install_path --version 2>/dev/null || true)"
else
  version=""
fi

echo "Installed $BINARY to $install_path"
if [ -n "$version" ]; then
  echo "Version: $version"
fi

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to PATH if your shell cannot find $BINARY." ;;
esac
