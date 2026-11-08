package abstraction

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文件覆盖 minio_storage.go 不依赖真实服务器的路径:
// - NewMinioStorageService 的 publicEndpoint 回落 + BucketExists 失败
// - Put/Get/Delete/Head 的客户端校验错误与网络错误
// - List 在 ctx 取消时返回 Err 条目
// - SignedURL 纯离线签名(成功 + 过期时间非法)
//
// 说明: minio.New 不接受带 scheme 的端点(会报
// "Endpoint url cannot have fully qualified paths"),
// 因此下面统一使用 "127.0.0.1:1" 这种裸 host:port.

func newOfflineMinioSvc(t *testing.T) *MinioStorageService {
	t.Helper()
	creds := credentials.NewStaticV4("minioadmin", "minioadmin", "")
	c, err := minio.New("127.0.0.1:1", &minio.Options{
		Creds:  creds,
		Secure: false,
		Region: "us-east-1",
	})
	require.NoError(t, err)
	return &MinioStorageService{
		client:         c,
		publicClient:   c,
		bucketName:     "mybilibili",
		publicEndpoint: "127.0.0.1:1",
	}
}

func TestNewMinioStorageService_PublicEndpointFallbackThenBucketCheck(t *testing.T) {
	cfg := MinioConfig{
		Endpoint:   "127.0.0.1:1", // 裸端点保证 minio.New 成功
		AccessKey:  "minioadmin",
		SecretKey:  "minioadmin",
		BucketName: "mybilibili",
		Region:     "us-east-1",
	}
	// PublicEndpoint 为空 → 回落为 Endpoint; 随后 BucketExists 联网失败
	_, err := NewMinioStorageService(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check bucket")
}

// 非法 bucket 名触发客户端校验错误(不联网、立即返回)
func TestMinioStorage_MethodErrorPaths_InvalidBucket(t *testing.T) {
	svc := newOfflineMinioSvc(t)
	ctx := context.Background()

	require.Error(t, svc.Put(ctx, "b", "k", strings.NewReader("x"), "text/plain"))

	_, err := svc.Get(ctx, "b", "k")
	require.Error(t, err)

	require.Error(t, svc.Delete(ctx, "b", "k"))

	_, err = svc.Head(ctx, "b", "k")
	require.Error(t, err)
}

// 合法参数下的惰性行为与联网错误
func TestMinioStorage_MethodLazyAndNetworkError(t *testing.T) {
	svc := newOfflineMinioSvc(t)
	ctx := context.Background()

	// GetObject 是惰性的: 合法参数不联网, 直接返回对象
	rc, err := svc.Get(ctx, "mybilibili", "k")
	require.NoError(t, err)
	require.NotNil(t, rc)
	_ = rc.Close()

	// List: ctx 已取消 → ListObjects 产出带 Err 的条目 → 立即返回错误
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = svc.List(cctx, "mybilibili", "")
	require.Error(t, err)
}

func TestMinioStorage_SignedURL_Offline(t *testing.T) {
	svc := newOfflineMinioSvc(t)

	// 签名的 URL 无需网络(region 已配置)
	u, err := svc.SignedURL(context.Background(), "mybilibili", "k", time.Minute)
	require.NoError(t, err)
	assert.Contains(t, u, "X-Amz-Signature")

	// 过期时间超过 7 天 → 错误
	_, err = svc.SignedURL(context.Background(), "mybilibili", "k", 10*24*time.Hour)
	require.Error(t, err)
}

// ---- 可选 e2e: 设置 MINIO_TEST_ENDPOINT 且可达时用真实 minio 覆盖成功路径 ----

func TestMinioStorage_E2E_WhenServerAvailable(t *testing.T) {
	ep := os.Getenv("MINIO_TEST_ENDPOINT")
	if ep == "" {
		t.Skip("MINIO_TEST_ENDPOINT not set; skipping real minio e2e")
	}
	cfg := MinioConfig{
		Endpoint:       ep,
		PublicEndpoint: ep,
		AccessKey:      os.Getenv("MINIO_TEST_ACCESS_KEY"),
		SecretKey:      os.Getenv("MINIO_TEST_SECRET_KEY"),
		BucketName:     "test-bucket-" + strings.ReplaceAll(ep, ":", "-"),
		Region:         "us-east-1",
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = "minioadmin"
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = "minioadmin"
	}
	svc, err := NewMinioStorageService(cfg)
	if err != nil {
		t.Skipf("minio at %s not reachable: %v", ep, err)
	}
	t.Cleanup(func() {
		_ = svc.Delete(context.Background(), cfg.BucketName, "k1")
		_ = svc.Delete(context.Background(), cfg.BucketName, "sub/k2")
	})

	ctx := context.Background()
	require.NoError(t, svc.Put(ctx, cfg.BucketName, "k1", strings.NewReader("hello minio"), "text/plain"))
	require.NoError(t, svc.Put(ctx, cfg.BucketName, "sub/k2", strings.NewReader("nested"), "text/plain"))

	rc, err := svc.Get(ctx, cfg.BucketName, "k1")
	require.NoError(t, err)
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "hello minio", string(data))
	require.NoError(t, rc.Close())

	fi, err := svc.Head(ctx, cfg.BucketName, "k1")
	require.NoError(t, err)
	require.Equal(t, "k1", fi.Key)
	require.Equal(t, int64(len("hello minio")), fi.Size)

	list, err := svc.List(ctx, cfg.BucketName, "sub/")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "sub/k2", list[0].Key)

	u, err := svc.SignedURL(ctx, cfg.BucketName, "k1", time.Minute)
	require.NoError(t, err)
	require.Contains(t, u, "X-Amz-Signature")

	require.NoError(t, svc.Delete(ctx, cfg.BucketName, "k1"))
	_, err = svc.Head(ctx, cfg.BucketName, "k1")
	require.Error(t, err)
}
// PublicEndpoint 非法但 Endpoint 合法: 第二个 client 创建失败
func TestNewMinioStorageService_PublicClientCreateError(t *testing.T) {
	cfg := MinioConfig{
		Endpoint:       "127.0.0.1:1", // 裸端点, 第一个 client 可创建
		PublicEndpoint: "://bad-endpoint",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		BucketName:     "mybilibili",
		Region:         "us-east-1",
	}
	_, err := NewMinioStorageService(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "public client")
}
