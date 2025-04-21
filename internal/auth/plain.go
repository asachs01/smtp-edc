package auth

import (
	"fmt"
)

// PlainAuthenticator implements the PLAIN authentication method
type PlainAuthenticator struct{}

// Type returns the authentication type
func (a *PlainAuthenticator) Type() string {
	return "PLAIN"
}

// Authenticate performs PLAIN authentication
func (a *PlainAuthenticator) Authenticate(username, password string) (string, error) {
	// PLAIN authentication format: \0username\0password
	authString := fmt.Sprintf("\x00%s\x00%s", username, password)
	return Base64Encode(authString), nil
}
