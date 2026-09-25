package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试会动全局注册，逐个用例恢复原状
func withRevocation(t *testing.T, s RevocationStore) {
	t.Helper()
	prev := CurrentRevocationStore()
	SetRevocationStore(s)
	t.Cleanup(func() { SetRevocationStore(prev) })
}

func withTokenStore(t *testing.T, s TokenStore) {
	t.Helper()
	prev := CurrentTokenStore()
	SetTokenStore(s)
	t.Cleanup(func() { SetTokenStore(prev) })
}

type errRevocation struct{}

func (errRevocation) Revoke(context.Context, string, time.Duration) error { return nil }
func (errRevocation) IsRevoked(context.Context, string) (bool, error) {
	return false, errors.New("redis down")
}

func runMW(t *testing.T, j *JWT, token string) (string, bool) {
	t.Helper()
	mw := IdentityMiddleware(j)
	var seen string
	var reached bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		seen = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/api/v1/user/me", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: token})
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	return seen, reached
}

func TestRevokedAccessTokenLosesIdentity(t *testing.T) {
	j := NewJWT(mwSecret)
	store := NewMemoryRevocationStore()
	withRevocation(t, store)

	tok, err := j.Generate(4)
	require.NoError(t, err)
	claims, err := j.Parse(tok)
	require.NoError(t, err)
	require.NotEmpty(t, claims.Jti)

	// 吊销前：身份正常
	seen, _ := runMW(t, j, tok)
	assert.Equal(t, "4", seen)

	require.NoError(t, store.Revoke(context.Background(), claims.Jti, time.Hour))

	seen, reached := runMW(t, j, tok)
	assert.True(t, reached, "请求仍应放行到业务层（公开接口照常工作）")
	assert.Equal(t, "", seen, "已吊销的令牌必须失去身份")
}

func TestRevocationStoreErrorFailsOpen(t *testing.T) {
	j := NewJWT(mwSecret)
	withRevocation(t, errRevocation{})

	tok, err := j.Generate(4)
	require.NoError(t, err)

	seen, _ := runMW(t, j, tok)
	assert.Equal(t, "4", seen, "Redis 抖动不能把在线用户踢下线")
}

func TestNoRevocationStoreStillWorks(t *testing.T) {
	j := NewJWT(mwSecret)
	withRevocation(t, nil)

	tok, err := j.Generate(4)
	require.NoError(t, err)

	seen, _ := runMW(t, j, tok)
	assert.Equal(t, "4", seen)
}

func TestLegacyTokenWithoutJtiCannotBeRevokedButStillWorks(t *testing.T) {
	withRevocation(t, NewMemoryRevocationStore())

	// 无 Jti 的存量令牌：吊销名单管不到它，但不能因此失效
	legacy := &Claims{UserId: 4}
	assert.False(t, isRevoked(context.Background(), legacy))
}

func TestConsumeRefreshOnce_AllowsFirstUseThenBlocksReplay(t *testing.T) {
	j := NewJWT(mwSecret)
	withTokenStore(t, NewMemoryTokenStore())

	tok, err := j.GenerateRefresh(4)
	require.NoError(t, err)

	assert.True(t, ConsumeRefreshOnce(context.Background(), j, tok), "首次使用必须放行")
	assert.False(t, ConsumeRefreshOnce(context.Background(), j, tok), "第二次必须判为重放")
}

func TestConsumeRefreshOnce_LegacyTokenWithoutJtiAllowed(t *testing.T) {
	j := NewJWT(mwSecret)
	withTokenStore(t, NewMemoryTokenStore())

	assert.True(t, ConsumeRefreshOnce(context.Background(), j, "not-a-real-jwt"),
		"解析不了/无 Jti 时宁可放行，也不把存量在线用户踢下线")
}

func TestConsumeRefreshOnce_StoreErrorFailsOpen(t *testing.T) {
	j := NewJWT(mwSecret)
	withTokenStore(t, errTokenStore{})

	tok, err := j.GenerateRefresh(4)
	require.NoError(t, err)
	assert.True(t, ConsumeRefreshOnce(context.Background(), j, tok))
}

type errTokenStore struct{}

func (errTokenStore) Consume(context.Context, string, time.Duration) (bool, error) {
	return false, errors.New("redis down")
}

func TestRevokeRequestSession_KillsBothTokens(t *testing.T) {
	j := NewJWT(mwSecret)
	rev := NewMemoryRevocationStore()
	toks := NewMemoryTokenStore()
	withRevocation(t, rev)
	withTokenStore(t, toks)

	access, err := j.Generate(4)
	require.NoError(t, err)
	refresh, refreshJTI, err := j.GenerateRefreshWithJTI(4)
	require.NoError(t, err)

	ac, err := j.Parse(access)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/user/logout", nil)
	req.AddCookie(&http.Cookie{Name: UserAccessTokenCookie, Value: access})
	req.AddCookie(&http.Cookie{Name: UserRefreshTokenCookie, Value: refresh})

	RevokeRequestSession(req, j)

	revoked, err := rev.IsRevoked(context.Background(), ac.Jti)
	require.NoError(t, err)
	assert.True(t, revoked, "访问令牌必须进吊销名单")

	// 刷新令牌被打上"已消费"标记 → 再用会被判重放
	assert.False(t, ConsumeRefreshOnce(context.Background(), j, refresh))
	_ = refreshJTI
}

func TestSetupStoresFromEnv_FallsBackToMemoryWhenRedisDown(t *testing.T) {
	prevR, prevT := CurrentRevocationStore(), CurrentTokenStore()
	t.Cleanup(func() { SetRevocationStore(prevR); SetTokenStore(prevT) })

	t.Setenv("REDIS_ADDR", "127.0.0.1:1") // 必然连不上
	err := SetupStoresFromEnv()
	assert.Error(t, err, "必须把降级情况告诉调用方")
	assert.NotNil(t, CurrentRevocationStore(), "降级也不能把功能整个关掉")
	assert.NotNil(t, CurrentTokenStore())
}
