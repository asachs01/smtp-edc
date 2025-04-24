package auth

import (
	"strings"
	"testing"
)

func TestNewAuthenticator(t *testing.T) {
	// Test creating a new authenticator with valid types
	testCases := []struct {
		authType string
		valid    bool
	}{
		{"plain", true},
		{"login", true},
		{"cram-md5", true},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.authType, func(t *testing.T) {
			auth, err := NewAuthenticator(tc.authType)
			if tc.valid {
				if err != nil {
					t.Errorf("Expected no error for auth type %s, got %v", tc.authType, err)
				}
				if auth == nil {
					t.Errorf("Expected non-nil authenticator for auth type %s", tc.authType)
				}
				// Compare auth types case-insensitively
				if !strings.EqualFold(auth.Type(), tc.authType) {
					t.Errorf("Expected auth type %s, got %s", tc.authType, auth.Type())
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid auth type %s", tc.authType)
				}
				if auth != nil {
					t.Errorf("Expected nil authenticator for invalid auth type %s", tc.authType)
				}
			}
		})
	}
}
