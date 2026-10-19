package auth_test

import (
	"context"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type spyMailer struct {
	otpTo, otp       string
	resetTo, resetCd string
	otpCalls         int
	resetCalls       int
	err              error
}

func (m *spyMailer) SendOTP(_ context.Context, email, _ string, otp string) error {
	m.otpCalls++
	m.otpTo, m.otp = email, otp
	return m.err
}

func (m *spyMailer) SendPasswordReset(_ context.Context, email, code string) error {
	m.resetCalls++
	m.resetTo, m.resetCd = email, code
	return m.err
}

func TestRegister_SendsOTPEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	mailer := &spyMailer{}
	uc := auth.NewRegisterUseCase(repo, &mockHasher{}, &mockOTPGen{}, mailer)

	res, err := uc.Execute(context.Background(), auth.RegisterCommand{
		Email:           "new@example.com",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mailer.otpCalls != 1 || mailer.otpTo != "new@example.com" || mailer.otp != "128333" {
		t.Errorf("expected OTP email to new@example.com with 128333, got calls=%d to=%s otp=%s", mailer.otpCalls, mailer.otpTo, mailer.otp)
	}
	if !res.OTPSent {
		t.Error("expected OTPSent true when mail succeeds")
	}
}

func TestRegister_MailFailureDoesNotFailRegistration(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	mailer := &spyMailer{err: context.DeadlineExceeded}
	uc := auth.NewRegisterUseCase(repo, &mockHasher{}, &mockOTPGen{}, mailer)

	res, err := uc.Execute(context.Background(), auth.RegisterCommand{
		Email:           "new@example.com",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	if err != nil {
		t.Fatalf("registration should still succeed when mail fails: %v", err)
	}
	if res.OTPSent {
		t.Error("expected OTPSent false when mail delivery fails")
	}
}

func TestForgotPassword_SendsResetEmail(t *testing.T) {
	u, _ := user.New("usr_1", "known@example.com", "Known", "hash")
	repo := &mockUserRepo{users: map[string]*user.User{"usr_1": u}}
	mailer := &spyMailer{}
	uc := auth.NewForgotPasswordUseCase(repo, &mockOTPGen{}, mailer)

	if err := uc.Execute(context.Background(), auth.ForgotPasswordCommand{Email: "known@example.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mailer.resetCalls != 1 || mailer.resetTo != "known@example.com" || mailer.resetCd != "128333" {
		t.Errorf("expected reset email to known@example.com with 128333, got calls=%d to=%s code=%s", mailer.resetCalls, mailer.resetTo, mailer.resetCd)
	}
}

func TestForgotPassword_UnknownEmailSendsNothing(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	mailer := &spyMailer{}
	uc := auth.NewForgotPasswordUseCase(repo, &mockOTPGen{}, mailer)

	// Silent success (anti-enumeration) and no email dispatched.
	if err := uc.Execute(context.Background(), auth.ForgotPasswordCommand{Email: "ghost@example.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mailer.resetCalls != 0 {
		t.Errorf("expected no email for unknown address, got %d calls", mailer.resetCalls)
	}
}
