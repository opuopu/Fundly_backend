package transaction

import (
	"context"
	"time"

	"github.com/Dodul01/fundly/internal/application/transfer"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type TransferInput struct {
	UserID        string `json:"-"`
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Note          string `json:"note"`
}

// TransferResponse describes a completed transfer, exposing both the source and destination
// accounts (a plain transaction response would surface only one side).
type TransferResponse struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Amount      int64       `json:"amount"`
	Currency    string      `json:"currency"`
	FromAccount *AccountRef `json:"from_account"`
	ToAccount   *AccountRef `json:"to_account"`
	Note        string      `json:"note,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// TransferUseCase is a thin adapter over the existing cross-account transfer service so the
// transactions handler can expose POST /transactions/transfer without reimplementing the
// debit/credit logic.
type TransferUseCase struct {
	svc      *transfer.Service
	accounts domainAccount.Repository
	cats     domainCategory.Repository
}

func NewTransferUseCase(svc *transfer.Service, accounts domainAccount.Repository, cats domainCategory.Repository) *TransferUseCase {
	return &TransferUseCase{svc: svc, accounts: accounts, cats: cats}
}

func (uc *TransferUseCase) Execute(ctx context.Context, input TransferInput) (*TransferResponse, error) {
	tx, err := uc.svc.Execute(ctx, transfer.Request{
		UserID:        input.UserID,
		FromAccountID: input.FromAccountID,
		ToAccountID:   input.ToAccountID,
		Amount:        input.Amount,
	})
	if err != nil {
		return nil, err
	}

	if input.Note != "" {
		tx.Note = input.Note
	}

	return &TransferResponse{
		ID:          tx.ID,
		Type:        string(tx.Type),
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		FromAccount: uc.accountRef(ctx, tx.FromAccountID),
		ToAccount:   uc.accountRef(ctx, tx.ToAccountID),
		Note:        tx.Note,
		CreatedAt:   tx.CreatedAt,
	}, nil
}

// accountRef resolves an account summary, returning just the ID if the lookup fails.
func (uc *TransferUseCase) accountRef(ctx context.Context, id string) *AccountRef {
	if id == "" {
		return nil
	}
	ref := &AccountRef{ID: id}
	if uc.accounts != nil {
		if acc, err := uc.accounts.GetByID(ctx, id); err == nil && acc != nil {
			ref.Name = acc.Name
			ref.Group = acc.Group
		}
	}
	return ref
}
