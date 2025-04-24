package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/asachs/smtp-edc/internal/auth"
	"github.com/asachs/smtp-edc/internal/message"
)

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

// ServerCapabilities represents SMTP server capabilities
type ServerCapabilities struct {
	Pipelining bool
	StartTLS   bool
	Auth       []string
	Size       int
	EightBit   bool
}

// SMTPClient represents an SMTP client connection
type SMTPClient struct {
	conn         net.Conn
	reader       *bufio.Reader
	writer       *bufio.Writer
	hostname     string
	server       string
	debug        bool
	tls          bool
	retry        RetryConfig
	timeout      time.Duration
	capabilities ServerCapabilities
	client       smtp.Client
}

// NewSMTPClient creates a new SMTP client connection
func NewSMTPClient(hostname string, debug bool) *SMTPClient {
	return &SMTPClient{
		hostname: hostname,
		debug:    debug,
		retry: RetryConfig{
			MaxAttempts: 3,
			Delay:       time.Second * 2,
		},
		timeout: time.Second * 30,
	}
}

// SetRetryConfig sets the retry configuration
func (c *SMTPClient) SetRetryConfig(maxAttempts int, delay time.Duration) {
	c.retry.MaxAttempts = maxAttempts
	c.retry.Delay = delay
}

// SetTimeout sets the connection timeout
func (c *SMTPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// withRetry executes a function with retry logic
func (c *SMTPClient) withRetry(operation string, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if c.debug {
				fmt.Printf("Attempt %d/%d for %s failed: %v\n",
					attempt, c.retry.MaxAttempts, operation, err)
			}
			if attempt < c.retry.MaxAttempts {
				time.Sleep(c.retry.Delay)
				continue
			}
			return fmt.Errorf("%s failed after %d attempts: %v",
				operation, c.retry.MaxAttempts, err)
		}
		return nil
	}
	return lastErr
}

// Connect establishes a connection to the SMTP server
func (c *SMTPClient) Connect(server string, port int) error {
	return c.withRetry("connect", func() error {
		// If we already have a connection (likely a mock in tests), use it
		if c.conn != nil {
			// Test the connection by trying to read the server greeting
			c.reader = bufio.NewReader(c.conn)
			c.writer = bufio.NewWriter(c.conn)
			c.server = server

			// Read server greeting to verify connection
			_, err := c.readResponse()
			if err != nil {
				return fmt.Errorf("failed to read server greeting: %v", err)
			}

			return nil
		}

		addr := fmt.Sprintf("%s:%d", server, port)

		// Create connection with timeout
		conn, err := net.DialTimeout("tcp", addr, c.timeout)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %v", err)
		}

		// Set read/write timeouts
		conn.SetDeadline(time.Now().Add(c.timeout))

		c.conn = conn
		c.reader = bufio.NewReader(conn)
		c.writer = bufio.NewWriter(conn)
		c.server = server

		// Read server greeting
		_, err = c.readResponse()
		if err != nil {
			c.conn.Close()
			return fmt.Errorf("failed to read server greeting: %v", err)
		}

		return nil
	})
}

// StartTLS initiates a TLS connection
func (c *SMTPClient) StartTLS() error {
	err := c.SendCommand("STARTTLS")
	if err != nil {
		return fmt.Errorf("failed to send STARTTLS command: %v", err)
	}

	// Read all response lines until we get a final response
	for {
		line, err := c.readResponse()
		if err != nil {
			return fmt.Errorf("server rejected STARTTLS: %v", err)
		}
		// Check if this is the final response line
		if len(line) >= 4 && line[3] == ' ' {
			if line[0] != '2' {
				return fmt.Errorf("server rejected STARTTLS: %s", line)
			}
			break
		}
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName:         c.server,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12, // Force TLS 1.2 or higher
	}

	if c.debug {
		fmt.Printf("Starting TLS handshake with server %s\n", c.server)
	}

	// Upgrade connection to TLS
	tlsConn := tls.Client(c.conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		if c.debug {
			fmt.Printf("TLS handshake failed: %v\n", err)
			fmt.Printf("TLS version attempted: %d\n", tlsConfig.MinVersion)
			fmt.Printf("Server name: %s\n", tlsConfig.ServerName)
		}
		return fmt.Errorf("TLS handshake failed: %v", err)
	}

	if c.debug {
		state := tlsConn.ConnectionState()
		fmt.Println("TLS handshake successful")
		fmt.Printf("TLS version: %s\n", tlsVersionString(state.Version))
		fmt.Printf("Cipher suite: %s\n", tls.CipherSuiteName(state.CipherSuite))
	}

	c.conn = tlsConn
	c.reader = bufio.NewReader(tlsConn)
	c.writer = bufio.NewWriter(tlsConn)
	c.tls = true

	return nil
}

