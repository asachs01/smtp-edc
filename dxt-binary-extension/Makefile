.PHONY: all build clean pack install test help

# Variables
DXT_NAME = smtp-edc-mcp
DXT_FILE = $(DXT_NAME).dxt
BUILD_SCRIPT = scripts/build.sh
SERVER_DIR = server

# Default target
all: clean build pack

# Help target
help:
	@echo "SMTP-EDC MCP Server Binary DXT Extension Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  make build    - Cross-compile binaries for all platforms"
	@echo "  make pack     - Create .dxt package (requires dxt CLI)"
	@echo "  make install  - Install the extension locally"
	@echo "  make clean    - Remove built files"
	@echo "  make test     - Test the binary locally"
	@echo "  make all      - Clean, build, and pack"
	@echo "  make help     - Show this help message"

# Build binaries
build:
	@echo "🔨 Building binaries..."
	@bash $(BUILD_SCRIPT)

# Pack into DXT
pack: build
	@echo "📦 Creating DXT package..."
	@if ! command -v dxt &> /dev/null; then \
		echo "❌ Error: dxt CLI not found. Install with: npm install -g @anthropic-ai/dxt"; \
		exit 1; \
	fi
	@dxt pack
	@if [ -f "dxt-binary-extension.dxt" ]; then \
		mv dxt-binary-extension.dxt $(DXT_FILE); \
	fi
	@echo "✅ Created $(DXT_FILE)"
	@ls -lh $(DXT_FILE) 2>/dev/null || true

# Install locally
install: pack
	@echo "📥 Installing extension..."
	@if ! command -v dxt &> /dev/null; then \
		echo "❌ Error: dxt CLI not found. Install with: npm install -g @anthropic-ai/dxt"; \
		exit 1; \
	fi
	@dxt install $(DXT_FILE)

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(SERVER_DIR)/*
	@rm -f $(DXT_FILE)
	@rm -f *.dxt
	@echo "✅ Clean complete"

# Test the binary
test:
	@echo "🧪 Testing binary..."
	@if [ -f "$(SERVER_DIR)/smtp-edc-mcp-server" ]; then \
		echo "Testing wrapper script:"; \
		./$(SERVER_DIR)/smtp-edc-mcp-server -help; \
	else \
		echo "❌ Binary not found. Run 'make build' first."; \
		exit 1; \
	fi

# Validate manifest
validate:
	@echo "🔍 Validating manifest.json..."
	@if command -v dxt &> /dev/null; then \
		dxt validate; \
	else \
		echo "⚠️  dxt CLI not found, skipping validation"; \
	fi

# Development build (only current platform)
dev:
	@echo "🔨 Building for current platform only..."
	@mkdir -p $(SERVER_DIR)
	@go build -o $(SERVER_DIR)/smtp-edc-mcp-server ../cmd/mcp-server
	@echo "✅ Built development binary"
	@./$(SERVER_DIR)/smtp-edc-mcp-server -help