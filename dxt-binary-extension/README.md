# SMTP-EDC MCP Server Binary Desktop Extension

This is a binary Desktop Extension (DXT) package for the SMTP-EDC MCP Server, designed for use with Claude Desktop and other AI assistants that support the DXT specification.

## Features

- **Cross-platform binaries** - Pre-compiled for macOS (Intel & Apple Silicon), Linux (amd64 & arm64), and Windows (amd64 & arm64)
- **Automatic platform detection** - The extension automatically runs the correct binary for your system
- **Small binary size** - Optimized with `-ldflags="-s -w"` to reduce file size
- **MCP Protocol support** - Full Model Context Protocol implementation for SMTP operations

## Building

### Prerequisites

1. Go 1.21 or later
2. Node.js and npm (for dxt CLI)
3. DXT CLI: `npm install -g @anthropic-ai/dxt`

### Build Commands

```bash
# Build everything (clean, build, pack)
make all

# Individual steps
make build    # Cross-compile binaries for all platforms
make pack     # Create .dxt package
make install  # Install locally for testing

# Development
make dev      # Build only for current platform
make test     # Test the binary
make clean    # Remove build artifacts
```

## Manual Build Process

```bash
# 1. Build the binaries
./scripts/build.sh

# 2. Create the DXT package
dxt pack

# 3. Install (optional, for testing)
dxt install smtp-edc-mcp.dxt
```

## Binary Structure

The extension includes platform-specific binaries:

```
server/
├── smtp-edc-mcp-server           # Shell wrapper (auto-detects platform)
├── smtp-edc-mcp-server.bat       # Windows batch wrapper
├── smtp-edc-mcp-server-darwin-amd64    # macOS Intel
├── smtp-edc-mcp-server-darwin-arm64    # macOS Apple Silicon
├── smtp-edc-mcp-server-linux-amd64     # Linux x64
├── smtp-edc-mcp-server-linux-arm64     # Linux ARM64
├── smtp-edc-mcp-server-windows-amd64.exe  # Windows x64
└── smtp-edc-mcp-server-windows-arm64.exe  # Windows ARM64
```

## Platform Detection

The extension automatically detects your platform and runs the appropriate binary:

- **Unix/Linux/macOS**: Uses the shell script wrapper `smtp-edc-mcp-server`
- **Windows**: Uses the batch file `smtp-edc-mcp-server.bat`

## MCP Tools Available

The binary provides these MCP tools:

- `test_connection` - Test SMTP server connection and capabilities
- `send_email` - Send emails via SMTP
- `validate_addresses` - Validate email addresses with MX record checking
- `load_template` - Load and process email templates

## Configuration

Users can configure the extension through the DXT settings:

- **debug_mode** - Enable debug logging
- **default_smtp_server** - Default SMTP server hostname
- **default_smtp_port** - Default SMTP port (587)
- **default_from_address** - Default sender email
- **timeout** - Connection timeout in seconds (30)
- **rate_limit** - Max emails per minute (10)

## Installation in Claude Desktop

1. Build the extension: `make all`
2. The `smtp-edc-mcp.dxt` file will be created
3. In Claude Desktop:
   - Go to Extensions settings
   - Click "Install from file"
   - Select the `smtp-edc-mcp.dxt` file

## Testing

Test the binary locally before packaging:

```bash
# Test the wrapper script
./server/smtp-edc-mcp-server -help

# Test with stdio transport (default for MCP)
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | ./server/smtp-edc-mcp-server -transport stdio
```

## Troubleshooting

### Binary not found
- Ensure the build script completed successfully
- Check that all platform binaries exist in the `server/` directory

### Permission denied
- Make sure the wrapper scripts are executable: `chmod +x server/smtp-edc-mcp-server*`

### DXT packaging fails
- Verify `dxt` CLI is installed: `npm install -g @anthropic-ai/dxt`
- Check manifest.json is valid: `dxt validate`

## Development

For development, you can build only for your current platform:

```bash
make dev
make test
```

This is faster than cross-compiling for all platforms.

## License

MIT - See the main project LICENSE file