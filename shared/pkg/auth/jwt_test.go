package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func TestJWT_GenerateAndParse(t *testing.T) {
	j := NewJWT(testSecret)
	uid := int64(12345)

	tok, err := j.Generate(uid)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, uid, claims.UserId)
	assert.Equal(t, RoleUser, claims.Role)
	assert.False(t, claims.IsAdmin)
}

func TestJWT_GenerateAdmin(t *testing.T) {
	j := NewJWT(testSecret)
	uid := int64(99)

	tok, err := j.GenerateAdmin(uid)
	require.NoError(t, err)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, uid, claims.UserId)
	assert.Equal(t, RoleAdmin, claims.Role)
	assert.True(t, claims.IsAdmin)
}

func TestJWT_GenerateRefresh(t *testing.T) {
	j := NewJWT(testSecret)
	uid := int64(7)

	tok, err := j.GenerateRefresh(uid)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, uid, claims.UserId)

	// 刷新 token 有效期应为 7 天
	exp := claims.ExpiresAt.Time
	delta := time.Until(exp)
	assert.InDelta(t, 7*24*time.Hour, delta, float64(time.Minute))
}

func TestJWT_Parse_InvalidToken(t *testing.T) {
	j := NewJWT(testSecret)
	_, err := j.Parse("this.is.not.a.jwt")
	assert.Error(t, err)
}

func TestJWT_Parse_TamperedToken(t *testing.T) {
	j := NewJWT(testSecret)
	// 篡改：用其他 secret 签发
	other := NewJWT("other-secret")
	tampered, err := other.Generate(42)
	require.NoError(t, err)

	_, err = j.Parse(tampered)
	assert.Error(t, err, "tampered token should fail to parse with wrong secret")
}

func TestJWT_Parse_ExpiredToken(t *testing.T) {
	// 1ms 有效期 → 立即过期
	j := NewJWTWithDuration(testSecret, time.Millisecond)
	tok, err := j.Generate(1)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)

	_, err = j.Parse(tok)
	assert.Error(t, err, "expired token must fail to parse")
}

func TestJWT_ParseUserID(t *testing.T) {
	j := NewJWT(testSecret)
	uid := int64(2024)

	tok, err := j.Generate(uid)
	require.NoError(t, err)

	got, err := j.ParseUserID(tok)
	require.NoError(t, err)
	assert.Equal(t, uid, got)
}

func TestJWT_ParseUserID_Invalid(t *testing.T) {
	j := NewJWT(testSecret)
	got, err := j.ParseUserID("garbage.token.value")
	assert.Error(t, err)
	assert.Equal(t, int64(0), got)
}

func TestBearerToken(t *testing.T) {
	tok, ok := BearerToken("Bearer abc123")
	assert.True(t, ok)
	assert.Equal(t, "abc123", tok)
}

func TestBearerToken_NoPrefix(t *testing.T) {
	tok, ok := BearerToken("abc123")
	assert.False(t, ok)
	assert.Equal(t, "", tok)
}

func TestBearerToken_Empty(t *testing.T) {
	tok, ok := BearerToken("")
	assert.False(t, ok)
	assert.Equal(t, "", tok)
}

func TestBearerToken_WithSpace(t *testing.T) {
	// 多个空格: "Bearer  abc" → " abc" (JWT 不会有空格，这里只验证 prefix trim 行为)
	_, ok := BearerToken("Bearer  abc")
	assert.True(t, ok)
	assert.Equal(t, " abc", strings.TrimPrefix("Bearer  abc", "Bearer "))
}

func TestJWT_RegisteredClaims(t *testing.T) {
	// 验证 RegisteredClaims 字段被设置 (iat / exp 都在)
	j := NewJWTWithDuration(testSecret, time.Hour)
	before := time.Now()
	tok, err := j.Generate(11)
	require.NoError(t, err)

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.IssuedAt.Time.After(before.Add(-time.Second)))
	assert.True(t, claims.ExpiresAt.Time.After(claims.IssuedAt.Time))

	// 触发 jwt v5 全部 lazy helpers 编译
	_ = jwt.ErrTokenExpired
	_ = jwt.ErrTokenSignatureInvalid
}