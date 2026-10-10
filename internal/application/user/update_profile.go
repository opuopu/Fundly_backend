package user

import (
	"context"
	"net/mail"
	"strings"

	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type UpdateProfileInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type UpdateProfileUseCase struct {
	userRepo domainUser.Repository
}

func NewUpdateProfileUseCase(userRepo domainUser.Repository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, userID string, input UpdateProfileInput) (*ProfileResponse, error) {
	if userID == "" {
		return nil, domainUser.ErrInvalidUserData
	}

	fullName := strings.TrimSpace(input.FullName)
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if fullName == "" && email == "" {
		return nil, domainUser.ErrInvalidUserData
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domainUser.ErrUserNotFound
	}

	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, domainUser.ErrInvalidEmail
		}

		if email != strings.ToLower(u.Email) {
			existing, err := uc.userRepo.FindByEmail(ctx, email)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != u.ID {
				return nil, domainUser.ErrEmailAlreadyExists
			}
		}
	}

	u.UpdateInfo(fullName, email)

	if err := uc.userRepo.Update(ctx, u); err != nil {
		return nil, err
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
