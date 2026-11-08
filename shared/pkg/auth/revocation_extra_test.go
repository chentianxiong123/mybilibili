package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAdmin_AnonymousRejected(t *testing.T) {
	var reached bool
	admin := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/admin/rooms", nil)
	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)

	assert.False(t, reached, "无 X-Admin-Id 的请求不得进入后台 handler")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}

func TestRequireAdmin_WithAdminHeaderPasses(t *testing.T) {
	var reached bool
	var seenID string
	admin := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		seenID = r.Header.Get("X-Admin-Id")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/admin/rooms", nil)
	req.Header.Set("X-Admin-Id", "7")
	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "7", seenID)
}

// RequireAdmin 与 IdentityMiddleware 组合：普通用户即使被注入 X-User-Id
// 也不能提权进后台。
func TestRequireAdmin_UserIdentityNotAdmin(t *testing.T) {
	j := NewJWT(mwSecret)
	tok, err := j.Generate(4)
	require.NoError(t, err)
	_ = tok

	// RequireAdmin 只看 X-Admin-Id：普通用户凭证换来的 X-User-Id 不会通过
	admin := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/api/v1/admin/rooms", nil)
	req.Header.Set("X-User-Id", "4")
	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "X-User-Id 不能当 X-Admin-Id 用")
}

// ---- RevokeRefreshToken ----

func TestRevokeRefreshToken_MarksConsumed(t *testing.T) {
	j := NewJWT(mwSecret)
	tok := NewMemoryTokenStore()
	withTokenStore(t, tok)

	refresh, jti, err := j.GenerateRefreshWithJTI(4)
	require.NoError(t, err)
	require.NotEmpty(t, jti)

	// 消费前：首次可用
	assert.True(t, ConsumeRefreshOnce(context.Background(), j, refresh))

	RevokeRefreshToken(context.Background(), j, refresh)

	// 单独作废后：再用必须判重放
	assert.False(t, ConsumeRefreshOnce(context.Background(), j, refresh))
}

func TestRevokeRefreshToken_NilOrEmptyNoOp(t *testing.T) {
	j := NewJWT(mwSecret)
	// 这些路径只早退，不 panic 即通过
	RevokeRefreshToken(context.Background(), nil, "whatever")
	RevokeRefreshToken(context.Background(), j, "")
	RevokeRefreshToken(context.Background(), j, "not-a-jwt")
}

func TestRevokeRefreshToken_NoTokenStoreRegistered(t *testing.T) {
	withTokenStore(t, nil)
	j := NewJWT(mwSecret)
	refresh, err := j.GenerateRefresh(4)
	require.NoError(t, err)
	// 无 store 时静默 no-op
	RevokeRefreshToken(context.Background(), j, refresh)
	assert.True(t, ConsumeRefreshOnce(context.Background(), j, refresh),
		"未登记过的 store 不影响放行判断")
}

// ---- RevokeRequestSession 边界 ----

func TestRevokeRequestSession_NilRequestNoPanic(t *testing.T) {
	j := NewJWT(mwSecret)
	RevokeRequestSession(nil, j)
}

func TestRevokeRequestSession_NilJWTNoPanic(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/user/logout", nil)
	RevokeRequestSession(req, nil)
}

func TestRevokeRequestSession_NoStoresRegistered(t *testing.T) {
	withRevocation(t, nil)
	withTokenStore(t, nil)
	j := NewJWT(mwSecret)
	access, err := j.Generate(4)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/user/logout", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: access})
	RevokeRequestSession(req, j) // 两个 store 都 nil → 全 no-op，不 panic
}

// ---- remaining 兜底 ----

func TestRemaining_NilAndPastFallback(t *testing.T) {
	// 无 exp → 1h 兜底
	assert.Equal(t, time.Hour, remaining(nil))
	// 已过期的 exp → 1min 兜底
	past := jwt.NewNumericDate(time.Now().Add(-time.Hour))
	assert.Equal(t, time.Minute, remaining(past))
}

// ---- memRevocation.Revoke / memTokenStore.Consume 边界 ----

