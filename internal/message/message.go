package message

import (
	"fmt"
	"mime"
	"strings"
	"time"
)

// Message represents an email message
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTMLBody    string
	Headers     map[string]string
	Attachments []Attachment
	Date        time.Time
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// NewMessage creates a new email message
func NewMessage() *Message {
	return &Message{
		Headers: make(map[string]string),
		Date:    time.Now(),
	}
}

// AddTo adds a recipient to the To field
func (m *Message) AddTo(recipient string) {
	m.To = append(m.To, recipient)
}

// AddCc adds a recipient to the Cc field
func (m *Message) AddCc(recipient string) {
	m.Cc = append(m.Cc, recipient)
}

// AddBcc adds a recipient to the Bcc field
func (m *Message) AddBcc(recipient string) {
	m.Bcc = append(m.Bcc, recipient)
}

// AddHeader adds a custom header to the message
func (m *Message) AddHeader(key, value string) {
	m.Headers[key] = value
}

// AddAttachment adds an attachment to the message
func (m *Message) AddAttachment(filename string, contentType string, data []byte) {
	m.Attachments = append(m.Attachments, Attachment{
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
	})
}

// Build constructs the complete email message
func (m *Message) Build() (string, error) {
	var builder strings.Builder

	// Add standard headers
	builder.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ", ")))
	if len(m.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(m.Cc, ", ")))
	}
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", m.Date.Format(time.RFC1123Z)))

	// Add custom headers
	for key, value := range m.Headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Handle message body and attachments
	if len(m.Attachments) > 0 || m.HTMLBody != "" {
		// Create multipart boundary
		boundary := fmt.Sprintf("_boundary_%d_", time.Now().UnixNano())
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		builder.WriteString("\r\n")

		// Add text body
		if m.Body != "" {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			builder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
			builder.WriteString("\r\n")
			builder.WriteString(m.Body)
			builder.WriteString("\r\n")
		}

		// Add HTML body if present
		if m.HTMLBody != "" {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			builder.WriteString("Content-Type: text/html; charset=utf-8\r\n")
			builder.WriteString("\r\n")
			builder.WriteString(m.HTMLBody)
			builder.WriteString("\r\n")
		}

		// Add attachments
		for _, attachment := range m.Attachments {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			builder.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.ContentType))
			builder.WriteString("Content-Transfer-Encoding: base64\r\n")
			builder.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n",
				mime.QEncoding.Encode("utf-8", attachment.Filename)))
			builder.WriteString("\r\n")
			builder.WriteString(attachment.EncodeBase64())
			builder.WriteString("\r\n")
		}

		// End multipart
		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Simple text message
		builder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(m.Body)
	}

	return builder.String(), nil
}
