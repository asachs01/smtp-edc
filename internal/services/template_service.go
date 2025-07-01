package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asachs/smtp-edc/internal/message"
)

// TemplateService provides Wails-compatible template management
type TemplateService struct {
	templateDir string
}

// TemplateData represents template data for frontend binding
type TemplateData struct {
	From    string                 `json:"from"`
	To      []string               `json:"to"`
	Cc      []string               `json:"cc"`
	Bcc     []string               `json:"bcc"`
	Subject string                 `json:"subject"`
	Data    map[string]interface{} `json:"data"`
}

// TemplateInfo represents template metadata
type TemplateInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Size        int64    `json:"size"`
	ModTime     string   `json:"modTime"`
	Variables   []string `json:"variables"`
	HasHTML     bool     `json:"hasHtml"`
	HasText     bool     `json:"hasText"`
	Description string   `json:"description"`
}

// TemplateResult represents the result of template operations
type TemplateResult struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Error     string   `json:"error,omitempty"`
	Content   string   `json:"content,omitempty"`
	Variables []string `json:"variables,omitempty"`
}

// NewTemplateService creates a new template service
func NewTemplateService() *TemplateService {
	homeDir, _ := os.UserHomeDir()
	templateDir := filepath.Join(homeDir, ".smtp-edc", "templates")

	return &TemplateService{
		templateDir: templateDir,
	}
}

// SetTemplateDirectory sets the template directory path
func (ts *TemplateService) SetTemplateDirectory(dir string) {
	ts.templateDir = dir
}

// GetTemplateDirectory returns the current template directory
func (ts *TemplateService) GetTemplateDirectory() string {
	return ts.templateDir
}

// ListTemplates lists available email templates
func (ts *TemplateService) ListTemplates() ([]*TemplateInfo, error) {
	// Ensure template directory exists
	if err := os.MkdirAll(ts.templateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create template directory: %v", err)
	}

	files, err := os.ReadDir(ts.templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %v", err)
	}

	var templates []*TemplateInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())
		if ext != ".html" && ext != ".txt" && ext != ".tmpl" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(ts.templateDir, file.Name())
		variables, _ := ts.extractVariables(fullPath)

		templates = append(templates, &TemplateInfo{
			Name:      file.Name(),
			Path:      fullPath,
			Size:      info.Size(),
			ModTime:   info.ModTime().Format("2006-01-02 15:04:05"),
			Variables: variables,
			HasHTML:   ext == ".html",
			HasText:   ext == ".txt" || ext == ".tmpl",
		})
	}

	return templates, nil
}

// LoadTemplate loads and parses an email template
func (ts *TemplateService) LoadTemplate(templateName string) (*TemplateResult, error) {
	templatePath := filepath.Join(ts.templateDir, templateName)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Template '%s' not found", templateName),
		}, nil
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to read template: %v", err),
		}, nil
	}

	variables, err := ts.extractVariables(templatePath)
	if err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract variables: %v", err),
		}, nil
	}

	return &TemplateResult{
		Success:   true,
		Message:   "Template loaded successfully",
		Content:   string(content),
		Variables: variables,
	}, nil
}

// SaveTemplate saves a template to the template directory
func (ts *TemplateService) SaveTemplate(templateName, content string) (*TemplateResult, error) {
	// Ensure template directory exists
	if err := os.MkdirAll(ts.templateDir, 0755); err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create template directory: %v", err),
		}, nil
	}

	templatePath := filepath.Join(ts.templateDir, templateName)

	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to save template: %v", err),
		}, nil
	}

	return &TemplateResult{
		Success: true,
		Message: "Template saved successfully",
	}, nil
}

// DeleteTemplate deletes a template from the template directory
func (ts *TemplateService) DeleteTemplate(templateName string) (*TemplateResult, error) {
	templatePath := filepath.Join(ts.templateDir, templateName)

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Template '%s' not found", templateName),
		}, nil
	}

	if err := os.Remove(templatePath); err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete template: %v", err),
		}, nil
	}

	return &TemplateResult{
		Success: true,
		Message: "Template deleted successfully",
	}, nil
}

