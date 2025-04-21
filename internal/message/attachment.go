package message

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// ReadFileAttachment reads a file and creates an attachment
func ReadFileAttachment(filename string) (*Attachment, error) {
	// Read file content
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Determine content type based on file extension
	contentType := determineContentType(filename)

	return &Attachment{
		Filename:    filepath.Base(filename),
		ContentType: contentType,
		Data:        data,
	}, nil
}

// EncodeBase64 encodes the attachment data in base64
func (a *Attachment) EncodeBase64() string {
	return base64.StdEncoding.EncodeToString(a.Data)
}

// determineContentType determines the MIME type based on file extension
func determineContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch strings.ToLower(ext) {
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
