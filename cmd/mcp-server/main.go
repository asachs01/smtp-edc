package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/asachs/smtp-edc/internal/mcp"
)

func main() {
	var (
		transport = flag.String("transport", "stdio", "Transport type: stdio or http")
		port      = flag.Int("port", 8080, "Port for HTTP transport")
		debug     = flag.Bool("debug", false, "Enable debug logging")
		help      = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	// Create and start the MCP server
	server := mcp.NewMCPServer(*debug)
	
	if *debug {
		log.Printf("Starting SMTP-EDC MCP Server with transport: %s\n", *transport)
	}

	// Start the server with the specified transport
	if err := server.Start(*transport); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}

func printHelp() {
	fmt.Println(`SMTP-EDC MCP Server

This server provides Model Context Protocol (MCP) access to SMTP-EDC functionality,
allowing AI assistants and other MCP clients to interact with SMTP servers.

Usage:
  smtp-edc mcp-server [options]

Options:
  -transport string
        Transport type: stdio or http (default "stdio")
  -port int
        Port for HTTP transport (default 8080)
  -debug
        Enable debug logging
  -help
        Show this help message

Available Tools:
  - smtp_test_connection: Test SMTP server connection and capabilities
  - smtp_send_email: Send an email via SMTP
  - smtp_validate_addresses: Validate email addresses and check MX records
  - smtp_load_template: Load and process email templates

Available Resources:
  - smtp-edc://config/current: Current configuration settings
  - smtp-edc://templates/list: Available email templates
  - smtp-edc://stats/auth: Authentication statistics

Examples:
  # Start MCP server with STDIO transport (for local tools)
  smtp-edc mcp-server -transport stdio

  # Start MCP server with HTTP transport
  smtp-edc mcp-server -transport http -port 8080

  # Enable debug logging
  smtp-edc mcp-server -debug

For more information about MCP, visit: https://modelcontextprotocol.io/
`)
}