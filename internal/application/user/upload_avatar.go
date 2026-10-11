package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

const MaxAvatarSize = 5 * 1024 * 1024 // 5MB

type AvatarStorage interface {
	SaveAvatar(ctx context.Context, userID, filename, contentType string, reader io.Reader) (string, error)
}

type UploadAvatarInput struct {
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type UploadAvatarResult struct {
	AvatarURL string `json:"avatar_url"`
}

type UploadAvatarUseCase struct {
	userRepo domainUser.Repository
	storage  AvatarStorage
}

func NewUploadAvatarUseCase(userRepo domainUser.Repository, storage AvatarStorage) *UploadAvatarUseCase {
	return &UploadAvatarUseCase{
		userRepo: userRepo,
		storage:  storage,
	}
}

func (uc *UploadAvatarUseCase) Execute(ctx context.Context, userID string, input UploadAvatarInput) (*UploadAvatarResult, error) {
	if userID == "" {
		return nil, domainUser.ErrInvalidUserData
	}
	if input.Size <= 0 || input.Reader == nil {
		return nil, domainUser.ErrInvalidUserData
	}
	if input.Size > MaxAvatarSize {
		return nil, domainUser.ErrImageTooLarge
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, domainUser.ErrUserNotFound
	}

	// Sniff magic bytes from the first 512 bytes
	header := make([]byte, 512)
	n, _ := io.ReadFull(input.Reader, header)
	if n == 0 {
		return nil, domainUser.ErrInvalidUserData
	}
	headerBuf := header[:n]
	detectedType := http.DetectContentType(headerBuf)

	var ext string
	var mimeType string

	switch {
	case detectedType == "image/jpeg" || strings.EqualFold(input.ContentType, "image/jpeg"):
		ext = ".jpg"
		mimeType = "image/jpeg"
	case detectedType == "image/png" || strings.EqualFold(input.ContentType, "image/png"):
		ext = ".png"
		mimeType = "image/png"
	case detectedType == "image/webp" || strings.EqualFold(input.ContentType, "image/webp"):
		ext = ".webp"
		mimeType = "image/webp"
	default:
		return nil, domainUser.ErrInvalidImageFormat
	}

	// Reassemble reader
	fullReader := io.MultiReader(bytes.NewReader(headerBuf), input.Reader)

	// Generate safe, collision-free filename
	filename := fmt.Sprintf("avatar_%s_%d%s", userID, time.Now().UnixNano(), ext)

	avatarURL, err := uc.storage.SaveAvatar(ctx, userID, filename, mimeType, fullReader)
	if err != nil {
		return nil, fmt.Errorf("saving avatar to storage: %w", err)
	}

	u.SetAvatar(avatarURL)
	if err := uc.userRepo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("updating user avatar in repository: %w", err)
	}

	return &UploadAvatarResult{
		AvatarURL: avatarURL,
	}, nil
}
