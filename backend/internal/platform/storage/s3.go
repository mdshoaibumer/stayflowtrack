package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stayflow/stayflow-track/internal/config"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store(cfg config.StorageConfig) (*S3Store, error) {
	opts := s3.Options{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}

	if cfg.Endpoint != "" {
		opts.BaseEndpoint = &cfg.Endpoint
	}

	if cfg.UsePathStyle {
		opts.UsePathStyle = true
	}

	client := s3.New(opts)

	return &S3Store{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *S3Store) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (*FileMetadata, error) {
	input := &s3.PutObjectInput{
		Bucket:        &s.bucket,
		Key:           &key,
		Body:          reader,
		ContentType:   &contentType,
		ContentLength: &size,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("upload to s3: %w", err)
	}

	return &FileMetadata{
		Key:         key,
		Bucket:      s.bucket,
		ContentType: contentType,
		Size:        size,
		UploadedAt:  time.Now(),
	}, nil
}

func (s *S3Store) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("download from s3: %w", err)
	}

	return output.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("delete from s3: %w", err)
	}

	return nil
}

func (s *S3Store) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	presigned, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}

	return presigned.URL, nil
}

// Compile-time check
var _ Store = (*S3Store)(nil)
