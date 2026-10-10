package user_test

import (
	"context"
	"errors"
	"testing"

	appuser "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type mockUserRepo struct {
	usersByID    map[string]*domainUser.User
	usersByEmail map[string]*domainUser.User
	err          error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domainUser.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domainUser.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.usersByEmail[email], nil
}

func (m *mockUserRepo) Save(ctx context.Context, user *domainUser.User) error {
	if m.err != nil {
		return m.err
	}
	m.usersByID[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *domainUser.User) error {
	return m.Save(ctx, user)
}

func TestGetProfileUseCase_Success(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}

	u, err := domainUser.New("usr_test_1", "test@fundly.com", "John Doe", "hashed_pwd")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	u.AvatarURL = "https://assets.fundly.com/avatars/usr_1.png"
	u.DefaultCurrency = "EUR"
	repo.usersByID[u.ID] = u

	uc := appuser.NewGetProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_test_1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.ID != "usr_test_1" {
		t.Errorf("expected ID 'usr_test_1', got '%s'", res.ID)
	}
	if res.Email != "test@fundly.com" {
		t.Errorf("expected email 'test@fundly.com', got '%s'", res.Email)
	}
	if res.FullName != "John Doe" {
		t.Errorf("expected full name 'John Doe', got '%s'", res.FullName)
	}
	if res.AvatarURL != "https://assets.fundly.com/avatars/usr_1.png" {
		t.Errorf("expected avatar url, got '%s'", res.AvatarURL)
	}
	if res.DefaultCurrency != "EUR" {
		t.Errorf("expected default currency 'EUR', got '%s'", res.DefaultCurrency)
	}
	if res.CreatedAt.IsZero() {
		t.Errorf("expected valid CreatedAt time, got zero")
	}
}

func TestGetProfileUseCase_DefaultCurrencyFallback(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}

	u, _ := domainUser.New("usr_test_2", "no_curr@fundly.com", "Jane Doe", "hashed_pwd")
	u.DefaultCurrency = ""
	repo.usersByID[u.ID] = u

	uc := appuser.NewGetProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_test_2")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.DefaultCurrency != "USD" {
		t.Errorf("expected fallback default currency 'USD', got '%s'", res.DefaultCurrency)
	}
}

func TestGetProfileUseCase_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}

	uc := appuser.NewGetProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "non_existent_user")
	if !errors.Is(err, domainUser.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestGetProfileUseCase_EmptyUserID(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}

	uc := appuser.NewGetProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "")
	if !errors.Is(err, domainUser.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}

func TestGetProfileUseCase_DatabaseError(t *testing.T) {
	dbErr := errors.New("db connection failure")
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
		err:       dbErr,
	}

	uc := appuser.NewGetProfileUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_test_1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected %v, got %v", dbErr, err)
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
}
