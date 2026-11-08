package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// IsUserRefresh 的 nil / 显式 typ / 旧令牌兼容 三个分支。
func TestClaims_IsUserRefresh_AllBranches(t *testing.T) {
	t.Run("nil claims are treated as legacy user refresh", func(t *testing.T) {
		var c *Claims
		assert.True(t, c.IsUserRefresh())
	})
	t.Run("empty typ is legacy user refresh", func(t *testing.T) {
		assert.True(t, (&Claims{}).IsUserRefresh())
	})
	t.Run("explicit refresh", func(t *testing.T) {
		assert.True(t, (&Claims{Typ: TokenTypeRefresh}).IsUserRefresh())
	})
	t.Run("access is not user refresh", func(t *testing.T) {
		assert.False(t, (&Claims{Typ: TokenTypeAccess}).IsUserRefresh())
	})
	t.Run("admin refresh is not user refresh", func(t *testing.T) {
		assert.False(t, (&Claims{Typ: TokenTypeAdminRefresh}).IsUserRefresh())
	})
}

func TestClaims_IsAdminRefresh_AllBranches(t *testing.T) {
	t.Run("nil is not admin refresh", func(t *testing.T) {
		var c *Claims
		assert.False(t, c.IsAdminRefresh())
	})
	t.Run("empty typ is not admin refresh", func(t *testing.T) {
		assert.False(t, (&Claims{}).IsAdminRefresh())
	})
	t.Run("user refresh is not admin refresh", func(t *testing.T) {
		assert.False(t, (&Claims{Typ: TokenTypeRefresh}).IsAdminRefresh())
	})
	t.Run("explicit admin refresh", func(t *testing.T) {
		assert.True(t, (&Claims{Typ: TokenTypeAdminRefresh}).IsAdminRefresh())
	})
}

func TestClaims_AccessRole_NilClaims(t *testing.T) {
	var c *Claims
	assert.Equal(t, RoleUser, c.AccessRole())
}

func TestJWTFromEnv_DefaultAndOverride(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	j := JWTFromEnv()
	require.NotNil(t, j)
	// 默认 secret 签发的令牌必须能被同一个默认值解析
	tok, err := j.Generate(1)
	require.NoError(t, err)
	claims, err := NewJWT(DefaultSecret).Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserId)

	t.Setenv("JWT_SECRET", "custom-env-secret")
	j2 := JWTFromEnv()
	require.NotNil(t, j2)
	tok2, err := j2.Generate(2)
	require.NoError(t, err)
	claims2, err := j2.Parse(tok2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), claims2.UserId)
	// 默认 secret 的实例不能解析 custom 签发的令牌
	_, err = NewJWT(DefaultSecret).Parse(tok2)
	assert.Error(t, err)
}

// GenerateAdminRefresh 必须不携带 Role/IsAdmin（见源码注释：刻意裸载荷），
// 且有效期 7 天、类型标记为 admin_refresh。
func TestGenerateAdminRefresh_NoRoleClaims(t *testing.T) {
	j := NewJWT(testSecret)
	tok, err := j.GenerateAdminRefresh(8)
	require.NoError(t, err)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Empty(t, claims.Role, "管理员刷新令牌刻意不带 Role")
	assert.False(t, claims.IsAdmin, "管理员刷新令牌刻意不带 IsAdmin")
	assert.True(t, claims.IsAdminRefresh())
	assert.False(t, claims.IsAccess())
	assert.False(t, claims.IsUserRefresh())
	assert.NotEmpty(t, claims.Jti)
	assert.InDelta(t, 7*24*time.Hour, time.Until(claims.ExpiresAt.Time), float64(time.Minute))
}

func TestGenerateAdminRefreshWithJTI_ReturnsMatchingJTI(t *testing.T) {
	j := NewJWT(testSecret)
	tok, jti, err := j.GenerateAdminRefreshWithJTI(9)
	require.NoError(t, err)
	require.NotEmpty(t, jti)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, jti, claims.Jti)
	assert.Equal(t, TokenTypeAdminRefresh, claims.Typ)
}

// Parse 的 "!token.Valid" 分支：HS256 签发但过期时 claims 类型转换成功而 token.Valid 为假。
func TestJWT_Parse_ExpiredValidClaims(t *testing.T) {
	j := NewJWTWithDuration(testSecret, time.Millisecond)
	tok, err := j.Generate(1)
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	claims, err := j.Parse(tok)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

// 未过期但类型断言失败不可达（Parse 只会产出 *Claims），故过期路径覆盖
// jwt.go:218-220 的 "!token.Valid" 分支。

// NewJTI 输出是 128bit 十六进制且彼此独立。
func TestNewJTI_Format(t *testing.T) {
	a, b := NewJTI(), NewJTI()
	assert.Len(t, a, 32)
	assert.NotEqual(t, a, b)
	_, err := jtiHex(a)
	assert.NoError(t, err)
}

func jtiHex(s string) (uint64, error) {
	if len(s) != 32 {
		return 0, assert.AnError
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return 0, assert.AnError
		}
	}
	return 0, nil
}

// GRPCAuthInterceptor 对刷新令牌（IsAccess 为假）必须拒绝，
// 7 天的刷新令牌不能当 24h 访问令牌打进服务端。
func TestGRPCAuthInterceptor_RejectsRefreshToken(t *testing.T) {
	j := NewJWT(testSecret)
	refresh, err := j.GenerateRefresh(5)
	require.NoError(t, err)
	adminRefresh, err := j.GenerateAdminRefresh(5)
	require.NoError(t, err)

	for name, tok := range map[string]string{
		"user refresh":    refresh,
		"admin refresh":   adminRefresh,
		"garbage":         "not-a-jwt",
	} {
		t.Run(name, func(t *testing.T) {
			md := metadata.New(map[string]string{"authorization": "Bearer " + tok})
			ctx := metadata.NewIncomingContext(context.Background(), md)
			called := false
			inner := func(ctx context.Context, _ interface{}) (interface{}, error) {
				called = true
				return nil, nil
			}
			_, err := j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, inner)
			require.Error(t, err)
			assert.False(t, called)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.Unauthenticated, st.Code())
		})
	}
}
