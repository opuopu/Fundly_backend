package email

import (
	"context"
	"log"
)

// LogMailer is the development fallback used when SMTP credentials are not configured. It
// prints the code to the server log instead of delivering a real email, so local flows work
// without a mail account.
type LogMailer struct{}

func NewLogMailer() *LogMailer { return &LogMailer{} }

func (LogMailer) SendOTP(_ context.Context, toEmail, _ string, otp string) error {
	log.Printf("📧 [dev mailer] OTP for %s: %s (configure SMTP_* to send real emails)", toEmail, otp)
	return nil
}

func (LogMailer) SendPasswordReset(_ context.Context, toEmail, code string) error {
	log.Printf("📧 [dev mailer] password reset code for %s: %s (configure SMTP_* to send real emails)", toEmail, code)
	return nil
}
