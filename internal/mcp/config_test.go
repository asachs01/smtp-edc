package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultMCPConfig(t *testing.T) {
	config := DefaultMCPConfig()
	
	if config.Transport.Type != "stdio" {
		t.Errorf("Expected default transport type 'stdio', got '%s'", config.Transport.Type)
	}
	
	if config.Transport.HTTP.Port != 8080 {
		t.Errorf("Expected default HTTP port 8080, got %d", config.Transport.HTTP.Port)
	}
	
	if config.Server.Name != "smtp-edc-mcp" {
		t.Errorf("Expected server name 'smtp-edc-mcp', got '%s'", config.Server.Name)
	}
	
	if config.Server.MaxClients != 10 {
		t.Errorf("Expected default max clients 10, got %d", config.Server.MaxClients)
	}
	
	if config.Server.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", config.Server.Timeout)
	}
	
	if config.Security.RequireAuth {
		t.Error("Expected RequireAuth to be false by default")
	}
}

func TestLoadMCPConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.json")
	
	testConfig := &MCPConfig{
		Transport: TransportConfig{
			Type: "http",
			HTTP: HTTPConfig{
				Port: 9090,
				Host: "0.0.0.0",
			},
		},
		Server: ServerConfig{
			Name:       "test-server",
			Version:    "2.0.0",
			MaxClients: 5,
			Timeout:    60,
		},
		Security: SecurityConfig{
			RequireAuth: true,
			APIKeys:     []string{"test-key"},
		},
	}
	
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	// Load the config
	loadedConfig, err := LoadMCPConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify loaded values
	if loadedConfig.Transport.Type != "http" {
		t.Errorf("Expected transport type 'http', got '%s'", loadedConfig.Transport.Type)
	}
	
	if loadedConfig.Transport.HTTP.Port != 9090 {
		t.Errorf("Expected HTTP port 9090, got %d", loadedConfig.Transport.HTTP.Port)
	}
	
	if loadedConfig.Server.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", loadedConfig.Server.Name)
	}
	
	if !loadedConfig.Security.RequireAuth {
		t.Error("Expected RequireAuth to be true")
	}
	
	if len(loadedConfig.Security.APIKeys) != 1 || loadedConfig.Security.APIKeys[0] != "test-key" {
		t.Errorf("Expected API keys ['test-key'], got %v", loadedConfig.Security.APIKeys)
	}
}

func TestLoadMCPConfig_NonExistentFile(t *testing.T) {
	config, err := LoadMCPConfig("/non/existent/file.json")
	if err != nil {
		t.Fatalf("Expected default config for non-existent file, got error: %v", err)
	}
	
	// Should return default config
	if config.Transport.Type != "stdio" {
		t.Errorf("Expected default transport type 'stdio', got '%s'", config.Transport.Type)
	}
}

func TestSaveMCPConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "saved-config.json")
	
	config := DefaultMCPConfig()
	config.Server.Name = "saved-test"
	
	if err := SaveMCPConfig(config, configFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}
	
	// Load and verify
	loadedConfig, err := LoadMCPConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	if loadedConfig.Server.Name != "saved-test" {
		t.Errorf("Expected server name 'saved-test', got '%s'", loadedConfig.Server.Name)
	}
}

func TestValidateMCPConfig(t *testing.T) {
	testCases := []struct {
		name      string
		config    *MCPConfig
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid config",
			config:    DefaultMCPConfig(),
			expectErr: false,
		},
		{
			name: "Invalid transport type",
			config: &MCPConfig{
				Transport: TransportConfig{Type: "invalid"},
				Server:    ServerConfig{MaxClients: 1, Timeout: 1},
			},
			expectErr: true,
			errMsg:    "invalid transport type",
		},
		{
			name: "Invalid HTTP port",
			config: &MCPConfig{
				Transport: TransportConfig{
					Type: "http",
					HTTP: HTTPConfig{Port: 0},
				},
				Server: ServerConfig{MaxClients: 1, Timeout: 1},
			},
			expectErr: true,
			errMsg:    "invalid HTTP port",
		},
		{
			name: "HTTPS without cert",
			config: &MCPConfig{
				Transport: TransportConfig{
					Type: "http",
					HTTP: HTTPConfig{
						Port:        8080,
						EnableHTTPS: true,
						CertFile:    "",
					},
				},
				Server: ServerConfig{MaxClients: 1, Timeout: 1},
			},
			expectErr: true,
			errMsg:    "cert file required",
		},
		{
			name: "Invalid max clients",
			config: &MCPConfig{
				Transport: TransportConfig{Type: "stdio"},
				Server:    ServerConfig{MaxClients: 0, Timeout: 1},
			},
			expectErr: true,
			errMsg:    "max clients must be at least 1",
		},
		{
			name: "Invalid timeout",
			config: &MCPConfig{
				Transport: TransportConfig{Type: "stdio"},
				Server:    ServerConfig{MaxClients: 1, Timeout: 0},
			},
			expectErr: true,
			errMsg:    "timeout must be at least 1",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMCPConfig(tc.config)
			if tc.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tc.errMsg != "" && !contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
	       len(s) > len(substr) && contains(s[1:], substr)
}