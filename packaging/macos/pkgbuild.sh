#!/usr/bin/env bash
set -euo pipefail

VERSION=${VERSION:-"0.0.0"}
ARCH=${ARCH:-"amd64"}
ROOT=$(pwd)
STAGE="$ROOT/dist/pkgroot"
PKG="$ROOT/dist/dis-agent-darwin-${ARCH}.pkg"

rm -rf "$STAGE"
mkdir -p "$STAGE/usr/local/bin"
mkdir -p "$STAGE/Library/Application Support/dis-agent"
mkdir -p "$STAGE/Library/LaunchDaemons"

# Map GoReleaser output folders to requested arch
case "$ARCH" in
  amd64) BIN_PATH="$ROOT/dist/agent_darwin_amd64_v1/dis-agent" ;;
  arm64) BIN_PATH="$ROOT/dist/agent_darwin_arm64_v8.0/dis-agent" ;;
  *) echo "Unsupported ARCH: $ARCH" ; exit 1 ;;
esac

if [ ! -f "$BIN_PATH" ]; then
  echo "Binary not found at $BIN_PATH"
  exit 1
fi

cp "$BIN_PATH" "$STAGE/usr/local/bin/dis-agent"
cp "$ROOT/packaging/examples/agent.yaml" "$STAGE/Library/Application Support/dis-agent/agent.yaml"
cp "$ROOT/packaging/examples/bridge_agents.yaml" "$STAGE/Library/Application Support/dis-agent/bridge_agents.yaml"
cp "$ROOT/packaging/launchd/com.dis.agent.plist" "$STAGE/Library/LaunchDaemons/com.dis.agent.plist"

pkgbuild \\
  --root "$STAGE" \\
  --identifier "com.dis.agent" \\
  --version "$VERSION" \\
  --install-location / \\
  "$PKG"

echo "Created $PKG"
