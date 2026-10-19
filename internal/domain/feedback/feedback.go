package feedback

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidSubject     = errors.New("subject cannot be empty")
	ErrInvalidDescription = errors.New("description cannot be empty")
)

type Feedback struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id,omitempty"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func New(id, userID, subject, description string) (*Feedback, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return nil, ErrInvalidSubject
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrInvalidDescription
	}

	return &Feedback{
		ID:          id,
		UserID:      userID,
		Subject:     subject,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
