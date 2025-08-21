#!/bin/bash

# Cleanup script for smtp-edc project
echo "Cleaning up smtp-edc project..."

# Remove backup files
echo "Removing backup files..."
find . -name "*.bak" -type f -delete 2>/dev/null

# Remove empty directories (but keep .gitkeep files)
echo "Removing empty directories..."
find ./pkg -type d -empty -delete 2>/dev/null
find ./prompt_logs -type d -empty -delete 2>/dev/null

# Remove compiled binaries that shouldn't be in git
echo "Removing compiled binaries..."
rm -f smtp-edc smtp-edc-ui

# Clean build artifacts
echo "Cleaning build artifacts..."
rm -rf build/bin/*
rm -rf bin/smtp-edc-cli bin/smtp-edc-mcp

# Clean Go build cache
echo "Cleaning Go build cache..."
go clean -cache 2>/dev/null || true

# Clean frontend build artifacts
echo "Cleaning frontend build artifacts..."
rm -rf frontend/dist
rm -rf frontend/node_modules/.cache
rm -rf frontend/node_modules/.vite-temp

echo "Cleanup complete!"