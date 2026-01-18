#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
CLI_DIR="$PROJECT_DIR/cli"
OUTPUT_DIR="${OUTPUT_DIR:-$PROJECT_DIR/dist}"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}"
DATE="${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

LDFLAGS="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

echo "🔨 Building cc-sandbox CLI v$VERSION"

mkdir -p "$OUTPUT_DIR"
cd "$CLI_DIR"

go mod download

TARGETS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for target in "${TARGETS[@]}"; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    OUTPUT_NAME="cc-sandbox-${GOOS}-${GOARCH}"
    [ "$GOOS" = "windows" ] && OUTPUT_NAME="${OUTPUT_NAME}.exe"

    echo "📦 Building for $GOOS/$GOARCH..."
    GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$LDFLAGS" -o "$OUTPUT_DIR/$OUTPUT_NAME" .
done

echo "✅ Build complete!"
ls -lh "$OUTPUT_DIR"/cc-sandbox-*

cd "$OUTPUT_DIR"
sha256sum cc-sandbox-* > checksums.txt
