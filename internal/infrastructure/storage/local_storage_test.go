package storage_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dodul01/fundly/internal/infrastructure/storage"
)

func TestLocalFileStorage_SaveAvatar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fundly_test_uploads_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.NewLocalFileStorage(tempDir, "/uploads")
	content := []byte("fake image data")
	reader := bytes.NewReader(content)

	urlPath, err := store.SaveAvatar(context.Background(), "usr_1", "test_avatar.png", "image/png", reader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedURL := "/uploads/avatars/test_avatar.png"
	if urlPath != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, urlPath)
	}

	savedFile := filepath.Join(tempDir, "avatars", "test_avatar.png")
	data, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("saved file content does not match")
	}
}
