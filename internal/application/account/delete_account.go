package account

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type DeleteAccountUseCase struct {
	repo domainAccount.Repository
}

func NewDeleteAccountUseCase(repo domainAccount.Repository) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{repo: repo}
}

func (uc *DeleteAccountUseCase) Execute(ctx context.Context, userID, accountID string) error {
	userID = strings.TrimSpace(userID)
	accountID = strings.TrimSpace(accountID)

	if userID == "" || accountID == "" {
		return domainAccount.ErrAccountNotFound
	}

	acc, err := uc.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}

	// Verify account belongs to requesting user
	if acc.UserID != userID {
		return domainAccount.ErrAccountNotFound
	}

	return uc.repo.Delete(ctx, accountID)
}
