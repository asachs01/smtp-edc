#!/bin/bash

# Docker-based build script for SMTP-EDC MCP Server DXT Binary Extension
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

echo -e "${GREEN}🚀 Building SMTP-EDC MCP Server Binary DXT Extension (Docker)${NC}"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf "$SERVER_DIR"/*
mkdir -p "$SERVER_DIR"

# Get version from git tag or use dev
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "📦 Version: $VERSION"
echo "🐳 Using Docker for cross-compilation..."

# Create a temporary Dockerfile for building
cat > "$DXT_DIR/Dockerfile.build" << 'EOF'
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT
ARG GOOS
ARG GOARCH
ARG OUTPUT

# Build the binary
RUN GOOS=${GOOS} GOARCH=${GOARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}" \
    -o ${OUTPUT} \
    ./cmd/mcp-server
EOF

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local output=$3
    
    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"
    
    docker build \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg COMMIT="$COMMIT" \
        --build-arg GOOS="$os" \
        --build-arg GOARCH="$arch" \
        --build-arg OUTPUT="/app/output" \
        -f "$DXT_DIR/Dockerfile.build" \
        -t smtp-edc-builder:${os}-${arch} \
        "$PROJECT_ROOT" > /dev/null 2>&1
    
    # Extract the binary
    docker create --name extract-${os}-${arch} smtp-edc-builder:${os}-${arch} > /dev/null 2>&1
    docker cp extract-${os}-${arch}:/app/output "$SERVER_DIR/$output"
    docker rm extract-${os}-${arch} > /dev/null 2>&1
    
    echo -e "${GREEN}✓ Built $output${NC}"
}

# Build for all platforms
build_platform "darwin" "amd64" "smtp-edc-mcp-server-darwin-amd64"
build_platform "darwin" "arm64" "smtp-edc-mcp-server-darwin-arm64"
build_platform "linux" "amd64" "smtp-edc-mcp-server-linux-amd64"
build_platform "linux" "arm64" "smtp-edc-mcp-server-linux-arm64"
build_platform "windows" "amd64" "smtp-edc-mcp-server-windows-amd64.exe"
build_platform "windows" "arm64" "smtp-edc-mcp-server-windows-arm64.exe"

# Clean up Docker images
docker rmi $(docker images -q smtp-edc-builder:*) 2>/dev/null || true
rm -f "$DXT_DIR/Dockerfile.build"

# Create platform-specific wrapper scripts (same as before)
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

echo -e "${GREEN}✅ Build complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Test the binaries: cd $DXT_DIR && ./server/smtp-edc-mcp-server -help"
echo "2. Package as DXT: cd $DXT_DIR && dxt pack"
echo "3. Install in Claude Desktop: dxt install smtp-edc-mcp.dxt"