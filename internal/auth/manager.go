package auth

import (
	"fmt"
	"time"

	"github.com/asachs/smtp-edc/internal/security"
)

// AuthManager manages authentication across different methods and provides security features
type AuthManager struct {
	credStore    *security.CredentialStore
	rateLimiter  *RateLimiter
	failureCache map[string]*AuthFailureInfo
	maxFailures  int
	lockoutTime  time.Duration
}

// AuthFailureInfo tracks authentication failures for rate limiting
type AuthFailureInfo struct {
	Count       int
	LastFailure time.Time
	LockedUntil time.Time
}

// AuthContext contains authentication context and security information
type AuthContext struct {
	Username      string
	AuthType      string
	Server        string
	Port          int
	TLSRequired   bool
	ClientIP      string
	UserAgent     string
	Timestamp     time.Time
	SessionID     string
	MaxRetries    int
	RetryDelay    time.Duration
	AttemptNumber int
}

// NewAuthManager creates a new authentication manager with security features
func NewAuthManager(credStore *security.CredentialStore) *AuthManager {
	return &AuthManager{
		credStore:    credStore,
		rateLimiter:  NewRateLimiter(100, time.Minute), // 100 attempts per minute
		failureCache: make(map[string]*AuthFailureInfo),
		maxFailures:  5,
		lockoutTime:  time.Minute * 15, // 15-minute lockout
	}
}

// SetRateLimit configures the rate limiting parameters
func (am *AuthManager) SetRateLimit(maxAttempts int, window time.Duration) {
	am.rateLimiter = NewRateLimiter(maxAttempts, window)
}

// SetFailureThreshold configures authentication failure thresholds
func (am *AuthManager) SetFailureThreshold(maxFailures int, lockoutTime time.Duration) {
	am.maxFailures = maxFailures
	am.lockoutTime = lockoutTime
}

// AuthenticateWithCredentialStore authenticates using stored credentials
func (am *AuthManager) AuthenticateWithCredentialStore(ctx *AuthContext, credentialName string) (Authenticator, error) {
	// Check rate limiting
	if !am.rateLimiter.Allow(ctx.ClientIP) {
		return nil, fmt.Errorf("rate limit exceeded for IP: %s", ctx.ClientIP)
	}

	// Check authentication failure lockout
	if am.isLocked(ctx.Username + "@" + ctx.Server) {
		return nil, fmt.Errorf("account temporarily locked due to multiple failed attempts")
	}

	// Get credentials from secure store
	cred, err := am.credStore.GetCredential(credentialName)
	if err != nil {
		am.recordFailure(ctx.Username + "@" + ctx.Server)
		return nil, fmt.Errorf("failed to retrieve credentials: %v", err)
	}

	// Validate that the credential matches the context
	if cred.Server != ctx.Server || cred.Port != ctx.Port {
		am.recordFailure(ctx.Username + "@" + ctx.Server)
		return nil, fmt.Errorf("credential server/port mismatch")
	}

	// Create authenticator based on stored auth type
	var authenticator Authenticator
	switch cred.AuthType {
	case "oauth2", "xoauth2":
		oauth2Config := &OAuth2Config{
			Provider: ProviderCustom,
		}
		// Try to determine provider from server
		if provider := am.detectOAuth2Provider(cred.Server); provider != ProviderUnknown {
			oauth2Config.Provider = provider
		}
		authenticator = NewOAuth2Authenticator(oauth2Config)
		// Set the access token from metadata if available
		if accessToken, exists := cred.Metadata["access_token"]; exists {
			if refreshToken, exists := cred.Metadata["refresh_token"]; exists {
				var expiresAt time.Time
				if expiresAtStr, exists := cred.Metadata["expires_at"]; exists {
					expiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
				}
				oauth2Auth := authenticator.(*OAuth2Authenticator)
				oauth2Auth.SetTokens(accessToken, refreshToken, expiresAt)
			}
		}
	default:
		authenticator, err = NewAuthenticator(cred.AuthType)
		if err != nil {
			am.recordFailure(ctx.Username + "@" + ctx.Server)
			return nil, fmt.Errorf("failed to create authenticator: %v", err)
		}
	}

	// Update context with credential information
	ctx.Username = cred.Username
	ctx.AuthType = cred.AuthType

	return authenticator, nil
}

