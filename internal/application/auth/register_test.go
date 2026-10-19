package auth_test

import (
	"context"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type mockUserRepo struct {
	users map[string]*user.User
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Save(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

type mockHasher struct{}

func (m *mockHasher) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

type mockOTPGen struct{}

func (m *mockOTPGen) Generate() string {
	return "128333"
}

func TestRegisterUseCaseSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	hasher := &mockHasher{}
	otpGen := &mockOTPGen{}

	uc := auth.NewRegisterUseCase(repo, hasher, otpGen, nil)

	cmd := auth.RegisterCommand{
		Email:           "newuser@example.com",
		FullName:        "Thom Haye",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	result, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error executing Register: %v", err)
	}

	if result.Email != "newuser@example.com" {
		t.Errorf("expected email newuser@example.com, got %s", result.Email)
	}
	if !result.OTPSent {
		t.Errorf("expected OTPSent to be true")
	}
	if result.OTPExpiresInSeconds != 300 {
		t.Errorf("expected OTP expiry 300, got %d", result.OTPExpiresInSeconds)
	}

	// Verify user persisted
	savedUser, _ := repo.FindByEmail(context.Background(), "newuser@example.com")
	if savedUser == nil {
		t.Fatalf("expected saved user to exist in repository")
	}
	if savedUser.PasswordHash != "hashed_Password123!" {
		t.Errorf("expected password to be hashed, got %s", savedUser.PasswordHash)
	}
	if savedUser.OTP != "128333" {
		t.Errorf("expected OTP 128333, got %s", savedUser.OTP)
	}
}

func TestRegisterUseCaseMismatchedPassword(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	uc := auth.NewRegisterUseCase(repo, &mockHasher{}, &mockOTPGen{}, nil)

	cmd := auth.RegisterCommand{
		Email:           "user@example.com",
		FullName:        "User",
		Password:        "Password123!",
		ConfirmPassword: "DifferentPassword123!",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if err == nil {
		t.Fatalf("expected error for mismatched passwords, got nil")
	}
}
