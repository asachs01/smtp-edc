#!/bin/bash

# Fix go.sum entries for MCP SDK
echo "Updating go.mod dependencies..."
go mod download github.com/modelcontextprotocol/go-sdk

echo "Running go mod tidy to ensure all dependencies are correct..."
go mod tidy

echo "Building the CLI with MCP support..."
go build -tags cli -o smtp-edc cmd/smtp-edc/main.go

echo "Done! You can now run: ./smtp-edc mcp-server"