package message

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}
	if msg.From != "test@example.com" {
		t.Errorf("Expected From to be test@example.com, got %s", msg.From)
	}
	if len(msg.To) != 1 {
		t.Errorf("Expected To to have 1 recipient, got %d", len(msg.To))
	}
	if len(msg.Cc) != 0 {
		t.Errorf("Expected Cc to be empty, got %d", len(msg.Cc))
	}
	if msg.Subject != "Test Subject" {
		t.Errorf("Expected Subject to be Test Subject, got %s", msg.Subject)
	}
	if msg.Body != "Test Body" {
		t.Errorf("Expected Body to be Test Body, got %s", msg.Body)
	}
}

func TestSetFrom(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.SetFrom("new@example.com")
	if msg.From != "new@example.com" {
		t.Errorf("Expected From to be 'new@example.com', got '%s'", msg.From)
	}
}

func TestAddTo(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.AddTo("new@example.com")
	if len(msg.To) != 2 {
		t.Errorf("Expected To to have 2 recipients, got %d", len(msg.To))
	}
	if msg.To[1] != "new@example.com" {
		t.Errorf("Expected To[1] to be 'new@example.com', got '%s'", msg.To[1])
	}
}

func TestAddCc(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.AddCc("cc@example.com")
	if len(msg.Cc) != 1 {
		t.Errorf("Expected Cc to have 1 recipient, got %d", len(msg.Cc))
	}
	if msg.Cc[0] != "cc@example.com" {
		t.Errorf("Expected Cc[0] to be 'cc@example.com', got '%s'", msg.Cc[0])
	}
}

func TestAddBcc(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.AddBcc("bcc@example.com")
	if len(msg.Bcc) != 1 {
		t.Errorf("Expected Bcc to have 1 recipient, got %d", len(msg.Bcc))
	}
	if msg.Bcc[0] != "bcc@example.com" {
		t.Errorf("Expected Bcc[0] to be 'bcc@example.com', got '%s'", msg.Bcc[0])
	}
}

func TestSetSubject(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.SetSubject("New Subject")
	if msg.Subject != "New Subject" {
		t.Errorf("Expected Subject to be 'New Subject', got '%s'", msg.Subject)
	}
}

func TestSetBody(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.SetBody("New Body")
	if msg.Body != "New Body" {
		t.Errorf("Expected Body to be 'New Body', got '%s'", msg.Body)
	}
}

func TestSetHTMLBody(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.SetHTMLBody("<p>Test HTML Body</p>")
	if msg.HTMLBody != "<p>Test HTML Body</p>" {
		t.Errorf("Expected HTMLBody to be '<p>Test HTML Body</p>', got '%s'", msg.HTMLBody)
	}
}

