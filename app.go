package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/asachs/smtp-edc/internal/services"
)

// App struct
type App struct {
	ctx             context.Context
	configService   *services.ConfigService
	smtpService     *services.SMTPService
	templateService *services.TemplateService
}

// NewApp creates a new App application struct
func NewApp() *App {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "smtp-edc-desktop"
	}

	return &App{
		configService:   services.NewConfigService(),
		smtpService:     services.NewSMTPService(hostname, false),
		templateService: services.NewTemplateService(),
	}
}

// startup is called when the app starts. The context provided
// is the same as the one passed to wails.Run
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "Hello, it's nice to meet you!"
	}
	return fmt.Sprintf("Hello %s, welcome to SMTP-EDC!", trimmedName)
}

// Configuration Methods

// LoadConfig loads configuration from file
func (a *App) LoadConfig(filename string) (*services.ConfigData, error) {
	return a.configService.LoadConfig(filename)
}

// SaveConfig saves configuration to file
func (a *App) SaveConfig(config *services.ConfigData, filename string) error {
	return a.configService.SaveConfig(config, filename)
}

// ValidateConfig validates configuration data
func (a *App) ValidateConfig(config *services.ConfigData) error {
	return a.configService.ValidateConfig(config)
}

// GetCurrentConfig returns the current configuration
func (a *App) GetCurrentConfig() *services.ConfigData {
	return a.configService.GetCurrentConfig()
}

// GetDefaultConfigPath returns the default config file path
func (a *App) GetDefaultConfigPath() string {
	return a.configService.GetDefaultConfigPath()
}

// ListConfigFiles lists available configuration files
func (a *App) ListConfigFiles() ([]string, error) {
	return a.configService.ListConfigFiles()
}

// SMTP Methods

// TestConnection tests SMTP server connection and capabilities
func (a *App) TestConnection(config *services.ConfigData) *services.ConnectionResult {
	return a.smtpService.TestConnection(config)
}

// SendEmail sends an email using the provided configuration and message
func (a *App) SendEmail(emailReq *services.EmailRequest) *services.SendResult {
	return a.smtpService.SendEmail(emailReq)
}

// ValidateEmailAddress validates a single email address
func (a *App) ValidateEmailAddress(email string) *services.SendResult {
	return a.smtpService.ValidateEmailAddress(email)
}

// ValidateEmailList validates a list of email addresses
func (a *App) ValidateEmailList(emails []string, validateMX bool) *services.SendResult {
	return a.smtpService.ValidateEmailList(emails, validateMX)
}

// GetAuthStats returns authentication statistics
func (a *App) GetAuthStats() map[string]interface{} {
	return a.smtpService.GetAuthStats()
}

// SetDebugMode enables or disables debug mode
func (a *App) SetDebugMode(debug bool) {
	a.smtpService.SetDebugMode(debug)
}

// GetDebugMode returns the current debug mode setting
func (a *App) GetDebugMode() bool {
	return a.smtpService.GetDebugMode()
}

// Template Methods

// ListTemplates lists available email templates
func (a *App) ListTemplates() ([]*services.TemplateInfo, error) {
	return a.templateService.ListTemplates()
}

// LoadTemplate loads and parses an email template
func (a *App) LoadTemplate(templateName string) (*services.TemplateResult, error) {
	return a.templateService.LoadTemplate(templateName)
}

// SaveTemplate saves a template to the template directory
func (a *App) SaveTemplate(templateName, content string) (*services.TemplateResult, error) {
	return a.templateService.SaveTemplate(templateName, content)
}

// DeleteTemplate deletes a template from the template directory
func (a *App) DeleteTemplate(templateName string) (*services.TemplateResult, error) {
	return a.templateService.DeleteTemplate(templateName)
}

// ExecuteTemplate processes a template with provided data
func (a *App) ExecuteTemplate(templateName, subjectTemplate string, templateData *services.TemplateData) (*services.EmailRequest, error) {
	return a.templateService.ExecuteTemplate(templateName, subjectTemplate, templateData)
}

// ValidateTemplate validates template syntax and variables
func (a *App) ValidateTemplate(content string) (*services.TemplateResult, error) {
	return a.templateService.ValidateTemplate(content)
}

// GetDefaultTemplates returns a list of default template examples
func (a *App) GetDefaultTemplates() map[string]string {
	return a.templateService.GetDefaultTemplates()
}

// CreateDefaultTemplates creates default template files
func (a *App) CreateDefaultTemplates() (*services.TemplateResult, error) {
	return a.templateService.CreateDefaultTemplates()
}

// SetTemplateDirectory sets the template directory path
func (a *App) SetTemplateDirectory(dir string) {
	a.templateService.SetTemplateDirectory(dir)
}

// GetTemplateDirectory returns the current template directory
func (a *App) GetTemplateDirectory() string {
	return a.templateService.GetTemplateDirectory()
}
