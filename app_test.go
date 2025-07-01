package main

import (
	"context"
	"testing"
)

func TestApp_Greet(t *testing.T) {
	ctx := context.Background()
	app := NewApp()
	app.startup(ctx)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "greet with name",
			input:    "John",
			expected: "Hello John, welcome to SMTP-EDC!",
		},
		{
			name:     "greet with empty name",
			input:    "",
			expected: "Hello, it's nice to meet you!",
		},
		{
			name:     "greet with whitespace name",
			input:    "   ",
			expected: "Hello, it's nice to meet you!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.Greet(tt.input)
			if result != tt.expected {
				t.Errorf("Greet(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestApp_ValidateEmailAddress(t *testing.T) {
	ctx := context.Background()
	app := NewApp()
	app.startup(ctx)

	tests := []struct {
		name        string
		email       string
		wantSuccess bool
		wantError   bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			wantSuccess: true,
			wantError:   false,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			wantSuccess: true,
			wantError:   false,
		},
		{
			name:        "invalid email without @",
			email:       "invalid-email",
			wantSuccess: false,
			wantError:   true,
		},
		{
			name:        "invalid email without domain",
			email:       "user@",
			wantSuccess: false,
			wantError:   true,
		},
		{
			name:        "empty email",
			email:       "",
			wantSuccess: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ValidateEmailAddress(tt.email)

			if result.Success != tt.wantSuccess {
				t.Errorf("ValidateEmailAddress(%q).Success = %v, want %v", tt.email, result.Success, tt.wantSuccess)
			}

			hasError := result.Error != ""
			if hasError != tt.wantError {
				t.Errorf("ValidateEmailAddress(%q) error = %v, want error = %v", tt.email, hasError, tt.wantError)
			}
		})
	}
}
