package user

import (
	"time"
)

type User struct {
	ID              string
	Email           string
	FullName        string
	PasswordHash    string
	AvatarURL       string
	DefaultCurrency string
	OTP             string
	OTPExpiresAt    time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// New creates a valid User entity.
func New(id, email, fullName, passwordHash string) (*User, error) {
	if id == "" || email == "" || fullName == "" || passwordHash == "" {
		return nil, ErrInvalidUserData
	}

	now := time.Now().UTC()
	return &User{
		ID:              id,
		Email:           email,
		FullName:        fullName,
		PasswordHash:    passwordHash,
		DefaultCurrency: "USD",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// NewSocialUser creates an email-verified User entity without password.
func NewSocialUser(id, email, fullName, avatarURL string) (*User, error) {
	if id == "" || email == "" {
		return nil, ErrInvalidUserData
	}
	if fullName == "" {
		fullName = email
	}

	now := time.Now().UTC()
	return &User{
		ID:              id,
		Email:           email,
		FullName:        fullName,
		AvatarURL:       avatarURL,
		DefaultCurrency: "USD",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (u *User) UpdateProfile(fullName, avatarURL string) {
	if fullName != "" {
		u.FullName = fullName
	}
	if avatarURL != "" {
		u.AvatarURL = avatarURL
	}
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) UpdateInfo(fullName, email string) {
	if fullName != "" {
		u.FullName = fullName
	}
	if email != "" {
		u.Email = email
	}
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) SetOTP(otp string, duration time.Duration) {
	u.OTP = otp
	u.OTPExpiresAt = time.Now().UTC().Add(duration)
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) VerifyOTP(otp string) error {
	if u.OTP == "" || time.Now().UTC().After(u.OTPExpiresAt) {
		return ErrOTPExpired
	}
	if u.OTP != otp {
		return ErrInvalidOTP
	}
	return nil
}

func (u *User) ClearOTP() {
	u.OTP = ""
	u.OTPExpiresAt = time.Time{}
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) IsVerified() bool {
	return u.OTP == ""
}

func (u *User) ChangePassword(newHash string) {
	u.PasswordHash = newHash
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) SetAvatar(url string) {
	u.AvatarURL = url
	u.UpdatedAt = time.Now().UTC()
}