// tlsVersionString converts a TLS version number to a string
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// Authenticate performs SMTP authentication
func (c *SMTPClient) Authenticate(authType, username, password string) error {
	// Create authenticator
	authenticator, err := auth.NewAuthenticator(authType)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %v", err)
	}

	// Send AUTH command
	cmd := fmt.Sprintf("AUTH %s", authenticator.Type())
	err = c.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to send AUTH command: %v", err)
	}

	// Handle different authentication methods
	switch authType {
	case "plain":
		response, err := authenticator.Authenticate(username, password)
		if err != nil {
			return fmt.Errorf("failed to generate PLAIN auth response: %v", err)
		}
		err = c.SendCommand(response)
		if err != nil {
			return fmt.Errorf("failed to send PLAIN auth response: %v", err)
		}
		_, err = c.readResponse()
		return err

	case "login":
		// First step: send username
		response, err := authenticator.Authenticate(username, password)
		if err != nil {
			return fmt.Errorf("failed to generate LOGIN auth response: %v", err)
		}
		err = c.SendCommand(response)
		if err != nil {
			return fmt.Errorf("failed to send LOGIN username: %v", err)
		}
		_, err = c.readResponse()
		if err != nil {
			return err
		}

		// Second step: send password
		passwordResponse := authenticator.(*auth.LoginAuthenticator).GetPassword(password)
		err = c.SendCommand(passwordResponse)
		if err != nil {
			return fmt.Errorf("failed to send LOGIN password: %v", err)
		}
		_, err = c.readResponse()
		return err

	case "cram-md5":
		// Get challenge from server
		challenge, err := c.readResponse()
		if err != nil {
			return fmt.Errorf("failed to read CRAM-MD5 challenge: %v", err)
		}

		// Generate and send response
		response, err := authenticator.(*auth.CRAMMD5Authenticator).GenerateResponse(challenge, username, password)
		if err != nil {
			return fmt.Errorf("failed to generate CRAM-MD5 response: %v", err)
		}
		err = c.SendCommand(response)
		if err != nil {
			return fmt.Errorf("failed to send CRAM-MD5 response: %v", err)
		}
		_, err = c.readResponse()
		return err

	default:
		return fmt.Errorf("unsupported authentication type: %s", authType)
	}
}

// Close closes the SMTP connection
func (c *SMTPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendCommand sends a command to the SMTP server
func (c *SMTPClient) SendCommand(cmd string) error {
	if c.debug {
		fmt.Printf("C: %s\n", cmd)
	}

	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return fmt.Errorf("failed to write command: %v", err)
	}

	err = c.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush command: %v", err)
	}

	return nil
}

// readResponse reads the server's response
func (c *SMTPClient) readResponse() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if c.debug {
		fmt.Printf("S: %s", line)
	}

	return line, nil
}

// Helo sends the HELO command to the server
func (c *SMTPClient) Helo() error {
	cmd := fmt.Sprintf("HELO %s", c.hostname)
	err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	_, err = c.readResponse()
	return err
}

