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
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

type Service struct {
	accounts     account.Repository
	transactions transaction.Repository
}

func NewService(accRepo account.Repository, txRepo transaction.Repository) *Service {
	return &Service{
		accounts:     accRepo,
		transactions: txRepo,
	}
}

func (s *Service) Execute(ctx context.Context, req Request) (*transaction.Transaction, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}

	fromAcc, err := s.accounts.GetByID(ctx, req.FromAccountID)
	if err != nil {
		return nil, fmt.Errorf("from account: %w", err)
	}

	toAcc, err := s.accounts.GetByID(ctx, req.ToAccountID)
	if err != nil {
		return nil, fmt.Errorf("to account: %w", err)
	}

	if fromAcc.Currency != toAcc.Currency {
		return nil, ErrMismatchedCurrency
	}

	if err := fromAcc.Withdraw(req.Amount); err != nil {
		return nil, err
	}

	if err := toAcc.Deposit(req.Amount); err != nil {
		return nil, err
	}

	if err := s.accounts.Update(ctx, fromAcc); err != nil {
		return nil, fmt.Errorf("updating sender account: %w", err)
	}

	if err := s.accounts.Update(ctx, toAcc); err != nil {
		return nil, fmt.Errorf("updating receiver account: %w", err)
	}

	tx := &transaction.Transaction{
		ID:            fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Currency:      fromAcc.Currency,
		Type:          transaction.TypeTransfer,
		Status:        transaction.StatusCompleted,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.transactions.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("recording transaction: %w", err)
	}

	return tx, nil
}
