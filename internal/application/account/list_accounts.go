package account

import (
	"context"
	"strings"
	"time"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type AccountResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Balance     int64     `json:"balance"`
	Currency    string    `json:"currency"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListAccountsUseCase struct {
	repo domainAccount.Repository
}

func NewListAccountsUseCase(repo domainAccount.Repository) *ListAccountsUseCase {
	return &ListAccountsUseCase{repo: repo}
}

func ToAccountResponse(acc *domainAccount.Account) AccountResponse {
	return AccountResponse{
		ID:          acc.ID,
		Name:        acc.Name,
		Group:       acc.Group,
		Balance:     acc.Balance,
		Currency:    acc.Currency,
		Description: acc.Description,
		CreatedAt:   acc.CreatedAt,
	}
}

func (uc *ListAccountsUseCase) Execute(ctx context.Context, userID, groupFilter string) ([]AccountResponse, error) {
	if userID == "" {
		return nil, domainAccount.ErrAccountNotFound
	}

	groupFilter = strings.ToLower(strings.TrimSpace(groupFilter))

	var accounts []*domainAccount.Account
	var err error

	if groupFilter != "" {
		accounts, err = uc.repo.GetByUserIDAndGroup(ctx, userID, groupFilter)
	} else {
		accounts, err = uc.repo.GetByUserID(ctx, userID)
	}

	if err != nil {
		return nil, err
	}

	res := make([]AccountResponse, 0, len(accounts))
	for _, acc := range accounts {
		res = append(res, ToAccountResponse(acc))
	}

	return res, nil
}
