#!/bin/sh
set -eu

REPO="rishabyd/codeberg-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="cb"

detect_os() {
  case "$(uname -s)" in
    Linux*) echo "linux" ;;
    *) echo "Unsupported operating system" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "Unsupported architecture" >&2; exit 1 ;;
  esac
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 1; }
}

fetch_latest_tag() {
  url="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO- "$url" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
    return
  fi
  echo "Need curl or wget" >&2
  exit 1
}

main() {
  require_cmd tar
  os=$(detect_os)
  arch=$(detect_arch)
  tag=$(fetch_latest_tag)
  [ -n "$tag" ] || { echo "Could not resolve latest release" >&2; exit 1; }

  version=$(printf "%s" "$tag" | sed 's/^v//')
  asset="${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"

  tmp_dir=$(mktemp -d)
  trap 'rm -rf "$tmp_dir"' EXIT
  archive_path="${tmp_dir}/${asset}"

  echo "Downloading ${asset}..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$archive_path"
  else
    wget -qO "$archive_path" "$url"
  fi

  tar -xzf "$archive_path" -C "$tmp_dir"
  chmod +x "${tmp_dir}/${BINARY_NAME}"

  target="${INSTALL_DIR}/${BINARY_NAME}"
  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmp_dir}/${BINARY_NAME}" "$target"
  else
    echo "Installing to ${target} (sudo may prompt)..."
    sudo mv "${tmp_dir}/${BINARY_NAME}" "$target"
  fi

  echo "Installed ${BINARY_NAME} ${tag} to ${target}"

  echo "Run: cb auth login"
}

main