// parseCapabilities parses EHLO response for server capabilities
func (c *SMTPClient) parseCapabilities(response string) {
	c.capabilities = ServerCapabilities{}
	lines := strings.Split(response, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "250-") || strings.HasPrefix(line, "250 ") {
			capability := strings.TrimPrefix(strings.TrimPrefix(line, "250-"), "250 ")
			switch {
			case strings.HasPrefix(capability, "PIPELINING"):
				c.capabilities.Pipelining = true
			case strings.HasPrefix(capability, "STARTTLS"):
				c.capabilities.StartTLS = true
			case strings.HasPrefix(capability, "AUTH"):
				c.capabilities.Auth = strings.Fields(capability)[1:]
			case strings.HasPrefix(capability, "SIZE"):
				if size := strings.Fields(capability); len(size) > 1 {
					c.capabilities.Size, _ = strconv.Atoi(size[1])
				}
			case strings.HasPrefix(capability, "8BITMIME"):
				c.capabilities.EightBit = true
			}
		}
	}
}

// Ehlo sends the EHLO command to the server and parses capabilities
func (c *SMTPClient) Ehlo() error {
	cmd := fmt.Sprintf("EHLO %s", c.hostname)
	err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	var response strings.Builder
	// Read all response lines until we get a final response
	for {
		line, err := c.readResponse()
		if err != nil {
			return err
		}
		response.WriteString(line)
		// Check if this is the final response line
		if len(line) >= 4 && line[3] == ' ' {
			if line[0] != '2' {
				return fmt.Errorf("server rejected EHLO: %s", line)
			}
			break
		}
	}

	// Parse capabilities from response
	c.parseCapabilities(response.String())
	return nil
}

// MailFrom sends the MAIL FROM command
func (c *SMTPClient) MailFrom(from string) error {
	cmd := fmt.Sprintf("MAIL FROM:<%s>", from)
	err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	_, err = c.readResponse()
	return err
}

// RcptTo sends the RCPT TO command
func (c *SMTPClient) RcptTo(to string) error {
	cmd := fmt.Sprintf("RCPT TO:<%s>", to)
	err := c.SendCommand(cmd)
	if err != nil {
		return err
	}

	_, err = c.readResponse()
	return err
}

// sendMessageNonPipelined sends a message without using pipelining
func (c *SMTPClient) sendMessageNonPipelined(msg *message.Message) error {
	return c.withRetry("send message", func() error {
		// Set sender
		if err := c.MailFrom(msg.From); err != nil {
			return fmt.Errorf("failed to set sender: %v", err)
		}

		// Set recipients (To, Cc, and Bcc)
		allRecipients := make([]string, 0)
		allRecipients = append(allRecipients, msg.To...)
		allRecipients = append(allRecipients, msg.Cc...)
		allRecipients = append(allRecipients, msg.Bcc...)

		// Remove duplicates
		seen := make(map[string]bool)
		uniqueRecipients := make([]string, 0)
		for _, recipient := range allRecipients {
			if !seen[recipient] {
				seen[recipient] = true
				uniqueRecipients = append(uniqueRecipients, recipient)
			}
		}

		// Send RCPT TO for each unique recipient
		for _, recipient := range uniqueRecipients {
			if err := c.RcptTo(recipient); err != nil {
				return fmt.Errorf("failed to set recipient %s: %v", recipient, err)
			}
		}

		// Send DATA command
		if err := c.SendCommand("DATA"); err != nil {
			return fmt.Errorf("failed to send DATA command: %v", err)
		}

		// Read server response
		_, err := c.readResponse()
		if err != nil {
			return fmt.Errorf("server rejected DATA command: %v", err)
		}

		// Build and send message
		messageData, err := msg.Build()
		if err != nil {
			return fmt.Errorf("failed to build message: %v", err)
		}

		// Send message data
		if err := c.SendCommand(messageData); err != nil {
			return fmt.Errorf("failed to send message: %v", err)
		}

		// Send end of message marker
		if err := c.SendCommand("."); err != nil {
			return fmt.Errorf("failed to send end of message marker: %v", err)
		}

		// Read final response
		_, err = c.readResponse()
		return err
	})
}

