package terms

import (
	"errors"
	"time"
)

var (
	ErrTermsNotFound  = errors.New("terms and conditions not found")
	ErrVersionMissing = errors.New("version is required")
)

type TermsAndConditions struct {
	ID        string    `json:"-"`
	Version   string    `json:"version"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Effective time.Time `json:"effective_date"`
}

// Acknowledgment records a user's acceptance of a specific terms version.
type Acknowledgment struct {
	UserID         string    `json:"-"`
	Version        string    `json:"version"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
}
