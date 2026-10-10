package user

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists with this email")
	ErrInvalidUserData   = errors.New("invalid user data")
)