// SendMessage sends a message, using pipelining if available
func (c *SMTPClient) SendMessage(msg *message.Message) error {
	if c.capabilities.Pipelining {
		return c.SendMessagePipelined(msg)
	}
	return c.sendMessageNonPipelined(msg)
}

// SendMessagePipelined sends a message using SMTP pipelining if supported
func (c *SMTPClient) SendMessagePipelined(msg *message.Message) error {
	if !c.capabilities.Pipelining {
		return c.SendMessage(msg)
	}

	return c.withRetry("send pipelined message", func() error {
		// Prepare all recipients
		allRecipients := make([]string, 0)
		allRecipients = append(allRecipients, msg.To...)
		allRecipients = append(allRecipients, msg.Cc...)
		allRecipients = append(allRecipients, msg.Bcc...)

		// Remove duplicates
		seen := make(map[string]bool)
		uniqueRecipients := make([]string, 0)
		for _, recipient := range allRecipients {
			if !seen[recipient] {
				seen[recipient] = true
				uniqueRecipients = append(uniqueRecipients, recipient)
			}
		}

		// Send MAIL FROM and all RCPT TO commands in one batch
		if err := c.SendCommand(fmt.Sprintf("MAIL FROM:<%s>", msg.From)); err != nil {
			return fmt.Errorf("failed to send MAIL FROM: %v", err)
		}

		for _, recipient := range uniqueRecipients {
			if err := c.SendCommand(fmt.Sprintf("RCPT TO:<%s>", recipient)); err != nil {
				return fmt.Errorf("failed to send RCPT TO: %v", err)
			}
		}

		if err := c.SendCommand("DATA"); err != nil {
			return fmt.Errorf("failed to send DATA: %v", err)
		}

		// Flush the writer to send all commands at once
		if err := c.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush commands: %v", err)
		}

		// Read responses for MAIL FROM and all RCPT TO commands
		_, err := c.readResponse() // MAIL FROM response
		if err != nil {
			return fmt.Errorf("MAIL FROM failed: %v", err)
		}

		for range uniqueRecipients {
			_, err := c.readResponse() // RCPT TO response
			if err != nil {
				return fmt.Errorf("RCPT TO failed: %v", err)
			}
		}

		// Read DATA response
		_, err = c.readResponse()
		if err != nil {
			return fmt.Errorf("DATA command failed: %v", err)
		}

		// Send message content
		messageData, err := msg.Build()
		if err != nil {
			return fmt.Errorf("failed to build message: %v", err)
		}

		if err := c.SendCommand(messageData); err != nil {
			return fmt.Errorf("failed to send message: %v", err)
		}

		if err := c.SendCommand("."); err != nil {
			return fmt.Errorf("failed to send end of message: %v", err)
		}

		// Read final response
		_, err = c.readResponse()
		return err
	})
}

// Quit sends the QUIT command
func (c *SMTPClient) Quit() error {
	err := c.SendCommand("QUIT")
	if err != nil {
		return err
	}

	_, err = c.readResponse()
	return err
}

// Send sends an email using the high-level smtp.SendMail function
func (c *SMTPClient) Send(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return c.withRetry("Send", func() error {
		return smtp.SendMail(addr, auth, from, to, msg)
	})
}

// SendRaw sends an email using the low-level SMTP commands
func (c *SMTPClient) SendRaw(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return c.withRetry("SendRaw", func() error {
		if err := c.client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %v", err)
		}

		if err := c.client.Hello(c.hostname); err != nil {
			return fmt.Errorf("hello failed: %v", err)
		}

		if err := c.client.Mail(from); err != nil {
			return fmt.Errorf("mail from failed: %v", err)
		}

		for _, addr := range to {
			if err := c.client.Rcpt(addr); err != nil {
				return fmt.Errorf("rcpt to failed: %v", err)
			}
		}

		w, err := c.client.Data()
		if err != nil {
			return fmt.Errorf("data failed: %v", err)
		}

		_, err = w.Write(msg)
		if err != nil {
			return fmt.Errorf("write failed: %v", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("close failed: %v", err)
		}

		return c.client.Quit()
	})
}
