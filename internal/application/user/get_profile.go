package user

import (
	"context"
	"time"

	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type ProfileResponse struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	AvatarURL       string    `json:"avatar_url"`
	DefaultCurrency string    `json:"default_currency"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetProfileUseCase struct {
	userRepo domainUser.Repository
}

func NewGetProfileUseCase(userRepo domainUser.Repository) *GetProfileUseCase {
	return &GetProfileUseCase{userRepo: userRepo}
}

func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string) (*ProfileResponse, error) {
	if userID == "" {
		return nil, domainUser.ErrInvalidUserData
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domainUser.ErrUserNotFound
	}

	currency := u.DefaultCurrency
	if currency == "" {
		currency = "USD"
	}

	return &ProfileResponse{
		ID:              u.ID,
		Email:           u.Email,
		FullName:        u.FullName,
		AvatarURL:       u.AvatarURL,
		DefaultCurrency: currency,
		CreatedAt:       u.CreatedAt,
	}, nil
}
