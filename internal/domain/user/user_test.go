package user_test

import (
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	u, err := user.New("usr-1", "alice@example.com", "Alice Smith", "hashed_secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", u.Email)
	}
	if u.FullName != "Alice Smith" {
		t.Errorf("expected full name Alice Smith, got %s", u.FullName)
	}
	if u.PasswordHash != "hashed_secret" {
		t.Errorf("expected password hash hashed_secret, got %s", u.PasswordHash)
	}

	_, err = user.New("usr-2", "", "Bob", "")
	if err != user.ErrInvalidUserData {
		t.Errorf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestUserOTPWorkflow(t *testing.T) {
	u, err := user.New("usr-1", "alice@example.com", "Alice Smith", "hashed_secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u.SetOTP("128333", 5*time.Minute)

	// Valid OTP verification
	if err := u.VerifyOTP("128333"); err != nil {
		t.Fatalf("expected OTP verification to succeed, got %v", err)
	}

	// Invalid OTP code
	if err := u.VerifyOTP("999999"); err != user.ErrInvalidOTP {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}

	// Cleared OTP
	u.ClearOTP()
	if err := u.VerifyOTP("128333"); err != user.ErrOTPExpired {
		t.Fatalf("expected ErrOTPExpired after ClearOTP, got %v", err)
	}
}

func TestNewSocialUser(t *testing.T) {
	u, err := user.NewSocialUser("usr-1", "social@example.com", "Social User", "https://avatar.url")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !u.IsVerified() {
		t.Errorf("social user should be verified by default")
	}
	if u.AvatarURL != "https://avatar.url" {
		t.Errorf("expected avatar https://avatar.url, got %s", u.AvatarURL)
	}

	u.UpdateProfile("Updated User", "https://new-avatar.url")
	if u.FullName != "Updated User" || u.AvatarURL != "https://new-avatar.url" {
		t.Errorf("expected updated profile, got %s, %s", u.FullName, u.AvatarURL)
	}

	_, err = user.NewSocialUser("", "", "", "")
	if err != user.ErrInvalidUserData {
		t.Errorf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestUserUpdateInfo(t *testing.T) {
	u, err := user.New("usr-1", "alice@example.com", "Alice Smith", "hashed_secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u.UpdateInfo("Alice Johnson", "alice.johnson@example.com")
	if u.FullName != "Alice Johnson" {
		t.Errorf("expected Alice Johnson, got %s", u.FullName)
	}
	if u.Email != "alice.johnson@example.com" {
		t.Errorf("expected alice.johnson@example.com, got %s", u.Email)
	}
}