func TestAddAttachment(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Write some test data
	testData := []byte("test content")
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	err = msg.AddAttachment(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	if len(msg.Attachments) != 1 {
		t.Errorf("Expected Attachments to have 1 file, got %d", len(msg.Attachments))
	}
	if msg.Attachments[0].Filename != filepath.Base(tmpFile.Name()) {
		t.Errorf("Expected Attachment[0].Filename to be '%s', got '%s'", filepath.Base(tmpFile.Name()), msg.Attachments[0].Filename)
	}
	if string(msg.Attachments[0].Content) != "test content" {
		t.Errorf("Expected Attachment[0].Content to be 'test content', got '%s'", string(msg.Attachments[0].Content))
	}
}

func TestSetDate(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	now := time.Now()
	msg.SetDate(now)
	if !msg.Date.Equal(now) {
		t.Errorf("Expected Date to be %v, got %v", now, msg.Date)
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name        string
		msg         *Message
		expectedErr error
	}{
		{
			name: "Valid message",
			msg: &Message{
				From:    "from@example.com",
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Date:    time.Now(),
			},
			expectedErr: nil,
		},
		{
			name: "Missing From",
			msg: &Message{
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Date:    time.Now(),
			},
			expectedErr: errors.New("from address is required"),
		},
		{
			name: "Missing To",
			msg: &Message{
				From:    "from@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
				Date:    time.Now(),
			},
			expectedErr: errors.New("at least one recipient is required"),
		},
		{
			name: "Missing Subject",
			msg: &Message{
				From: "from@example.com",
				To:   []string{"to@example.com"},
				Body: "Test Body",
				Date: time.Now(),
			},
			expectedErr: errors.New("subject is required"),
		},
		{
			name: "Missing Body",
			msg: &Message{
				From:    "from@example.com",
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Date:    time.Now(),
			},
			expectedErr: errors.New("body is required"),
		},
		{
			name: "Missing Date",
			msg: &Message{
				From:    "from@example.com",
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			expectedErr: errors.New("date is required"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if tc.expectedErr != nil && (err == nil || err.Error() != tc.expectedErr.Error()) {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestAddRecipients(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")

	// Test AddTo
	msg.AddTo("to@example.com")
	if len(msg.To) != 2 || msg.To[1] != "to@example.com" {
		t.Errorf("AddTo failed: got %v", msg.To)
	}

	// Test AddCc
	msg.AddCc("cc@example.com")
	if len(msg.Cc) != 1 || msg.Cc[0] != "cc@example.com" {
		t.Errorf("AddCc failed: got %v", msg.Cc)
	}

	// Test AddBcc
	msg.AddBcc("bcc@example.com")
	if len(msg.Bcc) != 1 || msg.Bcc[0] != "bcc@example.com" {
		t.Errorf("AddBcc failed: got %v", msg.Bcc)
	}
}

func TestAddHeader(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	msg.AddHeader("X-Custom", "value")
	if msg.Headers["X-Custom"] != "value" {
		t.Errorf("AddHeader failed: got %v", msg.Headers)
	}
}

func TestBuild(t *testing.T) {
	msg := NewMessage("from@example.com", []string{"to@example.com"}, "Test Subject", "Test Body")

	result, err := msg.Build()
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// Check for required headers
	requiredHeaders := []string{
		"From: from@example.com",
		"To: to@example.com",
		"Subject: Test Subject",
		"Content-Type: text/plain; charset=utf-8",
	}

	for _, header := range requiredHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Missing header: %s", header)
		}
	}

	// Check for body
	if !strings.Contains(result, "Test Body") {
		t.Error("Body not found in result")
	}
}

func TestBuildWithAttachment(t *testing.T) {
	msg := NewMessage("from@example.com", []string{"to@example.com"}, "Test Subject", "Test Body")

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Write some test data
	testData := []byte("test attachment data")
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	err = msg.AddAttachment(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	result, err := msg.Build()
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// Check for multipart content type
	if !strings.Contains(result, "Content-Type: multipart/mixed") {
		t.Error("Missing multipart content type")
	}

	// Check for attachment headers
	attachmentHeaders := []string{
		"Content-Type: application/octet-stream",
		"Content-Transfer-Encoding: base64",
		"Content-Disposition: attachment",
	}

	for _, header := range attachmentHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Missing attachment header: %s", header)
		}
	}
}

func TestAddAttachment_NonExistentFile(t *testing.T) {
	msg := NewMessage("test@example.com", []string{"recipient@example.com"}, "Test Subject", "Test Body")
	err := msg.AddAttachment("nonexistent.txt")
	if err == nil {
		t.Error("Expected error when adding non-existent file")
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
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temporary file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString("Test attachment content")
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

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

	contentType := parsedMsg.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/") {
		t.Fatalf("Expected content type to be multipart, got %s", contentType)
	}

	boundary := strings.Split(contentType, "boundary=")[1]
	mr := multipart.NewReader(parsedMsg.Body, boundary)
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
		err := ValidateEmail(email)
		if err != nil {
			t.Errorf("Expected email '%s' to be valid, but got error: %v", email, err)
		}
	}

	for _, email := range invalidEmails {
		err := ValidateEmail(email)
		if err == nil {
			t.Errorf("Expected email '%s' to be invalid, but got no error", email)
		}
	}
}
