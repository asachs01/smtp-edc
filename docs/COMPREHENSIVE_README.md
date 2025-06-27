# SMTP EDC (Enhanced Delivery Client)

![SMTP-EDC Logo](../smtp-edc.png)

[![Release](https://github.com/asachs01/smtp-edc/actions/workflows/release.yml/badge.svg)](https://github.com/asachs01/smtp-edc/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/asachs01/smtp-edc)](https://goreportcard.com/report/github.com/asachs01/smtp-edc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Deployment](#deployment)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

## Overview

SMTP EDC is a powerful, cross-platform SMTP testing and troubleshooting tool designed as an "Every Day Carry" (EDC) tool for system administrators. Built in Go, it provides comprehensive SMTP testing capabilities with both command-line and graphical interfaces, making it the modern alternative to traditional tools like Swaks.

### What Makes SMTP EDC Special

- **Zero Dependencies**: Single binary with no external dependencies
- **Cross-Platform**: Native support for Windows, macOS, and Linux
- **Dual Interface**: Both GUI and CLI for different user preferences
- **Enterprise Ready**: Advanced features for production environments
- **Developer Friendly**: Extensive configuration options and scripting support

## Features

### Core SMTP Testing
- **Protocol Compliance**: Full SMTP/ESMTP protocol support
- **Authentication Methods**: PLAIN, LOGIN, CRAM-MD5, DIGEST-MD5, NTLM, OAuth2
- **Security**: TLS/STARTTLS with configurable certificate validation
- **Message Types**: Plain text, HTML, multipart messages with attachments
- **Advanced Headers**: Custom headers, DKIM signing, SPF validation

### Debugging & Analysis
- **Protocol Logging**: Complete SMTP conversation logging
- **Performance Metrics**: Connection timing and throughput analysis
- **Error Analysis**: Detailed error reporting with suggestions
- **Network Diagnostics**: DNS, MX record validation, connectivity tests

### Enterprise Features
- **Queue Management**: Message queuing with retry logic
- **Batch Processing**: Send multiple messages efficiently
- **Template Engine**: Dynamic message generation
- **Monitoring Integration**: Prometheus metrics, structured logging
- **Configuration Management**: YAML/JSON configuration with environment variable support

## Quick Start

### 1. Install SMTP EDC

```bash
# Using Homebrew (macOS/Linux)
brew tap asachs01/smtp-edc
brew install smtp-edc

# Using Go
go install github.com/asachs/smtp-edc/cmd/smtp-edc@latest

# Download binary from releases
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-linux-amd64
chmod +x smtp-edc-linux-amd64
sudo mv smtp-edc-linux-amd64 /usr/local/bin/smtp-edc
```

### 2. Quick Test

```bash
# Basic connectivity test
smtp-edc --server smtp.gmail.com --port 587 --from you@example.com --to test@example.com --subject "Test" --body "Hello World"

# With authentication
smtp-edc --server smtp.gmail.com --port 587 --starttls --auth plain --username you@gmail.com --password your-app-password --from you@gmail.com --to test@example.com --subject "Test" --body "Hello World"
```

### 3. Launch GUI (Recommended for Beginners)

```bash
smtp-edc --gui
```

## Installation

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap asachs01/smtp-edc
brew install smtp-edc
```

#### Chocolatey (Windows)
```bash
choco install smtp-edc
```

#### APT (Debian/Ubuntu)
```bash
curl -fsSL https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc_amd64.deb -o smtp-edc.deb
sudo dpkg -i smtp-edc.deb
```

#### YUM/DNF (RedHat/CentOS/Fedora)
```bash
curl -fsSL https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-1.0.0-1.x86_64.rpm -o smtp-edc.rpm
sudo rpm -i smtp-edc.rpm
```

### Direct Download

Download the appropriate binary for your platform from the [releases page](https://github.com/asachs01/smtp-edc/releases).

```bash
# Linux AMD64
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-linux-amd64
chmod +x smtp-edc-linux-amd64
sudo mv smtp-edc-linux-amd64 /usr/local/bin/smtp-edc

# macOS AMD64
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-darwin-amd64
chmod +x smtp-edc-darwin-amd64
sudo mv smtp-edc-darwin-amd64 /usr/local/bin/smtp-edc

# Windows AMD64
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-windows-amd64.exe
```

### From Source

```bash
git clone https://github.com/asachs01/smtp-edc.git
cd smtp-edc
go build -o smtp-edc cmd/smtp-edc/main.go
go build -o smtp-edc-gui cmd/smtp-edc-gui/main.go
```

## Usage

### GUI Mode (Recommended)

The GUI provides an intuitive interface for users who prefer visual interaction:

```bash
smtp-edc --gui
# or
smtp-edc-gui
```

**GUI Features:**
- **Server Configuration Tab**: Set up SMTP server details
- **Authentication Tab**: Configure authentication methods
- **Message Composer**: Create and customize email messages
- **Attachments Manager**: Add and manage file attachments
- **Testing Tools**: Built-in connectivity and authentication testing
- **Results Viewer**: View detailed transaction logs and results
- **Configuration Profiles**: Save and load common configurations

### CLI Mode (Advanced Users)

#### Basic Usage Examples

```bash
# Simple email test
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "Test" --body "Hello World"

# With port specification
smtp-edc --server smtp.example.com --port 587 --from sender@example.com --to recipient@example.com --subject "Test" --body "Hello World"

# Multiple recipients
smtp-edc --server smtp.example.com --from sender@example.com --to "user1@example.com,user2@example.com" --cc "cc@example.com" --bcc "bcc@example.com" --subject "Test" --body "Hello World"
```

#### Authentication Examples

```bash
# PLAIN authentication
smtp-edc --server smtp.gmail.com --port 587 --starttls --auth plain --username user@gmail.com --password app-password --from user@gmail.com --to test@example.com --subject "Test" --body "Hello World"

# LOGIN authentication
smtp-edc --server smtp.office365.com --port 587 --starttls --auth login --username user@company.com --password password --from user@company.com --to test@example.com --subject "Test" --body "Hello World"

# CRAM-MD5 authentication
smtp-edc --server mail.example.com --port 587 --auth cram-md5 --username user --password password --from user@example.com --to test@example.com --subject "Test" --body "Hello World"
```

#### Security and TLS Examples

```bash
# STARTTLS (recommended)
smtp-edc --server smtp.example.com --port 587 --starttls --from sender@example.com --to recipient@example.com --subject "Test" --body "Hello World"

# Skip certificate verification (testing only)
smtp-edc --server smtp.example.com --port 587 --starttls --skip-verify --from sender@example.com --to recipient@example.com --subject "Test" --body "Hello World"

# Implicit TLS (port 465)
smtp-edc --server smtp.example.com --port 465 --tls --from sender@example.com --to recipient@example.com --subject "Test" --body "Hello World"
```

#### Advanced Message Features

```bash
# HTML email
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "HTML Test" --html "<h1>Hello World</h1><p>This is HTML content</p>"

# Email with attachments
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "File Attachment" --body "Please find attached files" --attach /path/to/file1.pdf --attach /path/to/file2.jpg

# Custom headers
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "Custom Headers" --body "Hello World" --headers "X-Custom-Header: Value, X-Priority: 1"

# Email from file
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "From File" --body-file /path/to/email-body.txt

# HTML from file
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "HTML From File" --html-file /path/to/email-body.html
```

#### Template Usage

```bash
# Using email templates
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --template /path/to/template.tmpl --template-data '{"Name":"John","Company":"Acme Corp"}'

# Subject template
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject-template "Welcome {{.Name}} to {{.Company}}" --template-data '{"Name":"John","Company":"Acme Corp"}'
```

#### Debugging and Testing

```bash
# Debug mode (show protocol conversation)
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --debug --subject "Debug Test" --body "Hello World"

# Verbose output
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --verbose --subject "Verbose Test" --body "Hello World"

# Connection timeout
smtp-edc --server smtp.example.com --timeout 10 --from sender@example.com --to recipient@example.com --subject "Timeout Test" --body "Hello World"

# Retry configuration
smtp-edc --server smtp.example.com --retries 5 --from sender@example.com --to recipient@example.com --subject "Retry Test" --body "Hello World"

# MX record validation
smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --validate-mx --subject "MX Test" --body "Hello World"
```

### Command-Line Options Reference

| Option | Short | Description | Example |
|--------|-------|-------------|---------|
| `--server` | `-s` | SMTP server hostname | `--server smtp.gmail.com` |
| `--port` | `-p` | SMTP server port | `--port 587` |
| `--from` | `-f` | Sender email address | `--from sender@example.com` |
| `--to` | `-t` | Recipient email addresses (comma-separated) | `--to "user1@example.com,user2@example.com"` |
| `--cc` | `-C` | CC recipient email addresses | `--cc cc@example.com` |
| `--bcc` | `-B` | BCC recipient email addresses | `--bcc bcc@example.com` |
| `--subject` | `-S` | Email subject | `--subject "Test Email"` |
| `--body` | `-b` | Email body text | `--body "Hello World"` |
| `--html` | `-H` | Email HTML body | `--html "<h1>Hello</h1>"` |
| `--auth` | `-a` | Authentication type | `--auth plain` |
| `--username` | `-u` | Authentication username | `--username user@example.com` |
| `--password` | `-P` | Authentication password | `--password secret` |
| `--starttls` | `-l` | Use STARTTLS | `--starttls` |
| `--tls` |  | Use implicit TLS | `--tls` |
| `--skip-verify` | `-k` | Skip TLS certificate verification | `--skip-verify` |
| `--debug` | `-D` | Enable debug output | `--debug` |
| `--verbose` | `-v` | Enable verbose output | `--verbose` |
| `--gui` | `-g` | Launch GUI interface | `--gui` |
| `--config` | `-c` | Configuration file path | `--config /path/to/config.yaml` |

## Configuration

SMTP EDC supports multiple configuration methods:

### 1. Command-Line Arguments
Pass options directly via command line (highest priority).

### 2. Configuration Files
Use YAML or JSON configuration files for complex setups.

#### YAML Configuration Example

```yaml
# smtp-edc.yaml
server:
  host: smtp.example.com
  port: 587
  connection:
    timeout: 30
    keep_alive: true
    retry:
      max_attempts: 3
      delay: 5

auth:
  type: plain
  username: user@example.com
  password: ${SMTP_PASSWORD}  # Environment variable
  tls:
    enabled: true
    start_tls: true
    verify: true

message:
  from: sender@example.com
  headers:
    X-Mailer: SMTP-EDC
    X-Priority: 3
  templates:
    path: ./templates
    default: default.tmpl

logging:
  level: info
  format: json
  transaction:
    enabled: true
    include_data: false
```

#### JSON Configuration Example

```json
{
  "server": {
    "host": "smtp.example.com",
    "port": 587,
    "connection": {
      "timeout": 30,
      "retry": {
        "max_attempts": 3,
        "delay": 5
      }
    }
  },
  "auth": {
    "type": "plain",
    "username": "user@example.com",
    "password": "${SMTP_PASSWORD}",
    "tls": {
      "enabled": true,
      "start_tls": true,
      "verify": true
    }
  },
  "message": {
    "from": "sender@example.com",
    "headers": {
      "X-Mailer": "SMTP-EDC",
      "X-Priority": "3"
    }
  }
}
```

### 3. Environment Variables
Set environment variables for sensitive information:

```bash
export SMTP_SERVER=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=user@example.com
export SMTP_PASSWORD=secret
export SMTP_FROM=sender@example.com
export SMTP_TO=recipient@example.com
export SMTP_AUTH_TYPE=plain
export SMTP_STARTTLS=true

# Run with environment variables
smtp-edc --subject "Environment Test" --body "Hello World"
```

## Troubleshooting

### Common SMTP Issues and Solutions

#### Connection Issues

**Problem**: `Connection refused` or `Connection timeout`

**Solutions**:
1. Verify server hostname and port:
   ```bash
   # Test connectivity
   telnet smtp.example.com 587
   # or
   nc -zv smtp.example.com 587
   ```

2. Check firewall rules:
   ```bash
   # Test with debug mode
   smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Debug" --body "Test"
   ```

3. Common ports:
   - Port 25: Unencrypted SMTP (legacy)
   - Port 587: SMTP with STARTTLS (recommended)
   - Port 465: SMTP with implicit TLS
   - Port 2525: Alternative submission port

**Problem**: `TLS handshake failed`

**Solutions**:
1. Verify STARTTLS support:
   ```bash
   smtp-edc --server smtp.example.com --port 587 --starttls --debug --from test@example.com --to test@example.com --subject "TLS Test" --body "Test"
   ```

2. Skip certificate verification (testing only):
   ```bash
   smtp-edc --server smtp.example.com --port 587 --starttls --skip-verify --from test@example.com --to test@example.com --subject "Skip Verify" --body "Test"
   ```

3. Use implicit TLS for port 465:
   ```bash
   smtp-edc --server smtp.example.com --port 465 --tls --from test@example.com --to test@example.com --subject "Implicit TLS" --body "Test"
   ```

#### Authentication Issues

**Problem**: `Authentication failed` or `535 Authentication credentials invalid`

**Solutions**:
1. Verify credentials:
   ```bash
   # Test different auth methods
   smtp-edc --server smtp.example.com --port 587 --starttls --auth plain --username user@example.com --password password --debug
   smtp-edc --server smtp.example.com --port 587 --starttls --auth login --username user@example.com --password password --debug
   ```

2. Check supported authentication methods:
   ```bash
   # Debug mode shows server capabilities
   smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Auth Debug" --body "Test"
   ```

3. Gmail specific (use App Passwords):
   ```bash
   smtp-edc --server smtp.gmail.com --port 587 --starttls --auth plain --username your-email@gmail.com --password your-app-password --from your-email@gmail.com --to test@example.com --subject "Gmail Test" --body "Test"
   ```

4. Office 365 specific:
   ```bash
   smtp-edc --server smtp.office365.com --port 587 --starttls --auth login --username your-email@company.com --password password --from your-email@company.com --to test@example.com --subject "O365 Test" --body "Test"
   ```

#### Message Delivery Issues

**Problem**: `550 Relay not permitted` or `550 Access denied`

**Solutions**:
1. Ensure sender domain matches authentication:
   ```bash
   smtp-edc --server smtp.example.com --port 587 --starttls --auth plain --username user@example.com --password password --from user@example.com --to test@example.com --subject "Relay Test" --body "Test"
   ```

2. Check SPF records:
   ```bash
   dig TXT example.com | grep -i spf
   ```

**Problem**: `552 Message size exceeds maximum permitted`

**Solutions**:
1. Check attachment sizes:
   ```bash
   ls -lh /path/to/attachment.pdf
   ```

2. Use smaller attachments or compress files:
   ```bash
   gzip /path/to/large-file.txt
   smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "Compressed" --body "Compressed attachment" --attach /path/to/large-file.txt.gz
   ```

### Diagnostic Commands

#### Network Connectivity
```bash
# Basic connectivity
ping smtp.example.com
telnet smtp.example.com 587
nc -zv smtp.example.com 587

# DNS resolution
nslookup smtp.example.com
dig smtp.example.com
dig MX example.com

# SSL/TLS testing
openssl s_client -connect smtp.example.com:587 -starttls smtp
openssl s_client -connect smtp.example.com:465
```

#### SMTP Server Capabilities
```bash
# Check server capabilities
smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Capabilities" --body "Test" 2>&1 | grep -E "(EHLO|250-)"
```

#### Performance Testing
```bash
# Connection timing
time smtp-edc --server smtp.example.com --port 587 --from test@example.com --to test@example.com --subject "Performance" --body "Test"

# Multiple recipients performance
smtp-edc --server smtp.example.com --port 587 --from test@example.com --to "user1@example.com,user2@example.com,user3@example.com" --subject "Bulk Test" --body "Bulk test message"
```

### Debug Output Analysis

When using `--debug` flag, SMTP EDC shows the complete SMTP conversation:

```
2024/01/01 10:00:00 Connecting to smtp.example.com:587
2024/01/01 10:00:00 Connected successfully
2024/01/01 10:00:00 -> 220 smtp.example.com ESMTP ready
2024/01/01 10:00:01 <- EHLO localhost
2024/01/01 10:00:01 -> 250-smtp.example.com Hello
2024/01/01 10:00:01 -> 250-STARTTLS
2024/01/01 10:00:01 -> 250-AUTH PLAIN LOGIN CRAM-MD5
2024/01/01 10:00:01 -> 250 8BITMIME
2024/01/01 10:00:01 <- STARTTLS
2024/01/01 10:00:01 -> 220 Ready to start TLS
2024/01/01 10:00:01 TLS handshake successful
2024/01/01 10:00:01 <- EHLO localhost
2024/01/01 10:00:01 -> 250-smtp.example.com Hello
2024/01/01 10:00:01 -> 250-AUTH PLAIN LOGIN CRAM-MD5
2024/01/01 10:00:01 -> 250 8BITMIME
2024/01/01 10:00:01 <- AUTH PLAIN AHVzZXJAZXhhbXBsZS5jb20AcGFzc3dvcmQ=
2024/01/01 10:00:02 -> 235 Authentication successful
```

Key indicators:
- `220` - Service ready
- `250` - Requested action completed
- `235` - Authentication successful
- `354` - Start mail input
- `550` - Access denied
- `535` - Authentication credentials invalid

## Deployment

### Docker Deployment

#### Basic Docker Usage

```dockerfile
# Dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY smtp-edc .
CMD ["./smtp-edc"]
```

```bash
# Build and run
docker build -t smtp-edc .
docker run -it smtp-edc --server smtp.example.com --from sender@example.com --to recipient@example.com --subject "Docker Test" --body "Hello from Docker"
```

#### Docker Compose for SMTP Testing Service

```yaml
# docker-compose.yml
version: '3.8'
services:
  smtp-edc:
    build: .
    environment:
      - SMTP_SERVER=smtp.example.com
      - SMTP_PORT=587
      - SMTP_USERNAME=user@example.com
      - SMTP_PASSWORD=password
      - SMTP_FROM=sender@example.com
      - SMTP_STARTTLS=true
    volumes:
      - ./config:/config
      - ./templates:/templates
      - ./logs:/logs
    command: >
      --config /config/smtp-edc.yaml
      --to recipient@example.com
      --subject "Automated Test"
      --body "This is an automated test"
```

```bash
# Deploy with Docker Compose
docker-compose up -d
```

### DigitalOcean Deployment

#### DigitalOcean App Platform

```yaml
# .do/app.yaml
name: smtp-edc-service
services:
- name: smtp-edc
  source_dir: /
  github:
    repo: your-username/smtp-edc-deployment
    branch: main
  run_command: ./smtp-edc --gui --port 8080
  environment_slug: go
  instance_count: 1
  instance_size: basic-xxs
  envs:
  - key: SMTP_SERVER
    value: smtp.example.com
  - key: SMTP_PORT
    value: 587
  - key: SMTP_USERNAME
    value: user@example.com
  - key: SMTP_PASSWORD
    type: SECRET
    value: your-password
```

#### DigitalOcean Droplet Deployment

```bash
# Create droplet
doctl compute droplet create smtp-edc-server \
  --image ubuntu-20-04-x64 \
  --size s-1vcpu-1gb \
  --region nyc3 \
  --ssh-keys your-ssh-key-id

# SSH and install
ssh root@your-droplet-ip
curl -LO https://github.com/asachs01/smtp-edc/releases/latest/download/smtp-edc-linux-amd64
chmod +x smtp-edc-linux-amd64
mv smtp-edc-linux-amd64 /usr/local/bin/smtp-edc

# Create systemd service
cat > /etc/systemd/system/smtp-edc.service << EOF
[Unit]
Description=SMTP EDC Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/smtp-edc --gui --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl enable smtp-edc
systemctl start smtp-edc
```

### Cloudflare Deployment

#### Cloudflare Workers

```javascript
// worker.js
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === '/smtp-test') {
      const body = await request.json();

      // SMTP testing logic here
      const result = await testSMTP(body);

      return Response.json(result);
    }

    return new Response('SMTP EDC Worker', { status: 200 });
  }
};

