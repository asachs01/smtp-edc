package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/asachs/smtp-edc/internal/message"
	"github.com/asachs/smtp-edc/internal/services"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Parameter structs for each tool
type TestConnectionArgs struct {
	Server     string `json:"server" jsonschema:"description=SMTP server hostname or IP,required"`
	Port       int    `json:"port,omitempty" jsonschema:"description=SMTP server port (default 587)"`
	Username   string `json:"username,omitempty" jsonschema:"description=Authentication username"`
	Password   string `json:"password,omitempty" jsonschema:"description=Authentication password"`
	AuthType   string `json:"authType,omitempty" jsonschema:"description=Authentication type,enum=plain,enum=login,enum=cram-md5"`
	StartTLS   bool   `json:"starttls,omitempty" jsonschema:"description=Use STARTTLS"`
	SkipVerify bool   `json:"skipVerify,omitempty" jsonschema:"description=Skip TLS certificate verification"`
}

type SendEmailArgs struct {
	Server     string   `json:"server" jsonschema:"description=SMTP server hostname or IP,required"`
	Port       int      `json:"port,omitempty" jsonschema:"description=SMTP server port (default 587)"`
	Username   string   `json:"username,omitempty" jsonschema:"description=Authentication username"`
	Password   string   `json:"password,omitempty" jsonschema:"description=Authentication password"`
	From       string   `json:"from" jsonschema:"description=Sender email address,required"`
	To         []string `json:"to" jsonschema:"description=Recipient email addresses,required"`
	CC         []string `json:"cc,omitempty" jsonschema:"description=CC recipients"`
	BCC        []string `json:"bcc,omitempty" jsonschema:"description=BCC recipients"`
	Subject    string   `json:"subject" jsonschema:"description=Email subject,required"`
	Body       string   `json:"body" jsonschema:"description=Email body content,required"`
	IsHTML     bool     `json:"isHTML,omitempty" jsonschema:"description=Whether the body is HTML"`
	AuthType   string   `json:"authType,omitempty" jsonschema:"description=Authentication type,enum=plain,enum=login,enum=cram-md5"`
	StartTLS   bool     `json:"starttls,omitempty" jsonschema:"description=Use STARTTLS"`
	SkipVerify bool     `json:"skipVerify,omitempty" jsonschema:"description=Skip TLS certificate verification"`
}

type ValidateAddressesArgs struct {
	Addresses []string `json:"addresses" jsonschema:"description=Email addresses to validate,required"`
	CheckMX   bool     `json:"checkMX,omitempty" jsonschema:"description=Check MX records for domains"`
}

type LoadTemplateArgs struct {
	TemplateName string                 `json:"templateName" jsonschema:"description=Name of the template to load,required"`
	Variables    map[string]interface{} `json:"variables,omitempty" jsonschema:"description=Variables to substitute in the template"`
}

// Service instances (package level for access by handlers)
var (
	configService   *services.ConfigService
	smtpService     *services.SMTPService
	templateService *services.TemplateService
)

// CreateMCPServer creates and configures the MCP server with all tools
func CreateMCPServer(debug bool) *mcp.Server {
	// Create server instance
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "smtp-edc",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	// Initialize service instances
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "smtp-edc-mcp"
	}
	configService = services.NewConfigService()
	smtpService = services.NewSMTPService(hostname, debug)
	templateService = services.NewTemplateService()

	// Register tools using the correct API pattern
	mcp.AddTool(server, 
		&mcp.Tool{
			Name:        "smtp_test_connection",
			Description: "Test SMTP server connection and capabilities",
		},
		testConnectionHandler,
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "smtp_send_email",
			Description: "Send an email via SMTP",
		},
		sendEmailHandler,
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "smtp_validate_addresses",
			Description: "Validate email addresses with optional MX record checking",
		},
		validateAddressesHandler,
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "smtp_load_template",
			Description: "Load and process email templates",
		},
		loadTemplateHandler,
	)

	// Add resources
	server.AddResource(
		&mcp.Resource{
			URI:         "smtp-edc://config/current",
			Name:        "Current Configuration",
			Description: "Current SMTP-EDC configuration settings",
			MIMEType:    "application/json",
		},
		configResourceHandler,
	)

	server.AddResource(
		&mcp.Resource{
			URI:         "smtp-edc://templates/list",
			Name:        "Email Templates",
			Description: "Available email templates",
			MIMEType:    "application/json",
		},
		templatesResourceHandler,
	)

	server.AddResource(
		&mcp.Resource{
			URI:         "smtp-edc://stats/auth",
			Name:        "Authentication Statistics",
			Description: "Authentication attempt statistics",
			MIMEType:    "application/json",
		},
		authStatsResourceHandler,
	)

	return server
}

// Tool handler functions
func testConnectionHandler(ctx context.Context, req *mcp.CallToolRequest, args TestConnectionArgs) (*mcp.CallToolResult, any, error) {
	// Set defaults
	if args.Port == 0 {
		args.Port = 587
	}
	if args.AuthType == "" {
		args.AuthType = "plain"
	}

	config := &services.ConfigData{
		Server:     args.Server,
		Port:       args.Port,
		Username:   args.Username,
		Password:   args.Password,
		AuthType:   args.AuthType,
		StartTLS:   args.StartTLS,
		SkipVerify: args.SkipVerify,
	}

	result := smtpService.TestConnection(config)
	
	// Convert result to JSON string for response
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func sendEmailHandler(ctx context.Context, req *mcp.CallToolRequest, args SendEmailArgs) (*mcp.CallToolResult, any, error) {
	// Set defaults
	if args.Port == 0 {
		args.Port = 587
	}
	if args.AuthType == "" {
		args.AuthType = "plain"
	}

	emailReq := &services.EmailRequest{
		Config: &services.ConfigData{
			Server:     args.Server,
			Port:       args.Port,
			Username:   args.Username,
			Password:   args.Password,
			AuthType:   args.AuthType,
			StartTLS:   args.StartTLS,
			SkipVerify: args.SkipVerify,
		},
		From:     args.From,
		To:       args.To,
		Cc:       args.CC,
		Bcc:      args.BCC,
		Subject:  args.Subject,
		Body:     args.Body,
		HTMLBody: "",
	}

	if args.IsHTML {
		emailReq.HTMLBody = args.Body
		emailReq.Body = ""
	}

	result := smtpService.SendEmail(emailReq)
	
	// Convert result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func validateAddressesHandler(ctx context.Context, req *mcp.CallToolRequest, args ValidateAddressesArgs) (*mcp.CallToolResult, any, error) {
	// Validate each address
	results := make(map[string]interface{})
	for _, addr := range args.Addresses {
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
		// TODO: Add MX record checking if args.CheckMX is true
	}
	
	// Convert results to JSON string
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

func loadTemplateHandler(ctx context.Context, req *mcp.CallToolRequest, args LoadTemplateArgs) (*mcp.CallToolResult, any, error) {
	// Template loading simplified - just return template name for now
	// TODO: Implement proper template loading with variables
	result := map[string]interface{}{
		"template":  args.TemplateName,
		"variables": args.Variables,
		"message":   "Template loading not yet implemented",
	}
	
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(resultJSON),
			},
		},
	}, nil, nil
}

// Resource handler functions
func configResourceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	config := configService.GetCurrentConfig()
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(configJSON),
			},
		},
	}, nil
}

func templatesResourceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	templates, err := templateService.ListTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	templatesJSON, err := json.Marshal(templates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal templates: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(templatesJSON),
			},
		},
	}, nil
}

func authStatsResourceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	stats := smtpService.GetAuthStats()
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(statsJSON),
			},
		},
	}, nil
}