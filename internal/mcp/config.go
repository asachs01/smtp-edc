package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPConfig represents the MCP server configuration
type MCPConfig struct {
	Transport TransportConfig `json:"transport"`
	Server    ServerConfig    `json:"server"`
	Security  SecurityConfig  `json:"security"`
}

// TransportConfig defines transport settings
type TransportConfig struct {
	Type string `json:"type"` // "stdio" or "http"
	HTTP HTTPConfig `json:"http,omitempty"`
}

// HTTPConfig defines HTTP transport settings
type HTTPConfig struct {
	Port        int    `json:"port"`
	Host        string `json:"host"`
	EnableCORS  bool   `json:"enableCors"`
	EnableHTTPS bool   `json:"enableHttps"`
	CertFile    string `json:"certFile,omitempty"`
	KeyFile     string `json:"keyFile,omitempty"`
}

// ServerConfig defines server settings
type ServerConfig struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Debug       bool     `json:"debug"`
	MaxClients  int      `json:"maxClients"`
	Timeout     int      `json:"timeout"` // seconds
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	RequireAuth bool     `json:"requireAuth"`
	APIKeys     []string `json:"apiKeys,omitempty"`
	AllowedIPs  []string `json:"allowedIps,omitempty"`
}

// DefaultMCPConfig returns the default MCP configuration
func DefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Transport: TransportConfig{
			Type: "stdio",
			HTTP: HTTPConfig{
				Port:       8080,
				Host:       "localhost",
				EnableCORS: true,
			},
		},
		Server: ServerConfig{
			Name:        "smtp-edc-mcp",
			Version:     "1.0.0",
			Description: "SMTP-EDC Model Context Protocol Server",
			Debug:       false,
			MaxClients:  10,
			Timeout:     30,
		},
		Security: SecurityConfig{
			RequireAuth: false,
			AllowedIPs:  []string{"127.0.0.1", "::1"},
		},
	}
}

// LoadMCPConfig loads MCP configuration from a file
func LoadMCPConfig(filename string) (*MCPConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultMCPConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing values
	if config.Transport.Type == "" {
		config.Transport.Type = "stdio"
	}
	if config.Server.Name == "" {
		config.Server.Name = "smtp-edc-mcp"
	}
	if config.Server.Version == "" {
		config.Server.Version = "1.0.0"
	}
	if config.Server.MaxClients == 0 {
		config.Server.MaxClients = 10
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30
	}
	if config.Transport.Type == "http" && config.Transport.HTTP.Port == 0 {
		config.Transport.HTTP.Port = 8080
	}
	if config.Transport.Type == "http" && config.Transport.HTTP.Host == "" {
		config.Transport.HTTP.Host = "localhost"
	}

	return &config, nil
}

// SaveMCPConfig saves MCP configuration to a file
func SaveMCPConfig(config *MCPConfig, filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultMCPConfigPath returns the default MCP config file path
func GetDefaultMCPConfigPath() string {
	// Check for config in current directory first
	if _, err := os.Stat("mcp-config.json"); err == nil {
		return "mcp-config.json"
	}

	// Check for config in user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".smtp-edc", "mcp-config.json")
		return configPath
	}

	// Default to current directory
	return "mcp-config.json"
}

// ValidateMCPConfig validates the MCP configuration
func ValidateMCPConfig(config *MCPConfig) error {
	// Validate transport type
	if config.Transport.Type != "stdio" && config.Transport.Type != "http" {
		return fmt.Errorf("invalid transport type: %s (must be 'stdio' or 'http')", config.Transport.Type)
	}

	// Validate HTTP config if using HTTP transport
	if config.Transport.Type == "http" {
		if config.Transport.HTTP.Port < 1 || config.Transport.HTTP.Port > 65535 {
			return fmt.Errorf("invalid HTTP port: %d", config.Transport.HTTP.Port)
		}

		if config.Transport.HTTP.EnableHTTPS {
			if config.Transport.HTTP.CertFile == "" {
				return fmt.Errorf("cert file required when HTTPS is enabled")
			}
			if config.Transport.HTTP.KeyFile == "" {
				return fmt.Errorf("key file required when HTTPS is enabled")
			}
		}
	}

	// Validate server config
	if config.Server.MaxClients < 1 {
		return fmt.Errorf("max clients must be at least 1")
	}
	if config.Server.Timeout < 1 {
		return fmt.Errorf("timeout must be at least 1 second")
	}

	return nil
}