async function testSMTP(config) {
  // Implement SMTP testing logic
  return {
    success: true,
    message: 'SMTP test completed'
  };
}
```

```bash
# Deploy to Cloudflare Workers
npm install -g wrangler
wrangler login
wrangler deploy
```

#### Cloudflare Pages with Functions

```javascript
// functions/api/smtp-test.js
export async function onRequestPost(context) {
  const { request } = context;
  const body = await request.json();

  // SMTP testing logic
  const result = await testSMTPConnection(body);

  return Response.json(result);
}
```

### Kubernetes Deployment

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smtp-edc-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: smtp-edc
  template:
    metadata:
      labels:
        app: smtp-edc
    spec:
      containers:
      - name: smtp-edc
        image: smtp-edc:latest
        ports:
        - containerPort: 8080
        env:
        - name: SMTP_SERVER
          valueFrom:
            configMapKeyRef:
              name: smtp-config
              key: server
        - name: SMTP_USERNAME
          valueFrom:
            secretKeyRef:
              name: smtp-secret
              key: username
        - name: SMTP_PASSWORD
          valueFrom:
            secretKeyRef:
              name: smtp-secret
              key: password
---
apiVersion: v1
kind: Service
metadata:
  name: smtp-edc-service
spec:
  selector:
    app: smtp-edc
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
```

