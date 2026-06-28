package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/stayflow/stayflow-track/internal/config"
)

// Sender sends transactional emails (password resets, account notifications).
type Sender struct {
	cfg     config.EmailConfig
	appName string
}

// New creates an email Sender. If cfg.Enabled is false, Send() will be a no-op.
func New(cfg config.EmailConfig, appName string) *Sender {
	return &Sender{cfg: cfg, appName: appName}
}

// IsEnabled returns whether email sending is configured and active.
func (s *Sender) IsEnabled() bool {
	return s.cfg.Enabled && s.cfg.SMTPHost != ""
}

// Send sends a single email. Returns nil if email is disabled (development).
func (s *Sender) Send(ctx context.Context, to, subject, htmlBody string) error {
	if !s.IsEnabled() {
		return nil
	}

	from := s.cfg.FromAddress
	fromHeader := fmt.Sprintf("%s <%s>", s.cfg.FromName, from)

	msg := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprintf("%d", s.cfg.SMTPPort))

	// Use a deadline derived from context, defaulting to 10 seconds
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}

	conn, err := net.DialTimeout("tcp", addr, time.Until(deadline))
	if err != nil {
		return fmt.Errorf("email dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(deadline)

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("email client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// Start TLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{
			ServerName: s.cfg.SMTPHost,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("email starttls: %w", err)
		}
	}

	// Authenticate if credentials provided
	if s.cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("email auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("email from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("email rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("email data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("email write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("email close: %w", err)
	}

	return client.Quit()
}

// SendPasswordReset sends a password reset email with a tokenized link.
func (s *Sender) SendPasswordReset(ctx context.Context, to, resetToken, appDomain string) error {
	resetLink := fmt.Sprintf("https://%s/reset-password?token=%s", appDomain, resetToken)

	subject := "Password Reset — " + s.appName
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 20px;">
  <h2>Password Reset Request</h2>
  <p>You requested a password reset for your %s account.</p>
  <p>Click the link below to reset your password. This link expires in 1 hour.</p>
  <p><a href="%s" style="display: inline-block; padding: 12px 24px; background: #2563eb; color: white; text-decoration: none; border-radius: 6px;">Reset Password</a></p>
  <p style="color: #6b7280; font-size: 14px;">If you didn't request this, please ignore this email. Your password will remain unchanged.</p>
  <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;">
  <p style="color: #9ca3af; font-size: 12px;">This is an automated email from %s. Do not reply to this message.</p>
</body>
</html>`, s.appName, resetLink, s.appName)

	return s.Send(ctx, to, subject, body)
}
