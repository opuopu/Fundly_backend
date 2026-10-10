package user_test

import (
	"context"
	"errors"
	"testing"

	appuser "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

func TestUpdateProfileUseCase_Success_Both(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u, _ := domainUser.New("usr_1", "old@example.com", "Old Name", "pwd_hash")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	uc := appuser.NewUpdateProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		FullName: "Thom Haye",
		Email:    "new@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.FullName != "Thom Haye" {
		t.Errorf("expected full_name 'Thom Haye', got '%s'", res.FullName)
	}
	if res.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", res.Email)
	}
}

func TestUpdateProfileUseCase_Success_OnlyFullName(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u, _ := domainUser.New("usr_1", "keep@example.com", "Old Name", "pwd_hash")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	uc := appuser.NewUpdateProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		FullName: "Updated Name",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.FullName != "Updated Name" {
		t.Errorf("expected 'Updated Name', got '%s'", res.FullName)
	}
	if res.Email != "keep@example.com" {
		t.Errorf("expected unchanged email, got '%s'", res.Email)
	}
}

func TestUpdateProfileUseCase_Success_OnlyEmail(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u, _ := domainUser.New("usr_1", "old@example.com", "My Name", "pwd_hash")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	uc := appuser.NewUpdateProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		Email: "newemail@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Email != "newemail@example.com" {
		t.Errorf("expected 'newemail@example.com', got '%s'", res.Email)
	}
	if res.FullName != "My Name" {
		t.Errorf("expected unchanged name, got '%s'", res.FullName)
	}
}

func TestUpdateProfileUseCase_EmailConflict(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u1, _ := domainUser.New("usr_1", "user1@example.com", "User One", "pwd_hash")
	repo.usersByID[u1.ID] = u1
	repo.usersByEmail[u1.Email] = u1

	u2, _ := domainUser.New("usr_2", "taken@example.com", "User Two", "pwd_hash")
	repo.usersByID[u2.ID] = u2
	repo.usersByEmail[u2.Email] = u2

	uc := appuser.NewUpdateProfileUseCase(repo)
	_, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		Email: "taken@example.com",
	})
	if !errors.Is(err, domainUser.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestUpdateProfileUseCase_SameEmailAllowed(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u1, _ := domainUser.New("usr_1", "user1@example.com", "User One", "pwd_hash")
	repo.usersByID[u1.ID] = u1
	repo.usersByEmail[u1.Email] = u1

	uc := appuser.NewUpdateProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		FullName: "Brand New Name",
		Email:    "user1@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error when keeping same email, got %v", err)
	}
	if res.FullName != "Brand New Name" {
		t.Errorf("expected updated name, got '%s'", res.FullName)
	}
}

func TestUpdateProfileUseCase_InvalidEmail(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	u, _ := domainUser.New("usr_1", "valid@example.com", "User", "pwd_hash")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	uc := appuser.NewUpdateProfileUseCase(repo)
	_, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		Email: "invalid-email-string",
	})
	if !errors.Is(err, domainUser.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestUpdateProfileUseCase_EmptyFields(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	uc := appuser.NewUpdateProfileUseCase(repo)
	_, err := uc.Execute(context.Background(), "usr_1", appuser.UpdateProfileInput{
		FullName: "   ",
		Email:    "",
	})
	if !errors.Is(err, domainUser.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestUpdateProfileUseCase_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}

	uc := appuser.NewUpdateProfileUseCase(repo)
	_, err := uc.Execute(context.Background(), "usr_nonexistent", appuser.UpdateProfileInput{
		FullName: "Test",
	})
	if !errors.Is(err, domainUser.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
