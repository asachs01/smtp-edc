package main

import (
	"context"
	"fmt"

	"github.com/asachs/smtp-edc/internal/client"
	"github.com/asachs/smtp-edc/internal/config"
	"github.com/asachs/smtp-edc/internal/message"
)

// App struct
type App struct {
	ctx    context.Context
	client *client.SMTPClient
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SMTPConnectionConfig represents SMTP connection configuration
type SMTPConnectionConfig struct {
	Server     string `json:"server"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	AuthType   string `json:"authType"`
	StartTLS   bool   `json:"startTLS"`
	SkipVerify bool   `json:"skipVerify"`
}

// EmailMessage represents an email message
type EmailMessage struct {
	From        string   `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	HTMLBody    string   `json:"htmlBody"`
	Attachments []string `json:"attachments"`
}

// TestResult represents the result of an SMTP test
type TestResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

// TestConnection tests the SMTP connection with given configuration
func (a *App) TestConnection(connectionConfig SMTPConnectionConfig) TestResult {
	smtpConfig := &config.SMTPConfig{
		Server:     connectionConfig.Server,
		Port:       connectionConfig.Port,
		Username:   connectionConfig.Username,
		Password:   connectionConfig.Password,
		AuthType:   connectionConfig.AuthType,
		StartTLS:   connectionConfig.StartTLS,
		SkipVerify: connectionConfig.SkipVerify,
	}

	client := client.NewSMTPClient("localhost", false) // hostname, debug

	// Test the connection
	err := client.Connect(smtpConfig.Server, smtpConfig.Port)
	if err != nil {
		return TestResult{
			Success: false,
			Error:   fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer client.Close()

	return TestResult{
		Success: true,
		Message: "Connection successful",
	}
}

// SendEmail sends an email using the configured SMTP settings
func (a *App) SendEmail(smtpConfig SMTPConnectionConfig, email EmailMessage) TestResult {
	cfg := &config.SMTPConfig{
		Server:     smtpConfig.Server,
		Port:       smtpConfig.Port,
		Username:   smtpConfig.Username,
		Password:   smtpConfig.Password,
		AuthType:   smtpConfig.AuthType,
		StartTLS:   smtpConfig.StartTLS,
		SkipVerify: smtpConfig.SkipVerify,
	}

	client := client.NewSMTPClient("localhost", false) // hostname, debug

	// Connect to server
	err := client.Connect(cfg.Server, cfg.Port)
	if err != nil {
		return TestResult{
			Success: false,
			Error:   fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer client.Close()

	// Create message
	msg := &message.Message{
		From:     email.From,
		To:       email.To,
		Cc:       email.Cc,
		Bcc:      email.Bcc,
		Subject:  email.Subject,
		Body:     email.Body,
		HTMLBody: email.HTMLBody,
	}

	// Add attachments if any
	for _, attachment := range email.Attachments {
		if err := msg.AddAttachment(attachment); err != nil {
			return TestResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to add attachment %s: %v", attachment, err),
			}
		}
	}

	// Send email
	err = client.SendMessage(msg)
	if err != nil {
		return TestResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to send email: %v", err),
		}
	}

	return TestResult{
		Success: true,
		Message: "Email sent successfully",
	}
}

// ValidateEmailAddress validates an email address
func (a *App) ValidateEmailAddress(email string) TestResult {
	err := message.ValidateEmail(email)
	if err != nil {
		return TestResult{
			Success: false,
			Error:   fmt.Sprintf("Invalid email address: %v", err),
		}
	}

	return TestResult{
		Success: true,
		Message: "Email address is valid",
	}
}
