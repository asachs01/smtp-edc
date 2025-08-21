# SMTP-EDC Desktop Extension (DXT)

A powerful SMTP testing and diagnostic tool packaged as a Desktop Extension (DXT) for AI assistants.

## Features

- **Test SMTP Connections**: Verify server connectivity and authentication
- **Send Test Emails**: Send emails with full control over parameters
- **Validate Email Addresses**: Check email format and MX records
- **Email Templates**: Load and process email templates with variables
- **Security Features**: Rate limiting, input sanitization, and audit logging
- **Local Operation**: Runs entirely on your local machine

## Installation

### Quick Install

1. Download the latest `smtp-edc.dxt` from the releases page
2. Open your AI assistant (Claude, etc.) that supports DXT
3. Navigate to Extensions or Integrations settings
4. Click "Install from file" and select the downloaded `.dxt` file

### Build from Source

```bash
# Clone the repository
git clone https://github.com/asachs01/smtp-edc.git
cd smtp-edc/dxt-extension

# Install build dependencies
npm install

# Build the extension
npm run build

# The extension will be created at dist/smtp-edc.dxt
```

## Available Tools

### smtp_test_connection

Test an SMTP server connection and authentication.

**Parameters:**
- `server` (required): SMTP server hostname or IP
- `port`: Server port (default: 587)
- `username`: Authentication username
- `password`: Authentication password
- `authType`: Authentication type (plain/login/cram-md5)
- `starttls`: Use STARTTLS (default: true)
- `skipVerify`: Skip TLS certificate verification (default: false)

**Example:**
```json
{
  "tool": "smtp_test_connection",
  "arguments": {
    "server": "smtp.gmail.com",
    "port": 587,
    "username": "user@gmail.com",
    "password": "app-password",
    "starttls": true
  }
}
```

### smtp_send_email

Send an email via SMTP.

**Parameters:**
- `server` (required): SMTP server hostname
- `from` (required): Sender email address
- `to` (required): Recipient email address(es)
- `subject` (required): Email subject
- `body` (required): Email body content
- `port`: Server port (default: 587)
- `username`: Authentication username
- `password`: Authentication password
- `cc`: CC recipients
- `bcc`: BCC recipients
- `isHTML`: Whether body is HTML (default: false)
- `authType`: Authentication type
- `starttls`: Use STARTTLS (default: true)
- `skipVerify`: Skip TLS verification (default: false)

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
    "body": "This is a test email sent via SMTP-EDC DXT.",
    "starttls": true
  }
}
```

### smtp_validate_addresses

Validate email addresses with optional MX record checking.

**Parameters:**
- `addresses` (required): Array of email addresses to validate
- `checkMX`: Check MX records for domains (default: false)

**Example:**
```json
{
  "tool": "smtp_validate_addresses",
  "arguments": {
    "addresses": ["user@example.com", "invalid-email"],
    "checkMX": true
  }
}
```

### smtp_load_template

Load and process an email template with variable substitution.

**Parameters:**
- `templateName` (required): Name of the template file
- `variables`: Object with variables to substitute

**Example:**
```json
{
  "tool": "smtp_load_template",
  "arguments": {
    "templateName": "welcome.html",
    "variables": {
      "name": "John Doe",
      "company": "Example Corp"
    }
  }
}
```

## Configuration

The extension can be configured through user settings:

- **default_server**: Default SMTP server hostname
- **default_port**: Default SMTP port (587)
- **debug_mode**: Enable debug logging
- **timeout**: Connection timeout in seconds (30)
- **rate_limit**: Maximum emails per minute (10, 0 = unlimited)

Configuration is stored in `~/.smtp-edc/config.json`:

```json
{
  "default_server": "smtp.example.com",
  "default_port": 587,
  "debug_mode": false,
  "timeout": 30,
  "rate_limit": 10
}
```

## Email Templates

Templates are stored in `~/.smtp-edc/templates/`. Create HTML or text files with variable placeholders:

```html
<!DOCTYPE html>
<html>
<body>
  <h1>Welcome {{name}}!</h1>
  <p>Thank you for joining {{company}}.</p>
</body>
</html>
```

Variables are replaced using `{{variable_name}}` syntax.

## Security Features

### Rate Limiting
- Configurable rate limit (default: 10 emails/minute)
- Prevents abuse and accidental bulk sending

### Input Sanitization
- Removes null bytes and control characters
- Limits input length to prevent overflow
- Sanitizes email subjects and bodies

### Connection Security
- Blocks connections to localhost and private IPs
- Blocks dangerous ports (SSH, Telnet, SMB, RDP)
- Validates server parameters

### Content Security
- Checks for sensitive data (credit cards, API keys, private keys)
- Validates attachment types and sizes
- Limits recipient count (max 100)

### Audit Logging
- Logs all connection attempts and email sends
- Masks sensitive data in logs
- Available in debug mode

## Troubleshooting

### Common Issues

**Connection Refused**
- Verify the SMTP server address and port
- Check firewall settings
- Ensure the server supports the authentication method

**Authentication Failed**
- Double-check username and password
- For Gmail, use App Passwords instead of regular password
- Verify the authentication type matches server requirements

**TLS/SSL Errors**
- Try toggling `starttls` setting
- Use `skipVerify: true` for self-signed certificates (not recommended for production)
- Check if the port matches the security setting (465 for SSL, 587 for STARTTLS)

**Rate Limit Exceeded**
- Wait 1 minute before sending more emails
- Adjust rate_limit in configuration
- Set rate_limit to 0 for unlimited (use with caution)

### Debug Mode

Enable debug mode for detailed logging:

```json
{
  "debug_mode": true
}
```

Debug logs will show:
- Connection attempts and results
- Authentication details (passwords masked)
- Email send operations
- Template processing
- Security validations

## Development

### Project Structure
```
dxt-extension/
├── manifest.json       # DXT manifest file
├── server/
│   ├── index.js       # Main MCP server
│   ├── security.js    # Security utilities
│   └── package.json   # Node.js dependencies
├── build.js           # Build script
├── test.js           # Test suite
└── README.md         # This file
```

### Testing

Run the test suite:
```bash
npm test
```

Test individual tools:
```bash
# Test connection
node server/index.js < test/test-connection.json

# Test email sending
node server/index.js < test/test-send.json
```

### Building

```bash
# Install dependencies
npm install

# Build the extension
npm run build

# Clean build artifacts
npm run clean
```

## License

MIT License - See LICENSE file for details

## Support

For issues, feature requests, or questions:
- Open an issue on [GitHub](https://github.com/asachs01/smtp-edc/issues)
- Check the [main project documentation](https://github.com/asachs01/smtp-edc)

## Changelog

### 1.0.0 (2024-01-21)
- Initial release
- Full SMTP testing capabilities
- Email template support
- Comprehensive security features
- MCP SDK integration