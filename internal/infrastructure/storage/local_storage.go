package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	baseDir string
	baseURL string
}

func NewLocalFileStorage(baseDir, baseURL string) *LocalFileStorage {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}
	return &LocalFileStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

func (s *LocalFileStorage) SaveAvatar(ctx context.Context, userID, filename, contentType string, reader io.Reader) (string, error) {
	return s.Save(ctx, "avatars", filename, reader)
}

func (s *LocalFileStorage) Save(ctx context.Context, subDir, filename string, reader io.Reader) (string, error) {
	targetDir := filepath.Join(s.baseDir, subDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("creating directory %s: %w", targetDir, err)
	}

	targetPath := filepath.Join(targetDir, filename)
	outFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("creating file %s: %w", targetPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return "", fmt.Errorf("writing file %s: %w", targetPath, err)
	}

	urlPath := fmt.Sprintf("%s/%s/%s", s.baseURL, subDir, filename)
	return urlPath, nil
}
