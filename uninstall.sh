#!/bin/sh
set -eu

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="cb"
CONFIG_DIR="$HOME/.config/cb"
GIT_HELPER_KEY="credential.https://codeberg.org.helper"

remove_binary() {
  target="${INSTALL_DIR}/${BINARY_NAME}"
  if [ ! -f "$target" ]; then
    return
  fi

  if [ -w "$INSTALL_DIR" ]; then
    rm -f "$target"
  else
    echo "Removing ${target} (sudo may prompt)..."
    sudo rm -f "$target"
  fi
}

remove_config() {
  if [ -d "$CONFIG_DIR" ]; then
    rm -rf "$CONFIG_DIR"
  fi
}

remove_git_helper() {
  if command -v git >/dev/null 2>&1; then
    git config --global --unset "$GIT_HELPER_KEY" >/dev/null 2>&1 || true
  fi
}

main() {
  printf "Uninstall cb? [y/N] "
  read -r answer < /dev/tty
  case "$answer" in
    [Yy]|[Yy][Ee][Ss]) ;;
    *) echo "Aborted."; exit 0 ;;
  esac

  remove_binary
  remove_git_helper
  remove_config
  echo "Uninstalled cb and removed local config from ${CONFIG_DIR}"
}

main
