package auth

import (
	"encoding/base64"
	"fmt"
)

// Authenticator defines the interface for SMTP authentication methods
type Authenticator interface {
	// Type returns the authentication type (e.g., "PLAIN", "LOGIN", "CRAM-MD5")
	Type() string
	// Authenticate performs the authentication process
	Authenticate(username, password string) (string, error)
}

// Base64Encode encodes a string to base64
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes a base64 string
func Base64Decode(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}
	return string(decoded), nil
}

// NewAuthenticator creates a new authenticator based on the type
func NewAuthenticator(authType string) (Authenticator, error) {
	switch authType {
	case "plain":
		return &PlainAuthenticator{}, nil
	case "login":
		return &LoginAuthenticator{}, nil
	case "cram-md5":
		return &CRAMMD5Authenticator{}, nil
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", authType)
	}
}
