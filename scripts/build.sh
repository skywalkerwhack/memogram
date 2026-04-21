#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
OUTPUT_DIR="${OUTPUT_DIR:-$ROOT_DIR/build}"
TARGETS="${TARGETS:-linux/amd64  freebsd/amd64}"

mkdir -p "$OUTPUT_DIR"
cd "$ROOT_DIR"

for target in $TARGETS; do
	GOOS=${target%/*}
	GOARCH=${target#*/}
	OUTPUT_BIN="$OUTPUT_DIR/memogram-$GOOS-$GOARCH"

	if [ "$GOOS" = "windows" ]; then
		OUTPUT_BIN="$OUTPUT_BIN.exe"
	fi

	echo "Building memogram for $GOOS/$GOARCH..."
	CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
		go build -trimpath -ldflags="-s -w" -o "$OUTPUT_BIN" ./cmd/memogram

	echo "Build successful: $OUTPUT_BIN"
done
