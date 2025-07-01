package services

import (
	"fmt"
	"time"

	"github.com/asachs/smtp-edc/internal/auth"
	"github.com/asachs/smtp-edc/internal/client"
	"github.com/asachs/smtp-edc/internal/config"
	"github.com/asachs/smtp-edc/internal/message"
	"github.com/asachs/smtp-edc/internal/security"
)

// SMTPService provides Wails-compatible SMTP operations
type SMTPService struct {
	hostname  string
	debug     bool
	authMgr   *auth.AuthManager
	credStore *security.CredentialStore
}

// ConnectionResult represents the result of a connection test
type ConnectionResult struct {
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
	Error        string          `json:"error,omitempty"`
	Timestamp    string          `json:"timestamp"`
	ServerInfo   *ServerInfo     `json:"serverInfo,omitempty"`
	Capabilities *CapabilityInfo `json:"capabilities,omitempty"`
}

// ServerInfo contains information about the SMTP server
type ServerInfo struct {
	Server    string   `json:"server"`
	Port      int      `json:"port"`
	TLSActive bool     `json:"tlsActive"`
	AuthTypes []string `json:"authTypes"`
}

// CapabilityInfo contains server capabilities
type CapabilityInfo struct {
	Pipelining bool     `json:"pipelining"`
	StartTLS   bool     `json:"startTLS"`
	Auth       []string `json:"auth"`
	Size       int      `json:"size"`
	EightBit   bool     `json:"eightBit"`
}

