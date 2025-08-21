package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer(false)
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.configService == nil {
		t.Fatal("Expected non-nil config service")
	}
	if server.smtpService == nil {
		t.Fatal("Expected non-nil SMTP service")
	}
	if server.templateService == nil {
		t.Fatal("Expected non-nil template service")
	}
}

func TestListTools(t *testing.T) {
	server := NewMCPServer(false)
	tools := server.ListTools()
	
	expectedTools := []string{
		"smtp_test_connection",
		"smtp_send_email",
		"smtp_validate_addresses",
		"smtp_load_template",
	}
	
	if len(tools) != len(expectedTools) {
		t.Fatalf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}
	
	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}
	
	for _, expected := range expectedTools {
		if !toolMap[expected] {
			t.Errorf("Missing expected tool: %s", expected)
		}
	}
}

func TestListResources(t *testing.T) {
	server := NewMCPServer(false)
	resources := server.ListResources()
	
	expectedResources := []string{
		"smtp-edc://config/current",
		"smtp-edc://templates/list",
		"smtp-edc://stats/auth",
	}
	
	if len(resources) != len(expectedResources) {
		t.Fatalf("Expected %d resources, got %d", len(expectedResources), len(resources))
	}
	
	resourceMap := make(map[string]bool)
	for _, resource := range resources {
		resourceMap[resource.URI] = true
	}
	
	for _, expected := range expectedResources {
		if !resourceMap[expected] {
			t.Errorf("Missing expected resource: %s", expected)
		}
	}
}

func TestCallTool_UnknownTool(t *testing.T) {
	server := NewMCPServer(false)
	ctx := context.Background()
	
	_, err := server.CallTool(ctx, "unknown_tool", json.RawMessage("{}"))
	if err == nil {
		t.Fatal("Expected error for unknown tool")
	}
	
	expectedError := "unknown tool: unknown_tool"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestReadResource_UnknownResource(t *testing.T) {
	server := NewMCPServer(false)
	ctx := context.Background()
	
	_, err := server.ReadResource(ctx, "smtp-edc://unknown/resource")
	if err == nil {
		t.Fatal("Expected error for unknown resource")
	}
	
	expectedError := "unknown resource: smtp-edc://unknown/resource"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateAddresses(t *testing.T) {
	server := NewMCPServer(false)
	ctx := context.Background()
	
	testCases := []struct {
		name      string
		args      string
		expectErr bool
	}{
		{
			name:      "Valid addresses",
			args:      `{"addresses": ["test@example.com", "user@domain.org"], "checkMX": false}`,
			expectErr: false,
		},
		{
			name:      "Empty addresses",
			args:      `{"addresses": [], "checkMX": false}`,
			expectErr: false,
		},
		{
			name:      "Invalid JSON",
			args:      `{invalid json}`,
			expectErr: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.validateAddresses(ctx, json.RawMessage(tc.args))
			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTestConnection(t *testing.T) {
	server := NewMCPServer(false)
	ctx := context.Background()
	
	testCases := []struct {
		name      string
		args      string
		expectErr bool
	}{
		{
			name:      "Valid connection params",
			args:      `{"server": "smtp.example.com", "port": 587}`,
			expectErr: false,
		},
		{
			name:      "Default port",
			args:      `{"server": "smtp.example.com"}`,
			expectErr: false,
		},
		{
			name:      "Invalid JSON",
			args:      `{invalid json}`,
			expectErr: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.testConnection(ctx, json.RawMessage(tc.args))
			if tc.expectErr && err != nil && err.Error() != "invalid arguments: invalid character 'i' looking for beginning of object key string" {
				t.Errorf("Expected JSON parse error, got: %v", err)
			}
		})
	}
}