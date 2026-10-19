package attachment

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrInvalidFile        = errors.New("invalid file payload")
	ErrFileTooLarge       = errors.New("file exceeds maximum allowed size")
	ErrUnsupportedType    = errors.New("unsupported file type: only images and pdf are allowed")
)

type Attachment struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	TransactionID string    `json:"transaction_id,omitempty"`
	FileName      string    `json:"file_name"`
	FileURL       string    `json:"url"`
	ContentType   string    `json:"content_type"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at"`
}

func New(id, userID, fileName, fileURL, contentType string, fileSize int64) *Attachment {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = "attachment"
	}
	return &Attachment{
		ID:          id,
		UserID:      userID,
		FileName:    fileName,
		FileURL:     fileURL,
		ContentType: contentType,
		FileSize:    fileSize,
		CreatedAt:   time.Now().UTC(),
	}
}
