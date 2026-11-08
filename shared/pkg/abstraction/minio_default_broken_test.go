package abstraction

import (
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMinioStorageService_DefaultConfigRejected 验证 FINDING-074 修复：
// DefaultMinioConfig() 的端点此前携带 scheme（http://...），minio-go v7 拒绝
// 带 scheme 的端点（"Endpoint url cannot have fully qualified paths"）。
// 修复后 normalizeEndpoint 剥离 scheme（Secure 表达），默认配置应可构造。
func TestNewMinioStorageService_DefaultConfigRejected(t *testing.T) {
	cfg := DefaultMinioConfig()
	// 修复后默认端点应为裸 host:port（无 scheme）。
	assert.NotContains(t, cfg.Endpoint, "://", "default endpoint should be bare host:port")

	// normalizeEndpoint：剥 http(s):// scheme + 路径，返回 bare host:port + Secure。
	bare, secure := normalizeEndpoint("http://127.0.0.1:9000")
	assert.Equal(t, "127.0.0.1:9000", bare)
	assert.False(t, secure)
	bare2, secure2 := normalizeEndpoint("https://minio.example.com:9000/path")
	assert.Equal(t, "minio.example.com:9000", bare2)
	assert.True(t, secure2)

	// 构造逻辑：带 scheme 的端点不再触发 "fully qualified paths" 拒绝。
	client, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	require.NoError(t, err, "bare host:port must be accepted by minio.New (FINDING-074 fixed)")
	require.NotNil(t, client)
}
