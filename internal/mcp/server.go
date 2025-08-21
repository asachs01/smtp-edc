package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/asachs/smtp-edc/internal/services"
)

// MCPServer represents the MCP server for SMTP-EDC
type MCPServer struct {
	configService   *services.ConfigService
	smtpService     *services.SMTPService
	templateService *services.TemplateService
	debug           bool
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(debug bool) *MCPServer {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "smtp-edc-mcp"
	}

	return &MCPServer{
		configService:   services.NewConfigService(),
		smtpService:     services.NewSMTPService(hostname, debug),
		templateService: services.NewTemplateService(),
		debug:           debug,
	}
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource definition
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// ListTools returns the available MCP tools
func (s *MCPServer) ListTools() []Tool {
	return []Tool{
		{
			Name:        "smtp_test_connection",
			Description: "Test SMTP server connection and capabilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"server": map[string]interface{}{
						"type":        "string",
						"description": "SMTP server hostname or IP",
					},
					"port": map[string]interface{}{
						"type":        "integer",
						"description": "SMTP server port",
						"default":     587,
					},
					"username": map[string]interface{}{
						"type":        "string",
						"description": "Authentication username",
					},
					"password": map[string]interface{}{
						"type":        "string",
						"description": "Authentication password",
					},
					"authType": map[string]interface{}{
						"type":        "string",
						"description": "Authentication type (plain, login, cram-md5)",
						"enum":        []string{"plain", "login", "cram-md5"},
						"default":     "plain",
					},
					"starttls": map[string]interface{}{
						"type":        "boolean",
						"description": "Use STARTTLS",
						"default":     false,
					},
					"skipVerify": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip TLS certificate verification",
						"default":     false,
					},
				},
				"required": []string{"server"},
			},
		},
		{
			Name:        "smtp_send_email",
			Description: "Send an email via SMTP",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"server": map[string]interface{}{
						"type":        "string",
						"description": "SMTP server hostname or IP",
					},
					"port": map[string]interface{}{
						"type":        "integer",
						"description": "SMTP server port",
						"default":     587,
					},
					"username": map[string]interface{}{
						"type":        "string",
						"description": "Authentication username",
					},
					"password": map[string]interface{}{
						"type":        "string",
						"description": "Authentication password",
					},
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Sender email address",
					},
					"to": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Recipient email addresses",
					},
					"cc": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "CC recipient email addresses",
					},
					"bcc": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "BCC recipient email addresses",
					},
					"subject": map[string]interface{}{
						"type":        "string",
						"description": "Email subject",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Email body (text or HTML)",
					},
					"isHTML": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the body is HTML",
						"default":     false,
					},
					"authType": map[string]interface{}{
						"type":        "string",
						"description": "Authentication type",
						"enum":        []string{"plain", "login", "cram-md5"},
						"default":     "plain",
					},
					"starttls": map[string]interface{}{
						"type":        "boolean",
						"description": "Use STARTTLS",
						"default":     false,
					},
					"skipVerify": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip TLS certificate verification",
						"default":     false,
					},
				},
				"required": []string{"server", "from", "to", "subject", "body"},
			},
		},
		{
			Name:        "smtp_validate_addresses",
			Description: "Validate email addresses and optionally check MX records",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"addresses": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Email addresses to validate",
					},
					"checkMX": map[string]interface{}{
						"type":        "boolean",
						"description": "Check MX records for domains",
						"default":     false,
					},
				},
				"required": []string{"addresses"},
			},
		},
		{
			Name:        "smtp_load_template",
			Description: "Load and process an email template",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"templateName": map[string]interface{}{
						"type":        "string",
						"description": "Name of the template to load",
					},
					"variables": map[string]interface{}{
						"type":        "object",
						"description": "Variables to substitute in the template",
					},
				},
				"required": []string{"templateName"},
			},
		},
	}
}

// ListResources returns the available MCP resources
func (s *MCPServer) ListResources() []Resource {
	return []Resource{
		{
			URI:         "smtp-edc://config/current",
			Name:        "Current Configuration",
			Description: "Current SMTP-EDC configuration settings",
			MimeType:    "application/json",
		},
		{
			URI:         "smtp-edc://templates/list",
			Name:        "Email Templates",
			Description: "Available email templates",
			MimeType:    "application/json",
		},
		{
			URI:         "smtp-edc://stats/auth",
			Name:        "Authentication Statistics",
			Description: "Authentication attempt statistics",
			MimeType:    "application/json",
		},
	}
}

