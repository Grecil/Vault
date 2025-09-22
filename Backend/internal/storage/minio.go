package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"filevault-backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/tags"
)

type MinIOStorage struct {
	client   *minio.Client
	bucket   string
	useSSL   bool
	endpoint string
}

func NewMinIOStorage(cfg *config.Config) (*MinIOStorage, error) {
	// Initialize MinIO client
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:   client,
		bucket:   cfg.MinIOBucket,
		useSSL:   cfg.MinIOUseSSL,
		endpoint: cfg.MinIOEndpoint,
	}

	// Ensure bucket exists
	if err := storage.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return storage, nil
}

func (m *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns the object key
func (m *MinIOStorage) UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// GetFileURL generates a presigned URL for file download
func (m *MinIOStorage) GetFileURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetUploadURL generates a presigned URL for file upload
func (m *MinIOStorage) GetUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, m.bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return url.String(), nil
}

// DeleteFile deletes a file from MinIO
func (m *MinIOStorage) DeleteFile(ctx context.Context, objectKey string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileInfo returns information about a file
func (m *MinIOStorage) GetFileInfo(ctx context.Context, objectKey string) (*minio.ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, m.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &info, nil
}

// ListFiles lists files with a prefix
func (m *MinIOStorage) ListFiles(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// SetObjectTags sets tags on an object
func (m *MinIOStorage) SetObjectTags(ctx context.Context, objectKey string, tagMap map[string]string) error {
	// Convert map to MinIO tags format
	objectTags, err := tags.NewTags(tagMap, false)
	if err != nil {
		return fmt.Errorf("failed to create tags: %w", err)
	}

	err = m.client.PutObjectTagging(ctx, m.bucket, objectKey, objectTags, minio.PutObjectTaggingOptions{})
	if err != nil {
		return fmt.Errorf("failed to set object tags: %w", err)
	}
	return nil
}

// GetObjectTags gets tags from an object
func (m *MinIOStorage) GetObjectTags(ctx context.Context, objectKey string) (map[string]string, error) {
	tags, err := m.client.GetObjectTagging(ctx, m.bucket, objectKey, minio.GetObjectTaggingOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object tags: %w", err)
	}
	return tags.ToMap(), nil
}

// RemoveObjectTags removes all tags from an object
func (m *MinIOStorage) RemoveObjectTags(ctx context.Context, objectKey string) error {
	err := m.client.RemoveObjectTagging(ctx, m.bucket, objectKey, minio.RemoveObjectTaggingOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove object tags: %w", err)
	}
	return nil
}

// GetPublicFileURL generates a clean public URL for tagged objects
func (m *MinIOStorage) GetPublicFileURL(objectKey string) string {
	// Generate clean public URL without authentication
	scheme := "http"
	if m.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, m.endpoint, m.bucket, objectKey)
}