func TestMemRevocation_RevokeEdgeCases(t *testing.T) {
	m := &memRevocation{}
	// 空 jti 或 ttl<=0 → no-op
	require.NoError(t, m.Revoke(context.Background(), "", time.Hour))
	require.NoError(t, m.Revoke(context.Background(), "j", 0))
	require.NoError(t, m.Revoke(context.Background(), "j", -time.Second))

	// 有效写入 → IsRevoked 为真
	require.NoError(t, m.Revoke(context.Background(), "j", time.Minute))
	ok, err := m.IsRevoked(context.Background(), "j")
	require.NoError(t, err)
	assert.True(t, ok)

	// 未写入 → 假
	ok, err = m.IsRevoked(context.Background(), "other")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestMemRevocation_ExpiredEntryAutoCleanup(t *testing.T) {
	m := &memRevocation{}
	// 直接写一个已过期的条目（memRevocation 是包内类型，测试可直接构造）
	m.m = map[string]time.Time{"gone": time.Now().Add(-time.Second)}
	ok, err := m.IsRevoked(context.Background(), "gone")
	require.NoError(t, err)
	assert.False(t, ok, "过期条目必须自动清理并判为未吊销")
}

func TestMemTokenStore_ConsumeEdgeCases(t *testing.T) {
	m := &memTokenStore{}
	// 空 key → false
	got, err := m.Consume(context.Background(), "", time.Minute)
	require.NoError(t, err)
	assert.False(t, got, "空 key 不得标记成功")

	// 首次成功、二次重放
	ok1, err := m.Consume(context.Background(), "k", time.Minute)
	require.NoError(t, err)
	assert.True(t, ok1)
	ok2, err := m.Consume(context.Background(), "k", time.Minute)
	require.NoError(t, err)
	assert.False(t, ok2)

	// ttl<=0 时按 1min 兜底，不 panic
	m2 := &memTokenStore{}
	ok, err := m2.Consume(context.Background(), "k2", 0)
	require.NoError(t, err)
	assert.True(t, ok)
}

// ---- LogSetupWarning ----

func TestLogSetupWarning_PrintsOnNilAndNonNil(t *testing.T) {
	// 只验证不 panic；nil 与带错误两种输入都应通过
	LogSetupWarning(nil)
	LogSetupWarning(errors.New("redis localhost:6379 不可达，吊销名单降级为进程内"))
}

// redisRevocation.IsRevoked 的 Redis 错误分支：
// 注册一个 IsRevoked 必报错的 store（模拟 Redis 抖动），中间件必须放行（fail-open）。
type failingRevocation struct{}

func (failingRevocation) Revoke(context.Context, string, time.Duration) error { return nil }
func (failingRevocation) IsRevoked(context.Context, string) (bool, error) {
	return false, errors.New("redis: connection refused")
}

func TestIsRevoked_RedisErrorFailsOpen(t *testing.T) {
	j := NewJWT(mwSecret)
	withRevocation(t, failingRevocation{})

	tok, err := j.Generate(4)
	require.NoError(t, err)

	// 走中间件：Jti 非空 + IsRevoked 出错 → 放行且身份注入
	seen, reached := runMW(t, j, tok)
	require.True(t, reached)
	assert.Equal(t, "4", seen, "吊销查询失败必须放行，不能把在线用户踢下线")

	// 直接调用 isRevoked 也验证一次
	assert.False(t, isRevoked(context.Background(), &Claims{Jti: "whatever"}))
}

// ---- SetupStoresFromEnv 成功路径需要 Redis，跳过；
// 覆盖默认地址分支（未设 REDIS_ADDR）。 ----

func TestSetupStoresFromEnv_DefaultAddrWhenUnset(t *testing.T) {
	prevR, prevT := CurrentRevocationStore(), CurrentTokenStore()
	t.Cleanup(func() { SetRevocationStore(prevR); SetTokenStore(prevT) })

	// 确保未设 REDIS_ADDR；默认 127.0.0.1:6379，CI/本机大概率无 Redis → 走降级
	// 若本机有 Redis，则走成功分支；两种结果都合法，只要不 panic 即可
	err := SetupStoresFromEnv()
	if err != nil {
		assert.Contains(t, err.Error(), "不可达", "降级路径必须把原因回传给调用方")
		assert.NotNil(t, CurrentRevocationStore(), "降级也要注册进程内实现")
		assert.NotNil(t, CurrentTokenStore())
	}
}
