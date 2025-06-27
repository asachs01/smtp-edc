package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2Provider represents different OAuth2 providers
type OAuth2Provider int

const (
	ProviderUnknown OAuth2Provider = iota
	ProviderGoogle
	ProviderMicrosoft
	ProviderYahoo
	ProviderCustom
)

// OAuth2Authenticator implements OAuth2 authentication for SMTP
type OAuth2Authenticator struct {
	Provider     OAuth2Provider
	ClientID     string
	ClientSecret string
	TokenURL     string
	AuthURL      string
	Scopes       []string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	Provider     OAuth2Provider `json:"provider"`
	ClientID     string         `json:"client_id"`
	ClientSecret string         `json:"client_secret"`
	TokenURL     string         `json:"token_url,omitempty"`
	AuthURL      string         `json:"auth_url,omitempty"`
	Scopes       []string       `json:"scopes,omitempty"`
}

// OAuth2Token represents an OAuth2 access token
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewOAuth2Authenticator creates a new OAuth2 authenticator
func NewOAuth2Authenticator(config *OAuth2Config) *OAuth2Authenticator {
	auth := &OAuth2Authenticator{
		Provider:     config.Provider,
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.TokenURL,
		AuthURL:      config.AuthURL,
		Scopes:       config.Scopes,
	}

	// Set default URLs and scopes based on provider
	switch config.Provider {
	case ProviderGoogle:
		if auth.TokenURL == "" {
			auth.TokenURL = "https://oauth2.googleapis.com/token"
		}
		if auth.AuthURL == "" {
			auth.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
		}
		if len(auth.Scopes) == 0 {
			auth.Scopes = []string{"https://mail.google.com/"}
		}

	case ProviderMicrosoft:
		if auth.TokenURL == "" {
			auth.TokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
		}
		if auth.AuthURL == "" {
			auth.AuthURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
		}
		if len(auth.Scopes) == 0 {
			auth.Scopes = []string{"https://outlook.office.com/SMTP.Send"}
		}

	case ProviderYahoo:
		if auth.TokenURL == "" {
			auth.TokenURL = "https://api.login.yahoo.com/oauth2/get_token"
		}
		if auth.AuthURL == "" {
			auth.AuthURL = "https://api.login.yahoo.com/oauth2/request_auth"
		}
		if len(auth.Scopes) == 0 {
			auth.Scopes = []string{"mail-w"}
		}
	}

	return auth
}

// Type returns the authentication type
func (a *OAuth2Authenticator) Type() string {
	return "XOAUTH2"
}

// Authenticate performs OAuth2 authentication
func (a *OAuth2Authenticator) Authenticate(username, password string) (string, error) {
	// For OAuth2, the "password" parameter is actually the access token
	if password != "" {
		a.AccessToken = password
	}

	if a.AccessToken == "" {
		return "", fmt.Errorf("access token is required for OAuth2 authentication")
	}

	// Check if token is expired and refresh if possible
	if time.Now().After(a.ExpiresAt) && a.RefreshToken != "" {
		if err := a.RefreshAccessToken(); err != nil {
			return "", fmt.Errorf("failed to refresh access token: %v", err)
		}
	}

	// Format OAuth2 SASL string: user=username\x01auth=Bearer token\x01\x01
	authString := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", username, a.AccessToken)
	return Base64Encode(authString), nil
}

// GetAuthorizationURL generates the authorization URL for OAuth2 flow
func (a *OAuth2Authenticator) GetAuthorizationURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", a.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(a.Scopes, " "))
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	if state != "" {
		params.Set("state", state)
	}

	return fmt.Sprintf("%s?%s", a.AuthURL, params.Encode())
}

// ExchangeCodeForToken exchanges authorization code for access token
func (a *OAuth2Authenticator) ExchangeCodeForToken(ctx context.Context, code, redirectURI string) (*OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", a.ClientID)
	data.Set("client_secret", a.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", a.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("OAuth2 error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	token := &OAuth2Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		Scope:        tokenResp.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	// Update authenticator with new token
	a.AccessToken = token.AccessToken
	a.RefreshToken = token.RefreshToken
	a.ExpiresAt = token.ExpiresAt

	return token, nil
}

// RefreshAccessToken refreshes an expired access token
func (a *OAuth2Authenticator) RefreshAccessToken() error {
	if a.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("client_id", a.ClientID)
	data.Set("client_secret", a.ClientSecret)
	data.Set("refresh_token", a.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", a.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %v", err)
	}

	if tokenResp.Error != "" {
		return fmt.Errorf("OAuth2 refresh error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	// Update tokens
	a.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		a.RefreshToken = tokenResp.RefreshToken
	}
	a.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// IsTokenValid checks if the current access token is valid
func (a *OAuth2Authenticator) IsTokenValid() bool {
	return a.AccessToken != "" && time.Now().Before(a.ExpiresAt)
}

// GetAccessToken returns the current access token
func (a *OAuth2Authenticator) GetAccessToken() string {
	return a.AccessToken
}

// SetTokens sets the access and refresh tokens
func (a *OAuth2Authenticator) SetTokens(accessToken, refreshToken string, expiresAt time.Time) {
	a.AccessToken = accessToken
	a.RefreshToken = refreshToken
	a.ExpiresAt = expiresAt
}

// ValidateConfiguration validates the OAuth2 configuration
func (a *OAuth2Authenticator) ValidateConfiguration() error {
	if a.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if a.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}
	if a.TokenURL == "" {
		return fmt.Errorf("token URL is required")
	}
	if a.AuthURL == "" {
		return fmt.Errorf("authorization URL is required")
	}
	if len(a.Scopes) == 0 {
		return fmt.Errorf("at least one scope is required")
	}
	return nil
}

// GetProviderName returns the human-readable name of the OAuth2 provider
func (a *OAuth2Authenticator) GetProviderName() string {
	switch a.Provider {
	case ProviderGoogle:
		return "Google"
	case ProviderMicrosoft:
		return "Microsoft"
	case ProviderYahoo:
		return "Yahoo"
	case ProviderCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}
