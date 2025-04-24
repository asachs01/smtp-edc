package message

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMessage(t *testing.T) {
	from := "test@example.com"
	to := []string{"recipient1@example.com", "recipient2@example.com"}
	subject := "Test Subject"
	body := "Test Body"

	msg := NewMessage(from, to, subject, body)

	if msg.From != from {
		t.Errorf("Expected From to be %s, but got %s", from, msg.From)
	}

	if len(msg.To) != len(to) {
		t.Errorf("Expected To length to be %d, but got %d", len(to), len(msg.To))
	}
	for i, recipient := range to {
		if msg.To[i] != recipient {
			t.Errorf("Expected To[%d] to be %s, but got %s", i, recipient, msg.To[i])
		}
	}

	if msg.Subject != subject {
		t.Errorf("Expected Subject to be %s, but got %s", subject, msg.Subject)
	}

	if msg.Body != body {
		t.Errorf("Expected Body to be %s, but got %s", body, msg.Body)
	}

	if msg.Attachments == nil {
		t.Errorf("Expected Attachments to be initialized, but it's nil")
	}
}

func TestAddAttachment(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("Test attachment content")
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	err = msg.AddAttachment(tmpFile.Name())
	if err != nil {
		t.Errorf("Failed to add attachment: %v", err)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, but got %d", len(msg.Attachments))
	}

	if msg.Attachments[0].Filename != filepath.Base(tmpFile.Name()) {
		t.Errorf("Expected attachment filename to be %s, but got %s", filepath.Base(tmpFile.Name()), msg.Attachments[0].Filename)
	}

	content, err := io.ReadAll(msg.Attachments[0].Content)
	if err != nil {
		t.Fatalf("Failed to read attachment content: %v", err)
	}

	if string(content) != "Test attachment content" {
		t.Errorf("Expected attachment content to be 'Test attachment content', but got '%s'", string(content))
	}
}

func TestAddAttachment_NonExistentFile(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	err := msg.AddAttachment("nonexistent.txt")
	if err == nil {
		t.Errorf("Expected an error when adding a non-existent file, but got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected error to be os.ErrNotExist, but got %v", err)
	}
}

func TestBuildMessage(t *testing.T) {
	from := "test@example.com"
	to := []string{"recipient1@example.com", "recipient2@example.com"}
	subject := "Test Subject"
	body := "Test Body"

	msg := NewMessage(from, to, subject, body)
	rawMsg, err := msg.BuildMessage()
	if err != nil {
		t.Fatalf("BuildMessage returned an error: %v", err)
	}

	parsedMsg, err := mail.ReadMessage(bytes.NewReader(rawMsg))
	if err != nil {
		t.Fatalf("Failed to parse built message: %v", err)
	}

	if parsedMsg.Header.Get("From") != from {
		t.Errorf("Expected 'From' header to be %s, but got %s", from, parsedMsg.Header.Get("From"))
	}
	if strings.Join(parsedMsg.Header["To"], ",") != strings.Join(to, ",") {
		t.Errorf("Expected 'To' header to be %v, but got %v", to, parsedMsg.Header["To"])
	}

	if parsedMsg.Header.Get("Subject") != subject {
		t.Errorf("Expected 'Subject' header to be %s, but got %s", subject, parsedMsg.Header.Get("Subject"))
	}

	bodyBytes, err := io.ReadAll(parsedMsg.Body)
	if err != nil {
		t.Fatalf("Failed to read message body: %v", err)
	}

	if strings.TrimSpace(string(bodyBytes)) != body {
		t.Errorf("Expected message body to be %s, but got %s", body, string(bodyBytes))
	}
}

func TestBuildMessage_WithAttachment(t *testing.T) {
	from := "test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Test Subject"
	body := "Test Body"

	msg := NewMessage(from, to, subject, body)

	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("Test attachment content")
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	err = msg.AddAttachment(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	rawMsg, err := msg.BuildMessage()
	if err != nil {
		t.Fatalf("BuildMessage returned an error: %v", err)
	}

	parsedMsg, err := mail.ReadMessage(bytes.NewReader(rawMsg))
	if err != nil {
		t.Fatalf("Failed to parse built message: %v", err)
	}

	mediaType, params, err := parsedMsg.Header.Get("Content-Type")
	if err != nil {
		t.Fatalf("Failed to get content type: %v", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("Expected content type to be multipart, got %s", mediaType)
	}

	mr := multipart.NewReader(parsedMsg.Body, params["boundary"])
	part, err := mr.NextPart()
	if err != nil {
		t.Fatalf("Failed to get the first part: %v", err)
	}
	if part.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("Expected Content-Type for first part to be text/plain; charset=utf-8")
	}
	bodyBytes, err := io.ReadAll(part)
	if err != nil {
		t.Fatalf("Failed to read message body: %v", err)
	}
	if string(bodyBytes) != body {
		t.Fatalf("Expected message body to be '%s', but got '%s'", body, string(bodyBytes))
	}
	part, err = mr.NextPart()
	if err != nil {
		t.Fatalf("Failed to get the second part: %v", err)
	}
	if part.Header.Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("Expected Content-Type for second part to be application/octet-stream")
	}
	if part.Header.Get("Content-Disposition") != "attachment; filename=\""+filepath.Base(tmpFile.Name())+"\"" {
		t.Fatalf("Expected Content-Disposition for second part to be attachment; filename=\"%s\"", filepath.Base(tmpFile.Name()))
	}
	attachmentContent, err := io.ReadAll(part)
	if err != nil {
		t.Fatalf("Failed to read attachment content: %v", err)
	}
	if string(attachmentContent) != "Test attachment content" {
		t.Fatalf("Expected attachment content to be '%s', but got '%s'", "Test attachment content", string(attachmentContent))
	}
}

