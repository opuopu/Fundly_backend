package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
	"github.com/google/uuid"
)

type UploadAttachmentInput struct {
	UserID      string
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type UploadAttachmentUseCase struct {
	repo    domainAttachment.Repository
	storage FileStorage
}

func NewUploadAttachmentUseCase(repo domainAttachment.Repository, storage FileStorage) *UploadAttachmentUseCase {
	return &UploadAttachmentUseCase{repo: repo, storage: storage}
}

func (uc *UploadAttachmentUseCase) Execute(ctx context.Context, input UploadAttachmentInput) (*AttachmentResponse, error) {
	if input.UserID == "" || input.Reader == nil || input.Size <= 0 {
		return nil, domainAttachment.ErrInvalidFile
	}
	if input.Size > MaxAttachmentSize {
		return nil, domainAttachment.ErrFileTooLarge
	}

	// Sniff the first 512 bytes to validate the real content type.
	header := make([]byte, 512)
	n, _ := io.ReadFull(input.Reader, header)
	if n == 0 {
		return nil, domainAttachment.ErrInvalidFile
	}
	headerBuf := header[:n]
	detected := http.DetectContentType(headerBuf)

	ext, mimeType, ok := resolveType(detected)
	if !ok {
		return nil, domainAttachment.ErrUnsupportedType
	}

	attachmentID := "att_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
	filename := fmt.Sprintf("%s_%d%s", attachmentID, time.Now().UnixNano(), ext)

	fullReader := io.MultiReader(bytes.NewReader(headerBuf), input.Reader)
	url, err := uc.storage.Save(ctx, "attachments", filename, fullReader)
	if err != nil {
		return nil, fmt.Errorf("saving attachment to storage: %w", err)
	}

	original := strings.TrimSpace(input.Filename)
	if original == "" {
		original = filename
	}

	a := domainAttachment.New(attachmentID, input.UserID, original, url, mimeType, input.Size)
	if err := uc.repo.Create(ctx, a); err != nil {
		return nil, err
	}

	res := toResponse(a)
	return &res, nil
}

// resolveType maps the server-sniffed content type to an extension and canonical MIME type.
// The decision is based solely on the sniffed bytes — the client-declared Content-Type is
// never trusted, so a mislabeled upload cannot bypass the allow-list.
func resolveType(detected string) (ext, mime string, ok bool) {
	switch {
	case detected == "image/jpeg":
		return ".jpg", "image/jpeg", true
	case detected == "image/png":
		return ".png", "image/png", true
	case detected == "image/webp":
		return ".webp", "image/webp", true
	case detected == "image/gif":
		return ".gif", "image/gif", true
	case detected == "application/pdf":
		return ".pdf", "application/pdf", true
	default:
		return "", "", false
	}
}
