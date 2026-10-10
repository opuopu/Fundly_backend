package stat

import "time"

type AccountStats struct {
	AccountID        string    `json:"account_id"`
	TotalInflow      int64     `json:"total_inflow"`
	TotalOutflow     int64     `json:"total_outflow"`
	TransactionCount int       `json:"transaction_count"`
	CalculatedAt     time.Time `json:"calculated_at"`
}
