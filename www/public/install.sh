#!/bin/sh
# Install the latest Airvoice CLI release.
#   curl -fsSL https://airvoice.yuler.cc/install.sh | sh
set -eu

repo="yuler/airvoice"

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

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}' | tr '[:upper:]' '[:lower:]'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}' | tr '[:upper:]' '[:lower:]'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$1" | awk '{print $NF}' | tr '[:upper:]' '[:lower:]'
  else
    echo "Airvoice install: sha256sum, shasum, or openssl is required to verify download integrity" >&2
    exit 1
  fi
}

if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1 && ! command -v openssl >/dev/null 2>&1; then
  echo "Airvoice install: sha256sum, shasum, or openssl is required to verify download integrity" >&2
  exit 1
fi

latest_url=$(curl -fsSLI --proto '=https' --proto-redir '=https' --tlsv1.2 -o /dev/null -w '%{url_effective}' "https://github.com/${repo}/releases/latest")
case "$latest_url" in
  *"/releases/tag/"*)
    tag=${latest_url##*/releases/tag/}
    tag=${tag%%[/?#]*}
    ;;
  *)
    echo "Airvoice install: failed to resolve latest release tag" >&2
    exit 1
    ;;
esac

if ! printf '%s\n' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "Airvoice install: invalid release tag ${tag}" >&2
  exit 1
fi

base="https://github.com/${repo}/releases/download/${tag}"

if [ "$(id -u)" -eq 0 ]; then
  dest="/usr/local/bin"
else
  dest="${HOME}/.local/bin"
fi
mkdir -p "$dest"

tmp_bin=$(mktemp "${dest}/.airvoice-install.XXXXXX")
tmp_checksums=$(mktemp "${dest}/.airvoice-checksums.XXXXXX")
trap 'rm -f "$tmp_bin" "$tmp_checksums"' EXIT

echo "Downloading ${asset} (${tag})..."
curl -fsSL --proto '=https' --proto-redir '=https' --tlsv1.2 "${base}/checksums-cli.txt" -o "$tmp_checksums"
curl -fsSL --proto '=https' --proto-redir '=https' --tlsv1.2 "${base}/${asset}" -o "$tmp_bin"

expected_sha=$(awk -v asset="$asset" '
  NF >= 2 {
    name = $2
    sub(/^\*/, "", name)
    if (name == asset) {
      print $1
      exit
    }
  }
' "$tmp_checksums" | tr '[:upper:]' '[:lower:]')

case "$expected_sha" in
  *[!0-9a-f]*|"")
    echo "Airvoice install: missing or invalid SHA-256 checksum for ${asset} in checksums-cli.txt" >&2
    exit 1
    ;;
esac
if [ "${#expected_sha}" -ne 64 ]; then
  echo "Airvoice install: invalid SHA-256 checksum length for ${asset}" >&2
  exit 1
fi

actual_sha=$(sha256_file "$tmp_bin")
if [ "$actual_sha" != "$expected_sha" ]; then
  echo "Airvoice install: SHA-256 checksum mismatch for ${asset}" >&2
  echo "  expected: ${expected_sha}" >&2
  echo "  actual:   ${actual_sha}" >&2
  exit 1
fi

magic=$(od -An -t x1 -N 4 "$tmp_bin" | tr -d ' \n')
case "$magic" in
  7f454c46|cffaedfe|feedfacf|cefaedfe|4d5a*) ;;
  *)
    echo "Airvoice install: download is not a binary. The release asset may be missing." >&2
    exit 1
    ;;
esac

chmod 755 "$tmp_bin"
mv -f "$tmp_bin" "${dest}/${bin}"
rm -f "$tmp_checksums"
trap - EXIT

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
