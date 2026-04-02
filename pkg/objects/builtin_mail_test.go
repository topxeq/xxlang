// pkg/objects/builtin_mail_test.go
// Tests for mail types (MailClient)
package objects

import (
	"strings"
	"testing"
)

func TestMailClient(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *MailClient
		validate func(t *testing.T, mc *MailClient)
	}{
		{
			name: "basic mail client",
			setup: func() *MailClient {
				return &MailClient{
					Host:     "smtp.example.com",
					Port:     587,
					User:     "user@example.com",
					Password: "password",
					From:     "user@example.com",
					FromName: "Test User",
					UseTLS:   true,
				}
			},
			validate: func(t *testing.T, mc *MailClient) {
				if mc.Host != "smtp.example.com" {
					t.Errorf("expected Host smtp.example.com, got %s", mc.Host)
				}
				if mc.Port != 587 {
					t.Errorf("expected Port 587, got %d", mc.Port)
				}
				if mc.User != "user@example.com" {
					t.Errorf("expected User user@example.com, got %s", mc.User)
				}
				if mc.Password != "password" {
					t.Errorf("expected Password, got %s", mc.Password)
				}
				if mc.From != "user@example.com" {
					t.Errorf("expected From user@example.com, got %s", mc.From)
				}
				if mc.FromName != "Test User" {
					t.Errorf("expected FromName Test User, got %s", mc.FromName)
				}
				if !mc.UseTLS {
					t.Errorf("expected UseTLS true, got %v", mc.UseTLS)
				}
			},
		},
		{
			name: "minimal mail client",
			setup: func() *MailClient {
				return &MailClient{
					Host:   "localhost",
					Port:   25,
					User:   "",
					From:   "",
					UseTLS: false,
				}
			},
			validate: func(t *testing.T, mc *MailClient) {
				if mc.Host != "localhost" {
					t.Errorf("expected Host localhost, got %s", mc.Host)
				}
				if mc.Port != 25 {
					t.Errorf("expected Port 25, got %d", mc.Port)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := tt.setup()

			// Test Type()
			if mc.Type() != "MAIL_CLIENT" {
				t.Errorf("expected type MAIL_CLIENT, got %s", mc.Type())
			}

			// Test TypeTag()
			if mc.TypeTag() != TypeTag(101) {
				t.Errorf("expected TypeTag 101, got %d", mc.TypeTag())
			}

			// Test Inspect()
			inspect := mc.Inspect()
			if inspect == "" {
				t.Errorf("expected non-empty Inspect string")
			}
			if !strings.Contains(inspect, "MailClient") {
				t.Errorf("expected Inspect to contain 'MailClient', got %s", inspect)
			}

			// Test ToBool()
			if !mc.ToBool().Value {
				t.Errorf("MailClient.ToBool() should return true")
			}

			// Test HashKey()
			key := mc.HashKey()
			if key.Type != "MAIL_CLIENT" {
				t.Errorf("expected HashKey.Type MAIL_CLIENT, got %s", key.Type)
			}

			// Custom validation
			if tt.validate != nil {
				tt.validate(t, mc)
			}
		})
	}
}

func TestMailClient_EdgeCases(t *testing.T) {
	// Test with zero values
	mc := &MailClient{}
	if mc.Type() != "MAIL_CLIENT" {
		t.Errorf("expected type MAIL_CLIENT for zero value")
	}

	// Test with very long values
	mc = &MailClient{
		Host:     "very-long-host-name-that-exceeds-normal-lengths.example.com",
		Port:     65535,
		User:     "verylongusername@example.comwithadditionallocalpart",
		Password: "very-long-password-with-many-special-chars!@#$%^&*()",
		From:     "from@example.com",
		FromName: "Very Long Name With Many Words And Characters",
		UseTLS:   true,
	}

	if mc.Type() != "MAIL_CLIENT" {
		t.Errorf("type check failed for large mail client")
	}
}
