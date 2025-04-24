package config

import (
	"errors"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// SMTPConfig represents the SMTP configuration
type SMTPConfig struct {
	Server     string            `yaml:"server"`
	Port       int               `yaml:"port"`
	Username   string            `yaml:"username"`
	Password   string            `yaml:"password"`
	AuthType   string            `yaml:"auth_type"`
	StartTLS   bool              `yaml:"starttls"`
	SkipVerify bool              `yaml:"skip_verify"`
	Templates  map[string]string `yaml:"templates"`
}

// LoadConfig loads the SMTP configuration from a file
func LoadConfig(filename string) (*SMTPConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config SMTPConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the SMTP configuration to a file
func SaveConfig(config *SMTPConfig, filename string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// Validate checks if the configuration is valid
func (c *SMTPConfig) Validate() error {
	if c.Server == "" {
		return errors.New("server is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	if c.AuthType == "" {
		return errors.New("auth_type is required")
	}
	return nil
}
