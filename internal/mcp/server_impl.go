package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asachs/smtp-edc/internal/message"
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
		&mcpsdk.ServerOptions{},
	)

	// Create service instances
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "smtp-edc-mcp"
	}
	configService := services.NewConfigService()
	smtpService := services.NewSMTPService(hostname, debug)
	templateService := services.NewTemplateService()

	// Register tools using a simpler approach with proper handler
	server.AddTool(
		&mcpsdk.Tool{
			Name:        "smtp_test_connection",
			Description: "Test SMTP server connection and capabilities",
			InputSchema: TestConnectionParams{},
		},
		mcpsdk.ToolHandlerFunc(func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			var params TestConnectionParams
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return nil, fmt.Errorf("failed to parse parameters: %w", err)
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

			result := smtpService.TestConnection(config)
			
			// Convert result to JSON string for response
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil
		}),
	)

	server.AddTool(
		&mcpsdk.Tool{
			Name:        "smtp_send_email",
			Description: "Send an email via SMTP",
			InputSchema: SendEmailParams{},
		},
		mcpsdk.ToolHandlerFunc(func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			var params SendEmailParams
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return nil, fmt.Errorf("failed to parse parameters: %w", err)
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
				From:    params.From,
				To:      params.To,
				Cc:      params.CC,
				Bcc:     params.BCC,
				Subject: params.Subject,
				Body:    params.Body,
				HTMLBody: "",
			}

			if params.IsHTML {
				emailReq.HTMLBody = params.Body
				emailReq.Body = ""
			}

			result := smtpService.SendEmail(emailReq)
			
			// Convert result to JSON string
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil
		}),
	)

	server.AddTool(
		&mcpsdk.Tool{
			Name:        "smtp_validate_addresses",
			Description: "Validate email addresses with optional MX record checking",
			InputSchema: ValidateAddressesParams{},
		},
		mcpsdk.ToolHandlerFunc(func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			var params ValidateAddressesParams
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return nil, fmt.Errorf("failed to parse parameters: %w", err)
			}

			// Validate each address
			results := make(map[string]interface{})
			for _, addr := range params.Addresses {
				err := message.ValidateEmail(addr)
				if err != nil {
					results[addr] = map[string]interface{}{
						"valid": false,
						"error": err.Error(),
					}
				} else {
					results[addr] = map[string]interface{}{
						"valid": true,
					}
				}
				// TODO: Add MX record checking if params.CheckMX is true
			}
			
			// Convert results to JSON string
			resultJSON, err := json.Marshal(results)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal results: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil
		}),
	)

	server.AddTool(
		&mcpsdk.Tool{
			Name:        "smtp_load_template",
			Description: "Load and process email templates",
			InputSchema: LoadTemplateParams{},
		},
		mcpsdk.ToolHandlerFunc(func(ctx context.Context, request mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			var params LoadTemplateParams
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return nil, fmt.Errorf("failed to parse parameters: %w", err)
			}

			// Template loading simplified - just return template name for now
			// TODO: Implement proper template loading with variables
			result := map[string]interface{}{
				"template": params.TemplateName,
				"variables": params.Variables,
				"message": "Template loading not yet implemented",
			}
			
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			return &mcpsdk.CallToolResult{
				Content: []mcpsdk.Content{
					&mcpsdk.TextContent{
						Text: string(resultJSON),
					},
				},
			}, nil
		}),
	)

	// Add resources with proper handler functions
	server.AddResource(
		&mcpsdk.Resource{
			URI:         "smtp-edc://config/current",
			Name:        "Current Configuration",
			Description: "Current SMTP-EDC configuration settings",
			MimeType:    "application/json",
		},
		mcpsdk.ResourceHandlerFunc(func(ctx context.Context, request mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			config := configService.GetCurrentConfig()
			configJSON, err := json.Marshal(config)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal config: %w", err)
			}

			return &mcpsdk.ReadResourceResult{
				Contents: []mcpsdk.ResourceContents{
					&mcpsdk.TextResourceContents{
						URI:      request.URI,
						MimeType: "application/json",
						Text:     string(configJSON),
					},
				},
			}, nil
		}),
	)

	server.AddResource(
		&mcpsdk.Resource{
			URI:         "smtp-edc://templates/list",
			Name:        "Email Templates",
			Description: "Available email templates",
			MimeType:    "application/json",
		},
		mcpsdk.ResourceHandlerFunc(func(ctx context.Context, request mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
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
						URI:      request.URI,
						MimeType: "application/json",
						Text:     string(templatesJSON),
					},
				},
			}, nil
		}),
	)

	server.AddResource(
		&mcpsdk.Resource{
			URI:         "smtp-edc://stats/auth",
			Name:        "Authentication Statistics",
			Description: "Authentication attempt statistics",
			MimeType:    "application/json",
		},
		mcpsdk.ResourceHandlerFunc(func(ctx context.Context, request mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			stats := smtpService.GetAuthStats()
			statsJSON, err := json.Marshal(stats)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal stats: %w", err)
			}

			return &mcpsdk.ReadResourceResult{
				Contents: []mcpsdk.ResourceContents{
					&mcpsdk.TextResourceContents{
						URI:      request.URI,
						MimeType: "application/json",
						Text:     string(statsJSON),
					},
				},
			}, nil
		}),
	)

	return server
}