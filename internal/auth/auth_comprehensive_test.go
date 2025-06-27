package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestPlainAuthenticatorComprehensive tests all aspects of PLAIN authentication
func TestPlainAuthenticatorComprehensive(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid credentials",
			username: "testuser",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "Empty username",
			username: "",
			password: "testpass",
			wantErr:  false, // PLAIN allows empty username
		},
		{
			name:     "Empty password",
			username: "testuser",
			password: "",
			wantErr:  false, // PLAIN allows empty password
		},
		{
			name:     "Special characters",
			username: "test@domain.com",
			password: "p@ss!w0rd#123",
			wantErr:  false,
		},
		{
			name:     "Unicode characters",
			username: "tëstüser",
			password: "pässwörd",
			wantErr:  false,
		},
		{
			name:     "Very long credentials",
			username: strings.Repeat("a", 1000),
			password: strings.Repeat("b", 1000),
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &PlainAuthenticator{}

			if auth.Type() != "PLAIN" {
				t.Errorf("Expected type PLAIN, got %s", auth.Type())
			}

			result, err := auth.Authenticate(tc.username, tc.password)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Decode and verify the result
			decoded, err := base64.StdEncoding.DecodeString(result)
			if err != nil {
				t.Errorf("Failed to decode result: %v", err)
				return
			}

			expected := fmt.Sprintf("\x00%s\x00%s", tc.username, tc.password)
			if string(decoded) != expected {
				t.Errorf("Expected %q, got %q", expected, string(decoded))
			}
		})
	}
}

// TestLoginAuthenticatorComprehensive tests all aspects of LOGIN authentication
func TestLoginAuthenticatorComprehensive(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid credentials",
			username: "testuser",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "Email as username",
			username: "test@example.com",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "Complex password",
			username: "testuser",
			password: "C0mpl3x!P@ssw0rd$123",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &LoginAuthenticator{}

			if auth.Type() != "LOGIN" {
				t.Errorf("Expected type LOGIN, got %s", auth.Type())
			}

			// Test first step (username)
			result, err := auth.Authenticate(tc.username, tc.password)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify username encoding
			decoded, err := base64.StdEncoding.DecodeString(result)
			if err != nil {
				t.Errorf("Failed to decode username: %v", err)
				return
			}

			if string(decoded) != tc.username {
				t.Errorf("Expected username %q, got %q", tc.username, string(decoded))
			}

			// Test second step (password)
			passwordResult := auth.GetPassword(tc.password)
			decodedPassword, err := base64.StdEncoding.DecodeString(passwordResult)
			if err != nil {
				t.Errorf("Failed to decode password: %v", err)
				return
			}

			if string(decodedPassword) != tc.password {
				t.Errorf("Expected password %q, got %q", tc.password, string(decodedPassword))
			}
		})
	}
}

