#!/bin/bash

# Simple verification script for the release process
set -e

echo "=== SMTP-EDC Release Verification ==="
echo "Verifying release components..."

# Check if required directories exist
echo "1. Checking directory structure..."
if [ ! -d "frontend" ]; then
    echo "ERROR: Frontend directory not found"
    exit 1
fi

if [ ! -d "cmd/smtp-edc" ]; then
    echo "ERROR: CLI source directory not found"
    exit 1
fi

if [ ! -f "app.go" ]; then
    echo "ERROR: app.go (UI backend) not found"
    exit 1
fi

if [ ! -f "main.go" ]; then
    echo "ERROR: main.go (UI main) not found"
    exit 1
fi

if [ ! -f "wails.json" ]; then
    echo "ERROR: wails.json configuration not found"
    exit 1
fi

if [ ! -f ".goreleaser.yml" ]; then
    echo "ERROR: GoReleaser configuration not found"
    exit 1
fi

echo "✓ Directory structure OK"

# Check if required tools are available (in CI environment)
echo "2. Checking required tools availability..."

if ! command -v go &> /dev/null; then
    echo "ERROR: Go not found"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "ERROR: Node.js not found"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "ERROR: npm not found"
    exit 1
fi

echo "✓ Required tools available"

# Check Go module
echo "3. Checking Go module..."
go mod download
go mod verify
echo "✓ Go module OK"

# Check frontend dependencies
echo "4. Checking frontend dependencies..."
cd frontend
npm ci
echo "✓ Frontend dependencies OK"

# Test frontend build
echo "5. Testing frontend build..."
npm run build
if [ ! -d "dist" ]; then
    echo "ERROR: Frontend dist directory not created"
    exit 1
fi
echo "✓ Frontend build OK"

cd ..

# Test CLI build
echo "6. Testing CLI build..."
go build -tags cli -o bin/smtp-edc-test ./cmd/smtp-edc
if [ ! -f "bin/smtp-edc-test" ]; then
    echo "ERROR: CLI binary not created"
    exit 1
fi

# Test CLI binary
./bin/smtp-edc-test --help > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "ERROR: CLI binary not working"
    exit 1
fi

echo "✓ CLI build OK"

# Cleanup test binary
rm -f bin/smtp-edc-test

echo ""
echo "=== All verification checks passed! ==="
echo "The release process is ready to be executed."