## API Documentation

### REST API Endpoints

SMTP EDC can be deployed as a web service providing REST API endpoints for SMTP testing.

#### Base URL
```
https://your-domain.com/api/v1
```

#### Authentication
```bash
# API Key authentication
curl -H "X-API-Key: your-api-key" https://your-domain.com/api/v1/test
```

#### Endpoints

##### POST /smtp/test
Test SMTP connection and send email.

**Request Body:**
```json
{
  "server": {
    "host": "smtp.example.com",
    "port": 587,
    "tls": {
      "enabled": true,
      "start_tls": true,
      "verify": true
    }
  },
  "auth": {
    "type": "plain",
    "username": "user@example.com",
    "password": "password"
  },
  "message": {
    "from": "sender@example.com",
    "to": ["recipient@example.com"],
    "cc": ["cc@example.com"],
    "bcc": ["bcc@example.com"],
    "subject": "Test Email",
    "body": "Hello World",
    "html": "<h1>Hello World</h1>",
    "attachments": [
      {
        "filename": "document.pdf",
        "content": "base64-encoded-content",
        "content_type": "application/pdf"
      }
    ],
    "headers": {
      "X-Custom-Header": "Custom Value"
    }
  },
  "options": {
    "debug": false,
    "timeout": 30,
    "retries": 3
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Email sent successfully",
  "details": {
    "message_id": "20240101120000.ABC123@smtp.example.com",
    "recipients_accepted": ["recipient@example.com"],
    "recipients_rejected": [],
    "connection_time": 1.234,
    "send_time": 0.567,
    "total_time": 1.801
  },
  "transaction_log": [
    "2024-01-01 12:00:00 -> 220 smtp.example.com ESMTP ready",
    "2024-01-01 12:00:00 <- EHLO localhost",
    "2024-01-01 12:00:00 -> 250-smtp.example.com Hello"
  ]
}
```

