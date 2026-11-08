package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 只有当 CI/本机提供真实 Redis 时（AUTH_TEST_REDIS_ADDR）才运行：
// 覆盖 redisRevocation / redisTokenStore 的 Redis 实现与
// SetupStoresFromEnv 的成功分支。
func TestRedisRevocationStore_Integration(t *testing.T) {
	addr := os.Getenv("AUTH_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("设置 AUTH_TEST_REDIS_ADDR 启用 Redis 集成测试")
	}

	prevR, prevT := CurrentRevocationStore(), CurrentTokenStore()
	t.Cleanup(func() { SetRevocationStore(prevR); SetTokenStore(prevT) })

	t.Setenv("REDIS_ADDR", addr)
	require.NoError(t, SetupStoresFromEnv(), "Redis 可达时必须成功注册 Redis store")

	ctx := context.Background()
	jti := "it-test-" + time.Now().Format("20060102150405.000000")
	key := "live-" + jti

	// 吊销名单
	rev := CurrentRevocationStore()
	require.NotNil(t, rev)
	require.NoError(t, rev.Revoke(ctx, jti, time.Minute))
	ok, err := rev.IsRevoked(ctx, jti)
	require.NoError(t, err)
	assert.True(t, ok, "刚吊销的 jti 必须命中名单")

	// 空 jti / 非正 TTL：no-op，不得误删 Redis 数据
	require.NoError(t, rev.Revoke(ctx, "", time.Minute))
	assert.NoError(t, rev.Revoke(ctx, jti, 0))
	assert.NoError(t, rev.Revoke(ctx, jti, -time.Second))

	// 一次性令牌 store
	ts := CurrentTokenStore()
	require.NotNil(t, ts)
	ok1, err := ts.Consume(ctx, key, time.Minute)
	require.NoError(t, err)
	assert.True(t, ok1, "首次消费成功")
	ok2, err := ts.Consume(ctx, key, time.Minute)
	require.NoError(t, err)
	assert.False(t, ok2, "第二次必须判为重放")
	// 空 key → 假
	ok3, err := ts.Consume(ctx, "", time.Minute)
	require.NoError(t, err)
	assert.False(t, ok3)
	// ttl<=0 → 兜底 1min，仍然成功
	ok4, err := ts.Consume(ctx, key+"-b", 0)
	require.NoError(t, err)
	assert.True(t, ok4)

	// 未吊销的 jti
	absent, err := rev.IsRevoked(ctx, "never-revoked-"+jti)
	require.NoError(t, err)
	assert.False(t, absent)
}

// 错误路径注入：把 store 指向一个必断的 Redis 连接，验证降级语义。
func TestRedisStore_ErrorPaths(t *testing.T) {
	// 指向一个已关闭的端口，确保连接失败
	prevR, prevT := CurrentRevocationStore(), CurrentTokenStore()
	t.Cleanup(func() { SetRevocationStore(prevR); SetTokenStore(prevT) })

	t.Setenv("REDIS_ADDR", "127.0.0.1:1")
	err := SetupStoresFromEnv()
	require.Error(t, err, "连不上 Redis 必须把错误回传")
	// 降级为进程内实现后，功能照常
	require.NotNil(t, CurrentRevocationStore())
	require.NotNil(t, CurrentTokenStore())
}
