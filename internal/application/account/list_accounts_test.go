package account_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type mockAccountRepo struct {
	accounts map[string]*domainAccount.Account
	err      error
}

func (m *mockAccountRepo) GetByID(ctx context.Context, id string) (*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	acc, ok := m.accounts[id]
	if !ok {
		return nil, domainAccount.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockAccountRepo) GetByUserID(ctx context.Context, userID string) ([]*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainAccount.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			res = append(res, acc)
		}
	}
	return res, nil
}

func (m *mockAccountRepo) GetByUserIDAndGroup(ctx context.Context, userID, group string) ([]*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainAccount.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID && strings.EqualFold(acc.Group, group) {
			res = append(res, acc)
		}
	}
	return res, nil
}

func (m *mockAccountRepo) Create(ctx context.Context, acc *domainAccount.Account) error {
	if m.err != nil {
		return m.err
	}
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepo) Update(ctx context.Context, acc *domainAccount.Account) error {
	if m.err != nil {
		return m.err
	}
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.accounts, id)
	return nil
}

func TestListAccountsUseCase_AllAccounts(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}

	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Daily wallet cash")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Debit Card", "card", "USD", 154050, "Checking card")
	accOther, _ := domainAccount.NewAccount("acc_03", "usr_2", "Other User Cash", "cash", "USD", 1000, "")

	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2
	repo.accounts[accOther.ID] = accOther

	uc := appaccount.NewListAccountsUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 accounts for usr_1, got %d", len(res))
	}
}

func TestListAccountsUseCase_WithGroupFilter(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}

	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Daily cash")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Debit Card", "card", "USD", 154050, "Debit")
	acc3, _ := domainAccount.NewAccount("acc_03", "usr_1", "Credit Card", "card", "USD", 50000, "Credit")

	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2
	repo.accounts[acc3.ID] = acc3

	uc := appaccount.NewListAccountsUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_1", "card")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 card accounts, got %d", len(res))
	}

	for _, a := range res {
		if a.Group != "card" {
			t.Errorf("expected group card, got %s", a.Group)
		}
	}
}

func TestListAccountsUseCase_EmptyResult(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}

	uc := appaccount.NewListAccountsUseCase(repo)
	res, err := uc.Execute(context.Background(), "usr_fresh", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res == nil {
		t.Errorf("expected non-nil empty slice")
	}
	if len(res) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(res))
	}
}

func TestListAccountsUseCase_EmptyUserID(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}

	uc := appaccount.NewListAccountsUseCase(repo)
	_, err := uc.Execute(context.Background(), "", "")
	if err == nil {
		t.Fatalf("expected error for empty userID, got nil")
	}
}

func TestListAccountsUseCase_RepositoryError(t *testing.T) {
	dbErr := errors.New("database connection failed")
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
		err:      dbErr,
	}

	uc := appaccount.NewListAccountsUseCase(repo)
	_, err := uc.Execute(context.Background(), "usr_1", "")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected %v, got %v", dbErr, err)
	}
}
