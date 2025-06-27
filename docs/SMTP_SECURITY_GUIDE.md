# SMTP EDC Security Guide

This document outlines security best practices, implementation details, and usage guidelines for the SMTP EDC security features.

## Table of Contents

1. [Overview](#overview)
2. [Credential Security](#credential-security)
3. [Authentication Methods](#authentication-methods)
4. [TLS/SSL Configuration](#tlsssl-configuration)
5. [Rate Limiting and Abuse Prevention](#rate-limiting-and-abuse-prevention)
6. [Secure Logging](#secure-logging)
7. [Security Hardening](#security-hardening)
8. [Incident Response](#incident-response)
9. [Compliance Considerations](#compliance-considerations)
10. [API Reference](#api-reference)

## Overview

SMTP EDC implements defense-in-depth security principles to protect SMTP communications and credentials. The security architecture includes:

- **Encrypted Credential Storage**: AES-256-GCM encryption with PBKDF2 key derivation
- **Multi-Factor Authentication**: Support for OAuth2, CRAM-MD5, and secure basic auth
- **TLS/SSL Verification**: Comprehensive certificate validation and cipher suite analysis
- **Rate Limiting**: Token bucket algorithm with configurable thresholds
- **Secure Logging**: Sensitive data masking with structured audit trails
- **Abuse Prevention**: Account lockout and anomaly detection

## Credential Security

### Secure Storage Architecture

The credential storage system uses industry-standard encryption:

```
┌─────────────────────────────────────┐
│           User Input                │
│      (Master Password)              │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│        PBKDF2 Key Derivation        │
│    (100,000 iterations, SHA-256)    │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│         AES-256-GCM                 │
│      Encryption/Decryption          │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│        Encrypted Storage            │
│      (File with 0600 perms)         │
└─────────────────────────────────────┘
```

### Best Practices

#### 1. Master Password Security
- Use a strong, unique master password (minimum 12 characters)
- Include uppercase, lowercase, numbers, and special characters
- Consider using a password manager to generate and store the master password
- Rotate the master password periodically (every 90 days recommended)

#### 2. Credential Management
```go
// Initialize credential store
store, err := security.NewCredentialStore("/secure/path/credentials.db")
if err != nil {
    log.Fatal("Failed to initialize credential store:", err)
}

// Set master key securely
err = store.InitializeMasterKey() // Prompts for password
if err != nil {
    log.Fatal("Failed to set master key:", err)
}

// Store credentials with metadata
cred := &security.Credential{
    Username:    "user@example.com",
    Password:    "secure_password_or_token",
    AuthType:    "oauth2",
    Server:      "smtp.gmail.com",
    Port:        587,
    Description: "Gmail SMTP for notifications",
    Metadata: map[string]string{
        "access_token":  "oauth_access_token",
        "refresh_token": "oauth_refresh_token",
        "expires_at":    "2024-12-31T23:59:59Z",
    },
}

err = store.StoreCredential("gmail-notifications", cred)
if err != nil {
    log.Fatal("Failed to store credential:", err)
}
```

#### 3. Credential Rotation
```go
// Set expiration for automatic rotation reminders
store.SetExpiration("gmail-notifications", time.Now().Add(90*24*time.Hour))

// Rotate master key
err = store.RotateMasterKey("new_secure_password")
if err != nil {
    log.Fatal("Failed to rotate master key:", err)
}
```

#### 4. Integrity Validation
```go
// Validate credential store integrity
err = store.ValidateIntegrity()
if err != nil {
    log.Error("Credential store integrity check failed:", err)
    // Implement recovery procedures
}

// Clean up expired credentials
err = store.Cleanup()
if err != nil {
    log.Error("Failed to clean expired credentials:", err)
}
```

## Authentication Methods

### Supported Methods

| Method | Security Level | Use Case | TLS Required |
|--------|----------------|----------|--------------|
| OAuth2 | High | Modern email providers | Yes |
| CRAM-MD5 | Medium | Legacy systems | No |
| PLAIN | Low | Development only | Yes |
| LOGIN | Low | Legacy compatibility | Yes |

### OAuth2 Implementation

OAuth2 provides the highest security for modern email providers:

```go
// Configure OAuth2 for Gmail
oauth2Config := &auth.OAuth2Config{
    Provider:     auth.ProviderGoogle,
    ClientID:     "your-client-id.apps.googleusercontent.com",
    ClientSecret: "your-client-secret",
    Scopes:       []string{"https://mail.google.com/"},
}

// Create authenticator
authenticator := auth.NewOAuth2Authenticator(oauth2Config)

// Get authorization URL
authURL := authenticator.GetAuthorizationURL("http://localhost:8080/callback", "state123")
fmt.Printf("Visit: %s\n", authURL)

// Exchange code for token (after user authorization)
token, err := authenticator.ExchangeCodeForToken(ctx, authCode, "http://localhost:8080/callback")
if err != nil {
    log.Fatal("Token exchange failed:", err)
}

// Store tokens securely
cred.Metadata["access_token"] = token.AccessToken
cred.Metadata["refresh_token"] = token.RefreshToken
cred.Metadata["expires_at"] = token.ExpiresAt.Format(time.RFC3339)
```

### Authentication Manager

Use the authentication manager for enhanced security:

```go
// Initialize authentication manager
authManager := auth.NewAuthManager(credentialStore)

// Configure security settings
authManager.SetRateLimit(100, time.Minute)           // 100 attempts per minute
authManager.SetFailureThreshold(5, 15*time.Minute)   // 5 failures = 15 min lockout

// Create authentication context
ctx := &auth.AuthContext{
    Username:     "user@example.com",
    AuthType:     "oauth2",
    Server:       "smtp.gmail.com",
    Port:         587,
    TLSRequired:  true,
    ClientIP:     "192.168.1.100",
    UserAgent:    "SMTP-EDC/1.0",
    SessionID:    "sess_12345",
    MaxRetries:   3,
    RetryDelay:   2 * time.Second,
}

// Authenticate using stored credentials
authenticator, err := authManager.AuthenticateWithCredentialStore(ctx, "gmail-notifications")
if err != nil {
    log.Error("Authentication failed:", err)
    return
}

// Validate authentication attempt result
err = authManager.ValidateAuthAttempt(ctx, success, authError)
if err != nil {
    log.Error("Authentication validation failed:", err)
}
```

## TLS/SSL Configuration

### Secure TLS Configuration

```go
// Create secure TLS configuration
tlsConfig := security.NewTLSConfig("smtp.gmail.com")

// Security settings
tlsConfig.MinVersion = tls.VersionTLS12           // Minimum TLS 1.2
tlsConfig.MaxVersion = tls.VersionTLS13           // Prefer TLS 1.3
tlsConfig.InsecureSkipVerify = false              // Always verify certificates
tlsConfig.RequireCertificate = true              // Require valid certificate
tlsConfig.PreferServerCiphers = false            // Client cipher preference
tlsConfig.CipherSuites = []uint16{               // Secure cipher suites only
    tls.TLS_AES_256_GCM_SHA384,
    tls.TLS_CHACHA20_POLY1305_SHA256,
    tls.TLS_AES_128_GCM_SHA256,
    tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
}

// Advanced security features
tlsConfig.MaxCertChainLength = 5                  // Limit certificate chain length
tlsConfig.AllowedSANs = []string{                // Restrict allowed SANs
    "smtp.gmail.com",
    "*.gmail.com",
}

// Certificate pinning (optional but recommended)
tlsConfig.PinCertificates = []string{            // SHA-256 fingerprints
    "a1b2c3d4e5f6...",  // Gmail SMTP certificate fingerprint
}
```

### TLS Verification and Troubleshooting

```go
// Create TLS verifier
verifier := security.NewTLSVerifier(tlsConfig)

// Test direct TLS connection
tlsInfo, err := verifier.VerifyTLSConnection("smtp.gmail.com", 465)
if err != nil {
    log.Error("TLS verification failed:", err)
    return
}

// Test SMTP STARTTLS
tlsInfo, err = verifier.TestSMTPSTARTTLS("smtp.gmail.com", 587)
if err != nil {
    log.Error("STARTTLS test failed:", err)
    return
}

// Analyze TLS security
fmt.Printf("TLS Version: %s\n", tlsInfo.Version)
fmt.Printf("Cipher Suite: %s\n", tlsInfo.CipherSuite)
fmt.Printf("Security Score: %d/100\n", tlsInfo.SecurityScore)
fmt.Printf("Certificate Valid: %v\n", tlsInfo.PeerCertificatesValid)

// Show security recommendations
for _, recommendation := range tlsInfo.Recommendations {
    fmt.Printf("RECOMMENDATION: %s\n", recommendation)
}

// Display certificate information
for i, cert := range tlsInfo.ServerCertificates {
    fmt.Printf("Certificate %d:\n", i)
    fmt.Printf("  Subject: %s\n", cert.Subject)
    fmt.Printf("  Issuer: %s\n", cert.Issuer)
    fmt.Printf("  Valid From: %s\n", cert.NotBefore)
    fmt.Printf("  Valid Until: %s\n", cert.NotAfter)
    fmt.Printf("  Key Algorithm: %s\n", cert.PublicKeyAlgorithm)
}
```

### Common TLS Issues and Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| Certificate not trusted | Self-signed or unknown CA | Add CA certificate to trust store |
| Hostname mismatch | Certificate CN/SAN doesn't match | Use correct hostname or update certificate |
| Weak cipher suite | Server using insecure ciphers | Update server configuration or use different server |
| TLS version too old | Server only supports TLS 1.0/1.1 | Upgrade server or temporarily lower MinVersion for testing |
| Certificate expired | Certificate past expiration date | Renew certificate |

## Rate Limiting and Abuse Prevention

### Rate Limiting Configuration

```go
// Create rate limiter
rateLimiter := auth.NewRateLimiter(100, time.Minute) // 100 attempts per minute

// Check if request is allowed
if !rateLimiter.Allow("192.168.1.100") {
    log.Warn("Rate limit exceeded for IP: 192.168.1.100")
    return fmt.Errorf("rate limit exceeded")
}

// Get remaining attempts
remaining := rateLimiter.GetRemainingAttempts("192.168.1.100")
resetTime := rateLimiter.GetResetTime("192.168.1.100")

fmt.Printf("Remaining attempts: %d\n", remaining)
fmt.Printf("Reset time: %s\n", resetTime)

// Administrative functions
rateLimiter.Reset("192.168.1.100")        // Reset specific IP
rateLimiter.Clear("192.168.1.100")        // Clear specific IP data
rateLimiter.ClearAll()                    // Clear all rate limit data
```

### Authentication Failure Handling

```go
// Configure failure thresholds
authManager.SetFailureThreshold(5, 15*time.Minute) // 5 failures = 15 min lockout

// Check if account is locked
failureInfo := authManager.GetFailureInfo("user@example.com", "smtp.gmail.com")
if failureInfo != nil {
    fmt.Printf("Failure count: %d\n", failureInfo.Count)
    fmt.Printf("Last failure: %s\n", failureInfo.LastFailure)
    fmt.Printf("Locked until: %s\n", failureInfo.LockedUntil)
}

// Administrative unlock
authManager.ClearFailures("user@example.com", "smtp.gmail.com")
```

### Monitoring and Alerting

```go
// Get authentication statistics
stats := authManager.GetAuthStats()
fmt.Printf("Total failed accounts: %v\n", stats["total_failed_accounts"])
fmt.Printf("Total failures: %v\n", stats["total_failures"])
fmt.Printf("Currently locked: %v\n", stats["currently_locked"])

// Monitor rate limiting
rateLimitStats := rateLimiter.GetStats()
fmt.Printf("Active buckets: %v\n", rateLimitStats["active_buckets"])
fmt.Printf("Empty buckets: %v\n", rateLimitStats["empty_buckets"])

// Implement alerting thresholds
if stats["currently_locked"].(int) > 10 {
    // Send alert - too many locked accounts
    sendSecurityAlert("High number of locked accounts", stats)
}
```

## Secure Logging

### Logger Configuration

```go
// Configure secure logger
logConfig := &security.LogConfig{
    Level:         security.LogInfo,
    MaskSensitive: true,              // Enable sensitive data masking
    AuditMode:     true,              // Enable audit logging
    MaxLogSize:    100 * 1024 * 1024, // 100MB max log size
    RetentionDays: 90,                // Keep logs for 90 days
    RotateDaily:   true,              // Rotate logs daily
    LogPath:       "/var/log/smtp-edc",
}

logger, err := security.NewSecureLogger(logConfig)
if err != nil {
    log.Fatal("Failed to initialize secure logger:", err)
}
defer logger.Close()
```

### Logging Best Practices

```go
// Use structured logging with context
metadata := map[string]interface{}{
    "session_id": "sess_12345",
    "user_id":    "user@example.com",
    "remote_ip":  "192.168.1.100",
    "server":     "smtp.gmail.com",
}

// Authentication events
logger.AuthLog("user@example.com", "login_attempt", "success", metadata)
logger.AuthLog("user@example.com", "password_change", "success", metadata)

// Security events
logger.SecurityLog("suspicious_activity", "medium", "Multiple failed logins", metadata)
logger.SecurityLog("rate_limit_exceeded", "high", "100+ requests in 1 minute", metadata)

// Connection events
logger.ConnectionLog("connect", "smtp.gmail.com", 587, true, metadata)
logger.TLSLog("handshake", "smtp.gmail.com", "TLS 1.3", "TLS_AES_256_GCM_SHA384", metadata)

// Application events
logger.Info("smtp-client", "Message sent successfully", metadata)
logger.Error("smtp-client", "Failed to send message: connection timeout", metadata)

// Audit events (always logged)
logger.Audit("credential-store", "Credential accessed", metadata)
logger.Audit("configuration", "TLS settings updated", metadata)
```

### Custom Sensitive Data Patterns

```go
// Add custom sensitive data patterns
err = logger.AddSensitivePattern(
    "credit_card",
    `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,
    "****-****-****-****",
    "financial",
)

err = logger.AddSensitivePattern(
    "social_security",
    `\b\d{3}-\d{2}-\d{4}\b`,
    "***-**-****",
    "pii",
)

err = logger.AddSensitivePattern(
    "api_key",
    `sk-[a-zA-Z0-9]{48}`,
    "sk-***MASKED***",
    "credentials",
)
```

## Security Hardening

### File System Security

```bash
# Set proper permissions for credential store
chmod 600 /path/to/credentials.db
chown smtp-edc:smtp-edc /path/to/credentials.db

# Set proper permissions for log directory
chmod 750 /var/log/smtp-edc
chown smtp-edc:smtp-edc /var/log/smtp-edc

# Set proper permissions for configuration files
chmod 640 /etc/smtp-edc/config.yaml
chown root:smtp-edc /etc/smtp-edc/config.yaml
```

### Network Security

```yaml
# Example firewall rules (iptables)
# Allow outbound SMTP connections
-A OUTPUT -p tcp --dport 25 -j ACCEPT
-A OUTPUT -p tcp --dport 587 -j ACCEPT
-A OUTPUT -p tcp --dport 465 -j ACCEPT

# Allow outbound DNS
-A OUTPUT -p udp --dport 53 -j ACCEPT
-A OUTPUT -p tcp --dport 53 -j ACCEPT

# Allow outbound HTTPS for OAuth2
-A OUTPUT -p tcp --dport 443 -j ACCEPT

# Block all other outbound connections
-A OUTPUT -j DROP
```

### Runtime Security

```go
// Implement security checks at runtime
func securityChecks() error {
    // Check file permissions
    if err := checkFilePermissions(); err != nil {
        return fmt.Errorf("file permission check failed: %v", err)
    }

    // Check for suspicious processes
    if err := checkProcesses(); err != nil {
        return fmt.Errorf("process check failed: %v", err)
    }

    // Validate system time (for certificate validation)
    if err := checkSystemTime(); err != nil {
        return fmt.Errorf("system time check failed: %v", err)
    }

    return nil
}

// Memory protection
func protectMemory() {
    // Clear sensitive data from memory after use
    defer func() {
        // Zero out sensitive variables
        for i := range password {
            password[i] = 0
        }
        for i := range apiKey {
            apiKey[i] = 0
        }
    }()
}
```

## Incident Response

### Security Event Detection

```go
// Monitor for security events
func monitorSecurityEvents(logger *security.SecureLogger) {
    // Monitor authentication failures
    go func() {
        for {
            stats := authManager.GetAuthStats()
            if stats["currently_locked"].(int) > 10 {
                logger.SecurityLog("mass_lockout", "critical",
                    "More than 10 accounts locked", nil)
                sendAlert("Mass account lockout detected")
            }
            time.Sleep(5 * time.Minute)
        }
    }()

    // Monitor rate limiting
    go func() {
        for {
            stats := rateLimiter.GetStats()
            if stats["empty_buckets"].(int) > 100 {
                logger.SecurityLog("rate_limit_abuse", "high",
                    "High number of rate-limited IPs", nil)
                sendAlert("Potential DDoS attack detected")
            }
            time.Sleep(1 * time.Minute)
        }
    }()
}
```

### Incident Response Procedures

1. **Immediate Response**
   - Isolate affected systems
   - Preserve logs and evidence
   - Notify security team

2. **Analysis**
   - Review audit logs
   - Identify attack vectors
   - Assess scope of compromise

3. **Containment**
   - Block malicious IPs
   - Disable compromised accounts
   - Rotate affected credentials

4. **Recovery**
   - Restore from secure backups
   - Update security configurations
   - Re-enable services with monitoring

5. **Post-Incident**
   - Document lessons learned
   - Update security procedures
   - Conduct security review

### Forensic Log Analysis

```bash
# Search for authentication failures
grep "AUTH.*FAILED" /var/log/smtp-edc/*.log | head -20

# Find rate limiting events
grep "rate_limit_exceeded" /var/log/smtp-edc/*.log | wc -l

# Analyze IP patterns
grep "ip=" /var/log/smtp-edc/*.log | cut -d'=' -f3 | cut -d' ' -f1 | sort | uniq -c | sort -nr

# Find TLS errors
grep "TLS.*ERROR" /var/log/smtp-edc/*.log | tail -10

# Search for specific user activity
grep "user=user@example.com" /var/log/smtp-edc/*.log | tail -20
```

## Compliance Considerations

### Data Protection Requirements

#### GDPR Compliance
- **Data Minimization**: Only collect necessary authentication data
- **Encryption**: All personal data encrypted at rest and in transit
- **Access Control**: Strict access controls on credential stores
- **Audit Trail**: Comprehensive logging of data access
- **Right to Erasure**: Ability to securely delete user credentials

#### HIPAA Compliance (Healthcare)
- **Access Control**: Role-based access to email systems
- **Audit Controls**: Detailed logging of all email access
- **Integrity**: Data integrity verification
- **Transmission Security**: End-to-end encryption

#### SOX Compliance (Financial)
- **Access Controls**: Segregation of duties
- **Audit Trail**: Immutable audit logs
- **Change Management**: Controlled configuration changes
- **Monitoring**: Continuous security monitoring

### Implementation Example

```go
// GDPR-compliant credential management
type GDPRCredentialStore struct {
    *security.CredentialStore
    auditLogger *security.SecureLogger
    retention   time.Duration
}

func (gcs *GDPRCredentialStore) GetCredential(name string) (*security.Credential, error) {
    // Log data access for GDPR audit requirements
    gcs.auditLogger.Audit("gdpr-compliance",
        fmt.Sprintf("Personal data accessed: %s", name),
        map[string]interface{}{
            "data_subject": name,
            "purpose": "email_authentication",
            "legal_basis": "legitimate_interest",
        })

    return gcs.CredentialStore.GetCredential(name)
}

func (gcs *GDPRCredentialStore) DeletePersonalData(userID string) error {
    // Right to erasure implementation
    gcs.auditLogger.Audit("gdpr-compliance",
        fmt.Sprintf("Personal data deletion requested: %s", userID),
        map[string]interface{}{
            "data_subject": userID,
            "request_type": "erasure",
        })

    return gcs.DeleteCredential(userID)
}
```

## API Reference

### Security Package Components

#### CredentialStore
```go
type CredentialStore struct {
    StorePath  string
    MasterKey  []byte
    Salt       []byte
}

// Methods
func NewCredentialStore(storePath string) (*CredentialStore, error)
func (cs *CredentialStore) InitializeMasterKey() error
func (cs *CredentialStore) StoreCredential(name string, cred *Credential) error
func (cs *CredentialStore) GetCredential(name string) (*Credential, error)
func (cs *CredentialStore) DeleteCredential(name string) error
func (cs *CredentialStore) ListCredentials() []string
func (cs *CredentialStore) RotateMasterKey(newPassword string) error
func (cs *CredentialStore) ValidateIntegrity() error
func (cs *CredentialStore) Cleanup() error
```

#### AuthManager
```go
type AuthManager struct {
    credStore    *CredentialStore
    rateLimiter  *RateLimiter
    failureCache map[string]*AuthFailureInfo
}

// Methods
func NewAuthManager(credStore *CredentialStore) *AuthManager
func (am *AuthManager) SetRateLimit(maxAttempts int, window time.Duration)
func (am *AuthManager) SetFailureThreshold(maxFailures int, lockoutTime time.Duration)
func (am *AuthManager) AuthenticateWithCredentialStore(ctx *AuthContext, credentialName string) (Authenticator, error)
func (am *AuthManager) CreateSecureAuthenticator(ctx *AuthContext, username, password string) (Authenticator, error)
func (am *AuthManager) ValidateAuthAttempt(ctx *AuthContext, success bool, err error) error
```

#### TLSVerifier
```go
type TLSVerifier struct {
    config  *TLSConfig
    metrics *TLSMetrics
}

// Methods
func NewTLSVerifier(config *TLSConfig) *TLSVerifier
func (tv *TLSVerifier) VerifyTLSConnection(serverAddr string, port int) (*TLSInfo, error)
func (tv *TLSVerifier) TestSMTPSTARTTLS(serverAddr string, port int) (*TLSInfo, error)
```

#### SecureLogger
```go
type SecureLogger struct {
    level         LogLevel
    maskSensitive bool
    auditMode     bool
    patterns      []*SensitivePattern
}

// Methods
func NewSecureLogger(config *LogConfig) (*SecureLogger, error)
func (sl *SecureLogger) Debug(component, message string, metadata ...map[string]interface{})
func (sl *SecureLogger) Info(component, message string, metadata ...map[string]interface{})
func (sl *SecureLogger) Warn(component, message string, metadata ...map[string]interface{})
func (sl *SecureLogger) Error(component, message string, metadata ...map[string]interface{})
func (sl *SecureLogger) AuthLog(username, event, result string, metadata ...map[string]interface{})
func (sl *SecureLogger) SecurityLog(event, severity, description string, metadata ...map[string]interface{})
func (sl *SecureLogger) Audit(component, message string, metadata ...map[string]interface{})
```

### Configuration Examples

#### Complete Security Configuration
```yaml
# smtp-edc-security.yaml
security:
  credentials:
    store_path: "/secure/credentials.db"
    master_key_prompt: true
    rotation_interval: "90d"
    backup_count: 5

  authentication:
    rate_limit:
      max_attempts: 100
      window: "1m"
    failure_threshold:
      max_failures: 5
      lockout_duration: "15m"

  tls:
    min_version: "1.2"
    max_version: "1.3"
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_CHACHA20_POLY1305_SHA256"
      - "TLS_AES_128_GCM_SHA256"
    certificate_pinning:
      enabled: true
      pins:
        - "sha256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

  logging:
    level: "info"
    mask_sensitive: true
    audit_mode: true
    max_log_size: "100MB"
    retention_days: 90
    log_path: "/var/log/smtp-edc"

  compliance:
    gdpr_mode: true
    audit_trail: true
    data_retention: "2y"
```

### Error Codes and Handling

| Error Code | Description | Recommended Action |
|------------|-------------|-------------------|
| SEC_001 | Credential store initialization failed | Check file permissions and disk space |
| SEC_002 | Master key validation failed | Verify master password |
| SEC_003 | Rate limit exceeded | Implement backoff strategy |
| SEC_004 | Account locked | Wait for lockout period or admin unlock |
| SEC_005 | TLS verification failed | Check certificate validity and configuration |
| SEC_006 | Weak authentication method | Upgrade to OAuth2 or CRAM-MD5 |
| SEC_007 | Sensitive data in logs | Review logging configuration |
| SEC_008 | Certificate expiring soon | Renew certificate |

---

**Note**: This guide represents current security best practices. Security requirements evolve rapidly, so regular review and updates are essential. Always consult with security professionals for production deployments.

**Version**: 1.0
**Last Updated**: December 2024
**Next Review**: March 2025
