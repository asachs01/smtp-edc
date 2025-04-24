package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempConfigFile := "temp_config.yaml"
	configContent := `
smtp:
  host: "test.smtp.com"
  port: 587
  username: "testuser"
  password: "testpassword"
  from: "test@example.com"
tls: true
auth:
  mechanisms: ["PLAIN", "LOGIN"]
`
	err := os.WriteFile(tempConfigFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}
	defer os.Remove(tempConfigFile)

	// Test loading config from the temporary file
	config, err := LoadConfig(tempConfigFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Expected config values
	expectedConfig := &Config{
		SMTP: SMTPConfig{
			Host:     "test.smtp.com",
			Port:     587,
			Username: "testuser",
			Password: "testpassword",
			From:     "test@example.com",
		},
		TLS: true,
		Auth: AuthConfig{
			Mechanisms: []string{"PLAIN", "LOGIN"},
		},
	}

	// Compare loaded config with expected config
	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config. Got: %+v, Expected: %+v", config, expectedConfig)
	}

	// Test loading config with an invalid file path
	_, err = LoadConfig("nonexistent_config.yaml")
	if err == nil {
		t.Errorf("LoadConfig did not return an error for an invalid file path")
	}
}

func TestValidate(t *testing.T) {
	validConfig := &Config{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "password",
			From:     "user@example.com",
		},
		TLS: true,
		Auth: AuthConfig{
			Mechanisms: []string{"PLAIN", "LOGIN"},
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Validate returned an error for a valid config: %v", err)
	}

	invalidConfig := &Config{
		SMTP: SMTPConfig{
			Host:     "",
			Port:     587,
			Username: "user",
			Password: "password",
			From:     "user@example.com",
		},
		TLS: true,
		Auth: AuthConfig{
			Mechanisms: []string{"PLAIN", "LOGIN"},
		},
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Errorf("Validate did not return an error for an invalid config")
	}

	invalidConfig2 := &Config{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "",
			Password: "password",
			From:     "user@example.com",
		},
		TLS: true,
		Auth: AuthConfig{
			Mechanisms: []string{"PLAIN", "LOGIN"},
		},
	}
	if err := invalidConfig2.Validate(); err == nil {
		t.Errorf("Validate did not return an error for an invalid config")
	}

	invalidConfig3 := &Config{
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "password",
			From:     "",
		},
		TLS: true,
		Auth: AuthConfig{
			Mechanisms: []string{"PLAIN", "LOGIN"},
		},
	}
	if err := invalidConfig3.Validate(); err == nil {
		t.Errorf("Validate did not return an error for an invalid config")
	}
}