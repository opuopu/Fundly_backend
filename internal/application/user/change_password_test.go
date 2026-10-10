package user_test

import (
	"context"
	"errors"
	"testing"

	appuser "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type mockPasswordHasher struct {
	hashErr    error
	compareErr error
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	return "hashed_" + password, nil
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	if m.compareErr != nil {
		return m.compareErr
	}
	if hashedPassword != "hashed_"+password {
		return errors.New("mismatch")
	}
	return nil
}

func TestChangePasswordUseCase_Success(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "hashed_OldPassword123!")
	repo.usersByID[u.ID] = u

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedUser, _ := repo.GetByID(context.Background(), "usr_1")
	if updatedUser.PasswordHash != "hashed_NewPassword456!" {
		t.Errorf("expected updated password hash, got %s", updatedUser.PasswordHash)
	}
}

func TestChangePasswordUseCase_WrongCurrentPassword(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "hashed_RealPassword123!")
	repo.usersByID[u.ID] = u

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "WrongPassword!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	})
	if !errors.Is(err, domainUser.ErrInvalidCurrentPassword) {
		t.Fatalf("expected ErrInvalidCurrentPassword, got %v", err)
	}
}

func TestChangePasswordUseCase_SocialUserNoPassword(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	u, _ := domainUser.NewSocialUser("usr_1", "social@example.com", "Social", "")
	repo.usersByID[u.ID] = u

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "AnyPassword123!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	})
	if !errors.Is(err, domainUser.ErrInvalidCurrentPassword) {
		t.Fatalf("expected ErrInvalidCurrentPassword for user without password, got %v", err)
	}
}

func TestChangePasswordUseCase_PasswordMismatch(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "DifferentPassword789!",
	})
	if !errors.Is(err, domainUser.ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestChangePasswordUseCase_PasswordTooWeak(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "short",
		ConfirmPassword: "short",
	})
	if !errors.Is(err, domainUser.ErrPasswordTooWeak) {
		t.Fatalf("expected ErrPasswordTooWeak, got %v", err)
	}
}

func TestChangePasswordUseCase_EmptyInput(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_1", appuser.ChangePasswordInput{
		CurrentPassword: "",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	})
	if !errors.Is(err, domainUser.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestChangePasswordUseCase_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	hasher := &mockPasswordHasher{}

	uc := appuser.NewChangePasswordUseCase(repo, hasher)
	err := uc.Execute(context.Background(), "usr_nonexistent", appuser.ChangePasswordInput{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword456!",
		ConfirmPassword: "NewPassword456!",
	})
	if !errors.Is(err, domainUser.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
