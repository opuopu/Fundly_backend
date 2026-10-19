package email

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// SMTPConfig holds the credentials for an SMTP relay (e.g. Gmail with an app password).
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string // provider app password, never the account login password
	From     string
	FromName string
}

// SMTPMailer sends transactional emails through an SMTP server. For Gmail use
// Host=smtp.gmail.com, Port=587, Username=<your address>, Password=<16-char app password>.
// net/smtp upgrades the connection with STARTTLS automatically on port 587.
type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendOTP(_ context.Context, toEmail, fullName, otp string) error {
	subject := "Your Fundly verification code"
	body := fmt.Sprintf("Hi %s,\n\nYour Fundly verification code is %s.\nIt expires in 5 minutes.\n\nIf you did not request this, you can safely ignore this email.",
		fallbackName(fullName), otp)
	return m.send(toEmail, subject, body)
}

func (m *SMTPMailer) SendPasswordReset(_ context.Context, toEmail, code string) error {
	subject := "Your Fundly password reset code"
	body := fmt.Sprintf("Your Fundly password reset code is %s.\nIt expires in 15 minutes.\n\nIf you did not request this, you can safely ignore this email.", code)
	return m.send(toEmail, subject, body)
}

func (m *SMTPMailer) send(toEmail, subject, body string) error {
	fromHeader := m.cfg.From
	if m.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.From)
	}

	msg := buildMessage(fromHeader, toEmail, subject, body)
	addr := m.cfg.Host + ":" + m.cfg.Port
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{toEmail}, msg); err != nil {
		log.Printf("⚠️ failed to send email to %s: %v", toEmail, err)
		return fmt.Errorf("sending email: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func fallbackName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "there"
	}
	return name
}
