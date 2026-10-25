package email

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
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
	plainBody := fmt.Sprintf("Hi %s,\n\nYour Fundly verification code is %s.\nIt expires in 5 minutes.\n\nIf you did not request this, you can safely ignore this email.",
		fallbackName(fullName), otp)
	htmlBody := renderCodeEmailHTML(codeEmailOptions{
		BadgeText:      "ACCOUNT VERIFICATION",
		Heading:        "Verify Your Account",
		Subheading:     "Register to continue your journey with us.",
		IntroParagraph: "Thank you for choosing Fundly. To complete your account registration and secure your profile, please enter the 6-digit verification code below:",
		Code:           otp,
		ExpiryMinutes:  5,
		Disclaimer:     "Didn't request this code? You can safely ignore this email. Your Fundly account will not be activated without verification.",
	})
	return m.send(toEmail, subject, plainBody, htmlBody)
}

func (m *SMTPMailer) SendPasswordReset(_ context.Context, toEmail, code string) error {
	subject := "Your Fundly password reset code"
	plainBody := fmt.Sprintf("Your Fundly password reset code is %s.\nIt expires in 15 minutes.\n\nIf you did not request this, you can safely ignore this email.", code)
	htmlBody := renderCodeEmailHTML(codeEmailOptions{
		BadgeText:      "PASSWORD RESET",
		Heading:        "Reset Your Password",
		Subheading:     "Confirm it's you to continue.",
		IntroParagraph: "We received a request to reset the password for your Fundly account. Please enter the 6-digit code below to continue:",
		Code:           code,
		ExpiryMinutes:  15,
		Disclaimer:     "Didn't request this code? You can safely ignore this email. Your password will not be changed without verification.",
	})
	return m.send(toEmail, subject, plainBody, htmlBody)
}

func (m *SMTPMailer) send(toEmail, subject, plainBody, htmlBody string) error {
	fromHeader := m.cfg.From
	if m.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.From)
	}

	msg, err := buildMessage(fromHeader, toEmail, subject, plainBody, htmlBody)
	if err != nil {
		return fmt.Errorf("building email: %w", err)
	}
	addr := m.cfg.Host + ":" + m.cfg.Port
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{toEmail}, msg); err != nil {
		log.Printf("⚠️ failed to send email to %s: %v", toEmail, err)
		return fmt.Errorf("sending email: %w", err)
	}
	return nil
}

// buildMessage assembles a multipart/alternative message with a plain-text part
// (for clients/spam filters that want it) and the branded HTML part.
func buildMessage(from, to, subject, plainBody, htmlBody string) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	headers := strings.Builder{}
	fmt.Fprintf(&headers, "From: %s\r\n", from)
	fmt.Fprintf(&headers, "To: %s\r\n", to)
	fmt.Fprintf(&headers, "Subject: %s\r\n", subject)
	headers.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&headers, "Content-Type: multipart/alternative; boundary=%s\r\n", w.Boundary())
	headers.WriteString("\r\n")

	plainPart, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=UTF-8"}})
	if err != nil {
		return nil, err
	}
	if _, err := plainPart.Write([]byte(plainBody)); err != nil {
		return nil, err
	}

	htmlPart, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=UTF-8"}})
	if err != nil {
		return nil, err
	}
	if _, err := htmlPart.Write([]byte(htmlBody)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return append([]byte(headers.String()), buf.Bytes()...), nil
}

func fallbackName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "there"
	}
	return name
}
