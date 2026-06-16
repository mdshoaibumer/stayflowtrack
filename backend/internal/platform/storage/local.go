package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stayflow/stayflow-track/internal/config"
)

type LocalStore struct {
	basePath string
}

func NewLocalStore(cfg config.StorageConfig) (*LocalStore, error) {
	basePath := cfg.Endpoint
	if basePath == "" {
		basePath = "./uploads"
	}

	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("create local storage directory: %w", err)
	}

	return &LocalStore{basePath: basePath}, nil
}

func (l *LocalStore) safePath(key string) (string, error) {
	fullPath := filepath.Join(l.basePath, filepath.Clean(key))
	absBase, _ := filepath.Abs(l.basePath)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
		return "", fmt.Errorf("invalid file path")
	}
	return fullPath, nil
}

func (l *LocalStore) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (*FileMetadata, error) {
	fullPath, err := l.safePath(key)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	written, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &FileMetadata{
		Key:         key,
		Bucket:      "local",
		ContentType: contentType,
		Size:        written,
		UploadedAt:  time.Now(),
	}, nil
}

func (l *LocalStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath, err := l.safePath(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return file, nil
}

func (l *LocalStore) Delete(ctx context.Context, key string) error {
	fullPath, err := l.safePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (l *LocalStore) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return fmt.Sprintf("/files/%s", key), nil
}

var _ Store = (*LocalStore)(nil)
