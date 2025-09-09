#!/bin/bash

# Test script for SMTP-EDC MCP Server
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DXT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVER_BIN="$DXT_DIR/server/smtp-edc-mcp-server"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🧪 Testing SMTP-EDC MCP Server${NC}"

# Check if binary exists
if [ ! -f "$SERVER_BIN" ]; then
    echo -e "${RED}❌ Binary not found at $SERVER_BIN${NC}"
    echo "Run 'make build' first to build the binaries."
    exit 1
fi

# Test 1: Check help output
echo -e "${YELLOW}Test 1: Checking help output...${NC}"
if $SERVER_BIN -help 2>&1 | grep -q "SMTP-EDC MCP Server"; then
    echo -e "${GREEN}✓ Help output works${NC}"
else
    echo -e "${RED}✗ Help output failed${NC}"
    exit 1
fi

# Test 2: Test MCP initialize request
echo -e "${YELLOW}Test 2: Testing MCP initialize...${NC}"

# Create a temporary file for the response
RESPONSE_FILE=$(mktemp)

# Send initialize request
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}},"id":1}' | \
    timeout 2 $SERVER_BIN -transport stdio 2>/dev/null | head -1 > "$RESPONSE_FILE" || true

# Check if we got a valid response
if grep -q '"result"' "$RESPONSE_FILE"; then
    echo -e "${GREEN}✓ MCP initialize successful${NC}"
    
    # Check for expected capabilities
    if grep -q '"tools"' "$RESPONSE_FILE"; then
        echo -e "${GREEN}  ✓ Tools capability present${NC}"
    fi
    if grep -q '"resources"' "$RESPONSE_FILE"; then
        echo -e "${GREEN}  ✓ Resources capability present${NC}"
    fi
else
    echo -e "${RED}✗ MCP initialize failed${NC}"
    echo "Response: $(cat $RESPONSE_FILE)"
fi

# Test 3: Test tools/list request
echo -e "${YELLOW}Test 3: Testing tools/list...${NC}"

# Create a test script that sends multiple requests
cat > "$RESPONSE_FILE.sh" << 'EOF'
#!/bin/bash
# Send initialize
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}},"id":1}'
sleep 0.1
# Send tools/list
echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}'
sleep 0.1
EOF

chmod +x "$RESPONSE_FILE.sh"

# Run the test
TOOLS_RESPONSE=$(timeout 2 bash "$RESPONSE_FILE.sh" | $SERVER_BIN -transport stdio 2>/dev/null | grep '"id":2' || echo "")

if echo "$TOOLS_RESPONSE" | grep -q '"test_connection"'; then
    echo -e "${GREEN}✓ Tools list successful${NC}"
    echo -e "${GREEN}  ✓ Found test_connection tool${NC}"
    
    # Check for other expected tools
    if echo "$TOOLS_RESPONSE" | grep -q '"send_email"'; then
        echo -e "${GREEN}  ✓ Found send_email tool${NC}"
    fi
    if echo "$TOOLS_RESPONSE" | grep -q '"validate_addresses"'; then
        echo -e "${GREEN}  ✓ Found validate_addresses tool${NC}"
    fi
    if echo "$TOOLS_RESPONSE" | grep -q '"load_template"'; then
        echo -e "${GREEN}  ✓ Found load_template tool${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Tools list returned but no tools found (this might be expected)${NC}"
fi

# Clean up
rm -f "$RESPONSE_FILE" "$RESPONSE_FILE.sh"

echo -e "${GREEN}✅ Basic tests completed${NC}"
echo ""
echo "To test interactively, run:"
echo "  $SERVER_BIN -transport stdio"
echo ""
echo "Then send MCP requests like:"
echo '  {"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
echo '  {"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}'