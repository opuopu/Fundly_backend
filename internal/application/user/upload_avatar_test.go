package user_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	appuser "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type mockAvatarStorage struct {
	saveURL string
	err     error
}

func (m *mockAvatarStorage) SaveAvatar(ctx context.Context, userID, filename, contentType string, reader io.Reader) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.saveURL != "" {
		return m.saveURL, nil
	}
	return "/uploads/avatars/" + filename, nil
}

func TestUploadAvatarUseCase_Success_PNG(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "pwd_hash")
	repo.usersByID[u.ID] = u

	storage := &mockAvatarStorage{}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	// Valid PNG magic bytes
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UploadAvatarInput{
		Filename:    "avatar.png",
		ContentType: "image/png",
		Size:        int64(len(pngHeader)),
		Reader:      bytes.NewReader(pngHeader),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.AvatarURL == "" {
		t.Errorf("expected non-empty avatar URL")
	}

	updatedUser, _ := repo.GetByID(context.Background(), "usr_1")
	if updatedUser.AvatarURL != res.AvatarURL {
		t.Errorf("user entity avatar url not updated, got %s", updatedUser.AvatarURL)
	}
}

func TestUploadAvatarUseCase_Success_JPEG(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "pwd_hash")
	repo.usersByID[u.ID] = u

	storage := &mockAvatarStorage{}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	// Valid JPEG magic bytes
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	res, err := uc.Execute(context.Background(), "usr_1", appuser.UploadAvatarInput{
		Filename:    "photo.jpeg",
		ContentType: "image/jpeg",
		Size:        int64(len(jpegHeader)),
		Reader:      bytes.NewReader(jpegHeader),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.AvatarURL == "" {
		t.Errorf("expected non-empty avatar URL")
	}
}

func TestUploadAvatarUseCase_ImageTooLarge(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domainUser.User)}
	storage := &mockAvatarStorage{}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	_, err := uc.Execute(context.Background(), "usr_1", appuser.UploadAvatarInput{
		Filename:    "big.png",
		ContentType: "image/png",
		Size:        6 * 1024 * 1024, // 6MB
		Reader:      bytes.NewReader([]byte("test")),
	})
	if !errors.Is(err, domainUser.ErrImageTooLarge) {
		t.Fatalf("expected ErrImageTooLarge, got %v", err)
	}
}

func TestUploadAvatarUseCase_InvalidFormat(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "pwd_hash")
	repo.usersByID[u.ID] = u

	storage := &mockAvatarStorage{}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	txtData := []byte("plain text file contents, not an image")
	_, err := uc.Execute(context.Background(), "usr_1", appuser.UploadAvatarInput{
		Filename:    "file.txt",
		ContentType: "text/plain",
		Size:        int64(len(txtData)),
		Reader:      bytes.NewReader(txtData),
	})
	if !errors.Is(err, domainUser.ErrInvalidImageFormat) {
		t.Fatalf("expected ErrInvalidImageFormat, got %v", err)
	}
}

func TestUploadAvatarUseCase_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{usersByID: make(map[string]*domainUser.User)}
	storage := &mockAvatarStorage{}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	_, err := uc.Execute(context.Background(), "usr_unknown", appuser.UploadAvatarInput{
		Filename:    "avatar.png",
		ContentType: "image/png",
		Size:        int64(len(pngHeader)),
		Reader:      bytes.NewReader(pngHeader),
	})
	if !errors.Is(err, domainUser.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUploadAvatarUseCase_StorageError(t *testing.T) {
	repo := &mockUserRepo{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "alice@example.com", "Alice", "pwd_hash")
	repo.usersByID[u.ID] = u

	storageErr := errors.New("disk full")
	storage := &mockAvatarStorage{err: storageErr}
	uc := appuser.NewUploadAvatarUseCase(repo, storage)

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	_, err := uc.Execute(context.Background(), "usr_1", appuser.UploadAvatarInput{
		Filename:    "avatar.png",
		ContentType: "image/png",
		Size:        int64(len(pngHeader)),
		Reader:      bytes.NewReader(pngHeader),
	})
	if err == nil {
		t.Fatalf("expected error from storage failure, got nil")
	}
}
