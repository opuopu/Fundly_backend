package account

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type UpdateAccountInput struct {
	UserID      string  `json:"-"`
	AccountID   string  `json:"-"`
	Name        *string `json:"name"`
	Group       *string `json:"group"`
	Description *string `json:"description"`
}

type UpdateAccountUseCase struct {
	repo domainAccount.Repository
}

func NewUpdateAccountUseCase(repo domainAccount.Repository) *UpdateAccountUseCase {
	return &UpdateAccountUseCase{repo: repo}
}

func (uc *UpdateAccountUseCase) Execute(ctx context.Context, input UpdateAccountInput) (*AccountResponse, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.AccountID = strings.TrimSpace(input.AccountID)

	if input.UserID == "" || input.AccountID == "" {
		return nil, domainAccount.ErrAccountNotFound
	}

	if input.Name == nil && input.Group == nil && input.Description == nil {
		return nil, domainAccount.ErrNoFieldsToUpdate
	}

	acc, err := uc.repo.GetByID(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}

	// Verify account belongs to requesting user
	if acc.UserID != input.UserID {
		return nil, domainAccount.ErrAccountNotFound
	}

	if err := acc.Update(input.Name, input.Group, input.Description); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, acc); err != nil {
		return nil, err
	}

	res := ToAccountResponse(acc)
	return &res, nil
}