##### GET /smtp/validate
Validate email address and check MX records.

**Query Parameters:**
- `email`: Email address to validate
- `check_mx`: Check MX records (true/false)

**Example:**
```bash
curl "https://your-domain.com/api/v1/smtp/validate?email=user@example.com&check_mx=true"
```

**Response:**
```json
{
  "valid": true,
  "email": "user@example.com",
  "domain": "example.com",
  "mx_records": [
    {
      "host": "mail.example.com",
      "priority": 10
    }
  ],
  "checks": {
    "syntax": true,
    "domain": true,
    "mx_record": true,
    "smtp_connection": true
  }
}
```

##### GET /smtp/capabilities
Get SMTP server capabilities.

**Query Parameters:**
- `host`: SMTP server hostname
- `port`: SMTP server port

**Example:**
```bash
curl "https://your-domain.com/api/v1/smtp/capabilities?host=smtp.example.com&port=587"
```

**Response:**
```json
{
  "host": "smtp.example.com",
  "port": 587,
  "capabilities": {
    "starttls": true,
    "auth_methods": ["PLAIN", "LOGIN", "CRAM-MD5"],
    "max_message_size": 26214400,
    "features": ["8BITMIME", "PIPELINING", "DSN"]
  },
  "connection_test": {
    "success": true,
    "connection_time": 0.234,
    "tls_version": "TLSv1.3",
    "cipher_suite": "TLS_AES_256_GCM_SHA384"
  }
}
```

