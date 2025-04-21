package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// CRAMMD5Authenticator implements the CRAM-MD5 authentication method
type CRAMMD5Authenticator struct{}

// Type returns the authentication type
func (a *CRAMMD5Authenticator) Type() string {
	return "CRAM-MD5"
}

// Authenticate performs CRAM-MD5 authentication
func (a *CRAMMD5Authenticator) Authenticate(username, password string) (string, error) {
	// CRAM-MD5 requires a challenge from the server
	// The actual authentication happens in a separate step
	return username, nil
}

// GenerateResponse generates the CRAM-MD5 response
func (a *CRAMMD5Authenticator) GenerateResponse(challenge, username, password string) (string, error) {
	// Decode the base64 challenge
	decodedChallenge, err := Base64Decode(challenge)
	if err != nil {
		return "", fmt.Errorf("failed to decode challenge: %v", err)
	}

	// Create HMAC-MD5 hash
	h := hmac.New(md5.New, []byte(password))
	h.Write([]byte(decodedChallenge))
	hash := hex.EncodeToString(h.Sum(nil))

	// Format: username hash
	response := fmt.Sprintf("%s %s", username, hash)
	return Base64Encode(response), nil
}
