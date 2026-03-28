// pkg/stdlib/mail.go
// Mail module for Xxlang - Email sending functionality.
package stdlib

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/topxeq/xxlang/pkg/objects"
)

func init() {
	Register(&Module{
		Name: "mail",
		Exports: map[string]objects.Object{
			// newClient creates a new mail client configuration.
			// Usage: newClient(host, port, user, password) -> MailClient
			//        newClient(host, port, user, password, options) -> MailClient
			// Options can include:
			//   - "-from=email@example.com"
			//   - "-fromName=Sender Name"
			//   - "-tls"
			// Example:
			//   client := mail.newClient("smtp.example.com", 587, "user@example.com", "password", "-tls")
			"newClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 4 {
					return Error("newClient() requires at least 4 arguments")
				}

				host, ok := args[0].(*objects.String)
				if !ok {
					return Error("host must be a string")
				}

				port, ok := args[1].(*objects.Int)
				if !ok {
					return Error("port must be an integer")
				}

				user, ok := args[2].(*objects.String)
				if !ok {
					return Error("user must be a string")
				}

				password, ok := args[3].(*objects.String)
				if !ok {
					return Error("password must be a string")
				}

				client := &objects.MailClient{
					Host:     host.Value,
					Port:     int(port.Value),
					User:     user.Value,
					Password: password.Value,
					From:     user.Value,
					UseTLS:   true,
				}

				// Parse options
				for i := 4; i < len(args); i++ {
					if opt, ok := args[i].(*objects.String); ok {
						if strings.HasPrefix(opt.Value, "-from=") {
							client.From = strings.TrimPrefix(opt.Value, "-from=")
						} else if strings.HasPrefix(opt.Value, "-fromName=") {
							client.FromName = strings.TrimPrefix(opt.Value, "-fromName=")
						} else if opt.Value == "-tls" {
							client.UseTLS = true
						} else if opt.Value == "-noTls" {
							client.UseTLS = false
						}
					}
				}

				return client
			}),

			// send sends an email.
			// Usage: send(options) -> bool
			// Options (all as named arguments in a map):
			//   - host: SMTP server host
			//   - port: SMTP server port
			//   - user: SMTP username
			//   - password: SMTP password
			//   - from: sender email address
			//   - to: recipient(s), comma or semicolon separated
			//   - cc: CC recipients (optional)
			//   - bcc: BCC recipients (optional)
			//   - subject: email subject
			//   - body: email body (HTML or plain text)
			//   - attachFiles: attachments in format "name:path;name2:path2" (optional)
			//   - tls: use TLS (optional, default true)
			// Example:
			//   result := mail.send({
			//       "host": "smtp.example.com",
			//       "port": 587,
			//       "user": "user@example.com",
			//       "password": "password",
			//       "to": "recipient@example.com",
			//       "subject": "Hello",
			//       "body": "<h1>Hello World</h1>"
			//   })
			"send": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) < 1 {
					return Error("send() requires at least 1 argument")
				}

				// Parse options from map or multiple string arguments
				var host, user, password, from, fromName, to, cc, bcc, subject, body, attachFiles string
				var port int = 587
				useTLS := true

				// Check if first argument is a map
				if opts, ok := args[0].(*objects.Map); ok {
					for _, pair := range opts.Pairs {
						key, ok := pair.Key.(*objects.String)
						if !ok {
							continue
						}
						value := pair.Value

						switch key.Value {
						case "host":
							if s, ok := value.(*objects.String); ok {
								host = s.Value
							}
						case "port":
							if i, ok := value.(*objects.Int); ok {
								port = int(i.Value)
							}
						case "user":
							if s, ok := value.(*objects.String); ok {
								user = s.Value
							}
						case "password":
							if s, ok := value.(*objects.String); ok {
								password = s.Value
							}
						case "from":
							if s, ok := value.(*objects.String); ok {
								from = s.Value
							}
						case "fromName":
							if s, ok := value.(*objects.String); ok {
								fromName = s.Value
							}
						case "to":
							if s, ok := value.(*objects.String); ok {
								to = s.Value
							}
						case "cc":
							if s, ok := value.(*objects.String); ok {
								cc = s.Value
							}
						case "bcc":
							if s, ok := value.(*objects.String); ok {
								bcc = s.Value
							}
						case "subject":
							if s, ok := value.(*objects.String); ok {
								subject = s.Value
							}
						case "body":
							if s, ok := value.(*objects.String); ok {
								body = s.Value
							}
						case "attachFiles":
							if s, ok := value.(*objects.String); ok {
								attachFiles = s.Value
							}
						case "tls":
							if b, ok := value.(*objects.Bool); ok {
								useTLS = b.Value
							}
						}
					}
				} else {
					// Parse as key-value string arguments
					for i := 0; i < len(args); i++ {
						if opt, ok := args[i].(*objects.String); ok {
							if strings.HasPrefix(opt.Value, "-host=") {
								host = strings.TrimPrefix(opt.Value, "-host=")
							} else if strings.HasPrefix(opt.Value, "-port=") {
								portStr := strings.TrimPrefix(opt.Value, "-port=")
								port = parseMailPortFromString(portStr)
							} else if strings.HasPrefix(opt.Value, "-user=") {
								user = strings.TrimPrefix(opt.Value, "-user=")
							} else if strings.HasPrefix(opt.Value, "-password=") {
								password = strings.TrimPrefix(opt.Value, "-password=")
							} else if strings.HasPrefix(opt.Value, "-from=") {
								from = strings.TrimPrefix(opt.Value, "-from=")
							} else if strings.HasPrefix(opt.Value, "-fromName=") {
								fromName = strings.TrimPrefix(opt.Value, "-fromName=")
							} else if strings.HasPrefix(opt.Value, "-to=") {
								to = strings.TrimPrefix(opt.Value, "-to=")
							} else if strings.HasPrefix(opt.Value, "-cc=") {
								cc = strings.TrimPrefix(opt.Value, "-cc=")
							} else if strings.HasPrefix(opt.Value, "-bcc=") {
								bcc = strings.TrimPrefix(opt.Value, "-bcc=")
							} else if strings.HasPrefix(opt.Value, "-subject=") {
								subject = strings.TrimPrefix(opt.Value, "-subject=")
							} else if strings.HasPrefix(opt.Value, "-body=") {
								body = strings.TrimPrefix(opt.Value, "-body=")
							} else if strings.HasPrefix(opt.Value, "-attachFiles=") {
								attachFiles = strings.TrimPrefix(opt.Value, "-attachFiles=")
							} else if opt.Value == "-tls" {
								useTLS = true
							} else if opt.Value == "-noTls" {
								useTLS = false
							}
						}
					}
				}

				// Validate required fields
				if host == "" {
					return Error("host is required")
				}
				if user == "" {
					return Error("user is required")
				}
				if password == "" {
					return Error("password is required")
				}
				if to == "" && cc == "" && bcc == "" {
					return Error("at least one recipient (to/cc/bcc) is required")
				}
				if subject == "" {
					return Error("subject is required")
				}
				if body == "" {
					return Error("body is required")
				}

				// Default from to user
				if from == "" {
					from = user
				}

				// Build email
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)

				// Build recipients list
				var recipients []string
				if to != "" {
					recipients = append(recipients, splitMailEmails(to)...)
				}
				if cc != "" {
					recipients = append(recipients, splitMailEmails(cc)...)
				}
				if bcc != "" {
					recipients = append(recipients, splitMailEmails(bcc)...)
				}

				// Build headers
				headers := make(textproto.MIMEHeader)
				headers.Set("From", formatMailAddress(from, fromName))
				if to != "" {
					headers.Set("To", to)
				}
				if cc != "" {
					headers.Set("Cc", cc)
				}
				headers.Set("Subject", subject)
				headers.Set("MIME-Version", "1.0")
				headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary()))

				// Write headers
				for key, values := range headers {
					for _, value := range values {
						buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
					}
				}
				buf.WriteString("\r\n")

				// Add body part
				bodyPart, err := writer.CreatePart(textproto.MIMEHeader{
					"Content-Type":              []string{"text/html; charset=UTF-8"},
					"Content-Transfer-Encoding": []string{"quoted-printable"},
				})
				if err != nil {
					return Error("failed to create body part: " + err.Error())
				}

				qpWriter := quotedprintable.NewWriter(bodyPart)
				qpWriter.Write([]byte(body))
				qpWriter.Close()

				// Add attachments
				if attachFiles != "" {
					attachList := strings.Split(strings.ReplaceAll(attachFiles, ";", ","), ",")
					for _, attach := range attachList {
						attach = strings.TrimSpace(attach)
						if attach == "" {
							continue
						}

						// Parse attachment format: "displayName:filePath"
						parts := strings.SplitN(attach, ":", 2)
						if len(parts) != 2 {
							continue
						}
						displayName := strings.TrimSpace(parts[0])
						filePath := strings.TrimSpace(parts[1])

						// Read file
						data, err := os.ReadFile(filePath)
						if err != nil {
							return Error("failed to read attachment " + filePath + ": " + err.Error())
						}

						// Create attachment part
						_, filename := filepath.Split(filePath)
						if displayName != "" {
							filename = displayName
						}

						part, err := writer.CreatePart(textproto.MIMEHeader{
							"Content-Type":              []string{fmt.Sprintf("application/octet-stream; name=%s", filename)},
							"Content-Transfer-Encoding": []string{"base64"},
							"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=%s", filename)},
						})
						if err != nil {
							return Error("failed to create attachment part: " + err.Error())
						}

						encoder := base64.NewEncoder(base64.StdEncoding, part)
						encoder.Write(data)
						encoder.Close()
					}
				}

				writer.Close()

				// Send email
				addr := fmt.Sprintf("%s:%d", host, port)
				var auth smtp.Auth
				if user != "" && password != "" {
					auth = smtp.PlainAuth("", user, password, host)
				}

				var client *smtp.Client
				var sendErr error

				if useTLS {
					// TLS connection
					tlsConfig := &tls.Config{
						InsecureSkipVerify: false,
						ServerName:         host,
					}

					conn, err := tls.Dial("tcp", addr, tlsConfig)
					if err != nil {
						return Error("failed to connect: " + err.Error())
					}

					client, sendErr = smtp.NewClient(conn, host)
					if sendErr != nil {
						return Error("failed to create client: " + sendErr.Error())
					}
				} else {
					client, sendErr = smtp.Dial(addr)
					if sendErr != nil {
						return Error("failed to connect: " + sendErr.Error())
					}
				}
				defer client.Close()

				// Authenticate
				if auth != nil {
					if err := client.Auth(auth); err != nil {
						return Error("authentication failed: " + err.Error())
					}
				}

				// Set sender
				if err := client.Mail(from); err != nil {
					return Error("failed to set sender: " + err.Error())
				}

				// Set recipients
				for _, recipient := range recipients {
					if err := client.Rcpt(recipient); err != nil {
						return Error("failed to set recipient " + recipient + ": " + err.Error())
					}
				}

				// Send data
				w, err := client.Data()
				if err != nil {
					return Error("failed to prepare data: " + err.Error())
				}

				_, err = w.Write(buf.Bytes())
				if err != nil {
					return Error("failed to write data: " + err.Error())
				}

				err = w.Close()
				if err != nil {
					return Error("failed to close data: " + err.Error())
				}

				client.Quit()

				return Bool(true)
			}),

			// isMailClient checks if an object is a MailClient.
			"isMailClient": BuiltinFunc(func(args ...objects.Object) objects.Object {
				if len(args) != 1 {
					return Error("isMailClient() takes exactly 1 argument")
				}
				_, ok := args[0].(*objects.MailClient)
				return Bool(ok)
			}),
		},
	})
}

// Helper functions

func splitMailEmails(s string) []string {
	s = strings.ReplaceAll(s, ";", ",")
	var result []string
	for _, email := range strings.Split(s, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			result = append(result, email)
		}
	}
	return result
}

func formatMailAddress(email, name string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

func parseMailPortFromString(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