// ExecuteTemplate processes a template with provided data
func (ts *TemplateService) ExecuteTemplate(templateName, subjectTemplate string, templateData *TemplateData) (*EmailRequest, error) {
	templatePath := filepath.Join(ts.templateDir, templateName)

	// Load the template
	tmpl, err := message.LoadTemplate(subjectTemplate, templatePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %v", err)
	}

	// Convert TemplateData to internal format
	internalData := &message.TemplateData{
		From:    templateData.From,
		To:      templateData.To,
		Cc:      templateData.Cc,
		Bcc:     templateData.Bcc,
		Subject: templateData.Subject,
		Data:    templateData.Data,
	}

	// Execute the template
	msg, err := tmpl.Execute(internalData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	// Convert message to EmailRequest
	emailReq := &EmailRequest{
		From:     msg.From,
		To:       msg.To,
		Cc:       msg.Cc,
		Bcc:      msg.Bcc,
		Subject:  msg.Subject,
		Body:     msg.Body,
		HTMLBody: msg.HTMLBody,
		Headers:  msg.Headers,
	}

	// Convert attachments to file paths (templates store them as file references)
	for _, attachment := range msg.Attachments {
		emailReq.Attachments = append(emailReq.Attachments, attachment.Filename)
	}

	return emailReq, nil
}

// ValidateTemplate validates template syntax and variables
func (ts *TemplateService) ValidateTemplate(content string) (*TemplateResult, error) {
	// TODO: Implement template validation logic
	// For now, just check if it's valid JSON or has template markers
	if content == "" {
		return &TemplateResult{
			Success: false,
			Error:   "Template content cannot be empty",
		}, nil
	}

	// Extract variables to validate template syntax
	variables := ts.findTemplateVariables(content)

	return &TemplateResult{
		Success:   true,
		Message:   fmt.Sprintf("Template is valid with %d variables", len(variables)),
		Variables: variables,
	}, nil
}

// GetDefaultTemplates returns a list of default template examples
func (ts *TemplateService) GetDefaultTemplates() map[string]string {
	return map[string]string{
		"welcome.html": `<!DOCTYPE html>
<html>
<head>
    <title>{{.Subject}}</title>
</head>
<body>
    <h1>Welcome {{.Data.name}}!</h1>
    <p>Thank you for joining our service.</p>
    <p>Your email: {{.Data.email}}</p>
</body>
</html>`,
		"notification.txt": `Hello {{.Data.name}},

This is a notification about {{.Data.event}}.

Details:
{{.Data.details}}

Best regards,
{{.Data.sender}}`,
		"test.tmpl": `Subject: Test Email for {{.Data.recipient}}
From: {{.From}}
To: {{.To}}

Hello {{.Data.recipient}},

This is a test email sent at {{.Data.timestamp}}.

Environment: {{.Data.environment}}
Version: {{.Data.version}}`,
	}
}

// CreateDefaultTemplates creates default template files
func (ts *TemplateService) CreateDefaultTemplates() (*TemplateResult, error) {
	// Ensure template directory exists
	if err := os.MkdirAll(ts.templateDir, 0755); err != nil {
		return &TemplateResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to create template directory: %v", err),
		}, nil
	}

	defaults := ts.GetDefaultTemplates()
	created := 0

	for name, content := range defaults {
		templatePath := filepath.Join(ts.templateDir, name)

		// Don't overwrite existing templates
		if _, err := os.Stat(templatePath); err == nil {
			continue
		}

		if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
			return &TemplateResult{
				Success: false,
				Error:   fmt.Sprintf("Failed to create template %s: %v", name, err),
			}, nil
		}
		created++
	}

	return &TemplateResult{
		Success: true,
		Message: fmt.Sprintf("Created %d default templates", created),
	}, nil
}

// extractVariables extracts template variables from a file
func (ts *TemplateService) extractVariables(templatePath string) ([]string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	return ts.findTemplateVariables(string(content)), nil
}

// findTemplateVariables finds template variables in content
func (ts *TemplateService) findTemplateVariables(content string) []string {
	// Simple regex-based variable extraction
	// Look for {{.Variable}} patterns
	variables := make(map[string]bool)
	var result []string

	// This is a simplified implementation
	// In a real implementation, you'd use proper template parsing
	for i := 0; i < len(content)-3; i++ {
		if content[i:i+2] == "{{" {
			end := i + 2
			for end < len(content)-1 && content[end:end+2] != "}}" {
				end++
			}
			if end < len(content)-1 {
				variable := content[i+2 : end]
				if variable[0] == '.' {
					variable = variable[1:]
				}
				if !variables[variable] && variable != "" {
					variables[variable] = true
					result = append(result, variable)
				}
			}
		}
	}

	return result
}