### SDK and Client Libraries

#### Go Client
```go
package main

import (
    "context"
    "fmt"
    "github.com/asachs/smtp-edc/pkg/client"
)

func main() {
    client := client.NewSMTPEDCClient("https://your-domain.com/api/v1")

    config := &client.TestConfig{
        Server: client.ServerConfig{
            Host: "smtp.example.com",
            Port: 587,
            TLS: client.TLSConfig{
                Enabled:  true,
                StartTLS: true,
                Verify:   true,
            },
        },
        Auth: client.AuthConfig{
            Type:     "plain",
            Username: "user@example.com",
            Password: "password",
        },
        Message: client.MessageConfig{
            From:    "sender@example.com",
            To:      []string{"recipient@example.com"},
            Subject: "Test Email",
            Body:    "Hello World",
        },
    }

    result, err := client.TestSMTP(context.Background(), config)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Success: %v\n", result.Success)
    fmt.Printf("Message: %s\n", result.Message)
}
```

#### Python Client
```python
import requests
import json

class SMTPEDCClient:
    def __init__(self, base_url, api_key=None):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        if api_key:
            self.session.headers.update({'X-API-Key': api_key})

    def test_smtp(self, config):
        url = f"{self.base_url}/smtp/test"
        response = self.session.post(url, json=config)
        return response.json()

    def validate_email(self, email, check_mx=True):
        url = f"{self.base_url}/smtp/validate"
        params = {'email': email, 'check_mx': check_mx}
        response = self.session.get(url, params=params)
        return response.json()

# Usage example
client = SMTPEDCClient("https://your-domain.com/api/v1", "your-api-key")

config = {
    "server": {
        "host": "smtp.example.com",
        "port": 587,
        "tls": {"enabled": True, "start_tls": True, "verify": True}
    },
    "auth": {
        "type": "plain",
        "username": "user@example.com",
        "password": "password"
    },
    "message": {
        "from": "sender@example.com",
        "to": ["recipient@example.com"],
        "subject": "Test Email",
        "body": "Hello World"
    }
}

result = client.test_smtp(config)
print(f"Success: {result['success']}")
print(f"Message: {result['message']}")
```

