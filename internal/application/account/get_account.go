package account

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type GetAccountUseCase struct {
	repo domainAccount.Repository
}

func NewGetAccountUseCase(repo domainAccount.Repository) *GetAccountUseCase {
	return &GetAccountUseCase{repo: repo}
}

func (uc *GetAccountUseCase) Execute(ctx context.Context, userID, accountID string) (*AccountResponse, error) {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)

	if userID == "" || accountID == "" {
		return nil, domainAccount.ErrAccountNotFound
	}

	acc, err := uc.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Verify that the account belongs to the requesting user
	if acc.UserID != userID {
		return nil, domainAccount.ErrAccountNotFound
	}

	res := ToAccountResponse(acc)
	return &res, nil
}
