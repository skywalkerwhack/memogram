#!/bin/sh

# memogram is a Telegram bot for saving messages into a Memos instance.
# Copyright (C) 2026  skywalkerwhack
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.


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
