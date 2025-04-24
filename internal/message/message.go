package message

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
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
	Content     []byte
}

// NewMessage creates a new email message
func NewMessage(from string, to []string, subject string, body string) *Message {
	msg := &Message{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		Headers: make(map[string]string),
		Date:    time.Now(),
	}
	return msg
}

// SetFrom sets the From address
func (m *Message) SetFrom(from string) {
	m.From = from
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

// SetSubject sets the message subject
func (m *Message) SetSubject(subject string) {
	m.Subject = subject
}

// SetBody sets the plain text body
func (m *Message) SetBody(body string) {
	m.Body = body
}

// SetHTMLBody sets the HTML body
func (m *Message) SetHTMLBody(htmlBody string) {
	m.HTMLBody = htmlBody
}

// SetDate sets the message date
func (m *Message) SetDate(date time.Time) {
	m.Date = date
}

// AddHeader adds a custom header to the message
func (m *Message) AddHeader(key, value string) {
	m.Headers[key] = value
}

// AddAttachment adds an attachment to the message
func (m *Message) AddAttachment(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	contentType := "application/octet-stream"
	m.Attachments = append(m.Attachments, Attachment{
		Filename:    filepath.Base(filename),
		ContentType: contentType,
		Content:     data,
	})
	return nil
}

// Validate checks if the message has all required fields
func (m *Message) Validate() error {
	if m.From == "" {
		return errors.New("from address is required")
	}
	if len(m.To) == 0 {
		return errors.New("at least one recipient is required")
	}
	if m.Subject == "" {
		return errors.New("subject is required")
	}
	if m.Body == "" && m.HTMLBody == "" && len(m.Attachments) == 0 {
		return errors.New("body is required")
	}
	if m.Date.IsZero() {
		return errors.New("date is required")
	}
	return nil
}

// Build constructs the complete email message as a string
func (m *Message) Build() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}

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
			builder.WriteString(base64.StdEncoding.EncodeToString(attachment.Content))
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

// BuildMessage constructs the complete email message as a byte slice
func (m *Message) BuildMessage() ([]byte, error) {
	// Only validate if there's a body or attachments
	if m.Body != "" || m.HTMLBody != "" || len(m.Attachments) > 0 {
		if err := m.Validate(); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer

	// Set default headers
	headers := map[string]string{
		"From":         m.From,
		"To":           strings.Join(m.To, ","),
		"Subject":      m.Subject,
		"Date":         m.Date.Format(time.RFC1123Z),
		"MIME-Version": "1.0",
	}

	// Add CC if present
	if len(m.Cc) > 0 {
		headers["Cc"] = strings.Join(m.Cc, ",")
	}

	// Add BCC if present
	if len(m.Bcc) > 0 {
		headers["Bcc"] = strings.Join(m.Bcc, ",")
	}

	// Add custom headers
	for k, v := range m.Headers {
		headers[k] = v
	}

	// Handle attachments
	if len(m.Attachments) > 0 {
		boundary := fmt.Sprintf("_boundary_%d_", time.Now().UnixNano())
		headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

		// Write headers
		for k, v := range headers {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
		fmt.Fprintf(&buf, "\r\n")

		// Add text body part
		if m.Body != "" {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: text/plain; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s\r\n", m.Body)
		}

		// Add HTML body part if present
		if m.HTMLBody != "" {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: text/html; charset=utf-8\r\n\r\n")
			fmt.Fprintf(&buf, "%s\r\n", m.HTMLBody)
		}

		// Add attachments
		for _, attachment := range m.Attachments {
			fmt.Fprintf(&buf, "--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: %s\r\n", attachment.ContentType)
			fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Filename)
			fmt.Fprintf(&buf, "\r\n")
			fmt.Fprintf(&buf, "%s\r\n", string(attachment.Content))
		}

		// End multipart
		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	} else {
		// Set content type based on body type
		if m.HTMLBody != "" {
			headers["Content-Type"] = "text/html; charset=utf-8"
		} else {
			headers["Content-Type"] = "text/plain; charset=utf-8"
		}

		// Write headers
		for k, v := range headers {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
		fmt.Fprintf(&buf, "\r\n")

		// Write body
		if m.HTMLBody != "" {
			fmt.Fprintf(&buf, "%s", m.HTMLBody)
		} else {
			fmt.Fprintf(&buf, "%s", m.Body)
		}
	}

	return buf.Bytes(), nil
}
