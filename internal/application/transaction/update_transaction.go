package transaction

import (
	"context"
	"time"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type UpdateTransactionInput struct {
	UserID      string  `json:"-"`
	ID          string  `json:"-"`
	Amount      *int64  `json:"amount"`
	Date        *string `json:"date"`
	CategoryID  *string `json:"category_id"`
	AccountID   *string `json:"account_id"`
	Note        *string `json:"note"`
	Description *string `json:"description"`
}

type UpdateTransactionUseCase struct {
	txRepo      domainTx.Repository
	accounts    domainAccount.Repository
	cats        domainCategory.Repository
	attachments AttachmentProvider
	atomic      Atomic
}

func NewUpdateTransactionUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, cats domainCategory.Repository, attachments AttachmentProvider, atomic Atomic) *UpdateTransactionUseCase {
	return &UpdateTransactionUseCase{txRepo: txRepo, accounts: accounts, cats: cats, attachments: attachments, atomic: atomic}
}

func (uc *UpdateTransactionUseCase) Execute(ctx context.Context, input UpdateTransactionInput) (*TransactionResponse, error) {
	var tx *domainTx.Transaction

	// Load, mutate, reconcile balances, and persist within a single transaction so a
	// partial failure cannot leave balances inconsistent with the stored entry.
	err := runAtomic(ctx, uc.atomic, func(ctx context.Context) error {
		loaded, err := uc.txRepo.GetByID(ctx, input.ID)
		if err != nil {
			return err
		}
		if loaded.UserID != input.UserID {
			return domainTx.ErrTransactionNotFound
		}
		tx = loaded

		oldAccountID := tx.AccountID()
		oldAmount := tx.Amount
		txType := tx.Type

		var datePtr *time.Time
		if input.Date != nil {
			d := parseDate(*input.Date)
			datePtr = &d
		}

		if err := tx.Update(input.Amount, datePtr, input.CategoryID, input.AccountID, input.Note, input.Description); err != nil {
			return err
		}

		newAccountID := tx.AccountID()

		// Reconcile account balances: reverse the original effect, then apply the new one.
		// Accounts are cached so the same account is loaded and persisted only once.
		cache := map[string]*domainAccount.Account{}
		load := func(id string) (*domainAccount.Account, error) {
			if acc, ok := cache[id]; ok {
				return acc, nil
			}
			acc, err := uc.accounts.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			if acc.UserID != input.UserID {
				return nil, domainAccount.ErrAccountNotFound
			}
			cache[id] = acc
			return acc, nil
		}

		oldAcc, err := load(oldAccountID)
		if err != nil {
			return err
		}
		if txType == domainTx.TypeIncome {
			if err := oldAcc.Withdraw(oldAmount); err != nil {
				return err
			}
		} else {
			if err := oldAcc.Deposit(oldAmount); err != nil {
				return err
			}
		}

		newAcc, err := load(newAccountID)
		if err != nil {
			return err
		}
		if txType == domainTx.TypeIncome {
			if err := newAcc.Deposit(tx.Amount); err != nil {
				return err
			}
		} else {
			if err := newAcc.Withdraw(tx.Amount); err != nil {
				return err
			}
		}

		for _, acc := range cache {
			if err := uc.accounts.Update(ctx, acc); err != nil {
				return err
			}
		}

		return uc.txRepo.Update(ctx, tx)
	})
	if err != nil {
		return nil, err
	}

	res := enricher{categories: uc.cats, accounts: uc.accounts, attachments: uc.attachments}.toDetailResponse(ctx, tx)
	return &res, nil
}