// TestCRAMMD5AuthenticatorComprehensive tests all aspects of CRAM-MD5 authentication
func TestCRAMMD5AuthenticatorComprehensive(t *testing.T) {
	testCases := []struct {
		name      string
		username  string
		password  string
		challenge string
		wantErr   bool
	}{
		{
			name:      "Valid CRAM-MD5",
			username:  "testuser",
			password:  "testpass",
			challenge: base64.StdEncoding.EncodeToString([]byte("<1234567890.12345@example.com>")),
			wantErr:   false,
		},
		{
			name:      "Different challenge",
			username:  "user@domain.com",
			password:  "secretpass",
			challenge: base64.StdEncoding.EncodeToString([]byte("<abcdef.ghijk@server.com>")),
			wantErr:   false,
		},
		{
			name:      "Empty challenge",
			username:  "testuser",
			password:  "testpass",
			challenge: base64.StdEncoding.EncodeToString([]byte("")),
			wantErr:   false,
		},
		{
			name:      "Invalid base64 challenge",
			username:  "testuser",
			password:  "testpass",
			challenge: "invalid-base64!",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &CRAMMD5Authenticator{}

			if auth.Type() != "CRAM-MD5" {
				t.Errorf("Expected type CRAM-MD5, got %s", auth.Type())
			}

			// Test initial authenticate (should return username)
			result, err := auth.Authenticate(tc.username, tc.password)
			if err != nil {
				t.Errorf("Initial authenticate failed: %v", err)
				return
			}

			if result != tc.username {
				t.Errorf("Expected username %q, got %q", tc.username, result)
			}

			// Test challenge response generation
			response, err := auth.GenerateResponse(tc.challenge, tc.username, tc.password)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify the response format
			decoded, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				t.Errorf("Failed to decode response: %v", err)
				return
			}

			parts := strings.Split(string(decoded), " ")
			if len(parts) != 2 {
				t.Errorf("Expected 2 parts in response, got %d", len(parts))
				return
			}

			if parts[0] != tc.username {
				t.Errorf("Expected username %q in response, got %q", tc.username, parts[0])
			}

			// Verify HMAC-MD5 hash
			challengeBytes, _ := base64.StdEncoding.DecodeString(tc.challenge)
			h := hmac.New(md5.New, []byte(tc.password))
			h.Write(challengeBytes)
			expectedHash := hex.EncodeToString(h.Sum(nil))

			if parts[1] != expectedHash {
				t.Errorf("Expected hash %q, got %q", expectedHash, parts[1])
			}
		})
	}
}

// TestBase64Functions tests the base64 utility functions
func TestBase64Functions(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Simple string",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Special characters",
			input:   "hello@world.com!",
			wantErr: false,
		},
		{
			name:    "Unicode characters",
			input:   "héllo wörld",
			wantErr: false,
		},
		{
			name:    "Binary data",
			input:   "\x00\x01\x02\x03",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test encoding
			encoded := Base64Encode(tc.input)
			if encoded == "" && tc.input != "" {
				t.Error("Encoding returned empty string for non-empty input")
			}

			// Test decoding
			decoded, err := Base64Decode(encoded)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected decode error")
				}
				return
			}

			if err != nil {
				t.Errorf("Decode error: %v", err)
				return
			}

			if decoded != tc.input {
				t.Errorf("Round trip failed: expected %q, got %q", tc.input, decoded)
			}
		})
	}

	// Test decoding invalid base64
	t.Run("Invalid base64", func(t *testing.T) {
		_, err := Base64Decode("invalid-base64!")
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})
}

// TestAuthenticatorFactory tests the authenticator factory function
func TestAuthenticatorFactory(t *testing.T) {
	testCases := []struct {
		authType    string
		expectType  string
		expectError bool
	}{
		{
			authType:    "plain",
			expectType:  "PLAIN",
			expectError: false,
		},
		{
			authType:    "PLAIN",
			expectType:  "PLAIN",
			expectError: false,
		},
		{
			authType:    "login",
			expectType:  "LOGIN",
			expectError: false,
		},
		{
			authType:    "LOGIN",
			expectType:  "LOGIN",
			expectError: false,
		},
		{
			authType:    "cram-md5",
			expectType:  "CRAM-MD5",
			expectError: false,
		},
		{
			authType:    "CRAM-MD5",
			expectType:  "CRAM-MD5",
			expectError: false,
		},
		{
			authType:    "invalid",
			expectError: true,
		},
		{
			authType:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.authType, func(t *testing.T) {
			auth, err := NewAuthenticator(tc.authType)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for auth type %s", tc.authType)
				}
				if auth != nil {
					t.Errorf("Expected nil authenticator for invalid type %s", tc.authType)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for auth type %s: %v", tc.authType, err)
				return
			}

			if auth == nil {
				t.Errorf("Authenticator is nil for type %s", tc.authType)
				return
			}

			if auth.Type() != tc.expectType {
				t.Errorf("Expected type %s, got %s", tc.expectType, auth.Type())
			}
		})
	}
}

