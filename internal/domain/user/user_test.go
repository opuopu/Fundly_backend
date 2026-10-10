package user_test

import (
	"testing"

	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	u, err := user.New("usr-1", "alice@example.com", "Alice Smith")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", u.Email)
	}
	if u.FullName != "Alice Smith" {
		t.Errorf("expected full name Alice Smith, got %s", u.FullName)
	}

	_, err = user.New("usr-2", "", "Bob")
	if err != user.ErrInvalidUserData {
		t.Errorf("expected ErrInvalidUserData, got %v", err)
	}
}
