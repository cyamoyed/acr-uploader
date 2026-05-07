#!/bin/bash

set -e

OUTPUT_DIR="${1:-bin}"

get_goos() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*)    echo "windows" ;;
        MINGW*)     echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

get_goarch() {
    case "$(uname -m)" in
        x86_64*)    echo "amd64" ;;
        aarch64*)   echo "arm64" ;;
        arm64*)     echo "arm64" ;;
        i386*)      echo "386" ;;
        *)          echo "amd64" ;;
    esac
}

GOOS=$(get_goos)
GOARCH=$(get_goarch)

echo "Detected platform: $GOOS/$GOARCH"

EXT=""
if [ "$GOOS" = "windows" ]; then
    EXT=".exe"
fi

OUTPUT_NAME="acr-uploader-${GOOS}-${GOARCH}${EXT}"
OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"

mkdir -p "$OUTPUT_DIR"

export GOOS="$GOOS"
export GOARCH="$GOARCH"

echo "Building $OUTPUT_PATH..."
go build -ldflags="-s -w" -gcflags="all=-l" -o "$OUTPUT_PATH" main.go

echo "Build successful! Output: $OUTPUT_PATH"