func TestBuildMessage_NoBody(t *testing.T) {
	from := "test@example.com"
	to := []string{"recipient@example.com"}
	subject := "Test Subject"

	msg := NewMessage(from, to, subject, "")
	rawMsg, err := msg.BuildMessage()
	if err != nil {
		t.Fatalf("BuildMessage returned an error: %v", err)
	}

	parsedMsg, err := mail.ReadMessage(bytes.NewReader(rawMsg))
	if err != nil {
		t.Fatalf("Failed to parse built message: %v", err)
	}
	bodyBytes, err := io.ReadAll(parsedMsg.Body)
	if err != nil {
		t.Fatalf("Failed to read message body: %v", err)
	}

	if len(string(bodyBytes)) != 0 {
		t.Errorf("Expected message body to be empty, but got %s", string(bodyBytes))
	}
}

func TestValidateEmail(t *testing.T) {
	validEmails := []string{"test@example.com", "test.test@subdomain.example.co.uk", "123@example.com"}
	invalidEmails := []string{"test", "test@", "@example.com", "test@@example.com"}

	for _, email := range validEmails {
		err := validateEmail(email)
		if err != nil {
			t.Errorf("Expected email '%s' to be valid, but got error: %v", email, err)
		}
	}

	for _, email := range invalidEmails {
		err := validateEmail(email)
		if err == nil {
			t.Errorf("Expected email '%s' to be invalid, but got no error", email)
		}
	}
}

func TestBuildPart(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		contentType    string
		filename       string
		expectedHeader textproto.MIMEHeader
	}{
		{
			name:        "simple text part",
			body:        "simple text body",
			contentType: "text/plain; charset=utf-8",
			filename:    "",
			expectedHeader: textproto.MIMEHeader{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
		},
		{
			name:        "attachment part",
			body:        "attachment body",
			contentType: "application/octet-stream",
			filename:    "test.txt",
			expectedHeader: textproto.MIMEHeader{
				"Content-Type":        []string{"application/octet-stream"},
				"Content-Disposition": []string{"attachment; filename=\"test.txt\""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			err := buildPart(&buffer, tt.body, tt.contentType, tt.filename)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			part, err := mail.ReadMessage(&buffer)
			if err != nil {
				t.Fatalf("Failed to parse part: %v", err)
			}

			if len(part.Header) != len(tt.expectedHeader) {
				t.Fatalf("Expected %d header entries, got %d", len(tt.expectedHeader), len(part.Header))
			}

			for key, expectedValues := range tt.expectedHeader {
				actualValues, exists := part.Header[key]
				if !exists {
					t.Fatalf("Expected header '%s' to exist", key)
				}
				if len(actualValues) != len(expectedValues) {
					t.Fatalf("Expected %d values for header '%s', got %d", len(expectedValues), key, len(actualValues))
				}
				for i, expectedValue := range expectedValues {
					if actualValues[i] != expectedValue {
						t.Fatalf("Expected header '%s' value at index %d to be '%s', got '%s'", key, i, expectedValue, actualValues[i])
					}
				}
			}
		})
	}
}

func TestBuildMultipart(t *testing.T) {
	tests := []struct {
		name           string
		parts          []string
		expectedLength int
	}{
		{
			name:           "single part",
			parts:          []string{"part1"},
			expectedLength: 1,
		},
		{
			name:           "multiple parts",
			parts:          []string{"part1", "part2", "part3"},
			expectedLength: 3,
		},
		{
			name:           "empty parts",
			parts:          []string{},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			writer := multipart.NewWriter(&buffer)
			err := buildMultipart(writer, tt.parts)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			err = writer.Close()
			if err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

			mr := multipart.NewReader(&buffer, writer.Boundary())
			count := 0
			for {
				_, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Unexpected error while reading part: %v", err)
				}
				count++
			}

			if count != tt.expectedLength {
				t.Fatalf("Expected %d parts, got %d", tt.expectedLength, count)
			}
		})
	}
}