#### JavaScript/Node.js Client
```javascript
const axios = require('axios');

class SMTPEDCClient {
    constructor(baseURL, apiKey) {
        this.client = axios.create({
            baseURL: baseURL,
            headers: apiKey ? { 'X-API-Key': apiKey } : {}
        });
    }

    async testSMTP(config) {
        try {
            const response = await this.client.post('/smtp/test', config);
            return response.data;
        } catch (error) {
            throw new Error(`SMTP test failed: ${error.response?.data?.message || error.message}`);
        }
    }

    async validateEmail(email, checkMX = true) {
        try {
            const response = await this.client.get('/smtp/validate', {
                params: { email, check_mx: checkMX }
            });
            return response.data;
        } catch (error) {
            throw new Error(`Email validation failed: ${error.response?.data?.message || error.message}`);
        }
    }
}

// Usage example
const client = new SMTPEDCClient('https://your-domain.com/api/v1', 'your-api-key');

const config = {
    server: {
        host: 'smtp.example.com',
        port: 587,
        tls: { enabled: true, start_tls: true, verify: true }
    },
    auth: {
        type: 'plain',
        username: 'user@example.com',
        password: 'password'
    },
    message: {
        from: 'sender@example.com',
        to: ['recipient@example.com'],
        subject: 'Test Email',
        body: 'Hello World'
    }
};

(async () => {
    try {
        const result = await client.testSMTP(config);
        console.log(`Success: ${result.success}`);
        console.log(`Message: ${result.message}`);
    } catch (error) {
        console.error('Error:', error.message);
    }
})();
```

