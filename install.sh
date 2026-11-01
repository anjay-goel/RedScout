#!/bin/sh
set -eu

repo="anjay-goel/RedScout"
base="${REDSCOUT_DOWNLOAD_URL:-https://github.com/$repo/releases/latest/download}"
dir="${REDSCOUT_INSTALL_DIR:-$HOME/.local/bin}"

fail() {
	echo "redscout: $1" >&2
	exit 1
}

case "$(uname -s)" in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fail "unsupported system $(uname -s)" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fail "unsupported CPU $(uname -m)" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is needed to install redscout"

file="redscout-$os-$arch"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$base/$file" -o "$tmp/$file" || fail "couldn't download $file"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || fail "couldn't download checksums.txt"

cd "$tmp"
grep " $file\$" checksums.txt >expected || fail "no checksum for $file"
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum -c expected >/dev/null || fail "checksum mismatch for $file"
else
	shasum -a 256 -c expected >/dev/null || fail "checksum mismatch for $file"
fi

mkdir -p "$dir"
chmod 755 "$file"
mv -f "$file" "$dir/redscout"

echo "Installed redscout to $dir/redscout"
case ":$PATH:" in
*":$dir:"*) ;;
*) echo "Add $dir to your PATH" ;;
esac
