package attachment

import (
	"context"
	"io"
	"time"

	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
)

// MaxAttachmentSize caps receipt/proof uploads at 10MB per the API spec.
const MaxAttachmentSize = 10 * 1024 * 1024

// FileStorage persists an uploaded file and returns its served URL.
type FileStorage interface {
	Save(ctx context.Context, subDir, filename string, reader io.Reader) (string, error)
}

type AttachmentResponse struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
}

func toResponse(a *domainAttachment.Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:          a.ID,
		FileName:    a.FileName,
		ContentType: a.ContentType,
		FileSize:    a.FileSize,
		URL:         a.FileURL,
		CreatedAt:   a.CreatedAt,
	}
}
