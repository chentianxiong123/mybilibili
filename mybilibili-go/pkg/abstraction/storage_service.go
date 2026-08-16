package abstraction

import (
	"context"
	"io"
	"time"
)

type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

type StorageService interface {
	Put(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	Head(ctx context.Context, bucket, key string) (*FileInfo, error)
	List(ctx context.Context, bucket, prefix string) ([]FileInfo, error)
	SignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error)
}
