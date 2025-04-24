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
server: "test.smtp.com"
port: 587
username: "testuser"
password: "testpassword"
auth_type: "plain"
starttls: true
skip_verify: false
templates:
  welcome: "welcome.txt"
  reset: "reset.txt"
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
	expectedConfig := &SMTPConfig{
		Server:     "test.smtp.com",
		Port:       587,
		Username:   "testuser",
		Password:   "testpassword",
		AuthType:   "plain",
		StartTLS:   true,
		SkipVerify: false,
		Templates: map[string]string{
			"welcome": "welcome.txt",
			"reset":   "reset.txt",
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

func TestSaveConfig(t *testing.T) {
	// Create a test config
	testConfig := &SMTPConfig{
		Server:     "test.smtp.com",
		Port:       587,
		Username:   "testuser",
		Password:   "testpassword",
		AuthType:   "plain",
		StartTLS:   true,
		SkipVerify: false,
		Templates: map[string]string{
			"welcome": "welcome.txt",
		},
	}

	// Test saving to YAML
	tempYAMLFile := "temp_config.yaml"
	defer os.Remove(tempYAMLFile)

	err := SaveConfig(testConfig, tempYAMLFile)
	if err != nil {
		t.Fatalf("SaveConfig failed for YAML: %v", err)
	}

	// Test loading the saved YAML config
	loadedConfig, err := LoadConfig(tempYAMLFile)
	if err != nil {
		t.Fatalf("Failed to load saved YAML config: %v", err)
	}

	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded YAML config does not match saved config. Got: %+v, Expected: %+v", loadedConfig, testConfig)
	}

	// Test saving to JSON
	tempJSONFile := "temp_config.json"
	defer os.Remove(tempJSONFile)

	err = SaveConfig(testConfig, tempJSONFile)
	if err != nil {
		t.Fatalf("SaveConfig failed for JSON: %v", err)
	}

	// Test loading the saved JSON config
	loadedConfig, err = LoadConfig(tempJSONFile)
	if err != nil {
		t.Fatalf("Failed to load saved JSON config: %v", err)
	}

	if !reflect.DeepEqual(loadedConfig, testConfig) {
		t.Errorf("Loaded JSON config does not match saved config. Got: %+v, Expected: %+v", loadedConfig, testConfig)
	}
}

func TestValidate(t *testing.T) {
	validConfig := &SMTPConfig{
		Server:     "smtp.example.com",
		Port:       587,
		Username:   "user",
		Password:   "password",
		AuthType:   "plain",
		StartTLS:   true,
		SkipVerify: false,
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Validate returned an error for a valid config: %v", err)
	}

	invalidConfig := &SMTPConfig{
		Server:     "",
		Port:       587,
		Username:   "user",
		Password:   "password",
		AuthType:   "plain",
		StartTLS:   true,
		SkipVerify: false,
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Errorf("Validate did not return an error for an invalid config")
	}

	invalidConfig2 := &SMTPConfig{
		Server:     "smtp.example.com",
		Port:       587,
		Username:   "",
		Password:   "password",
		AuthType:   "plain",
		StartTLS:   true,
		SkipVerify: false,
	}
	if err := invalidConfig2.Validate(); err == nil {
		t.Errorf("Validate did not return an error for an invalid config")
	}
}
