# SMTP EDC MCP Tools Reference

## Overview

SMTP EDC provides Model Context Protocol (MCP) integration, allowing AI assistants like Claude to interact with SMTP servers for testing, troubleshooting, and automation. This document describes all available MCP tools and their parameters.

## Table of Contents

- [MCP Server Setup](#mcp-server-setup)
- [Available Tools](#available-tools)
- [Tool Parameters](#tool-parameters)
- [Usage Examples](#usage-examples)
- [Integration Guide](#integration-guide)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## MCP Server Setup

### Installation and Configuration

1. **Install SMTP EDC MCP Server**
```bash
# Install via package manager
brew install smtp-edc-mcp

# Or download binary
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-mcp-server
chmod +x smtp-edc-mcp-server
```

2. **Configuration File**
```json
{
  "servers": {
    "smtp-edc": {
      "command": "smtp-edc-mcp-server",
      "args": ["--config", "/path/to/smtp-edc-mcp-config.json"],
      "env": {
        "SMTP_EDC_LOG_LEVEL": "info"
      }
    }
  }
}
```

3. **MCP Server Configuration**
```json
{
  "server": {
    "name": "smtp-edc-mcp",
    "version": "1.0.0",
    "description": "SMTP Enhanced Delivery Client MCP Server"
  },
  "default_settings": {
    "timeout": 30,
    "retries": 3,
    "debug": false
  },
  "security": {
    "max_attachment_size": "10MB",
    "allowed_domains": ["*"],
    "rate_limit": {
      "requests_per_minute": 60,
      "burst": 10
    }
  }
}
```

### Claude Desktop Integration

Add to your Claude Desktop configuration:
```json
{
  "mcpServers": {
    "smtp-edc": {
      "command": "smtp-edc-mcp-server",
      "args": ["--config", "/path/to/smtp-edc-mcp-config.json"]
    }
  }
}
```

## Available Tools

### 1. smtp_test_connection

Tests basic SMTP server connectivity and capabilities.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port (default: 25)
- `timeout` (optional): Connection timeout in seconds (default: 30)
- `tls` (optional): Enable TLS/STARTTLS (default: false)
- `skip_verify` (optional): Skip certificate verification (default: false)

**Example:**
```json
{
  "tool": "smtp_test_connection",
  "parameters": {
    "server": "smtp.gmail.com",
    "port": 587,
    "tls": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "connection_time": 0.234,
  "capabilities": {
    "starttls": true,
    "auth_methods": ["PLAIN", "LOGIN", "CRAM-MD5"],
    "max_message_size": 26214400,
    "features": ["8BITMIME", "PIPELINING", "DSN"]
  },
  "server_banner": "220 smtp.gmail.com ESMTP ready"
}
```

### 2. smtp_authenticate

Tests SMTP authentication with various methods.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port (default: 587)
- `auth_type` (required): Authentication method (plain, login, cram-md5, oauth2)
- `username` (required): Authentication username
- `password` (conditional): Password (required for plain, login, cram-md5)
- `oauth_token` (conditional): OAuth token (required for oauth2)
- `tls` (optional): Enable TLS/STARTTLS (default: true)

**Example:**
```json
{
  "tool": "smtp_authenticate",
  "parameters": {
    "server": "smtp.gmail.com",
    "port": 587,
    "auth_type": "plain",
    "username": "user@gmail.com",
    "password": "app-password",
    "tls": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "auth_method": "PLAIN",
  "connection_time": 0.456,
  "auth_time": 0.123,
  "server_response": "235 Authentication successful"
}
```

### 3. smtp_send_email

Sends an email through the specified SMTP server.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port (default: 587)
- `auth` (optional): Authentication configuration object
- `message` (required): Message configuration object
- `options` (optional): Additional options object

**Message Object:**
- `from` (required): Sender email address
- `to` (required): Array of recipient email addresses
- `cc` (optional): Array of CC recipients
- `bcc` (optional): Array of BCC recipients
- `subject` (required): Email subject
- `body` (optional): Plain text body
- `html` (optional): HTML body
- `attachments` (optional): Array of attachment objects
- `headers` (optional): Custom headers object

**Example:**
```json
{
  "tool": "smtp_send_email",
  "parameters": {
    "server": "smtp.gmail.com",
    "port": 587,
    "auth": {
      "type": "plain",
      "username": "sender@gmail.com",
      "password": "app-password"
    },
    "message": {
      "from": "sender@gmail.com",
      "to": ["recipient@example.com"],
      "subject": "Test Email",
      "body": "This is a test email sent via MCP.",
      "html": "<h1>Test Email</h1><p>This is a test email sent via MCP.</p>"
    },
    "options": {
      "tls": true,
      "debug": false
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "20240101120000.ABC123@smtp.gmail.com",
  "recipients_accepted": ["recipient@example.com"],
  "recipients_rejected": [],
  "send_time": 1.234,
  "size_bytes": 1024
}
```

### 4. smtp_validate_email

Validates email addresses using various methods.

**Parameters:**
- `email` (required): Email address to validate
- `check_syntax` (optional): Check email syntax (default: true)
- `check_mx` (optional): Check MX records (default: true)
- `check_smtp` (optional): Test SMTP connection (default: false)
- `smtp_server` (conditional): SMTP server for connection test

**Example:**
```json
{
  "tool": "smtp_validate_email",
  "parameters": {
    "email": "user@example.com",
    "check_syntax": true,
    "check_mx": true,
    "check_smtp": true,
    "smtp_server": "smtp.example.com"
  }
}
```

**Response:**
```json
{
  "valid": true,
  "email": "user@example.com",
  "domain": "example.com",
  "checks": {
    "syntax": true,
    "mx_record": true,
    "smtp_connection": true
  },
  "mx_records": [
    {
      "host": "mail.example.com",
      "priority": 10
    }
  ],
  "smtp_response": "250 OK"
}
```

### 5. smtp_get_capabilities

Retrieves SMTP server capabilities and features.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port (default: 25)
- `tls` (optional): Use TLS/STARTTLS (default: false)
- `auth` (optional): Authentication for EHLO after auth

**Example:**
```json
{
  "tool": "smtp_get_capabilities",
  "parameters": {
    "server": "smtp.office365.com",
    "port": 587,
    "tls": true
  }
}
```

**Response:**
```json
{
  "server": "smtp.office365.com",
  "port": 587,
  "capabilities": {
    "starttls": true,
    "auth_methods": ["PLAIN", "LOGIN", "XOAUTH2"],
    "max_message_size": 26214400,
    "features": ["8BITMIME", "PIPELINING", "DSN", "SMTPUTF8"],
    "extensions": {
      "PIPELINING": true,
      "SIZE": 26214400,
      "8BITMIME": true,
      "DSN": true,
      "STARTTLS": true,
      "AUTH": ["PLAIN", "LOGIN", "XOAUTH2"]
    }
  },
  "server_banner": "220 smtp.office365.com Microsoft ESMTP MAIL Service ready"
}
```

### 6. smtp_troubleshoot

Performs comprehensive SMTP troubleshooting with natural language analysis.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port
- `issue_description` (optional): Natural language description of the issue
- `auth` (optional): Authentication configuration
- `test_email` (optional): Test email configuration
- `include_network_tests` (optional): Include network connectivity tests (default: true)

**Example:**
```json
{
  "tool": "smtp_troubleshoot",
  "parameters": {
    "server": "smtp.example.com",
    "port": 587,
    "issue_description": "Emails are not being delivered and I'm getting authentication errors",
    "auth": {
      "type": "plain",
      "username": "user@example.com",
      "password": "password"
    },
    "include_network_tests": true
  }
}
```

**Response:**
```json
{
  "analysis": {
    "overall_status": "issues_found",
    "confidence": 0.85,
    "primary_issue": "authentication_failure",
    "secondary_issues": ["tls_configuration"]
  },
  "tests_performed": [
    {
      "test": "network_connectivity",
      "status": "passed",
      "result": "Connection successful",
      "time": 0.123
    },
    {
      "test": "tls_handshake",
      "status": "warning",
      "result": "Certificate verification failed",
      "recommendation": "Check certificate validity or use --skip-verify for testing"
    },
    {
      "test": "authentication",
      "status": "failed",
      "result": "535 Authentication credentials invalid",
      "recommendation": "Verify username and password, check if account is locked"
    }
  ],
  "recommendations": [
    {
      "priority": "high",
      "issue": "Authentication failure",
      "solution": "Verify credentials and ensure the account has SMTP access enabled",
      "commands": [
        "smtp-edc --server smtp.example.com --port 587 --auth plain --username user@example.com --debug"
      ]
    },
    {
      "priority": "medium",
      "issue": "TLS certificate verification",
      "solution": "Update certificate store or configure proper certificate validation",
      "commands": [
        "smtp-edc --server smtp.example.com --port 587 --starttls --skip-verify"
      ]
    }
  ],
  "next_steps": [
    "Contact email administrator to verify account status",
    "Test with a different authentication method",
    "Check server logs for additional error information"
  ]
}
```

### 7. smtp_bulk_test

Tests multiple SMTP configurations or recipients in batch.

**Parameters:**
- `configurations` (required): Array of SMTP configuration objects
- `parallel` (optional): Execute tests in parallel (default: false)
- `max_concurrent` (optional): Maximum concurrent connections (default: 5)
- `fail_fast` (optional): Stop on first failure (default: false)

**Example:**
```json
{
  "tool": "smtp_bulk_test",
  "parameters": {
    "configurations": [
      {
        "name": "Gmail Test",
        "server": "smtp.gmail.com",
        "port": 587,
        "auth": {"type": "plain", "username": "user1@gmail.com", "password": "pass1"},
        "message": {"from": "user1@gmail.com", "to": ["test1@example.com"], "subject": "Bulk Test 1"}
      },
      {
        "name": "Office365 Test",
        "server": "smtp.office365.com",
        "port": 587,
        "auth": {"type": "login", "username": "user2@company.com", "password": "pass2"},
        "message": {"from": "user2@company.com", "to": ["test2@example.com"], "subject": "Bulk Test 2"}
      }
    ],
    "parallel": true,
    "max_concurrent": 2
  }
}
```

**Response:**
```json
{
  "summary": {
    "total_tests": 2,
    "successful": 1,
    "failed": 1,
    "total_time": 2.345
  },
  "results": [
    {
      "name": "Gmail Test",
      "success": true,
      "message_id": "20240101120000.ABC123@smtp.gmail.com",
      "time": 1.234
    },
    {
      "name": "Office365 Test",
      "success": false,
      "error": "535 Authentication credentials invalid",
      "time": 1.111
    }
  ]
}
```

### 8. smtp_monitor_queue

Monitors SMTP queue status and performance metrics.

**Parameters:**
- `server` (required): SMTP server hostname
- `port` (optional): SMTP server port
- `auth` (optional): Authentication configuration
- `duration` (optional): Monitoring duration in seconds (default: 60)
- `interval` (optional): Check interval in seconds (default: 10)

**Example:**
```json
{
  "tool": "smtp_monitor_queue",
  "parameters": {
    "server": "smtp.example.com",
    "port": 587,
    "duration": 300,
    "interval": 30
  }
}
```

**Response:**
```json
{
  "monitoring_period": {
    "start": "2024-01-01T12:00:00Z",
    "end": "2024-01-01T12:05:00Z",
    "duration": 300
  },
  "metrics": {
    "total_connections": 15,
    "successful_connections": 14,
    "failed_connections": 1,
    "average_connection_time": 0.456,
    "messages_sent": 128,
    "messages_failed": 2,
    "throughput": {
      "messages_per_second": 0.427,
      "bytes_per_second": 1024.5
    }
  },
  "issues_detected": [
    {
      "timestamp": "2024-01-01T12:02:30Z",
      "issue": "temporary_failure",
      "details": "450 Temporary failure, please try again later"
    }
  ]
}
```

## Tool Parameters Reference

### Common Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `server` | string | Yes | - | SMTP server hostname |
| `port` | integer | No | 25/587 | SMTP server port |
| `timeout` | integer | No | 30 | Connection timeout in seconds |
| `debug` | boolean | No | false | Enable debug output |
| `tls` | boolean | No | false | Enable TLS/STARTTLS |
| `skip_verify` | boolean | No | false | Skip certificate verification |

### Authentication Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `auth_type` | string | Yes | Authentication method (plain, login, cram-md5, oauth2) |
| `username` | string | Yes | Authentication username |
| `password` | string | Conditional | Password (required for plain, login, cram-md5) |
| `oauth_token` | string | Conditional | OAuth token (required for oauth2) |

### Message Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `from` | string | Yes | Sender email address |
| `to` | array | Yes | Recipient email addresses |
| `cc` | array | No | CC recipient addresses |
| `bcc` | array | No | BCC recipient addresses |
| `subject` | string | Yes | Email subject |
| `body` | string | No | Plain text body |
| `html` | string | No | HTML body |
| `headers` | object | No | Custom headers |
| `attachments` | array | No | File attachments |

### Attachment Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | string | Yes | Attachment filename |
| `content` | string | Yes | Base64-encoded content |
| `content_type` | string | No | MIME content type |
| `disposition` | string | No | Content disposition (attachment, inline) |

## Usage Examples

### Basic SMTP Testing with Claude

```
Human: Can you test if smtp.gmail.com is working properly with port 587?

Claude: I'll test the SMTP connection to Gmail for you using the smtp_test_connection tool.

{
  "tool": "smtp_test_connection",
  "parameters": {
    "server": "smtp.gmail.com",
    "port": 587,
    "tls": true
  }
}

The connection test was successful! Here are the results:
- Connection time: 0.234 seconds
- TLS/STARTTLS: Supported
- Authentication methods: PLAIN, LOGIN, CRAM-MD5
- Maximum message size: 25MB
- Additional features: 8BITMIME, PIPELINING, DSN

Gmail's SMTP server is working properly and accepting connections on port 587.
```

### Email Validation

```
Human: Please validate this email address: user@example.com

Claude: I'll validate that email address for you, including checking the domain's MX records.

{
  "tool": "smtp_validate_email",
  "parameters": {
    "email": "user@example.com",
    "check_syntax": true,
    "check_mx": true,
    "check_smtp": false
  }
}

The email validation results:
✅ Syntax: Valid
✅ Domain: example.com exists
✅ MX Records: Found (mail.example.com with priority 10)

The email address user@example.com appears to be valid and deliverable.
```

### Troubleshooting SMTP Issues

```
Human: I'm having trouble sending emails through our company SMTP server. Emails aren't being delivered and I'm getting authentication errors.

Claude: I'll help troubleshoot your SMTP issues. Let me run a comprehensive analysis of your server.

{
  "tool": "smtp_troubleshoot",
  "parameters": {
    "server": "smtp.yourcompany.com",
    "port": 587,
    "issue_description": "Emails not being delivered, authentication errors",
    "include_network_tests": true
  }
}

I found several issues with your SMTP configuration:

🔴 **Primary Issue: Authentication Failure**
- Error: "535 Authentication credentials invalid"
- Recommendation: Verify your username and password, and check if your account has SMTP access enabled

⚠️ **Secondary Issue: TLS Configuration**
- Warning: Certificate verification failed
- Recommendation: Update your certificate store or contact your IT administrator

**Next Steps:**
1. Contact your email administrator to verify your account status
2. Try using a different authentication method (LOGIN instead of PLAIN)
3. Test with debug mode to see the full SMTP conversation

Would you like me to test with specific credentials or run additional diagnostics?
```

### Sending Test Emails

```
Human: Send a test email from test@mycompany.com to recipient@example.com with the subject "MCP Test" using our SMTP server at smtp.mycompany.com

Claude: I'll send that test email for you using your company's SMTP server.

{
  "tool": "smtp_send_email",
  "parameters": {
    "server": "smtp.mycompany.com",
    "port": 587,
    "auth": {
      "type": "plain",
      "username": "test@mycompany.com",
      "password": "[password would be provided securely]"
    },
    "message": {
      "from": "test@mycompany.com",
      "to": ["recipient@example.com"],
      "subject": "MCP Test",
      "body": "This is a test email sent via MCP SMTP tools.",
      "html": "<h2>MCP Test</h2><p>This is a test email sent via MCP SMTP tools.</p>"
    },
    "options": {
      "tls": true,
      "debug": false
    }
  }
}

✅ **Email sent successfully!**
- Message ID: 20240101120000.ABC123@smtp.mycompany.com
- Recipient accepted: recipient@example.com
- Send time: 1.234 seconds
- Message size: 1,024 bytes

The test email has been delivered to the SMTP server and should arrive at the recipient shortly.
```

## Integration Guide

### Adding SMTP EDC MCP to Your Workflow

1. **Install and Configure**
```bash
# Install MCP server
brew install smtp-edc-mcp

# Create configuration
cat > ~/.config/smtp-edc-mcp/config.json << EOF
{
  "server": {
    "name": "smtp-edc-mcp",
    "version": "1.0.0"
  },
  "default_settings": {
    "timeout": 30,
    "retries": 3
  }
}
EOF
```

2. **Add to Claude Desktop**
```json
{
  "mcpServers": {
    "smtp-edc": {
      "command": "smtp-edc-mcp-server",
      "args": ["--config", "~/.config/smtp-edc-mcp/config.json"]
    }
  }
}
```

3. **Test Integration**
```bash
# Verify MCP server is running
smtp-edc-mcp-server --test-connection

# Check available tools
smtp-edc-mcp-server --list-tools
```

### Custom Tool Development

You can extend SMTP EDC MCP with custom tools:

```go
// custom_tool.go
package main

import (
    "context"
    "github.com/asachs/smtp-edc/pkg/mcp"
)

func init() {
    mcp.RegisterTool("smtp_custom_check", &mcp.Tool{
        Name:        "smtp_custom_check",
        Description: "Custom SMTP validation tool",
        Parameters: map[string]interface{}{
            "server": {
                "type":        "string",
                "description": "SMTP server hostname",
                "required":    true,
            },
        },
        Handler: handleCustomCheck,
    })
}

func handleCustomCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    server := params["server"].(string)

    // Custom validation logic
    result := performCustomCheck(server)

    return map[string]interface{}{
        "server": server,
        "custom_check_result": result,
    }, nil
}
```

## Error Handling

### Common Error Codes

| Error Code | Description | Resolution |
|------------|-------------|------------|
| `SMTP_CONNECTION_FAILED` | Cannot connect to server | Check hostname, port, and network connectivity |
| `SMTP_AUTH_FAILED` | Authentication failed | Verify credentials and authentication method |
| `SMTP_TLS_FAILED` | TLS handshake failed | Check certificate validity or use skip-verify |
| `SMTP_TIMEOUT` | Operation timed out | Increase timeout value or check network |
| `SMTP_INVALID_EMAIL` | Invalid email format | Validate email address syntax |
| `SMTP_QUOTA_EXCEEDED` | Rate limit exceeded | Reduce request frequency |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "SMTP_AUTH_FAILED",
    "message": "Authentication failed: 535 Authentication credentials invalid",
    "details": {
      "server_response": "535 Authentication credentials invalid",
      "auth_method": "PLAIN",
      "username": "user@example.com"
    },
    "suggestions": [
      "Verify username and password",
      "Check if account has SMTP access enabled",
      "Try a different authentication method"
    ]
  }
}
```

## Best Practices

### Security
- Never log passwords or sensitive authentication data
- Use environment variables for credentials
- Implement rate limiting to prevent abuse
- Validate all input parameters
- Use secure connections (TLS/STARTTLS) when possible

### Performance
- Implement connection pooling for bulk operations
- Use appropriate timeouts based on network conditions
- Cache server capabilities to reduce redundant requests
- Implement exponential backoff for retries

### Monitoring
- Log all SMTP transactions for debugging
- Monitor connection success rates
- Track response times and performance metrics
- Set up alerts for authentication failures

### User Experience
- Provide clear error messages with actionable suggestions
- Include relevant context in troubleshooting responses
- Support natural language problem descriptions
- Offer multiple resolution paths for common issues

---

This MCP tools reference provides comprehensive documentation for integrating SMTP EDC with AI assistants through the Model Context Protocol. For additional support, please refer to the main documentation or submit issues to the project repository.
