# SMTP EDC Troubleshooting Guide

## Table of Contents

- [Quick Diagnostic Checklist](#quick-diagnostic-checklist)
- [Connection Issues](#connection-issues)
- [Authentication Problems](#authentication-problems)
- [TLS/SSL Issues](#tlsssl-issues)
- [Message Delivery Problems](#message-delivery-problems)
- [Performance Issues](#performance-issues)
- [Provider-Specific Guides](#provider-specific-guides)
- [Advanced Diagnostics](#advanced-diagnostics)
- [Common Error Codes](#common-error-codes)
- [Natural Language Troubleshooting](#natural-language-troubleshooting)

## Quick Diagnostic Checklist

Before diving into specific issues, run this quick diagnostic checklist:

### 1. Basic Connectivity Test
```bash
# Test basic connection
smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Connectivity Test" --body "Basic test"
```

### 2. Network Connectivity
```bash
# Test network connectivity
ping smtp.example.com
telnet smtp.example.com 587
nc -zv smtp.example.com 587
```

### 3. DNS Resolution
```bash
# Check DNS resolution
nslookup smtp.example.com
dig smtp.example.com
dig MX example.com
```

### 4. Server Capabilities
```bash
# Check server capabilities
smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Capabilities" --body "Test" 2>&1 | grep -E "(EHLO|250-)"
```

## Connection Issues

### Problem: "Connection refused" or "Connection timeout"

#### Symptoms
- Cannot establish connection to SMTP server
- Connection attempts timeout
- "Connection refused" error messages

#### Diagnostic Steps

1. **Verify Server and Port**
```bash
# Test different common SMTP ports
smtp-edc --server smtp.example.com --port 25 --debug --from test@example.com --to test@example.com --subject "Port 25" --body "Test"
smtp-edc --server smtp.example.com --port 587 --debug --from test@example.com --to test@example.com --subject "Port 587" --body "Test"
smtp-edc --server smtp.example.com --port 465 --debug --from test@example.com --to test@example.com --subject "Port 465" --body "Test"
smtp-edc --server smtp.example.com --port 2525 --debug --from test@example.com --to test@example.com --subject "Port 2525" --body "Test"
```

2. **Network Connectivity Test**
```bash
# Basic network tests
ping smtp.example.com
traceroute smtp.example.com
nmap -p 25,587,465,2525 smtp.example.com
```

3. **Firewall Check**
```bash
# Test with different timeout values
smtp-edc --server smtp.example.com --port 587 --timeout 60 --debug --from test@example.com --to test@example.com --subject "Extended Timeout" --body "Test"
```

#### Common Solutions

| Issue | Solution | Command |
|-------|----------|---------|
| Wrong port | Try standard SMTP ports | See port testing commands above |
| Firewall blocking | Contact IT/Network admin | - |
| Server down | Check server status | `ping smtp.example.com` |
| DNS issues | Use IP address | `smtp-edc --server 192.168.1.100 --port 587...` |

#### Port Reference

| Port | Protocol | Usage | Security |
|------|----------|-------|----------|
| 25 | SMTP | Legacy/Server-to-server | Usually unencrypted |
| 587 | SMTP | Submission (recommended) | STARTTLS |
| 465 | SMTPS | Legacy secure SMTP | Implicit TLS |
| 2525 | SMTP | Alternative submission | Often STARTTLS |

### Problem: "Network unreachable" or DNS errors

#### Diagnostic Steps
```bash
# DNS troubleshooting
nslookup smtp.example.com
dig smtp.example.com A
dig smtp.example.com MX
host smtp.example.com

# Try with different DNS servers
nslookup smtp.example.com 8.8.8.8
nslookup smtp.example.com 1.1.1.1
```

#### Solutions
1. **DNS Configuration Issues**
   - Check local DNS settings
   - Try different DNS servers
   - Use IP address directly if DNS fails

2. **Network Configuration**
   - Check routing tables
   - Verify VPN/proxy settings
   - Test from different network

## Authentication Problems

### Problem: "Authentication failed" (535 Error)

#### Symptoms
- "535 Authentication credentials invalid"
- "535 Incorrect authentication data"
- "535 Login Failure"

#### Diagnostic Steps

1. **Test Different Authentication Methods**
```bash
# Try PLAIN authentication
smtp-edc --server smtp.example.com --port 587 --starttls --auth plain --username user@example.com --password password --debug --from user@example.com --to test@example.com --subject "PLAIN Auth" --body "Test"

# Try LOGIN authentication
smtp-edc --server smtp.example.com --port 587 --starttls --auth login --username user@example.com --password password --debug --from user@example.com --to test@example.com --subject "LOGIN Auth" --body "Test"

# Try CRAM-MD5 authentication
smtp-edc --server smtp.example.com --port 587 --starttls --auth cram-md5 --username user@example.com --password password --debug --from user@example.com --to test@example.com --subject "CRAM-MD5 Auth" --body "Test"
```

2. **Check Server Capabilities**
```bash
# See what authentication methods are supported
smtp-edc --debug --server smtp.example.com --port 587 --from test@example.com --to test@example.com --subject "Auth Check" --body "Test" 2>&1 | grep -i auth
```

#### Common Solutions by Provider

##### Gmail
```bash
# Use App Password (not regular password)
smtp-edc --server smtp.gmail.com --port 587 --starttls --auth plain --username your-email@gmail.com --password your-16-digit-app-password --from your-email@gmail.com --to test@example.com --subject "Gmail Test" --body "Test"
```

**Setup Steps for Gmail:**
1. Enable 2-factor authentication
2. Generate App Password: Google Account > Security > 2-Step Verification > App passwords
3. Use the 16-digit app password

##### Office 365
```bash
# Often requires LOGIN method
smtp-edc --server smtp.office365.com --port 587 --starttls --auth login --username your-email@company.com --password password --from your-email@company.com --to test@example.com --subject "O365 Test" --body "Test"
```

##### Yahoo Mail
```bash
# Use App Password
smtp-edc --server smtp.mail.yahoo.com --port 587 --starttls --auth plain --username your-email@yahoo.com --password app-password --from your-email@yahoo.com --to test@example.com --subject "Yahoo Test" --body "Test"
```

#### Troubleshooting Checklist

- [ ] Credentials are correct (test login via webmail)
- [ ] Account is not locked or suspended
- [ ] Two-factor authentication is properly configured
- [ ] App-specific passwords are used where required
- [ ] SMTP access is enabled for the account
- [ ] Correct authentication method is used
- [ ] Username format is correct (email vs. username)

### Problem: OAuth Authentication Issues

#### Modern Authentication (OAuth2)
```bash
# For services supporting OAuth2
smtp-edc --server smtp.office365.com --port 587 --starttls --auth oauth2 --oauth-token "your-oauth-token" --from user@company.com --to test@example.com --subject "OAuth Test" --body "Test"
```

#### Getting OAuth Tokens
1. **Office 365/Azure AD**
   - Register application in Azure AD
   - Request appropriate scopes (Mail.Send)
   - Use token endpoint to get access token

2. **Gmail OAuth**
   - Set up OAuth 2.0 in Google Cloud Console
   - Request https://mail.google.com/ scope
   - Exchange authorization code for tokens

## TLS/SSL Issues

### Problem: "TLS handshake failed" or certificate errors

#### Symptoms
- "x509: certificate signed by unknown authority"
- "TLS handshake timeout"
- "certificate verify failed"

#### Diagnostic Steps

1. **Test TLS Connection**
```bash
# Test STARTTLS
smtp-edc --server smtp.example.com --port 587 --starttls --debug --from test@example.com --to test@example.com --subject "STARTTLS Test" --body "Test"

# Test implicit TLS
smtp-edc --server smtp.example.com --port 465 --tls --debug --from test@example.com --to test@example.com --subject "TLS Test" --body "Test"

# Skip certificate verification (testing only)
smtp-edc --server smtp.example.com --port 587 --starttls --skip-verify --debug --from test@example.com --to test@example.com --subject "Skip Verify" --body "Test"
```

2. **Check Certificate Details**
```bash
# Check certificate with OpenSSL
openssl s_client -connect smtp.example.com:587 -starttls smtp -servername smtp.example.com
openssl s_client -connect smtp.example.com:465 -servername smtp.example.com

# Check certificate expiration
echo | openssl s_client -connect smtp.example.com:587 -starttls smtp 2>/dev/null | openssl x509 -noout -dates
```

#### Solutions

1. **Certificate Trust Issues**
```bash
# Update certificate store (Linux)
sudo apt-get update && sudo apt-get install ca-certificates
sudo update-ca-certificates

# Update certificate store (macOS)
brew install ca-certificates

# For testing, skip verification (NOT for production)
smtp-edc --server smtp.example.com --port 587 --starttls --skip-verify --from test@example.com --to test@example.com --subject "Test" --body "Test"
```

2. **TLS Version Issues**
```bash
# Some servers require specific TLS versions
openssl s_client -connect smtp.example.com:587 -starttls smtp -tls1_2
openssl s_client -connect smtp.example.com:587 -starttls smtp -tls1_3
```

### Problem: "STARTTLS not available"

#### Diagnostic Steps
```bash
# Check if server supports STARTTLS
telnet smtp.example.com 587
# Type: EHLO test.com
# Look for "250-STARTTLS" in response
```

#### Solutions
- Use implicit TLS on port 465
- Use unencrypted connection on port 25 (not recommended)
- Contact server administrator

## Message Delivery Problems

### Problem: "550 Relay not permitted" or "550 Access denied"

#### Symptoms
- Messages rejected by server
- "550 5.7.1 Unable to relay"
- "550 Access denied"

#### Diagnostic Steps
```bash
# Test with authenticated sender
smtp-edc --server smtp.example.com --port 587 --starttls --auth plain --username user@example.com --password password --from user@example.com --to external@example.com --subject "Relay Test" --body "Test"

# Check if sender domain matches authentication
smtp-edc --server smtp.example.com --port 587 --starttls --auth plain --username user@domain1.com --password password --from user@domain2.com --to test@example.com --subject "Domain Mismatch" --body "Test"
```

#### Solutions
1. **Authentication Required**
   - Always authenticate when sending through SMTP servers
   - Use sender address that matches authentication credentials

2. **SPF/DKIM Configuration**
   ```bash
   # Check SPF records
   dig TXT example.com | grep -i spf
   nslookup -type=TXT example.com | grep -i spf

   # Check DKIM records
   dig TXT default._domainkey.example.com
   ```

### Problem: "552 Message size exceeds maximum permitted"

#### Diagnostic Steps
```bash
# Check message size limits
smtp-edc --debug --server smtp.example.com --from test@example.com --to test@example.com --subject "Size Check" --body "Test" 2>&1 | grep -i size

# Test with large attachment
smtp-edc --server smtp.example.com --from test@example.com --to test@example.com --subject "Large File" --body "Large attachment test" --attach /path/to/large/file.pdf
```

#### Solutions
1. **Reduce Message Size**
   ```bash
   # Compress attachments
   gzip large-file.txt
   smtp-edc --server smtp.example.com --from test@example.com --to test@example.com --subject "Compressed" --body "Compressed attachment" --attach large-file.txt.gz
   ```

2. **Split Large Messages**
   - Send multiple smaller messages
   - Use file sharing services for large files
   - Contact administrator to increase size limits

### Problem: Messages appear to send but aren't received

#### Diagnostic Steps
1. **Check Server Response**
```bash
# Look for message ID in response
smtp-edc --debug --server smtp.example.com --auth plain --username user@example.com --password password --from user@example.com --to recipient@example.com --subject "Delivery Test" --body "Test" 2>&1 | grep -i "message-id\|250"
```

2. **Test with Multiple Recipients**
```bash
# Test different domains
smtp-edc --server smtp.example.com --auth plain --username user@example.com --password password --from user@example.com --to "test1@gmail.com,test2@yahoo.com,test3@outlook.com" --subject "Multi-domain Test" --body "Test"
```

#### Solutions
1. **Check Spam Folders**
2. **Verify Recipient Addresses**
3. **Check Server Logs**
4. **Test Reverse DNS**
```bash
# Check reverse DNS
dig -x IP_ADDRESS
nslookup IP_ADDRESS
```

## Performance Issues

### Problem: Slow connection or sending

#### Diagnostic Steps
```bash
# Measure connection time
time smtp-edc --server smtp.example.com --port 587 --from test@example.com --to test@example.com --subject "Performance Test" --body "Test"

# Test with different timeout values
smtp-edc --server smtp.example.com --port 587 --timeout 10 --from test@example.com --to test@example.com --subject "Short Timeout" --body "Test"
smtp-edc --server smtp.example.com --port 587 --timeout 60 --from test@example.com --to test@example.com --subject "Long Timeout" --body "Test"
```

#### Solutions
1. **Network Optimization**
   - Use geographically closer SMTP servers
   - Check for network congestion
   - Test during different times

2. **Connection Reuse**
   - Implement connection pooling for bulk sending
   - Keep connections alive for multiple messages

### Problem: High CPU or memory usage

#### Diagnostic Steps
```bash
# Monitor resource usage
top -p $(pgrep smtp-edc)
htop
ps aux | grep smtp-edc

# Test with verbose output
smtp-edc --verbose --server smtp.example.com --from test@example.com --to test@example.com --subject "Resource Test" --body "Test"
```

## Provider-Specific Guides

### Gmail SMTP Configuration

```bash
# Standard Gmail SMTP
smtp-edc --server smtp.gmail.com --port 587 --starttls --auth plain --username your-email@gmail.com --password your-app-password --from your-email@gmail.com --to recipient@example.com --subject "Gmail Test" --body "Hello from Gmail"
```

**Requirements:**
- Enable 2-factor authentication
- Generate app password
- Use app password (not regular password)

**Common Issues:**
- "Less secure app access" disabled (use app passwords)
- Account locked due to suspicious activity
- Incorrect app password format

### Office 365 SMTP Configuration

```bash
# Office 365 SMTP
smtp-edc --server smtp.office365.com --port 587 --starttls --auth login --username your-email@company.com --password password --from your-email@company.com --to recipient@example.com --subject "O365 Test" --body "Hello from Office 365"
```

**Requirements:**
- SMTP authentication enabled
- Modern authentication (OAuth2) may be required
- Correct tenant configuration

**Common Issues:**
- Modern authentication required
- Conditional access policies blocking
- Incorrect authentication method (use LOGIN)

### Yahoo Mail SMTP Configuration

```bash
# Yahoo Mail SMTP
smtp-edc --server smtp.mail.yahoo.com --port 587 --starttls --auth plain --username your-email@yahoo.com --password app-password --from your-email@yahoo.com --to recipient@example.com --subject "Yahoo Test" --body "Hello from Yahoo"
```

**Requirements:**
- Generate app password
- Enable less secure apps (deprecated)
- Use app password for authentication

### Amazon SES SMTP Configuration

```bash
# Amazon SES SMTP
smtp-edc --server email-smtp.us-east-1.amazonaws.com --port 587 --starttls --auth plain --username SMTP_USERNAME --password SMTP_PASSWORD --from verified@yourdomain.com --to recipient@example.com --subject "SES Test" --body "Hello from Amazon SES"
```

**Requirements:**
- Verify sender email/domain
- Create SMTP credentials
- Configure sending limits

### SendGrid SMTP Configuration

```bash
# SendGrid SMTP
smtp-edc --server smtp.sendgrid.net --port 587 --starttls --auth plain --username apikey --password your-api-key --from verified@yourdomain.com --to recipient@example.com --subject "SendGrid Test" --body "Hello from SendGrid"
```

**Requirements:**
- API key with mail send permissions
- Username is always "apikey"
- Password is your API key

## Advanced Diagnostics

### Network Analysis

```bash
# Packet capture for SMTP traffic
sudo tcpdump -i any -s 0 -w smtp-capture.pcap host smtp.example.com and port 587

# Analyze with Wireshark or tshark
tshark -r smtp-capture.pcap -Y "smtp" -V

# Network trace with mtr
mtr smtp.example.com
```

### SMTP Protocol Analysis

```bash
# Manual SMTP session
telnet smtp.example.com 587
```

**Manual SMTP Commands:**
```
EHLO test.com
STARTTLS
(TLS negotiation happens here)
EHLO test.com
AUTH PLAIN
(base64 encoded: \0username\0password)
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: Test

This is a test message.
.
QUIT
```

### Log Analysis

```bash
# SMTP EDC debug output analysis
smtp-edc --debug --server smtp.example.com --from test@example.com --to test@example.com --subject "Debug" --body "Test" > smtp-debug.log 2>&1

# Parse debug output
grep -E "^[0-9]{4}/[0-9]{2}/[0-9]{2}" smtp-debug.log | head -20
grep -E "(->|<-)" smtp-debug.log
grep -E "ERROR|WARN" smtp-debug.log
```

## Common Error Codes

### SMTP Response Codes

| Code | Category | Description | Common Causes |
|------|----------|-------------|---------------|
| 220 | Success | Service ready | Normal connection |
| 250 | Success | Requested action okay | Command successful |
| 354 | Intermediate | Start mail input | Ready to receive message |
| 421 | Temporary | Service not available | Server busy/temporary failure |
| 450 | Temporary | Mailbox unavailable | Temporary issue, retry later |
| 451 | Temporary | Local error | Server processing error |
| 452 | Temporary | Insufficient storage | Server storage full |
| 500 | Permanent | Syntax error | Invalid command |
| 501 | Permanent | Parameter error | Invalid parameters |
| 502 | Permanent | Command not implemented | Unsupported command |
| 503 | Permanent | Bad command sequence | Commands out of order |
| 504 | Permanent | Parameter not implemented | Unsupported parameter |
| 521 | Permanent | Domain does not accept mail | Domain blocks email |
| 530 | Permanent | Authentication required | Must authenticate first |
| 535 | Permanent | Authentication failed | Invalid credentials |
| 550 | Permanent | Mailbox unavailable | Address doesn't exist/access denied |
| 551 | Permanent | User not local | Forwarding required |
| 552 | Permanent | Storage allocation exceeded | Message too large |
| 553 | Permanent | Mailbox name invalid | Invalid address format |
| 554 | Permanent | Transaction failed | Message rejected |

### SMTP EDC Specific Errors

| Error | Description | Solution |
|-------|-------------|----------|
| `connection_timeout` | Connection timed out | Increase timeout, check network |
| `tls_handshake_failed` | TLS negotiation failed | Check certificate, try skip-verify |
| `auth_method_not_supported` | Authentication method unsupported | Try different auth method |
| `invalid_email_format` | Email address format invalid | Check email address syntax |
| `attachment_too_large` | Attachment exceeds size limit | Reduce file size or compress |
| `dns_resolution_failed` | Cannot resolve hostname | Check DNS, try IP address |

## Natural Language Troubleshooting

### "My emails aren't being sent"

**Diagnostic Approach:**
```bash
# Step 1: Test basic connectivity
smtp-edc --server smtp.yourprovider.com --port 587 --debug --from your-email@domain.com --to test@example.com --subject "Connection Test" --body "Testing connectivity"

# Step 2: Check authentication
smtp-edc --server smtp.yourprovider.com --port 587 --starttls --auth plain --username your-email@domain.com --password your-password --debug --from your-email@domain.com --to test@example.com --subject "Auth Test" --body "Testing authentication"

# Step 3: Send actual message
smtp-edc --server smtp.yourprovider.com --port 587 --starttls --auth plain --username your-email@domain.com --password your-password --from your-email@domain.com --to recipient@example.com --subject "Your Subject" --body "Your message content"
```

### "I get authentication errors"

**Diagnostic Steps:**
1. **Verify credentials work in webmail**
2. **Check for 2FA/App passwords**
3. **Test different auth methods**
4. **Verify server settings**

```bash
# Test each authentication method
for auth in plain login cram-md5; do
  echo "Testing $auth authentication..."
  smtp-edc --server smtp.example.com --port 587 --starttls --auth $auth --username user@example.com --password password --debug --from user@example.com --to test@example.com --subject "Auth Test: $auth" --body "Testing $auth authentication"
done
```

### "Emails are going to spam"

**Diagnostic Approach:**
1. **Check SPF records**
2. **Verify DKIM signing**
3. **Test reverse DNS**
4. **Review message content**

```bash
# Check email authentication records
dig TXT yourdomain.com | grep -i spf
dig TXT default._domainkey.yourdomain.com

# Test with different content
smtp-edc --server smtp.example.com --auth plain --username user@example.com --password password --from user@example.com --to test@gmail.com --subject "Plain Test" --body "This is a simple test message without any promotional content."
```

### "Connections are slow or timing out"

**Performance Optimization:**
```bash
# Test different servers/ports
for port in 25 587 465 2525; do
  echo "Testing port $port..."
  time smtp-edc --server smtp.example.com --port $port --timeout 30 --from test@example.com --to test@example.com --subject "Port $port Test" --body "Testing port $port"
done

# Test network path
mtr smtp.example.com
traceroute smtp.example.com
```

### "I can't send large attachments"

**Size Limit Testing:**
```bash
# Check server size limits
smtp-edc --debug --server smtp.example.com --from test@example.com --to test@example.com --subject "Size Limit Check" --body "Checking size limits" 2>&1 | grep -i size

# Create test files of different sizes
dd if=/dev/zero of=test-1mb.txt bs=1024 count=1024
dd if=/dev/zero of=test-5mb.txt bs=1024 count=5120
dd if=/dev/zero of=test-10mb.txt bs=1024 count=10240

# Test with each size
for file in test-1mb.txt test-5mb.txt test-10mb.txt; do
  echo "Testing with $(ls -lh $file | awk '{print $5}') file..."
  smtp-edc --server smtp.example.com --auth plain --username user@example.com --password password --from user@example.com --to test@example.com --subject "Size Test: $file" --body "Testing with $file" --attach $file
done
```

## Getting Help

### Collecting Diagnostic Information

When reporting issues, collect this information:

```bash
# System information
uname -a
smtp-edc --version

# Network connectivity
ping smtp.example.com
telnet smtp.example.com 587
nslookup smtp.example.com

# Full debug output
smtp-edc --debug --verbose --server smtp.example.com --port 587 --starttls --auth plain --username user@example.com --password [REDACTED] --from user@example.com --to test@example.com --subject "Debug Report" --body "Debug information for support" > debug-output.log 2>&1

# Sanitize the output (remove passwords)
sed 's/password [^ ]*/password [REDACTED]/g' debug-output.log > sanitized-debug.log
```

### Support Channels

- **GitHub Issues**: [Report bugs and feature requests](https://github.com/asachs01/smtp-edc/issues)
- **Documentation**: [Project Wiki](https://github.com/asachs01/smtp-edc/wiki)
- **Community**: [GitHub Discussions](https://github.com/asachs01/smtp-edc/discussions)

---

This troubleshooting guide provides systematic approaches to diagnosing and resolving common SMTP issues. For additional help, please consult the project documentation or submit an issue with detailed diagnostic information.
