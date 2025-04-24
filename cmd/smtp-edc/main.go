package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/asachs/smtp-edc/internal/client"
	"github.com/asachs/smtp-edc/internal/config"
	"github.com/asachs/smtp-edc/internal/message"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg *config.SMTPConfig
)

func init() {
	// Set default values
	viper.SetDefault("port", 25)
	viper.SetDefault("retries", 3)
	viper.SetDefault("timeout", 30)
	viper.SetDefault("starttls", false)
	viper.SetDefault("skip_verify", false)
	viper.SetDefault("debug", false)
	viper.SetDefault("validate_mx", false)

	// Bind environment variables
	viper.BindEnv("server", "SMTP_SERVER")
	viper.BindEnv("port", "SMTP_PORT")
	viper.BindEnv("username", "SMTP_USERNAME")
	viper.BindEnv("password", "SMTP_PASSWORD")
	viper.BindEnv("from", "SMTP_FROM")
	viper.BindEnv("to", "SMTP_TO")
	viper.BindEnv("cc", "SMTP_CC")
	viper.BindEnv("bcc", "SMTP_BCC")
	viper.BindEnv("subject", "SMTP_SUBJECT")
	viper.BindEnv("auth_type", "SMTP_AUTH_TYPE")
	viper.BindEnv("starttls", "SMTP_STARTTLS")
	viper.BindEnv("skip_verify", "SMTP_SKIP_VERIFY")
	viper.BindEnv("debug", "SMTP_DEBUG")

	// Define flags
	pflag.StringP("config", "c", "", "Path to config file (JSON or YAML)")
	pflag.StringP("server", "s", "", "SMTP server address")
	pflag.IntP("port", "p", 25, "SMTP server port")
	pflag.StringP("from", "f", "", "Sender email address")
	pflag.StringP("to", "t", "", "Recipient email addresses (comma-separated)")
	pflag.StringP("cc", "C", "", "CC recipient email addresses (comma-separated)")
	pflag.StringP("bcc", "B", "", "BCC recipient email addresses (comma-separated)")
	pflag.StringP("subject", "S", "", "Email subject")
	pflag.StringP("subject_template", "T", "", "Email subject template")
	pflag.StringP("body", "b", "", "Email body text")
	pflag.StringP("body_file", "F", "", "File containing email body text")
	pflag.StringP("html", "H", "", "Email HTML body")
	pflag.StringP("html_file", "L", "", "File containing email HTML body")
	pflag.StringP("template", "e", "", "Path to email template file")
	pflag.StringP("template_data", "d", "", "JSON data for template (format: '{\"key\":\"value\"}')")
	pflag.StringP("auth_type", "a", "", "Authentication type (plain, login, cram-md5)")
	pflag.StringP("username", "u", "", "Authentication username")
	pflag.StringP("password", "P", "", "Authentication password")
	pflag.BoolP("starttls", "l", false, "Use STARTTLS")
	pflag.BoolP("skip_verify", "k", false, "Skip TLS certificate verification")
	pflag.BoolP("debug", "D", false, "Enable debug output")
	pflag.StringP("attachments", "A", "", "Comma-separated list of files to attach")
	pflag.StringP("headers", "h", "", "Custom headers (format: 'Key1: Value1, Key2: Value2')")
	pflag.IntP("retries", "r", 3, "Number of retry attempts for failed operations")
	pflag.IntP("timeout", "o", 30, "Connection timeout in seconds")
	pflag.BoolP("validate_mx", "m", false, "Validate email addresses by checking MX records")

	// Bind flags to Viper
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Set config file
	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}
	}
}

// parseAddressList splits a comma-separated list of email addresses
func parseAddressList(list string) []string {
	if list == "" {
		return nil
	}
	addresses := strings.Split(list, ",")
	// Trim spaces from addresses
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}
	return addresses
}

// parseHeaders parses the custom headers string into a map
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	// Split header pairs
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		// Split key and value
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}
	return headers
}

