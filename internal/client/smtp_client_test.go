package client

import (
	"bytes"
	"errors"
	"io"
	"net/smtp"
	"testing"
)

// MockSMTPClient is a mock implementation of the SMTPClient interface for testing.
type MockSMTPClient struct {
	SendMailFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	CloseFunc    func() error
	AuthFunc     func(a smtp.Auth) error
	HelloFunc    func(localName string) error
	MailFunc     func(from string) error
	RcptFunc     func(to string) error
	DataFunc     func() (io.WriteCloser, error)
	QuitFunc     func() error
}

func (m *MockSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return m.SendMailFunc(addr, a, from, to, msg)
}

func (m *MockSMTPClient) Close() error {
	return m.CloseFunc()
}

func (m *MockSMTPClient) Auth(a smtp.Auth) error {
	return m.AuthFunc(a)
}

func (m *MockSMTPClient) Hello(localName string) error {
	return m.HelloFunc(localName)
}

func (m *MockSMTPClient) Mail(from string) error {
	return m.MailFunc(from)
}

func (m *MockSMTPClient) Rcpt(to string) error {
	return m.RcptFunc(to)
}

func (m *MockSMTPClient) Data() (io.WriteCloser, error) {
	return m.DataFunc()
}

func (m *MockSMTPClient) Quit() error {
	return m.QuitFunc()
}

func TestNewSMTPClient(t *testing.T) {
	t.Run("Successful connection", func(t *testing.T) {
		_, err := NewSMTPClient("localhost:25", false)
		if err != nil {
			t.Errorf("NewSMTPClient returned an error: %v", err)
		}
	})
}

func TestSend(t *testing.T) {
	testCases := []struct {
		name        string
		mockClient  *MockSMTPClient
		expectedErr error
	}{
		{
			name: "Successful send",
			mockClient: &MockSMTPClient{
				SendMailFunc: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name: "SendMail error",
			mockClient: &MockSMTPClient{
				SendMailFunc: func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
					return errors.New("sendmail error")
				},
			},
			expectedErr: errors.New("sendmail error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{
				client: tc.mockClient,
			}
			err := c.Send("localhost:25", nil, "from@example.com", []string{"to@example.com"}, []byte("test message"))
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if tc.expectedErr != nil && (err == nil || err.Error() != tc.expectedErr.Error()) {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestClose(t *testing.T) {
	testCases := []struct {
		name        string
		mockClient  *MockSMTPClient
		expectedErr error
	}{
		{
			name: "Successful close",
			mockClient: &MockSMTPClient{
				CloseFunc: func() error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name: "Close error",
			mockClient: &MockSMTPClient{
				CloseFunc: func() error {
					return errors.New("close error")
				},
			},
			expectedErr: errors.New("close error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{
				client: tc.mockClient,
			}
			err := c.Close()
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if tc.expectedErr != nil && (err == nil || err.Error() != tc.expectedErr.Error()) {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestSendRaw(t *testing.T) {
	testCases := []struct {
		name        string
		mockClient  *MockSMTPClient
		expectedErr error
	}{
		{
			name: "Successful send raw",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: nil,
		},
		{
			name: "Auth error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return errors.New("auth error")
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: errors.New("auth error"),
		},
		{
			name: "Hello error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return errors.New("hello error")
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: errors.New("hello error"),
		},
		{
			name: "Mail error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return errors.New("mail error")
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: errors.New("mail error"),
		},
		{
			name: "Rcpt error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return errors.New("rcpt error")
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: errors.New("rcpt error"),
		},
		{
			name: "Data error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return nil, errors.New("data error")
				},
				QuitFunc: func() error {
					return nil
				},
			},
			expectedErr: errors.New("data error"),
		},
		{
			name: "Quit error",
			mockClient: &MockSMTPClient{
				AuthFunc: func(a smtp.Auth) error {
					return nil
				},
				HelloFunc: func(localName string) error {
					return nil
				},
				MailFunc: func(from string) error {
					return nil
				},
				RcptFunc: func(to string) error {
					return nil
				},
				DataFunc: func() (io.WriteCloser, error) {
					return &mockWriteCloser{}, nil
				},
				QuitFunc: func() error {
					return errors.New("quit error")
				},
			},
			expectedErr: errors.New("quit error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{
				client: tc.mockClient,
			}
			err := c.SendRaw("localhost:25", nil, "from@example.com", []string{"to@example.com"}, []byte("test message"))
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if tc.expectedErr != nil && (err == nil || err.Error() != tc.expectedErr.Error()) {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

type mockWriteCloser struct{}

func (mwc *mockWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (mwc *mockWriteCloser) Close() error {
	return nil
}