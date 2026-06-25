#!/bin/bash

# Build script for Fingerprint Qubu Service
# Builds Windows executables with an embedded build version.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== Fingerprint Qubu Service Build Script ===${NC}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Working directory: $SCRIPT_DIR"

CURRENT_VERSION=""
if [ -f "bin/version.txt" ]; then
    CURRENT_VERSION=$(grep -E "^version=" "bin/version.txt" | sed 's/^version=//' | tr -d '\r\n')
fi
[[ -z "$CURRENT_VERSION" ]] && CURRENT_VERSION="v0.0"

# Version priority: CLI arg > VERSION env > interactive prompt > current version.txt
if [[ -n "$1" ]]; then
    BUILD_VERSION="$1"
elif [[ -n "$VERSION" ]]; then
    BUILD_VERSION="$VERSION"
elif [[ -t 0 ]]; then
    echo ""
    read -p "Masukkan version hasil build [$CURRENT_VERSION] (kosong = tidak ubah): " BUILD_VERSION
    if [[ -z "$BUILD_VERSION" ]]; then
        BUILD_VERSION="$CURRENT_VERSION"
    fi
else
    BUILD_VERSION="$CURRENT_VERSION"
fi

if [[ "$BUILD_VERSION" != v* ]]; then
    BUILD_VERSION="v${BUILD_VERSION}"
fi

echo -e "${GREEN}Version build: $BUILD_VERSION${NC}"
echo ""

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

echo -e "${GREEN}Go version: $(go version)${NC}"

mkdir -p bin

echo -e "${YELLOW}Running go mod tidy...${NC}"
go mod tidy

LDFLAGS="-s -w -X main.version=${BUILD_VERSION}"
export GOOS=windows

echo -e "${YELLOW}Building fingerprint-service.exe...${NC}"
go build -ldflags "$LDFLAGS" -o bin/fingerprint-service.exe ./cmd/server

echo -e "${YELLOW}Building cli.exe...${NC}"
go build -ldflags "$LDFLAGS" -o bin/cli.exe ./cmd/cli

if [ ! -f "bin/fingerprint-service.exe" ] || [ ! -f "bin/cli.exe" ]; then
    echo -e "${RED}Error: Build failed - executable not found${NC}"
    exit 1
fi

echo "version=$BUILD_VERSION" > bin/version.txt

# Package bin/ contents into a ZIP archive
TEMP_DIR=$(mktemp -d)
echo "Temporary directory: $TEMP_DIR"

echo -e "${YELLOW}Packaging files dari bin/...${NC}"
for item in bin/*; do
    [[ -e "$item" ]] || continue
    base="$(basename "$item")"
    case "$base" in
        logs) continue ;;
    esac
    [[ "$base" == *.zip ]] && continue
    cp -r "$item" "$TEMP_DIR/"
done

ZIP_NAME="installer-fingerprint-qubu-service-${BUILD_VERSION}.zip"
OUTPUT_DIR="${OUTPUT_DIR:-$SCRIPT_DIR/dist}"

if [ ! -d "$OUTPUT_DIR" ]; then
    echo "Creating output directory: $OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
fi

echo -e "${YELLOW}Creating ZIP archive...${NC}"

convert_to_win_path() {
    local unix_path="$1"
    if command -v cygpath &> /dev/null; then
        cygpath -w "$unix_path"
    else
        if [[ "$unix_path" =~ ^/([a-z])/ ]]; then
            echo "$unix_path" | sed "s|^/\\([a-z]\\)/|\\1:/|" | sed 's|/|\\|g'
        elif [[ "$unix_path" =~ ^/tmp/ ]]; then
            local win_temp
            win_temp=$(powershell.exe -Command "[System.IO.Path]::GetTempPath()" | tr -d '\r\n')
            local rel_path
            rel_path=$(echo "$unix_path" | sed 's|^/tmp/||')
            echo "$win_temp$rel_path" | sed 's|/|\\|g'
        else
            echo "$unix_path" | sed 's|/|\\|g'
        fi
    fi
}

TEMP_DIR_WIN=$(convert_to_win_path "$TEMP_DIR")
OUTPUT_DIR_WIN=$(convert_to_win_path "$OUTPUT_DIR")
TARGET_PATH_WIN="$OUTPUT_DIR_WIN\\$ZIP_NAME"

TEMP_DIR_PS=$(echo "$TEMP_DIR_WIN" | sed 's|\\|\\\\|g')
TARGET_PATH_PS=$(echo "$TARGET_PATH_WIN" | sed 's|\\|\\\\|g')
powershell.exe -Command "Compress-Archive -Path '$TEMP_DIR_PS\*' -DestinationPath '$TARGET_PATH_PS' -Force"

TARGET_PATH="$OUTPUT_DIR/$ZIP_NAME"
rm -rf "$TEMP_DIR"
echo "Temporary files cleaned up."

echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
echo "Output:"
echo "  bin/fingerprint-service.exe  (version: $BUILD_VERSION)"
echo "  bin/cli.exe                  (version: $BUILD_VERSION)"
echo "  bin/version.txt"
echo "  $TARGET_PATH"
echo ""
echo "Isi ZIP: executable, version.txt, dan script service (*.bat) dari folder bin."
echo "Jika nssm.exe ada di bin/, file tersebut ikut ter-pack. Jika tidak, jalankan download-nssm.bat setelah extract."
echo ""
echo "Cek version:"
echo "  bin/fingerprint-service.exe -version"
echo "  bin/cli.exe -version"
echo ""
echo "Install sebagai Windows Service (Administrator):"
echo "  cd bin && install.bat"
