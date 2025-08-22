package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asachs/smtp-edc/internal/services"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Parameter structs for each tool
type TestConnectionParams struct {
	Server     string `json:"server" jsonschema:"required,description=SMTP server hostname or IP"`
	Port       int    `json:"port,omitempty" jsonschema:"description=SMTP server port (default 587)"`
	Username   string `json:"username,omitempty" jsonschema:"description=Authentication username"`
	Password   string `json:"password,omitempty" jsonschema:"description=Authentication password"`
	AuthType   string `json:"authType,omitempty" jsonschema:"enum=plain;login;cram-md5,description=Authentication type"`
	StartTLS   bool   `json:"starttls,omitempty" jsonschema:"description=Use STARTTLS"`
	SkipVerify bool   `json:"skipVerify,omitempty" jsonschema:"description=Skip TLS certificate verification"`
}

type SendEmailParams struct {
	Server     string   `json:"server" jsonschema:"required,description=SMTP server hostname or IP"`
	Port       int      `json:"port,omitempty" jsonschema:"description=SMTP server port (default 587)"`
	Username   string   `json:"username,omitempty" jsonschema:"description=Authentication username"`
	Password   string   `json:"password,omitempty" jsonschema:"description=Authentication password"`
	From       string   `json:"from" jsonschema:"required,description=Sender email address"`
	To         []string `json:"to" jsonschema:"required,description=Recipient email addresses"`
	CC         []string `json:"cc,omitempty" jsonschema:"description=CC recipients"`
	BCC        []string `json:"bcc,omitempty" jsonschema:"description=BCC recipients"`
	Subject    string   `json:"subject" jsonschema:"required,description=Email subject"`
	Body       string   `json:"body" jsonschema:"required,description=Email body content"`
	IsHTML     bool     `json:"isHTML,omitempty" jsonschema:"description=Whether the body is HTML"`
	AuthType   string   `json:"authType,omitempty" jsonschema:"enum=plain;login;cram-md5,description=Authentication type"`
	StartTLS   bool     `json:"starttls,omitempty" jsonschema:"description=Use STARTTLS"`
	SkipVerify bool     `json:"skipVerify,omitempty" jsonschema:"description=Skip TLS certificate verification"`
}

type ValidateAddressesParams struct {
	Addresses []string `json:"addresses" jsonschema:"required,description=Email addresses to validate"`
	CheckMX   bool     `json:"checkMX,omitempty" jsonschema:"description=Check MX records for domains"`
}

type LoadTemplateParams struct {
	TemplateName string                 `json:"templateName" jsonschema:"required,description=Name of the template to load"`
	Variables    map[string]interface{} `json:"variables,omitempty" jsonschema:"description=Variables to substitute in the template"`
}

// CreateMCPServer creates and configures the MCP server with all tools
func CreateMCPServer(debug bool) *mcpsdk.Server {
	// Create server instance
	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    "smtp-edc",
			Version: "1.0.0",
		},
		&mcpsdk.ServerOptions{
			Prompts:   nil,
			Resources: nil,
			Tools:     nil,
		},
	)

	// Create service instances
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "smtp-edc-mcp"
	}
	configService := services.NewConfigService()
	smtpService := services.NewSMTPService(hostname, debug)
	templateService := services.NewTemplateService()

	// Register tools
	mcpsdk.AddTool(server, 
		&mcpsdk.Tool{
			Name:        "smtp_test_connection",
			Description: "Test SMTP server connection and capabilities",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, params TestConnectionParams) (*mcpsdk.CallToolResult, any, error) {
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

			result := smtpService.TestConnection(config)
			
			// Convert result to JSON string for response
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil, nil
		},
	)

	mcpsdk.AddTool(server,
		&mcpsdk.Tool{
			Name:        "smtp_send_email",
			Description: "Send an email via SMTP",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, params SendEmailParams) (*mcpsdk.CallToolResult, any, error) {
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
					Body:    params.Body,
					IsHTML:  params.IsHTML,
				},
			}

			result := smtpService.SendEmail(emailReq)
			
			// Convert result to JSON string
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil, nil
		},
	)

	mcpsdk.AddTool(server,
		&mcpsdk.Tool{
			Name:        "smtp_validate_addresses",
			Description: "Validate email addresses with optional MX record checking",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, params ValidateAddressesParams) (*mcpsdk.CallToolResult, any, error) {
			results := smtpService.ValidateAddresses(params.Addresses, params.CheckMX)
			
			// Convert results to JSON string
			resultJSON, err := json.Marshal(results)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal results: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil, nil
		},
	)

	mcpsdk.AddTool(server,
		&mcpsdk.Tool{
			Name:        "smtp_load_template",
			Description: "Load and process email templates",
		},
		func(ctx context.Context, req *mcpsdk.CallToolRequest, params LoadTemplateParams) (*mcpsdk.CallToolResult, any, error) {
			content, err := templateService.LoadTemplate(params.TemplateName, params.Variables)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load template: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: content,
					},
				},
			}, nil, nil
		},
	)

	// Add resources
	mcpsdk.AddResource(server,
		&mcpsdk.Resource{
			URI:         "smtp-edc://config/current",
			Name:        "Current Configuration",
			Description: "Current SMTP-EDC configuration settings",
			MimeType:    "application/json",
		},
		func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			config := configService.GetCurrentConfig()
			configJSON, err := json.Marshal(config)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal config: %w", err)
			}

			return &mcpsdk.ReadResourceResult{
				Contents: []mcpsdk.ResourceContents{
					&mcpsdk.TextResourceContents{
						URI:      req.URI,
						MimeType: "application/json",
						Text:     string(configJSON),
					},
				},
			}, nil
		},
	)

	mcpsdk.AddResource(server,
		&mcpsdk.Resource{
			URI:         "smtp-edc://templates/list",
			Name:        "Email Templates",
			Description: "Available email templates",
			MimeType:    "application/json",
		},
		func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			templates, err := templateService.ListTemplates()
			if err != nil {
				return nil, fmt.Errorf("failed to list templates: %w", err)
			}

			templatesJSON, err := json.Marshal(templates)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal templates: %w", err)
			}

			return &mcpsdk.ReadResourceResult{
				Contents: []mcpsdk.ResourceContents{
					&mcpsdk.TextResourceContents{
						URI:      req.URI,
						MimeType: "application/json",
						Text:     string(templatesJSON),
					},
				},
			}, nil
		},
	)

	mcpsdk.AddResource(server,
		&mcpsdk.Resource{
			URI:         "smtp-edc://stats/auth",
			Name:        "Authentication Statistics",
			Description: "Authentication attempt statistics",
			MimeType:    "application/json",
		},
		func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			stats := smtpService.GetAuthStats()
			statsJSON, err := json.Marshal(stats)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal stats: %w", err)
			}

			return &mcpsdk.ReadResourceResult{
				Contents: []mcpsdk.ResourceContents{
					&mcpsdk.TextResourceContents{
						URI:      req.URI,
						MimeType: "application/json",
						Text:     string(statsJSON),
					},
				},
			}, nil
		},
	)

	return server
}