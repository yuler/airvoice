#!/bin/sh
# Install the latest Airvoice CLI release.
#   curl -fsSL https://airvoice.yuler.cc/install.sh | sh
set -eu

repo="yuler/airvoice"
base="https://github.com/${repo}/releases/latest/download"

os=$(uname -s)
arch=$(uname -m)

case "$os" in
  Linux) target="linux" ;;
  Darwin) target="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) target="windows" ;;
  *)
    echo "Airvoice install: unsupported OS ${os}" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) cpu="amd64" ;;
  arm64|aarch64) cpu="arm64" ;;
  *)
    echo "Airvoice install: unsupported architecture ${arch}" >&2
    exit 1
    ;;
esac

if [ "$target" = "windows" ]; then
  if [ "$cpu" != "amd64" ]; then
    echo "Airvoice install: Windows builds are amd64 only" >&2
    exit 1
  fi
  asset="airvoice-cli-windows-amd64.exe"
  bin="airvoice.exe"
else
  asset="airvoice-cli-${target}-${cpu}"
  bin="airvoice"
fi

if [ "$(id -u)" -eq 0 ]; then
  dest="/usr/local/bin"
else
  dest="${HOME}/.local/bin"
fi
mkdir -p "$dest"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

echo "Downloading ${asset}..."
curl -fsSL "${base}/${asset}" -o "$tmp"

magic=$(od -An -t x1 -N 4 "$tmp" | tr -d ' \n')
case "$magic" in
  7f454c46|cffaedfe|feedfacf|cefaedfe|4d5a*) ;;
  *)
    echo "Airvoice install: download is not a binary. The release asset may be missing." >&2
    exit 1
    ;;
esac

mv "$tmp" "${dest}/${bin}"
trap - EXIT
chmod 755 "${dest}/${bin}"

echo "Installed ${dest}/${bin}"
if [ "$target" != "windows" ]; then
  "${dest}/${bin}" version
fi

case ":${PATH}:" in
  *":${dest}:"*) ;;
  *)
    echo
    echo "Add this to your shell profile:"
    echo "  export PATH=\"${dest}:\$PATH\""
    ;;
esac

echo
echo "Next: run airvoice, then scan the QR code with the phone app."
