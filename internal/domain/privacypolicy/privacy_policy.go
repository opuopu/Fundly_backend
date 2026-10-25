package privacypolicy

import (
	"errors"
	"time"
)

var (
	ErrPolicyNotFound = errors.New("privacy policy not found")
	ErrVersionMissing = errors.New("version is required")
)

type PrivacyPolicy struct {
	ID        string    `json:"-"`
	Version   string    `json:"version"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Effective time.Time `json:"effective_date"`
}

// Acknowledgment records a user's acceptance of a specific privacy policy version.
type Acknowledgment struct {
	UserID         string    `json:"-"`
	Version        string    `json:"version"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
}
