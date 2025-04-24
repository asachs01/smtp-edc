package client

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/asachs/smtp-edc/internal/message"
)

// MockSMTPClient is a mock implementation of the SMTPClient interface for testing.
type MockSMTPClient struct {
	ConnectFunc      func(server string, port int) error
	StartTLSFunc     func() error
	AuthenticateFunc func(authType, username, password string) error
	SendMessageFunc  func(msg *message.Message) error
	CloseFunc        func() error
}

func (m *MockSMTPClient) Connect(server string, port int) error {
	return m.ConnectFunc(server, port)
}

func (m *MockSMTPClient) StartTLS() error {
	return m.StartTLSFunc()
}

func (m *MockSMTPClient) Authenticate(authType, username, password string) error {
	return m.AuthenticateFunc(authType, username, password)
}

func (m *MockSMTPClient) SendMessage(msg *message.Message) error {
	return m.SendMessageFunc(msg)
}

func (m *MockSMTPClient) Close() error {
	return m.CloseFunc()
}

func TestNewSMTPClient(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		debug    bool
		want     *SMTPClient
	}{
		{
			name:     "valid config",
			hostname: "smtp.example.com",
			debug:    false,
			want: &SMTPClient{
				hostname: "smtp.example.com",
				debug:    false,
				retry: RetryConfig{
					MaxAttempts: 3,
					Delay:       2 * time.Second,
				},
				timeout: 30 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSMTPClient(tt.hostname, tt.debug)
			if got.hostname != tt.want.hostname {
				t.Errorf("NewSMTPClient().hostname = %v, want %v", got.hostname, tt.want.hostname)
			}
			if got.debug != tt.want.debug {
				t.Errorf("NewSMTPClient().debug = %v, want %v", got.debug, tt.want.debug)
			}
			if got.retry.MaxAttempts != tt.want.retry.MaxAttempts {
				t.Errorf("NewSMTPClient().retry.MaxAttempts = %v, want %v", got.retry.MaxAttempts, tt.want.retry.MaxAttempts)
			}
			if got.retry.Delay != tt.want.retry.Delay {
				t.Errorf("NewSMTPClient().retry.Delay = %v, want %v", got.retry.Delay, tt.want.retry.Delay)
			}
			if got.timeout != tt.want.timeout {
				t.Errorf("NewSMTPClient().timeout = %v, want %v", got.timeout, tt.want.timeout)
			}
		})
	}
}

func TestSetRetryConfig(t *testing.T) {
	client := NewSMTPClient("localhost", false)
	maxAttempts := 5
	delay := 3 * time.Second

	client.SetRetryConfig(maxAttempts, delay)

	if client.retry.MaxAttempts != maxAttempts {
		t.Errorf("SetRetryConfig() MaxAttempts = %v, want %v", client.retry.MaxAttempts, maxAttempts)
	}
	if client.retry.Delay != delay {
		t.Errorf("SetRetryConfig() Delay = %v, want %v", client.retry.Delay, delay)
	}
}

func TestSetTimeout(t *testing.T) {
	client := NewSMTPClient("localhost", false)
	timeout := 60 * time.Second

	client.SetTimeout(timeout)

	if client.timeout != timeout {
		t.Errorf("SetTimeout() = %v, want %v", client.timeout, timeout)
	}
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		port    int
		setup   func(*SMTPClient)
		wantErr bool
	}{
		{
			name:   "successful connect",
			server: "smtp.example.com",
			port:   587,
			setup: func(c *SMTPClient) {
				c.retry.MaxAttempts = 1
				// Mock the connection
				c.conn = &mockConn{
					readFunc: func(b []byte) (n int, err error) {
						copy(b, "220 smtp.example.com ESMTP ready\r\n")
						return len("220 smtp.example.com ESMTP ready\r\n"), nil
					},
					writeFunc: func(b []byte) (n int, err error) {
						return len(b), nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:   "connect error",
			server: "invalid.example.com",
			port:   587,
			setup: func(c *SMTPClient) {
				c.retry.MaxAttempts = 1
				// Mock the connection to return an error
				c.conn = &mockConn{
					readFunc: func(b []byte) (n int, err error) {
						return 0, errors.New("connection refused")
					},
					writeFunc: func(b []byte) (n int, err error) {
						return 0, errors.New("connection refused")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSMTPClient("localhost", false)
			if tt.setup != nil {
				tt.setup(client)
			}

			err := client.Connect(tt.server, tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*SMTPClient)
		wantErr bool
	}{
		{
			name: "successful close",
			setup: func(c *SMTPClient) {
				c.conn = &mockConn{
					closeFunc: func() error { return nil },
				}
			},
			wantErr: false,
		},
		{
			name: "close error",
			setup: func(c *SMTPClient) {
				c.conn = &mockConn{
					closeFunc: func() error { return errors.New("close failed") },
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSMTPClient("localhost", false)
			if tt.setup != nil {
				tt.setup(client)
			}

			err := client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mockConn implements net.Conn interface for testing
type mockConn struct {
	closeFunc func() error
	readFunc  func(b []byte) (n int, err error)
	writeFunc func(b []byte) (n int, err error)
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readFunc != nil {
		return m.readFunc(b)
	}
	return 0, nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.writeFunc != nil {
		return m.writeFunc(b)
	}
	return len(b), nil
}

func (m *mockConn) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }
