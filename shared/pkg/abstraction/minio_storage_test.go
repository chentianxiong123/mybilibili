package abstraction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MinioStorageService 需要真 minio 服务器或重写 client 字段.
// 这里测能脱离网络的纯路径:
// - DefaultMinioConfig 默认值
// - NewMinioStorageService 连不上的错误路径 (端点 127.0.0.1:1)
// - 各种 nil/scheme 解析

func TestDefaultMinioConfig(t *testing.T) {
	c := DefaultMinioConfig()
	assert.Equal(t, "http://127.0.0.1:9000", c.Endpoint)
	assert.Equal(t, "http://127.0.0.1:9000", c.PublicEndpoint)
	assert.Equal(t, "minioadmin", c.AccessKey)
	assert.Equal(t, "minioadmin", c.SecretKey)
	assert.Equal(t, "mybilibili", c.BucketName)
	assert.Equal(t, "us-east-1", c.Region)
}

func TestNewMinioStorageService_EndpointUnreachable(t *testing.T) {
	cfg := MinioConfig{
		Endpoint:       "http://127.0.0.1:1",
		PublicEndpoint: "http://127.0.0.1:1",
		AccessKey:      "x",
		SecretKey:      "x",
		BucketName:     "x",
	}
	// minio.New 是延迟连接, BucketExists 才真连; 期望错误
	_, err := NewMinioStorageService(cfg)
	assert.Error(t, err)
}

func TestNewMinioStorageService_EndpointParseFail(t *testing.T) {
	cfg := MinioConfig{
		Endpoint:       "://bad-url",
		PublicEndpoint: "://bad-url",
	}
	// minio.New 解析 URL 失败应报错
	_, err := NewMinioStorageService(cfg)
	assert.Error(t, err)
}

func TestMinioStorageService_FileInfo_Fields(t *testing.T) {
	// FileInfo 是纯结构体, 测试字段写入.
	now := time.Now()
	fi := FileInfo{
		Key:          "a/b.txt",
		Size:         100,
		ContentType:  "text/plain",
		ETag:         "abc",
		LastModified: now,
	}
	assert.Equal(t, "a/b.txt", fi.Key)
	assert.Equal(t, int64(100), fi.Size)
	assert.Equal(t, now, fi.LastModified)
}