// BenchmarkAuthenticators benchmarks different authentication methods
func BenchmarkPlainAuthenticate(b *testing.B) {
	auth := &PlainAuthenticator{}
	username := "testuser"
	password := "testpass"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := auth.Authenticate(username, password)
		if err != nil {
			b.Fatalf("Authentication failed: %v", err)
		}
	}
}

func BenchmarkLoginAuthenticate(b *testing.B) {
	auth := &LoginAuthenticator{}
	username := "testuser"
	password := "testpass"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := auth.Authenticate(username, password)
		if err != nil {
			b.Fatalf("Authentication failed: %v", err)
		}
		_ = auth.GetPassword(password)
	}
}

func BenchmarkCRAMMD5Response(b *testing.B) {
	auth := &CRAMMD5Authenticator{}
	username := "testuser"
	password := "testpass"
	challenge := base64.StdEncoding.EncodeToString([]byte("<test@example.com>"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := auth.GenerateResponse(challenge, username, password)
		if err != nil {
			b.Fatalf("Response generation failed: %v", err)
		}
	}
}

func BenchmarkBase64Encode(b *testing.B) {
	input := "testuser\x00testpass"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Base64Encode(input)
	}
}

func BenchmarkBase64Decode(b *testing.B) {
	encoded := Base64Encode("testuser\x00testpass")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Base64Decode(encoded)
		if err != nil {
			b.Fatalf("Decode failed: %v", err)
		}
	}
}

// TestConcurrentAuthentication tests thread safety
func TestConcurrentAuthentication(t *testing.T) {
	auths := []Authenticator{
		&PlainAuthenticator{},
		&LoginAuthenticator{},
		&CRAMMD5Authenticator{},
	}

	const numGoroutines = 100
	const numIterations = 10

	for _, auth := range auths {
		t.Run(auth.Type(), func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines*numIterations)

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()

					for j := 0; j < numIterations; j++ {
						username := fmt.Sprintf("user%d", goroutineID)
						password := fmt.Sprintf("pass%d", j)

						_, err := auth.Authenticate(username, password)
						if err != nil {
							errors <- fmt.Errorf("goroutine %d iteration %d: %v", goroutineID, j, err)
						}
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			for err := range errors {
				t.Error(err)
			}
		})
	}
}

// Helper function to verify HMAC-MD5 manually
func verifyHMACMD5(challenge, username, password string) (string, error) {
	challengeBytes, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		return "", err
	}

	h := hmac.New(md5.New, []byte(password))
	h.Write(challengeBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	response := fmt.Sprintf("%s %s", username, hash)
	return base64.StdEncoding.EncodeToString([]byte(response)), nil
}

// TestCRAMMD5Compatibility tests CRAM-MD5 compatibility with standard implementations
func TestCRAMMD5Compatibility(t *testing.T) {
	testVectors := []struct {
		name      string
		username  string
		password  string
		challenge string
		expected  string
	}{
		{
			name:      "RFC example",
			username:  "tim",
			password:  "tanstaaftanstaaf",
			challenge: base64.StdEncoding.EncodeToString([]byte("<1896.697170952@postoffice.reston.mci.net>")),
		},
	}

	auth := &CRAMMD5Authenticator{}

	for _, tv := range testVectors {
		t.Run(tv.name, func(t *testing.T) {
			result, err := auth.GenerateResponse(tv.challenge, tv.username, tv.password)
			if err != nil {
				t.Fatalf("GenerateResponse failed: %v", err)
			}

			// Verify it matches our manual calculation
			expected, err := verifyHMACMD5(tv.challenge, tv.username, tv.password)
			if err != nil {
				t.Fatalf("Manual verification failed: %v", err)
			}

			if result != expected {
				t.Errorf("CRAM-MD5 mismatch: got %s, expected %s", result, expected)
			}
		})
	}
}
