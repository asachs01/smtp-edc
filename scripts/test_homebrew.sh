#!/bin/bash

# Script to test Homebrew installation locally
# This script helps verify the formula works before releasing

set -e

echo "🧪 Testing Homebrew formula for smtp-edc..."

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew is not installed. Please install Homebrew first."
    exit 1
fi

# Remove any existing tap and installation
echo "🧹 Cleaning up any existing installation..."
brew uninstall smtp-edc 2>/dev/null || true
brew untap asachs01/smtp-edc 2>/dev/null || true

# Add the tap from current directory
echo "🔗 Adding local tap..."
brew tap asachs01/smtp-edc $(pwd)

# Install smtp-edc
echo "📦 Installing smtp-edc..."
brew install smtp-edc

# Test the installation
echo "🚀 Testing smtp-edc installation..."
if smtp-edc --version; then
    echo "✅ smtp-edc installed and working correctly!"
else
    echo "❌ smtp-edc installation failed or not working properly."
    exit 1
fi

echo "🎉 Homebrew formula test completed successfully!"
