package abstraction

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceDiscovery(t *testing.T) {
	d, err := NewServiceDiscovery(ServiceDiscoveryConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, d)

	_, err = NewServiceDiscovery(ServiceDiscoveryConfig{Type: "file"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	_, err = NewServiceDiscovery(ServiceDiscoveryConfig{Type: "etcd"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "etcd discovery not implemented")

	_, err = NewServiceDiscovery(ServiceDiscoveryConfig{Type: "unknown"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown service discovery type")
}

func TestNewMessageQueue(t *testing.T) {
	q, err := NewMessageQueue(MessageQueueConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, q)

	q, err = NewMessageQueue(MessageQueueConfig{Type: "file", Path: t.TempDir()})
	require.NoError(t, err)
	assert.NotNil(t, q)

	_, err = NewMessageQueue(MessageQueueConfig{Type: "redis-stream"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis stream queue not implemented")

	// nats 真实连接测试要求本地有 nats-server, 此处仅断言类型被分发
	_, err = NewMessageQueue(MessageQueueConfig{Type: "nats", NATSURL: "nats://127.0.0.1:1"})
	// 应为连接错误或超时错误, 不为 nil
	if err == nil {
		t.Skip("nats to bad addr unexpectedly succeeded (server may be up)")
	}

	_, err = NewMessageQueue(MessageQueueConfig{Type: "bogus"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message queue type")
}

func TestNewCacheStore(t *testing.T) {
	c, err := NewCacheStore(CacheStoreConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, c)
	_ = c.Close()

	// redis: 用 miniredis 验证成功连接
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	c, err = NewCacheStore(CacheStoreConfig{Type: "redis", Addr: mr.Addr()})
	require.NoError(t, err)
	assert.NotNil(t, c)
	_ = c.Close()

	_, err = NewCacheStore(CacheStoreConfig{Type: "sqlite"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sqlite cache not implemented")

	_, err = NewCacheStore(CacheStoreConfig{Type: "weird"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown cache store type")
}

func TestNewServiceCaller(t *testing.T) {
	c, err := NewServiceCaller(ServiceCallerConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, c)

	c, err = NewServiceCaller(ServiceCallerConfig{Type: "ollama"})
	require.NoError(t, err)
	assert.NotNil(t, c)

	_, err = NewServiceCaller(ServiceCallerConfig{Type: "grpc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grpc caller not implemented")

	_, err = NewServiceCaller(ServiceCallerConfig{Type: "http"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http caller not implemented")

	_, err = NewServiceCaller(ServiceCallerConfig{Type: "nope"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown service caller type")
}

func TestNewStorageService(t *testing.T) {
	s, err := NewStorageService(StorageServiceConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, s)

	_, err = NewStorageService(StorageServiceConfig{Type: "minio"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	_, err = NewStorageService(StorageServiceConfig{Type: "s3"})
	assert.Error(t, err)

	_, err = NewStorageService(StorageServiceConfig{Type: "local"})
	assert.Error(t, err)

	_, err = NewStorageService(StorageServiceConfig{Type: "?"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown storage service type")
}

func TestNewSearchEngine(t *testing.T) {
	s, err := NewSearchEngine(SearchEngineConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, s)

	_, err = NewSearchEngine(SearchEngineConfig{Type: "bleve"})
	assert.Error(t, err)

	_, err = NewSearchEngine(SearchEngineConfig{Type: "elasticsearch"})
	assert.Error(t, err)

	_, err = NewSearchEngine(SearchEngineConfig{Type: "bad"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown search engine type")
}

func TestNewDocumentStore(t *testing.T) {
	s, err := NewDocumentStore(DocumentStoreConfig{Type: "memory"})
	require.NoError(t, err)
	assert.NotNil(t, s)

	_, err = NewDocumentStore(DocumentStoreConfig{Type: "pg-jsonb"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DSN required")

	_, err = NewDocumentStore(DocumentStoreConfig{Type: "sqlite"})
	assert.Error(t, err)

	_, err = NewDocumentStore(DocumentStoreConfig{Type: "mongodb"})
	assert.Error(t, err)

	_, err = NewDocumentStore(DocumentStoreConfig{Type: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document store type")
}

// 防止 time 包未被引用 (timeout 配置用到)
var _ = time.Second
