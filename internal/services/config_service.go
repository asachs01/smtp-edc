package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asachs/smtp-edc/internal/config"
)

// ConfigService provides Wails-compatible configuration management
type ConfigService struct {
	defaultConfigPath string
	currentConfig     *config.SMTPConfig
}

// ConfigData represents configuration data for frontend binding
type ConfigData struct {
	Server     string            `json:"server"`
	Port       int               `json:"port"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	AuthType   string            `json:"authType"`
	StartTLS   bool              `json:"startTLS"`
	SkipVerify bool              `json:"skipVerify"`
	Templates  map[string]string `json:"templates"`
}

// NewConfigService creates a new configuration service
func NewConfigService() *ConfigService {
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, ".smtp-edc", "config.yaml")

	return &ConfigService{
		defaultConfigPath: defaultPath,
		currentConfig:     &config.SMTPConfig{},
	}
}

// LoadConfig loads configuration from a file
func (cs *ConfigService) LoadConfig(filename string) (*ConfigData, error) {
	if filename == "" {
		filename = cs.defaultConfigPath
	}

	cfg, err := config.LoadConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	cs.currentConfig = cfg
	return cs.configToData(cfg), nil
}

// SaveConfig saves configuration to a file
func (cs *ConfigService) SaveConfig(configData *ConfigData, filename string) error {
	if filename == "" {
		filename = cs.defaultConfigPath
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	cfg := cs.dataToConfig(configData)
	if err := config.SaveConfig(cfg, filename); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	cs.currentConfig = cfg
	return nil
}

// ValidateConfig validates configuration data
func (cs *ConfigService) ValidateConfig(configData *ConfigData) error {
	cfg := cs.dataToConfig(configData)
	return cfg.Validate()
}

// GetCurrentConfig returns the current configuration
func (cs *ConfigService) GetCurrentConfig() *ConfigData {
	return cs.configToData(cs.currentConfig)
}

// GetDefaultConfigPath returns the default config file path
func (cs *ConfigService) GetDefaultConfigPath() string {
	return cs.defaultConfigPath
}

// SetDefaultConfigPath sets the default config file path
func (cs *ConfigService) SetDefaultConfigPath(path string) {
	cs.defaultConfigPath = path
}

// ListConfigFiles lists available configuration files
func (cs *ConfigService) ListConfigFiles() ([]string, error) {
	configDir := filepath.Dir(cs.defaultConfigPath)

	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read config directory: %v", err)
	}

	var configFiles []string
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml") {
			configFiles = append(configFiles, filepath.Join(configDir, file.Name()))
		}
	}

	return configFiles, nil
}

// ExportConfigAsJSON exports configuration as JSON string
func (cs *ConfigService) ExportConfigAsJSON() (string, error) {
	data := cs.configToData(cs.currentConfig)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %v", err)
	}
	return string(jsonData), nil
}

// ImportConfigFromJSON imports configuration from JSON string
func (cs *ConfigService) ImportConfigFromJSON(jsonStr string) (*ConfigData, error) {
	var configData ConfigData
	if err := json.Unmarshal([]byte(jsonStr), &configData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON config: %v", err)
	}

	if err := cs.ValidateConfig(&configData); err != nil {
		return nil, fmt.Errorf("invalid config data: %v", err)
	}

	cs.currentConfig = cs.dataToConfig(&configData)
	return &configData, nil
}

// configToData converts internal config to frontend-compatible data
func (cs *ConfigService) configToData(cfg *config.SMTPConfig) *ConfigData {
	if cfg == nil {
		return &ConfigData{
			Port:       25,
			AuthType:   "plain",
			StartTLS:   false,
			SkipVerify: false,
			Templates:  make(map[string]string),
		}
	}

	return &ConfigData{
		Server:     cfg.Server,
		Port:       cfg.Port,
		Username:   cfg.Username,
		Password:   cfg.Password,
		AuthType:   cfg.AuthType,
		StartTLS:   cfg.StartTLS,
		SkipVerify: cfg.SkipVerify,
		Templates:  cfg.Templates,
	}
}

// dataToConfig converts frontend data to internal config
func (cs *ConfigService) dataToConfig(data *ConfigData) *config.SMTPConfig {
	if data == nil {
		return &config.SMTPConfig{}
	}

	return &config.SMTPConfig{
		Server:     data.Server,
		Port:       data.Port,
		Username:   data.Username,
		Password:   data.Password,
		AuthType:   data.AuthType,
		StartTLS:   data.StartTLS,
		SkipVerify: data.SkipVerify,
		Templates:  data.Templates,
	}
}
