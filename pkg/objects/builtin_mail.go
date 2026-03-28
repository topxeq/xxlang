// pkg/objects/builtin_mail.go
// Email types for Xxlang (MailClient object type)
// The mail functions have been moved to the mail module.
package objects

import (
	"fmt"
)

// MailClient represents an email client configuration
type MailClient struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// Type returns the object type
func (m *MailClient) Type() ObjectType { return "MAIL_CLIENT" }

// TypeTag returns the type tag
func (m *MailClient) TypeTag() TypeTag { return TypeTag(101) }

// Inspect returns a string representation
func (m *MailClient) Inspect() string {
	return fmt.Sprintf("MailClient(host=%s:%d, user=%s)", m.Host, m.Port, m.User)
}

// ToBool returns true
func (m *MailClient) ToBool() *Bool { return TRUE }

// HashKey returns a hash key
func (m *MailClient) HashKey() HashKey {
	return HashKey{Type: "MAIL_CLIENT", Value: 0}
}
