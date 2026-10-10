package user

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidUserData     = errors.New("invalid user data")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrPasswordTooWeak     = errors.New("password does not meet security requirements")
	ErrOTPExpired          = errors.New("otp has expired")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrAccountNotVerified        = errors.New("account is not verified, please verify OTP")
	ErrInvalidSocialToken        = errors.New("invalid or expired social authentication token")
	ErrUnsupportedSocialProvider = errors.New("unsupported social authentication provider")
	ErrInvalidRefreshToken       = errors.New("invalid or expired refresh token")
	ErrPasswordMismatch          = errors.New("passwords do not match")
	ErrInvalidCurrentPassword    = errors.New("current password is incorrect")
)