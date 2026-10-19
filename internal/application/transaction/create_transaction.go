package transaction

import (
	"context"
	"strings"
	"time"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
	"github.com/google/uuid"
)

type CreateTransactionInput struct {
	UserID        string   `json:"-"`
	Type          string   `json:"type"`
	Date          string   `json:"date"` // YYYY-MM-DD (optional, defaults to now)
	Amount        int64    `json:"amount"`
	Currency      string   `json:"currency"`
	CategoryID    string   `json:"category_id"`
	AccountID     string   `json:"account_id"`
	Note          string   `json:"note"`
	Description   string   `json:"description"`
	AttachmentIDs []string `json:"attachment_ids"`
}

type CreateTransactionUseCase struct {
	txRepo      domainTx.Repository
	accounts    domainAccount.Repository
	cats        domainCategory.Repository
	attachments AttachmentProvider
	atomic      Atomic
}

func NewCreateTransactionUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, cats domainCategory.Repository, attachments AttachmentProvider, atomic Atomic) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{txRepo: txRepo, accounts: accounts, cats: cats, attachments: attachments, atomic: atomic}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*TransactionResponse, error) {
	txID := "tx_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

	date := parseDate(input.Date)

	entry, err := domainTx.NewEntry(
		txID,
		input.UserID,
		input.Type,
		date,
		input.Amount,
		input.Currency,
		input.CategoryID,
		input.AccountID,
		input.Note,
		input.Description,
	)
	if err != nil {
		return nil, err
	}

	// The balance update and transaction insert (and optional attachment linking) must be
	// all-or-nothing, so they run inside a single database transaction.
	err = runAtomic(ctx, uc.atomic, func(ctx context.Context) error {
		// Load and verify ownership of the affected account.
		acc, err := uc.accounts.GetByID(ctx, entry.AccountID())
		if err != nil {
			return err
		}
		if acc.UserID != input.UserID {
			return domainAccount.ErrAccountNotFound
		}

		// Apply the balance effect: income credits, expense debits.
		if entry.Type == domainTx.TypeIncome {
			if err := acc.Deposit(entry.Amount); err != nil {
				return err
			}
		} else {
			if err := acc.Withdraw(entry.Amount); err != nil {
				return err
			}
		}

		if err := uc.accounts.Update(ctx, acc); err != nil {
			return err
		}

		if err := uc.txRepo.Create(ctx, entry); err != nil {
			return err
		}

		// Associate any pre-uploaded attachments with the new transaction.
		if uc.attachments != nil && len(input.AttachmentIDs) > 0 {
			if err := uc.attachments.LinkToTransaction(ctx, input.UserID, entry.ID, input.AttachmentIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := enricher{categories: uc.cats, accounts: uc.accounts, attachments: uc.attachments}.toDetailResponse(ctx, entry)
	return &res, nil
}

// parseDate accepts YYYY-MM-DD; anything unparseable yields the zero time so the domain
// defaults it to now.
func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}