## Contributing

We welcome contributions to SMTP EDC! Here's how you can help:

### Development Setup

1. **Fork and Clone**
```bash
git clone https://github.com/your-username/smtp-edc.git
cd smtp-edc
```

2. **Install Dependencies**
```bash
go mod download
```

3. **Run Tests**
```bash
go test ./...
go test -race ./...
go test -bench=. ./...
```

4. **Build and Test**
```bash
# Build CLI
go build -o smtp-edc cmd/smtp-edc/main.go

# Build GUI
go build -o smtp-edc-gui cmd/smtp-edc-gui/main.go

# Run tests
./scripts/test_basic.sh
./scripts/test_auth.sh
./scripts/test_gui.sh
```

### Code Style Guidelines

- Follow Go best practices and `gofmt` formatting
- Write comprehensive tests for new features
- Use meaningful variable and function names
- Add documentation for public APIs
- Include examples in documentation

### Submitting Changes

1. Create a feature branch: `git checkout -b feature/amazing-feature`
2. Make your changes and add tests
3. Run the test suite: `go test ./...`
4. Commit your changes: `git commit -m 'Add amazing feature'`
5. Push to the branch: `git push origin feature/amazing-feature`
6. Open a Pull Request

### Reporting Issues

When reporting issues, please include:

- SMTP EDC version: `smtp-edc --version`
- Operating system and version
- Go version (if building from source)
- Complete command used
- Error messages and debug output
- Steps to reproduce the issue

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## Support and Community

- **Documentation**: [GitHub Wiki](https://github.com/asachs01/smtp-edc/wiki)
- **Issues**: [GitHub Issues](https://github.com/asachs01/smtp-edc/issues)
- **Discussions**: [GitHub Discussions](https://github.com/asachs01/smtp-edc/discussions)
- **Email**: [smtp-edc@example.com](mailto:smtp-edc@example.com)

## Acknowledgments

- Inspired by the excellent [Swaks](https://jetmore.org/john/code/swaks/) tool
- Built with the powerful [Go](https://golang.org/) programming language
- GUI powered by [Fyne](https://fyne.io/)
- Community contributions and feedback

---

Made with ❤️ by the SMTP EDC team and contributors.
