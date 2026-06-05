package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/stayflow/stayflow-track/internal/config"
)

// FileMetadata contains metadata about an uploaded file.
type FileMetadata struct {
	Key         string
	Bucket      string
	ContentType string
	Size        int64
	UploadedAt  time.Time
}

// Store defines the interface for file storage operations.
type Store interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (*FileMetadata, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// New creates a new storage implementation based on config.
func New(cfg config.StorageConfig) (Store, error) {
	switch cfg.Provider {
	case "s3":
		return NewS3Store(cfg)
	case "local":
		return NewLocalStore(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