// EmailRequest represents an email sending request
type EmailRequest struct {
	Config      *ConfigData       `json:"config"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc"`
	Bcc         []string          `json:"bcc"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	HTMLBody    string            `json:"htmlBody"`
	Attachments []string          `json:"attachments"`
	Headers     map[string]string `json:"headers"`
}

// SendResult represents the result of an email send operation
type SendResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
	MessageID string `json:"messageId,omitempty"`
}

// NewSMTPService creates a new SMTP service
func NewSMTPService(hostname string, debug bool) *SMTPService {
	credStore, _ := security.NewCredentialStore("/tmp/smtp-edc-credentials.db")
	authMgr := auth.NewAuthManager(credStore)

	return &SMTPService{
		hostname:  hostname,
		debug:     debug,
		authMgr:   authMgr,
		credStore: credStore,
	}
}

// TestConnection tests SMTP server connection and capabilities
func (ss *SMTPService) TestConnection(configData *ConfigData) *ConnectionResult {
	if configData == nil {
		return &ConnectionResult{
			Success:   false,
			Error:     "configuration is required",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	cfg := ss.configDataToSMTPConfig(configData)
	client := client.NewSMTPClient(ss.hostname, ss.debug)

	// Set timeouts and retry configuration
	client.SetTimeout(time.Second * 30)
	client.SetRetryConfig(3, time.Second*2)

	// Test connection
	err := client.Connect(cfg.Server, cfg.Port)
	if err != nil {
		return &ConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("Connection failed: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}
	defer client.Close()

	// Perform EHLO to get capabilities
	err = client.Ehlo()
	if err != nil {
		return &ConnectionResult{
			Success:   false,
			Error:     fmt.Sprintf("EHLO failed: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Start TLS if required
	tlsActive := false
	if cfg.StartTLS {
		err = client.StartTLS()
		if err != nil {
			return &ConnectionResult{
				Success:   false,
				Error:     fmt.Sprintf("STARTTLS failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
		tlsActive = true

		// Re-run EHLO after TLS
		err = client.Ehlo()
		if err != nil {
			return &ConnectionResult{
				Success:   false,
				Error:     fmt.Sprintf("EHLO after TLS failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}

	// Test authentication if credentials provided
	if cfg.Username != "" && cfg.Password != "" {
		err = client.Authenticate(cfg.AuthType, cfg.Username, cfg.Password)
		if err != nil {
			return &ConnectionResult{
				Success:   false,
				Error:     fmt.Sprintf("Authentication failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}

	return &ConnectionResult{
		Success:   true,
		Message:   "Connection successful",
		Timestamp: time.Now().Format(time.RFC3339),
		ServerInfo: &ServerInfo{
			Server:    cfg.Server,
			Port:      cfg.Port,
			TLSActive: tlsActive,
			AuthTypes: []string{cfg.AuthType},
		},
		Capabilities: &CapabilityInfo{
			Pipelining: false, // TODO: Extract from client capabilities
			StartTLS:   cfg.StartTLS,
			Auth:       []string{cfg.AuthType},
			Size:       0,     // TODO: Extract from client capabilities
			EightBit:   false, // TODO: Extract from client capabilities
		},
	}
}

// SendEmail sends an email using the provided configuration and message
func (ss *SMTPService) SendEmail(emailReq *EmailRequest) *SendResult {
	if emailReq == nil || emailReq.Config == nil {
		return &SendResult{
			Success:   false,
			Error:     "email request and configuration are required",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	cfg := ss.configDataToSMTPConfig(emailReq.Config)
	client := client.NewSMTPClient(ss.hostname, ss.debug)

	// Set timeouts and retry configuration
	client.SetTimeout(time.Second * 30)
	client.SetRetryConfig(3, time.Second*2)

	// Connect to server
	err := client.Connect(cfg.Server, cfg.Port)
	if err != nil {
		return &SendResult{
			Success:   false,
			Error:     fmt.Sprintf("Connection failed: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}
	defer client.Close()

	// Start TLS if required
	if cfg.StartTLS {
		err = client.StartTLS()
		if err != nil {
			return &SendResult{
				Success:   false,
				Error:     fmt.Sprintf("STARTTLS failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}

	// Authenticate if credentials provided
	if cfg.Username != "" && cfg.Password != "" {
		err = client.Authenticate(cfg.AuthType, cfg.Username, cfg.Password)
		if err != nil {
			return &SendResult{
				Success:   false,
				Error:     fmt.Sprintf("Authentication failed: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}

	// Create message
	msg := message.NewMessage(emailReq.From, emailReq.To, emailReq.Subject, emailReq.Body)

	// Set additional recipients
	for _, cc := range emailReq.Cc {
		msg.AddCc(cc)
	}
	for _, bcc := range emailReq.Bcc {
		msg.AddBcc(bcc)
	}

	// Set HTML body if provided
	if emailReq.HTMLBody != "" {
		msg.SetHTMLBody(emailReq.HTMLBody)
	}

	// Add custom headers
	for key, value := range emailReq.Headers {
		msg.AddHeader(key, value)
	}

	// Add attachments
	for _, attachment := range emailReq.Attachments {
		if err := msg.AddAttachment(attachment); err != nil {
			return &SendResult{
				Success:   false,
				Error:     fmt.Sprintf("Failed to add attachment %s: %v", attachment, err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
		}
	}

	// Validate message
	if err := msg.Validate(); err != nil {
		return &SendResult{
			Success:   false,
			Error:     fmt.Sprintf("Message validation failed: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Send email
	err = client.SendMessage(msg)
	if err != nil {
		return &SendResult{
			Success:   false,
			Error:     fmt.Sprintf("Failed to send email: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return &SendResult{
		Success:   true,
		Message:   "Email sent successfully",
		Timestamp: time.Now().Format(time.RFC3339),
		MessageID: fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), ss.hostname),
	}
}

// ValidateEmailAddress validates a single email address
func (ss *SMTPService) ValidateEmailAddress(email string) *SendResult {
	err := message.ValidateEmail(email)
	if err != nil {
		return &SendResult{
			Success:   false,
			Error:     fmt.Sprintf("Invalid email address: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return &SendResult{
		Success:   true,
		Message:   "Email address is valid",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// ValidateEmailList validates a list of email addresses
func (ss *SMTPService) ValidateEmailList(emails []string, validateMX bool) *SendResult {
	err := message.ValidateAddressList(emails, validateMX)
	if err != nil {
		return &SendResult{
			Success:   false,
			Error:     fmt.Sprintf("Invalid email addresses: %v", err),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return &SendResult{
		Success:   true,
		Message:   fmt.Sprintf("All %d email addresses are valid", len(emails)),
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// GetAuthStats returns authentication statistics
func (ss *SMTPService) GetAuthStats() map[string]interface{} {
	return ss.authMgr.GetAuthStats()
}

// SetDebugMode enables or disables debug mode
func (ss *SMTPService) SetDebugMode(debug bool) {
	ss.debug = debug
}

// GetDebugMode returns the current debug mode setting
func (ss *SMTPService) GetDebugMode() bool {
	return ss.debug
}

// configDataToSMTPConfig converts ConfigData to internal SMTPConfig
func (ss *SMTPService) configDataToSMTPConfig(data *ConfigData) *config.SMTPConfig {
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
