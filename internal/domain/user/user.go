package user

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New(id, email, fullName string) (*User, error) {
	if email == "" || fullName == "" {
		return nil, ErrInvalidUserData
	}
	now := time.Now().UTC()
	return &User{
		ID:        id,
		Email:     email,
		FullName:  fullName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
