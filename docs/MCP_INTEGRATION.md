# Model Context Protocol (MCP) Integration

## Overview

SMTP-EDC now includes Model Context Protocol (MCP) server support, enabling AI assistants and other MCP clients to interact with SMTP functionality through a standardized protocol.

## What is MCP?

The Model Context Protocol (MCP) is an open-source protocol that standardizes how applications and AI models interact with external context, tools, and data sources. It provides a modular client-server architecture for dynamic integration of capabilities into AI-driven workflows.

## Installation and Setup

### Starting the MCP Server

The MCP server can be started in two modes:

#### STDIO Transport (Recommended for local development)
```bash
# Start MCP server with STDIO transport
smtp-edc mcp-server -transport stdio

# With debug logging
smtp-edc mcp-server -transport stdio -debug
```

#### HTTP Transport (For remote access)
```bash
# Start MCP server with HTTP transport on default port 8080
smtp-edc mcp-server -transport http

# Custom port
smtp-edc mcp-server -transport http -port 9090

# With debug logging
smtp-edc mcp-server -transport http -debug
```

### Configuration

MCP server can be configured using a JSON configuration file. Create a `mcp-config.json` file:

```json
{
  "transport": {
    "type": "stdio",
    "http": {
      "port": 8080,
      "host": "localhost",
      "enableCors": true,
      "enableHttps": false
    }
  },
  "server": {
    "name": "smtp-edc-mcp",
    "version": "1.0.0",
    "description": "SMTP-EDC Model Context Protocol Server",
    "debug": false,
    "maxClients": 10,
    "timeout": 30
  },
  "security": {
    "requireAuth": false,
    "apiKeys": [],
    "allowedIps": ["127.0.0.1", "::1"]
  }
}
```

## Available MCP Tools

The MCP server exposes the following tools for SMTP operations:

### 1. smtp_test_connection
Test SMTP server connection and capabilities.

**Parameters:**
- `server` (string, required): SMTP server hostname or IP
- `port` (integer): SMTP server port (default: 587)
- `username` (string): Authentication username
- `password` (string): Authentication password
- `authType` (string): Authentication type (plain, login, cram-md5)
- `starttls` (boolean): Use STARTTLS
- `skipVerify` (boolean): Skip TLS certificate verification

**Example:**
```json
{
  "tool": "smtp_test_connection",
  "arguments": {
    "server": "smtp.gmail.com",
    "port": 587,
    "username": "user@gmail.com",
    "password": "app-password",
    "authType": "plain",
    "starttls": true
  }
}
```

### 2. smtp_send_email
Send an email via SMTP.

**Parameters:**
- `server` (string, required): SMTP server hostname or IP
- `port` (integer): SMTP server port (default: 587)
- `username` (string): Authentication username
- `password` (string): Authentication password
- `from` (string, required): Sender email address
- `to` (array, required): Recipient email addresses
- `cc` (array): CC recipient email addresses
- `bcc` (array): BCC recipient email addresses
- `subject` (string, required): Email subject
- `body` (string, required): Email body (text or HTML)
- `isHTML` (boolean): Whether the body is HTML
- `authType` (string): Authentication type
- `starttls` (boolean): Use STARTTLS
- `skipVerify` (boolean): Skip TLS certificate verification

**Example:**
```json
{
  "tool": "smtp_send_email",
  "arguments": {
    "server": "smtp.gmail.com",
    "port": 587,
    "username": "sender@gmail.com",
    "password": "app-password",
    "from": "sender@gmail.com",
    "to": ["recipient@example.com"],
    "subject": "Test Email",
    "body": "This is a test email sent via MCP.",
    "authType": "plain",
    "starttls": true
  }
}
```

### 3. smtp_validate_addresses
Validate email addresses and optionally check MX records.

**Parameters:**
- `addresses` (array, required): Email addresses to validate
- `checkMX` (boolean): Check MX records for domains

**Example:**
```json
{
  "tool": "smtp_validate_addresses",
  "arguments": {
    "addresses": ["test@example.com", "invalid-email"],
    "checkMX": true
  }
}
```

### 4. smtp_load_template
Load and process an email template.

**Parameters:**
- `templateName` (string, required): Name of the template to load
- `variables` (object): Variables to substitute in the template

**Example:**
```json
{
  "tool": "smtp_load_template",
  "arguments": {
    "templateName": "welcome",
    "variables": {
      "name": "John Doe",
      "company": "Example Corp"
    }
  }
}
```

## Available MCP Resources

The MCP server provides access to the following resources:

### 1. smtp-edc://config/current
Returns the current SMTP-EDC configuration settings.

### 2. smtp-edc://templates/list
Returns a list of available email templates.

### 3. smtp-edc://stats/auth
Returns authentication attempt statistics.

## Integration with AI Assistants

### Claude Desktop

To use SMTP-EDC with Claude Desktop, add the following to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "smtp-edc": {
      "command": "smtp-edc",
      "args": ["mcp-server", "-transport", "stdio"],
      "env": {}
    }
  }
}
```

### Custom MCP Clients

For custom MCP client implementations, connect to the server using either:

1. **STDIO**: Launch the process and communicate via stdin/stdout
2. **HTTP**: Connect to the HTTP endpoint (default: `http://localhost:8080`)

## Security Considerations

### Authentication
- Enable `requireAuth` in the configuration to require API key authentication
- API keys should be transmitted in the `Authorization` header

### Network Security
- Use HTTPS for production deployments
- Configure `allowedIps` to restrict access to specific IP addresses
- Use firewall rules to limit network access

### Credential Management
- Never hardcode credentials in configuration files
- Use environment variables or secure credential stores
- Rotate credentials regularly

## Examples

### Testing SMTP Connection via MCP
```python
# Python example using MCP client
import mcp_client

client = mcp_client.connect("stdio", ["smtp-edc", "mcp-server"])
result = client.call_tool("smtp_test_connection", {
    "server": "smtp.gmail.com",
    "port": 587,
    "username": "test@gmail.com",
    "password": "app-password",
    "starttls": True
})
print(result)
```

### Sending Email via MCP
```javascript
// JavaScript example using MCP client
const mcp = require('mcp-client');

const client = await mcp.connect('http', 'http://localhost:8080');
const result = await client.callTool('smtp_send_email', {
  server: 'smtp.gmail.com',
  port: 587,
  from: 'sender@gmail.com',
  to: ['recipient@example.com'],
  subject: 'Test Email',
  body: 'Hello from MCP!',
  username: 'sender@gmail.com',
  password: 'app-password',
  starttls: true
});
console.log(result);
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure the MCP server is running
   - Check the port is not blocked by firewall
   - Verify the transport type matches your client configuration

2. **Authentication Failed**
   - Verify API keys are correctly configured
   - Check that `requireAuth` setting matches client expectations

3. **Tool Not Found**
   - Ensure you're using the correct tool name
   - Check that the MCP server version supports the tool

### Debug Mode

Enable debug mode for detailed logging:
```bash
smtp-edc mcp-server -debug
```

## API Reference

For detailed API documentation and protocol specifications, refer to:
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [SMTP-EDC API Documentation](./API_REFERENCE.md)

## Contributing

To contribute to the MCP integration:

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests for new functionality
5. Update documentation
6. Submit a pull request

## License

The MCP integration is part of SMTP-EDC and is licensed under the MIT License.