// readFile reads the contents of a file
func readFile(filename string) (string, error) {
	if filename == "" {
		return "", nil
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return string(content), nil
}

func main() {
	// Validate required fields
	server := viper.GetString("server")
	from := viper.GetString("from")
	to := viper.GetString("to")
	cc := viper.GetString("cc")
	bcc := viper.GetString("bcc")

	if server == "" || from == "" || (to == "" && cc == "" && bcc == "") {
		fmt.Println("Error: server, from, and at least one recipient (to, cc, or bcc) are required")
		fmt.Println("Current values:")
		fmt.Printf("  Server: %s\n", server)
		fmt.Printf("  From: %s\n", from)
		fmt.Printf("  To: %s\n", to)
		fmt.Printf("  CC: %s\n", cc)
		fmt.Printf("  BCC: %s\n", bcc)
		pflag.Usage()
		os.Exit(1)
	}

	// Validate email addresses
	if err := message.ValidateEmail(from); err != nil {
		log.Fatalf("Invalid sender address: %v", err)
	}

	toAddrs := parseAddressList(to)
	ccAddrs := parseAddressList(cc)
	bccAddrs := parseAddressList(bcc)

	if err := message.ValidateAddressList(toAddrs, viper.GetBool("validate_mx")); err != nil {
		log.Fatalf("Invalid To address: %v", err)
	}
	if err := message.ValidateAddressList(ccAddrs, viper.GetBool("validate_mx")); err != nil {
		log.Fatalf("Invalid Cc address: %v", err)
	}
	if err := message.ValidateAddressList(bccAddrs, viper.GetBool("validate_mx")); err != nil {
		log.Fatalf("Invalid Bcc address: %v", err)
	}

	var msg *message.Message

	// Handle templates
	if templateFile := viper.GetString("template"); templateFile != "" {
		// Parse template data
		var data map[string]interface{}
		if templateData := viper.GetString("template_data"); templateData != "" {
			if err := json.Unmarshal([]byte(templateData), &data); err != nil {
				log.Fatalf("Failed to parse template data: %v", err)
			}
		}

		// Load template
		tmpl, err := message.LoadTemplate(viper.GetString("subject_template"), templateFile, "")
		if err != nil {
			log.Fatalf("Failed to load template: %v", err)
		}

		// Execute template
		msg, err = tmpl.Execute(&message.TemplateData{
			From:    viper.GetString("from"),
			To:      toAddrs,
			Cc:      ccAddrs,
			Bcc:     bccAddrs,
			Subject: viper.GetString("subject"),
			Data:    data,
		})
		if err != nil {
			log.Fatalf("Failed to execute template: %v", err)
		}
	} else {
		// Create message without template
		msg = message.NewMessage(viper.GetString("from"), toAddrs, viper.GetString("subject"), viper.GetString("body"))
		msg.Cc = ccAddrs
		msg.Bcc = bccAddrs

		// Read body from file if specified
		if bodyFile := viper.GetString("body_file"); bodyFile != "" {
			body, err := readFile(bodyFile)
			if err != nil {
				log.Fatal(err)
			}
			msg.Body = body
		}

		// Read HTML body from file if specified
		if htmlFile := viper.GetString("html_file"); htmlFile != "" {
			htmlBody, err := readFile(htmlFile)
			if err != nil {
				log.Fatal(err)
			}
			msg.HTMLBody = htmlBody
		}
	}

	// Add custom headers
	for key, value := range parseHeaders(viper.GetString("headers")) {
		msg.AddHeader(key, value)
	}

	// Add attachments
	if attachments := viper.GetString("attachments"); attachments != "" {
		for _, attachment := range parseAddressList(attachments) {
			if _, err := message.ReadFileAttachment(attachment); err != nil {
				log.Fatalf("Failed to read attachment %s: %v", attachment, err)
			}
			msg.AddAttachment(attachment)
		}
	}

	// Create SMTP client
	client := client.NewSMTPClient("localhost", viper.GetBool("debug"))
	client.SetRetryConfig(viper.GetInt("retries"), time.Duration(viper.GetInt("timeout"))*time.Second)

	// Connect to server
	if err := client.Connect(viper.GetString("server"), viper.GetInt("port")); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send EHLO
	if err := client.Ehlo(); err != nil {
		log.Fatalf("Failed to send EHLO: %v", err)
	}

	// Start TLS if requested
	if viper.GetBool("starttls") {
		if err := client.StartTLS(); err != nil {
			log.Fatalf("Failed to start TLS: %v", err)
		}
		// Send EHLO again after STARTTLS
		if err := client.Ehlo(); err != nil {
			log.Fatalf("Failed to send EHLO after STARTTLS: %v", err)
		}
	}

	// Authenticate if requested
	if authType := viper.GetString("auth_type"); authType != "" {
		username := viper.GetString("username")
		password := viper.GetString("password")
		if username == "" || password == "" {
			log.Fatal("Username and password are required for authentication")
		}
		if err := client.Authenticate(authType, username, password); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}
	}

	// Send message
	if err := client.SendMessage(msg); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Quit
	if err := client.Quit(); err != nil {
		log.Fatalf("Failed to quit: %v", err)
	}

	fmt.Println("Message sent successfully")
}