// CallTool executes a tool with the given arguments
func (s *MCPServer) CallTool(ctx context.Context, name string, arguments json.RawMessage) (interface{}, error) {
	switch name {
	case "smtp_test_connection":
		return s.testConnection(ctx, arguments)
	case "smtp_send_email":
		return s.sendEmail(ctx, arguments)
	case "smtp_validate_addresses":
		return s.validateAddresses(ctx, arguments)
	case "smtp_load_template":
		return s.loadTemplate(ctx, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// ReadResource reads a resource by URI
func (s *MCPServer) ReadResource(ctx context.Context, uri string) (interface{}, error) {
	switch uri {
	case "smtp-edc://config/current":
		return s.configService.GetCurrentConfig(), nil
	case "smtp-edc://templates/list":
		return s.templateService.ListTemplates()
	case "smtp-edc://stats/auth":
		return s.smtpService.GetAuthStats(), nil
	default:
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}
}

// testConnection handles the smtp_test_connection tool
func (s *MCPServer) testConnection(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Server     string `json:"server"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		AuthType   string `json:"authType"`
		StartTLS   bool   `json:"starttls"`
		SkipVerify bool   `json:"skipVerify"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Set defaults
	if params.Port == 0 {
		params.Port = 587
	}
	if params.AuthType == "" {
		params.AuthType = "plain"
	}

	config := &services.ConfigData{
		Server:     params.Server,
		Port:       params.Port,
		Username:   params.Username,
		Password:   params.Password,
		AuthType:   params.AuthType,
		StartTLS:   params.StartTLS,
		SkipVerify: params.SkipVerify,
	}

	result := s.smtpService.TestConnection(config)
	return result, nil
}

// sendEmail handles the smtp_send_email tool
func (s *MCPServer) sendEmail(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Server     string   `json:"server"`
		Port       int      `json:"port"`
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		From       string   `json:"from"`
		To         []string `json:"to"`
		CC         []string `json:"cc"`
		BCC        []string `json:"bcc"`
		Subject    string   `json:"subject"`
		Body       string   `json:"body"`
		IsHTML     bool     `json:"isHTML"`
		AuthType   string   `json:"authType"`
		StartTLS   bool     `json:"starttls"`
		SkipVerify bool     `json:"skipVerify"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Set defaults
	if params.Port == 0 {
		params.Port = 587
	}
	if params.AuthType == "" {
		params.AuthType = "plain"
	}

	emailReq := &services.EmailRequest{
		Config: &services.ConfigData{
			Server:     params.Server,
			Port:       params.Port,
			Username:   params.Username,
			Password:   params.Password,
			AuthType:   params.AuthType,
			StartTLS:   params.StartTLS,
			SkipVerify: params.SkipVerify,
		},
		Message: &services.MessageData{
			From:    params.From,
			To:      params.To,
			CC:      params.CC,
			BCC:     params.BCC,
			Subject: params.Subject,
		},
	}

	if params.IsHTML {
		emailReq.Message.HTMLBody = params.Body
	} else {
		emailReq.Message.TextBody = params.Body
	}

	result := s.smtpService.SendEmail(emailReq)
	return result, nil
}

// validateAddresses handles the smtp_validate_addresses tool
func (s *MCPServer) validateAddresses(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Addresses []string `json:"addresses"`
		CheckMX   bool     `json:"checkMX"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	result := s.smtpService.ValidateEmailList(params.Addresses, params.CheckMX)
	return result, nil
}

// loadTemplate handles the smtp_load_template tool
func (s *MCPServer) loadTemplate(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		TemplateName string                 `json:"templateName"`
		Variables    map[string]interface{} `json:"variables"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Load the template
	templateResult, err := s.templateService.LoadTemplate(params.TemplateName)
	if err != nil {
		return nil, err
	}

	// If variables provided, execute the template
	if params.Variables != nil {
		templateData := &services.TemplateData{
			Variables: params.Variables,
		}
		emailReq, err := s.templateService.ExecuteTemplate(params.TemplateName, "", templateData)
		if err != nil {
			return nil, err
		}
		return emailReq, nil
	}

	return templateResult, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(transport string) error {
	if s.debug {
		log.Println("Starting SMTP-EDC MCP Server...")
	}

	switch strings.ToLower(transport) {
	case "stdio":
		return s.startSTDIO()
	case "http":
		return s.startHTTP()
	default:
		return fmt.Errorf("unsupported transport: %s", transport)
	}
}

// startSTDIO starts the server with STDIO transport
func (s *MCPServer) startSTDIO() error {
	// Implementation will use the MCP SDK once properly integrated
	// For now, this is a placeholder
	if s.debug {
		log.Println("MCP Server running on STDIO transport")
	}
	
	// The actual implementation would use the MCP SDK's STDIO transport
	// This is a simplified version for the initial implementation
	return fmt.Errorf("STDIO transport implementation pending MCP SDK integration")
}

// startHTTP starts the server with HTTP transport
func (s *MCPServer) startHTTP() error {
	// Implementation will use the MCP SDK once properly integrated
	// For now, this is a placeholder
	if s.debug {
		log.Println("MCP Server running on HTTP transport")
	}
	
	// The actual implementation would use the MCP SDK's HTTP transport
	// This is a simplified version for the initial implementation
	return fmt.Errorf("HTTP transport implementation pending MCP SDK integration")
}