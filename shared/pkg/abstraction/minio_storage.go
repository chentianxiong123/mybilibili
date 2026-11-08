package abstraction

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorageService struct {
	client        *minio.Client
	publicClient  *minio.Client
	bucketName    string
	publicEndpoint string
}

type MinioConfig struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	BucketName     string
	Region         string
	// UseSSL 强制 HTTPS（FINDING-015：端点裸化后 Secure 无法由 scheme 前缀判定，
	// 需独立开关）。
	UseSSL bool
}

// normalizeEndpoint 将 minio-go 要求的主机端点规范化：剥掉 http(s):// scheme
// 与路径，返回裸 host:port 与 Secure 标志。minio-go v7 拒绝带 scheme 的端点
// （"Endpoint url cannot have fully qualified paths"），scheme 须由 Secure 表达。
func normalizeEndpoint(endpoint string) (bare string, secure bool) {
	bare = strings.TrimSpace(endpoint)
	switch {
	case strings.HasPrefix(bare, "https://"):
		secure = true
		bare = strings.TrimPrefix(bare, "https://")
	case strings.HasPrefix(bare, "http://"):
		bare = strings.TrimPrefix(bare, "http://")
	}
	if i := strings.IndexByte(bare, '/'); i >= 0 {
		bare = bare[:i]
	}
	return bare, secure
}

func DefaultMinioConfig() MinioConfig {
	return MinioConfig{
		Endpoint:       "127.0.0.1:9000",
		PublicEndpoint: "127.0.0.1:9000",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		BucketName:     "mybilibili",
		Region:         "us-east-1",
	}
}

func NewMinioStorageService(cfg MinioConfig) (*MinioStorageService, error) {
	endpoint, secure := normalizeEndpoint(cfg.Endpoint)
	// FINDING-015 修复：UseSSL 强制 HTTPS（端点裸化后 scheme 前缀恒 false）。
	if cfg.UseSSL {
		secure = true
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create client: %w", err)
	}

	publicEndpoint := cfg.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}
	publicBare, publicSecure := normalizeEndpoint(publicEndpoint)

	publicClient, err := minio.New(publicBare, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: publicSecure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create public client: %w", err)
	}

	svc := &MinioStorageService{
		client:         client,
		publicClient:   publicClient,
		bucketName:     cfg.BucketName,
		publicEndpoint: publicBare,
	}

	exists, err := client.BucketExists(context.Background(), cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("minio: check bucket: %w", err)
	}
	if !exists {
		err = client.MakeBucket(context.Background(), cfg.BucketName, minio.MakeBucketOptions{Region: cfg.Region})
		if err != nil {
			return nil, fmt.Errorf("minio: create bucket: %w", err)
		}
		log.Printf("minio: created bucket %s", cfg.BucketName)
	}

	return svc, nil
}

func (s *MinioStorageService) Put(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, key, body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinioStorageService) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *MinioStorageService) Delete(ctx context.Context, bucket, key string) error {
	return s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s *MinioStorageService) Head(ctx context.Context, bucket, key string) (*FileInfo, error) {
	obj, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Key:          obj.Key,
		Size:         obj.Size,
		ContentType:  obj.ContentType,
		ETag:         obj.ETag,
		LastModified: obj.LastModified,
	}, nil
}

func (s *MinioStorageService) List(ctx context.Context, bucket, prefix string) ([]FileInfo, error) {
	var result []FileInfo
	for obj := range s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		result = append(result, FileInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			ETag:         obj.ETag,
			LastModified: obj.LastModified,
		})
	}
	return result, nil
}

func (s *MinioStorageService) SignedURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	u, err := s.publicClient.PresignedGetObject(ctx, bucket, key, expire, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}