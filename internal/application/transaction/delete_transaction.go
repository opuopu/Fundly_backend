package transaction

import (
	"context"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type DeleteTransactionUseCase struct {
	txRepo   domainTx.Repository
	accounts domainAccount.Repository
	atomic   Atomic
}

func NewDeleteTransactionUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, atomic Atomic) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{txRepo: txRepo, accounts: accounts, atomic: atomic}
}

func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, userID, id string) error {
	return runAtomic(ctx, uc.atomic, func(ctx context.Context) error {
		tx, err := uc.txRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if tx.UserID != userID {
			return domainTx.ErrTransactionNotFound
		}

		// Reverse the balance effect before removing the entry. A missing account is
		// tolerated (it may have been deleted), so deletion still succeeds.
		if acc, accErr := uc.accounts.GetByID(ctx, tx.AccountID()); accErr == nil && acc != nil && acc.UserID == userID {
			if tx.Type == domainTx.TypeIncome {
				_ = acc.Withdraw(tx.Amount)
			} else {
				_ = acc.Deposit(tx.Amount)
			}
			if err := uc.accounts.Update(ctx, acc); err != nil {
				return err
			}
		}

		return uc.txRepo.Delete(ctx, id)
	})
}
