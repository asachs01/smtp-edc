#!/bin/bash

# Build script for SMTP-EDC MCP Server DXT Binary Extension
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DXT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_DIR="$DXT_DIR/server"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Building SMTP-EDC MCP Server Binary DXT Extension${NC}"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf "$SERVER_DIR"/*
mkdir -p "$SERVER_DIR"

# Get version from git tag or use dev
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags to reduce binary size and embed version info
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

echo "📦 Version: $VERSION"
echo "🔨 Building binaries with optimizations..."

# Change to project root for building
cd "$PROJECT_ROOT"

# Build for macOS (Intel)
echo -e "${YELLOW}Building for macOS (Intel)...${NC}"
GOOS=darwin GOARCH=amd64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-darwin-amd64" \
  ./cmd/mcp-server

# Build for macOS (Apple Silicon)
echo -e "${YELLOW}Building for macOS (Apple Silicon)...${NC}"
GOOS=darwin GOARCH=arm64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-darwin-arm64" \
  ./cmd/mcp-server

# Build for Linux (amd64)
echo -e "${YELLOW}Building for Linux (amd64)...${NC}"
GOOS=linux GOARCH=amd64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-linux-amd64" \
  ./cmd/mcp-server

# Build for Linux (arm64)
echo -e "${YELLOW}Building for Linux (arm64)...${NC}"
GOOS=linux GOARCH=arm64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-linux-arm64" \
  ./cmd/mcp-server

# Build for Windows (amd64)
echo -e "${YELLOW}Building for Windows (amd64)...${NC}"
GOOS=windows GOARCH=amd64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-windows-amd64.exe" \
  ./cmd/mcp-server

# Build for Windows (arm64)
echo -e "${YELLOW}Building for Windows (arm64)...${NC}"
GOOS=windows GOARCH=arm64 go build \
  -ldflags="$LDFLAGS" \
  -o "$SERVER_DIR/smtp-edc-mcp-server-windows-arm64.exe" \
  ./cmd/mcp-server

# Create platform-specific wrapper scripts
echo "📝 Creating platform wrapper scripts..."

# Create the main entry point that will detect platform
cat > "$SERVER_DIR/smtp-edc-mcp-server" << 'EOF'
#!/bin/sh

# Detect platform and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

# Map OS names
case "$OS" in
    Linux*)
        BINARY="./smtp-edc-mcp-server-linux-${ARCH}"
        ;;
    Darwin*)
        BINARY="./smtp-edc-mcp-server-darwin-${ARCH}"
        ;;
    CYGWIN*|MINGW*|MSYS*)
        BINARY="./smtp-edc-mcp-server-windows-${ARCH}.exe"
        ;;
    *)
        echo "Unsupported operating system: $OS" >&2
        exit 1
        ;;
esac

# Get the directory where this script is located
DIR="$(cd "$(dirname "$0")" && pwd)"

# Check if binary exists
if [ ! -f "$DIR/$BINARY" ]; then
    echo "Binary not found: $DIR/$BINARY" >&2
    echo "Platform: $OS $ARCH" >&2
    exit 1
fi

# Execute the appropriate binary with all arguments
exec "$DIR/$BINARY" "$@"
EOF

# Create Windows batch wrapper
cat > "$SERVER_DIR/smtp-edc-mcp-server.bat" << 'EOF'
@echo off
setlocal enabledelayedexpansion

rem Detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set ARCH=arm64
) else (
    echo Unsupported architecture: %PROCESSOR_ARCHITECTURE% >&2
    exit /b 1
)

set BINARY=smtp-edc-mcp-server-windows-%ARCH%.exe

rem Check if binary exists
if not exist "%~dp0%BINARY%" (
    echo Binary not found: %~dp0%BINARY% >&2
    exit /b 1
)

rem Execute the binary with all arguments
"%~dp0%BINARY%" %*
EOF

# Make scripts executable
chmod +x "$SERVER_DIR/smtp-edc-mcp-server"
chmod +x "$SERVER_DIR"/smtp-edc-mcp-server-*

# Display binary sizes
echo -e "${GREEN}📊 Binary sizes:${NC}"
ls -lh "$SERVER_DIR"/smtp-edc-mcp-server-* | awk '{print $9 ": " $5}'

# Copy icon if it exists
if [ -f "$PROJECT_ROOT/dxt-extension/icon.png" ]; then
    echo "🎨 Copying icon..."
    cp "$PROJECT_ROOT/dxt-extension/icon.png" "$DXT_DIR/icon.png"
elif [ -f "$PROJECT_ROOT/assets/icon.png" ]; then
    echo "🎨 Copying icon from assets..."
    cp "$PROJECT_ROOT/assets/icon.png" "$DXT_DIR/icon.png"
else
    echo -e "${YELLOW}⚠️  No icon found, creating placeholder...${NC}"
    # Create a simple placeholder icon using ImageMagick if available
    if command -v convert &> /dev/null; then
        convert -size 128x128 xc:steelblue \
                -gravity center -pointsize 48 -fill white \
                -annotate +0+0 "MCP" \
                "$DXT_DIR/icon.png"
    fi
fi

echo -e "${GREEN}✅ Build complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Test the binaries: cd $DXT_DIR && ./server/smtp-edc-mcp-server -help"
echo "2. Package as DXT: cd $DXT_DIR && dxt pack"
echo "3. Install in Claude Desktop: dxt install smtp-edc-mcp.dxt"