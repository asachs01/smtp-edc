package message

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"text/template/parse"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Data    map[string]interface{}
}

// Template represents an email template
type Template struct {
	subject *template.Template
	text    *template.Template
	html    *template.Template
}

// LoadTemplate loads a template from files
func LoadTemplate(subjectTemplate, textTemplate, htmlTemplate string) (*Template, error) {
	t := &Template{}
	var err error

	// Load subject template
	if subjectTemplate != "" {
		t.subject, err = template.New(filepath.Base(subjectTemplate)).ParseFiles(subjectTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subject template: %v", err)
		}
	}

	// Load text template
	if textTemplate != "" {
		t.text, err = template.New(filepath.Base(textTemplate)).ParseFiles(textTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse text template: %v", err)
		}
	}

	// Load HTML template
	if htmlTemplate != "" {
		t.html, err = template.New(filepath.Base(htmlTemplate)).ParseFiles(htmlTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML template: %v", err)
		}
	}

	return t, nil
}

// LoadTemplateFromString loads a template from strings
func LoadTemplateFromString(subjectTemplate, textTemplate, htmlTemplate string) (*Template, error) {
	t := &Template{}
	var err error

	// Load subject template
	if subjectTemplate != "" {
		t.subject, err = template.New("subject").Parse(subjectTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subject template: %v", err)
		}
	}

	// Load text template
	if textTemplate != "" {
		t.text, err = template.New("text").Parse(textTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse text template: %v", err)
		}
	}

	// Load HTML template
	if htmlTemplate != "" {
		t.html, err = template.New("html").Parse(htmlTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML template: %v", err)
		}
	}

	return t, nil
}

// Execute renders the template with the given data
func (t *Template) Execute(data *TemplateData) (*Message, error) {
	msg := NewMessage(data.From, data.To, data.Subject, "")
	msg.Cc = data.Cc
	msg.Bcc = data.Bcc

	// Render subject
	if t.subject != nil {
		var subject bytes.Buffer
		if err := t.subject.Execute(&subject, data); err != nil {
			return nil, fmt.Errorf("failed to render subject: %v", err)
		}
		msg.Subject = subject.String()
	} else {
		msg.Subject = data.Subject
	}

	// Render text body
	if t.text != nil {
		var body bytes.Buffer
		if err := t.text.Execute(&body, data); err != nil {
			return nil, fmt.Errorf("failed to render text body: %v", err)
		}
		msg.Body = body.String()
	}

	// Render HTML body
	if t.html != nil {
		var body bytes.Buffer
		if err := t.html.Execute(&body, data); err != nil {
			return nil, fmt.Errorf("failed to render HTML body: %v", err)
		}
		msg.HTMLBody = body.String()
	}

	return msg, nil
}

// GetTemplateFields returns a list of fields used in the template
func (t *Template) GetTemplateFields() []string {
	fields := make(map[string]bool)

	// Helper function to extract fields from a template
	extractFields := func(t *template.Template) {
		if t == nil {
			return
		}
		var walk func(node parse.Node)
		walk = func(node parse.Node) {
			if node == nil {
				return
			}
			switch n := node.(type) {
			case *parse.ActionNode:
				for _, cmd := range n.Pipe.Cmds {
					for _, arg := range cmd.Args {
						if field, ok := arg.(*parse.FieldNode); ok {
							fields[strings.Join(field.Ident, ".")] = true
						}
					}
				}
			case *parse.ListNode:
				if n != nil {
					for _, node := range n.Nodes {
						walk(node)
					}
				}
			case *parse.IfNode:
				walk(n.Pipe)
				walk(n.List)
				walk(n.ElseList)
			case *parse.RangeNode:
				walk(n.Pipe)
				walk(n.List)
				walk(n.ElseList)
			case *parse.WithNode:
				walk(n.Pipe)
				walk(n.List)
				walk(n.ElseList)
			}
		}
		walk(t.Tree.Root)
	}

	extractFields(t.subject)
	extractFields(t.text)
	extractFields(t.html)

	// Convert map to slice
	result := make([]string, 0, len(fields))
	for field := range fields {
		result = append(result, field)
	}
	return result
}
