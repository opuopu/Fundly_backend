package transaction

import (
	"time"
)

type Type string

const (
	TypeDeposit  Type = "DEPOSIT"
	TypeWithdraw Type = "WITHDRAW"
	TypeTransfer Type = "TRANSFER"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

type Transaction struct {
	ID            string    `json:"id"`
	FromAccountID string    `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id,omitempty"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Type          Type      `json:"type"`
	Status        Status    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
