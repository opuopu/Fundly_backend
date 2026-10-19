package transfer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dodul01/fundly/internal/domain/account"
	"github.com/Dodul01/fundly/internal/domain/transaction"
)

var (
	ErrMismatchedCurrency = errors.New("currencies must match for transfer")
	ErrSameAccount        = errors.New("cannot transfer to the same account")
)

type Request struct {
	UserID        string `json:"-"`
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

// Atomic runs fn inside a single database transaction, rolling back on any error. May be
// nil (e.g. in unit tests), in which case fn runs directly.
type Atomic interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	accounts     account.Repository
	transactions transaction.Repository
	atomic       Atomic
}

func NewService(accRepo account.Repository, txRepo transaction.Repository, atomic Atomic) *Service {
	return &Service{
		accounts:     accRepo,
		transactions: txRepo,
		atomic:       atomic,
	}
}

func (s *Service) Execute(ctx context.Context, req Request) (*transaction.Transaction, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}

	tx := &transaction.Transaction{
		ID:            fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		UserID:        req.UserID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Type:          transaction.TypeTransfer,
		Status:        transaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),
	}

	// Debit, credit and the transaction record must commit atomically.
	run := func(ctx context.Context) error {
		fromAcc, err := s.accounts.GetByID(ctx, req.FromAccountID)
		if err != nil {
			return fmt.Errorf("from account: %w", err)
		}

		toAcc, err := s.accounts.GetByID(ctx, req.ToAccountID)
		if err != nil {
			return fmt.Errorf("to account: %w", err)
		}

		// Ownership guard: both accounts must belong to the requesting user. A foreign
		// account is reported as not-found, mirroring the create-transaction behaviour.
		if req.UserID != "" && (fromAcc.UserID != req.UserID || toAcc.UserID != req.UserID) {
			return account.ErrAccountNotFound
		}

		if fromAcc.Currency != toAcc.Currency {
			return ErrMismatchedCurrency
		}

		if err := fromAcc.Withdraw(req.Amount); err != nil {
			return err
		}
		if err := toAcc.Deposit(req.Amount); err != nil {
			return err
		}

		if err := s.accounts.Update(ctx, fromAcc); err != nil {
			return fmt.Errorf("updating sender account: %w", err)
		}
		if err := s.accounts.Update(ctx, toAcc); err != nil {
			return fmt.Errorf("updating receiver account: %w", err)
		}

		tx.Currency = fromAcc.Currency
		if err := s.transactions.Create(ctx, tx); err != nil {
			return fmt.Errorf("recording transaction: %w", err)
		}
		return nil
	}

	if s.atomic != nil {
		if err := s.atomic.Do(ctx, run); err != nil {
			return nil, err
		}
	} else if err := run(ctx); err != nil {
		return nil, err
	}

	return tx, nil
}
