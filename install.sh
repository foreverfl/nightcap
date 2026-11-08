#!/bin/sh
# Installs the nightcap binary for this OS into ~/.local/bin; NIGHTCAP_VERSION=v0.1.0 picks a release, otherwise the latest.
set -eu

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "nightcap: unsupported OS $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) arch=arm64 ;;
  x86_64 | amd64) arch=amd64 ;;
  *) echo "nightcap: unsupported CPU $(uname -m)" >&2; exit 1 ;;
esac

if [ -n "${NIGHTCAP_VERSION:-}" ]; then
  url="https://github.com/foreverfl/nightcap/releases/download/$NIGHTCAP_VERSION/nightcap_${os}_${arch}"
else
  url="https://github.com/foreverfl/nightcap/releases/latest/download/nightcap_${os}_${arch}"
fi

bin_dir="$HOME/.local/bin"
mkdir -p "$bin_dir"
curl -fsSL "$url" -o "$bin_dir/nightcap"
chmod +x "$bin_dir/nightcap"
echo "installed $("$bin_dir/nightcap" version) to ~/.local/bin"

# Add ~/.local/bin to PATH in the shell rc, unless PATH already has it.
case ":$PATH:" in
  *":$bin_dir:"*) exit 0 ;;
esac
case "${SHELL:-}" in
  */zsh) rc="$HOME/.zshrc" ;;
  *) rc="$HOME/.bashrc" ;;
esac
line='export PATH="$HOME/.local/bin:$PATH"'
if ! grep -qsF "$line" "$rc"; then
  echo "$line" >> "$rc"
fi
echo "added ~/.local/bin to ~/${rc#"$HOME"/}, open a new shell"
