package account

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	"github.com/google/uuid"
)

type CreateAccountInput struct {
	UserID         string `json:"-"`
	Group          string `json:"group"`
	Name           string `json:"name"`
	InitialBalance int64  `json:"initial_balance"`
	Currency       string `json:"currency"`
	Description    string `json:"description"`
}

type CreateAccountUseCase struct {
	repo domainAccount.Repository
}

func NewCreateAccountUseCase(repo domainAccount.Repository) *CreateAccountUseCase {
	return &CreateAccountUseCase{repo: repo}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, input CreateAccountInput) (*AccountResponse, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, domainAccount.ErrAccountNotFound
	}

	accountID := "acc_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

	acc, err := domainAccount.NewAccount(
		accountID,
		input.UserID,
		input.Name,
		input.Group,
		input.Currency,
		input.InitialBalance,
		input.Description,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, acc); err != nil {
		return nil, err
	}

	res := ToAccountResponse(acc)
	return &res, nil
}
