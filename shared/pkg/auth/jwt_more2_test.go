package auth

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ---- ConsumeRefreshOnce：未注册 TokenStore 的早退分支 ----

func TestConsumeRefreshOnce_NoStoreRegistered(t *testing.T) {
	withTokenStore(t, nil)
	j := NewJWT(mwSecret)
	tok, err := j.GenerateRefresh(4)
	require.NoError(t, err)
	// 无 store：令牌解析有效、有 Jti，但名单缺席 → 放行（兼容存量行为）
	assert.True(t, ConsumeRefreshOnce(context.Background(), j, tok))
}

func TestConsumeRefreshOnce_NilJWT(t *testing.T) {
	// j == nil → 直接放行
	assert.True(t, ConsumeRefreshOnce(context.Background(), nil, "whatever"))
}

// ---- GRPC 拦截器：合法 access 令牌（对照测试，防止 Refresh 拒绝用例误报） ----

func TestGRPCAuthInterceptor_AcceptsAccessToken(t *testing.T) {
	j := NewJWT(testSecret)
	tok, err := j.Generate(66)
	require.NoError(t, err)
	md := metadata.New(map[string]string{"authorization": "Bearer " + tok})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var got int64
	_, err = j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, _ interface{}) (interface{}, error) {
		c, ok := ClaimsFromContext(ctx)
		require.True(t, ok)
		got = c.UserId
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, int64(66), got)
}

// ---- 手工构造带自定义 role 的 JWT，验证 AccessRole 在 Role 分支的行为 ----

func signClaims(t *testing.T, secret string, c *Claims, exp time.Time) string {
	t.Helper()
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
	if c.IssuedAt == nil {
		c.IssuedAt = jwt.NewNumericDate(time.Now())
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	require.NoError(t, err)
	return tok
}

func TestAccessRole_CustomRoleWinsOverIsAdmin(t *testing.T) {
	// IsAdmin 与 Role 同时为真：AccessRole 按定义返回 RoleAdmin（IsAdmin 优先）
	c := &Claims{UserId: 1, IsAdmin: true, Role: "moderator"}
	assert.Equal(t, RoleAdmin, c.AccessRole())

	// 只有 Role 字符串（非 admin）：按 Role 原样返回
	c2 := &Claims{UserId: 2, Role: "moderator"}
	assert.Equal(t, "moderator", c2.AccessRole())

	// 什么都没有：默认 user
	c3 := &Claims{UserId: 3}
	assert.Equal(t, RoleUser, c3.AccessRole())
}

// 用 signClaims 生成带 Role 的 access 风格令牌，Parse 后 AccessRole 应还原。
func TestAccessRole_RoundTripsThroughParse(t *testing.T) {
	tok := signClaims(t, testSecret, &Claims{
		UserId: 12, Role: "moderator", Jti: "x",
	}, time.Now().Add(time.Hour))

	j := NewJWT(testSecret)
	c, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, "moderator", c.AccessRole())
	assert.True(t, c.IsAccess(), "未设置 typ 时视作访问令牌")
}

// 并发安全：同一 JWT 实例被并发解析（gRPC 拦截器场景）。
func TestJWT_Parse_Concurrent(t *testing.T) {
	j := NewJWT(testSecret)
	tok, err := j.Generate(1)
	require.NoError(t, err)

	var n int64
	done := make(chan struct{})
	var once sync.Once
	for i := 0; i < 8; i++ {
		go func() {
			for {
				if _, e := j.Parse(tok); e != nil {
					t.Errorf("Parse: %v", e)
					return
				}
				if atomic.AddInt64(&n, 1) >= 40 {
					once.Do(func() { close(done) })
					return
				}
			}
		}()
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent parse did not finish")
	}
}

// memTokenStore.Consume 并发重放：同一 key 并发消费，恰好一次成功。
func TestMemTokenStore_ConcurrentConsume(t *testing.T) {
	m := &memTokenStore{}
	var okCount int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := m.Consume(context.Background(), "same-key", time.Minute)
			require.NoError(t, err)
			if ok {
				atomic.AddInt64(&okCount, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(1), okCount, "一次性语义：并发重放只有首个成功")
}

// 并发的两个 goroutine 竞争注册全局 store，验证注册/查询互斥无数据竞争。
func TestGlobalStoreRegistry_ConcurrentAccess(t *testing.T) {
	prevR, prevT := CurrentRevocationStore(), CurrentTokenStore()
	t.Cleanup(func() { SetRevocationStore(prevR); SetTokenStore(prevT) })

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			SetRevocationStore(NewMemoryRevocationStore())
		}(i)
		go func(i int) {
			defer wg.Done()
			_ = CurrentTokenStore()
			SetTokenStore(NewMemoryTokenStore())
		}(i)
	}
	wg.Wait()
	assert.NotNil(t, CurrentRevocationStore())
	assert.NotNil(t, CurrentTokenStore())
}

// memRevocation 并发读写不冲突。
func TestMemRevocation_ConcurrentAccess(t *testing.T) {
	m := &memRevocation{}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, m.Revoke(context.Background(), "j", time.Hour))
		}()
		go func() {
			defer wg.Done()
			_, err := m.IsRevoked(context.Background(), "j")
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}