// CreateSecureAuthenticator creates an authenticator with enhanced security validation
func (am *AuthManager) CreateSecureAuthenticator(ctx *AuthContext, username, password string) (Authenticator, error) {
	// Validate input parameters
	if err := am.validateAuthContext(ctx); err != nil {
		return nil, fmt.Errorf("invalid authentication context: %v", err)
	}

	// Check rate limiting
	if !am.rateLimiter.Allow(ctx.ClientIP) {
		return nil, fmt.Errorf("rate limit exceeded for IP: %s", ctx.ClientIP)
	}

	// Check authentication failure lockout
	key := username + "@" + ctx.Server
	if am.isLocked(key) {
		return nil, fmt.Errorf("account temporarily locked due to multiple failed attempts")
	}

	// Create authenticator with security validation
	var authenticator Authenticator
	var err error

	switch ctx.AuthType {
	case "plain":
		if !ctx.TLSRequired {
			return nil, fmt.Errorf("PLAIN authentication requires TLS encryption")
		}
		authenticator, err = NewAuthenticator("plain")
	case "login":
		if !ctx.TLSRequired {
			return nil, fmt.Errorf("LOGIN authentication requires TLS encryption")
		}
		authenticator, err = NewAuthenticator("login")
	case "cram-md5":
		authenticator, err = NewAuthenticator("cram-md5")
	case "oauth2", "xoauth2":
		oauth2Config := &OAuth2Config{
			Provider: am.detectOAuth2Provider(ctx.Server),
		}
		authenticator = NewOAuth2Authenticator(oauth2Config)
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", ctx.AuthType)
	}

	if err != nil {
		am.recordFailure(key)
		return nil, fmt.Errorf("failed to create authenticator: %v", err)
	}

	return authenticator, nil
}

// ValidateAuthAttempt validates an authentication attempt and records results
func (am *AuthManager) ValidateAuthAttempt(ctx *AuthContext, success bool, err error) error {
	key := ctx.Username + "@" + ctx.Server

	if success {
		// Clear failure count on successful authentication
		delete(am.failureCache, key)
		return nil
	}

	// Record failure
	am.recordFailure(key)

	// Return appropriate error based on failure type
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	return fmt.Errorf("authentication failed for user %s", ctx.Username)
}

// recordFailure records authentication failure for rate limiting
func (am *AuthManager) recordFailure(key string) {
	now := time.Now()
	failure, exists := am.failureCache[key]

	if !exists {
		failure = &AuthFailureInfo{
			Count:       1,
			LastFailure: now,
		}
	} else {
		failure.Count++
		failure.LastFailure = now
	}

	// Apply lockout if max failures exceeded
	if failure.Count >= am.maxFailures {
		failure.LockedUntil = now.Add(am.lockoutTime)
	}

	am.failureCache[key] = failure
}

// isLocked checks if an account is currently locked
func (am *AuthManager) isLocked(key string) bool {
	failure, exists := am.failureCache[key]
	if !exists {
		return false
	}

	return time.Now().Before(failure.LockedUntil)
}

// detectOAuth2Provider attempts to detect OAuth2 provider from server hostname
func (am *AuthManager) detectOAuth2Provider(server string) OAuth2Provider {
	switch {
	case contains(server, "gmail") || contains(server, "google"):
		return ProviderGoogle
	case contains(server, "outlook") || contains(server, "hotmail") || contains(server, "office365"):
		return ProviderMicrosoft
	case contains(server, "yahoo") || contains(server, "ymail"):
		return ProviderYahoo
	default:
		return ProviderCustom
	}
}

// validateAuthContext validates the authentication context
func (am *AuthManager) validateAuthContext(ctx *AuthContext) error {
	if ctx.Username == "" {
		return fmt.Errorf("username is required")
	}
	if ctx.AuthType == "" {
		return fmt.Errorf("authentication type is required")
	}
	if ctx.Server == "" {
		return fmt.Errorf("server is required")
	}
	if ctx.Port <= 0 || ctx.Port > 65535 {
		return fmt.Errorf("invalid port number")
	}
	if ctx.ClientIP == "" {
		ctx.ClientIP = "unknown"
	}
	if ctx.Timestamp.IsZero() {
		ctx.Timestamp = time.Now()
	}
	if ctx.MaxRetries <= 0 {
		ctx.MaxRetries = 3
	}
	if ctx.RetryDelay <= 0 {
		ctx.RetryDelay = time.Second * 2
	}

	return nil
}

// GetFailureInfo returns authentication failure information for debugging
func (am *AuthManager) GetFailureInfo(username, server string) *AuthFailureInfo {
	key := username + "@" + server
	if failure, exists := am.failureCache[key]; exists {
		// Return a copy to prevent external modification
		return &AuthFailureInfo{
			Count:       failure.Count,
			LastFailure: failure.LastFailure,
			LockedUntil: failure.LockedUntil,
		}
	}
	return nil
}

// ClearFailures clears authentication failure cache (admin function)
func (am *AuthManager) ClearFailures(username, server string) {
	if username == "" && server == "" {
		// Clear all failures
		am.failureCache = make(map[string]*AuthFailureInfo)
	} else {
		key := username + "@" + server
		delete(am.failureCache, key)
	}
}

// GetAuthStats returns authentication statistics
func (am *AuthManager) GetAuthStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalLocked := 0
	totalFailures := 0

	for _, failure := range am.failureCache {
		totalFailures += failure.Count
		if time.Now().Before(failure.LockedUntil) {
			totalLocked++
		}
	}

	stats["total_failed_accounts"] = len(am.failureCache)
	stats["total_failures"] = totalFailures
	stats["currently_locked"] = totalLocked
	stats["rate_limit_window"] = am.rateLimiter.window
	stats["rate_limit_max"] = am.rateLimiter.maxAttempts

	return stats
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0))
}

// indexOfSubstring finds the index of a substring (simple implementation)
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
