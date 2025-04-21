package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Server     string            `json:"server" yaml:"server"`
	Port       int               `json:"port" yaml:"port"`
	Username   string            `json:"username" yaml:"username"`
	Password   string            `json:"password" yaml:"password"`
	AuthType   string            `json:"auth_type" yaml:"auth_type"`
	StartTLS   bool              `json:"starttls" yaml:"starttls"`
	SkipVerify bool              `json:"skip_verify" yaml:"skip_verify"`
	Templates  map[string]string `json:"templates" yaml:"templates"`
}

// LoadConfig loads configuration from environment variables and config file
func LoadConfig(configFile string) (*SMTPConfig, error) {
	config := &SMTPConfig{
		Port: 25, // Default port
	}

	// Load from config file if specified
	if configFile != "" {
		if err := loadConfigFile(configFile, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if server := os.Getenv("SMTP_SERVER"); server != "" {
		config.Server = server
	}
	if port := os.Getenv("SMTP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Port)
	}
	if username := os.Getenv("SMTP_USERNAME"); username != "" {
		config.Username = username
	}
	if password := os.Getenv("SMTP_PASSWORD"); password != "" {
		config.Password = password
	}
	if authType := os.Getenv("SMTP_AUTH_TYPE"); authType != "" {
		config.AuthType = authType
	}
	if startTLS := os.Getenv("SMTP_STARTTLS"); startTLS != "" {
		config.StartTLS = strings.ToLower(startTLS) == "true"
	}
	if skipVerify := os.Getenv("SMTP_SKIP_VERIFY"); skipVerify != "" {
		config.SkipVerify = strings.ToLower(skipVerify) == "true"
	}

	return config, nil
}

// loadConfigFile loads configuration from a JSON or YAML file
func loadConfigFile(filename string, config *SMTPConfig) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %v", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse YAML config: %v", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", filepath.Ext(filename))
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *SMTPConfig, filename string) error {
	var data []byte
	var err error

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config file format: %s", filepath.Ext(filename))
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
