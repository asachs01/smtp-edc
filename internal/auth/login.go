package auth

// LoginAuthenticator implements the LOGIN authentication method
type LoginAuthenticator struct{}

// Type returns the authentication type
func (a *LoginAuthenticator) Type() string {
	return "LOGIN"
}

// Authenticate performs LOGIN authentication
func (a *LoginAuthenticator) Authenticate(username, password string) (string, error) {
	// LOGIN authentication requires separate base64 encoding of username and password
	encodedUsername := Base64Encode(username)

	// Return the encoded username, the password will be sent in a separate step
	return encodedUsername, nil
}

// GetPassword returns the base64 encoded password for the second step of LOGIN auth
func (a *LoginAuthenticator) GetPassword(password string) string {
	return Base64Encode(password)
}
