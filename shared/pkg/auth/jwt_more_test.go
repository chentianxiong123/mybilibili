package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAdminIDFromHeader_Invalid(t *testing.T) {
	id := AdminIDFromHeader(func(key string) string {
		if key == "X-Admin-Id" {
			return "not-a-number"
		}
		return ""
	})
	assert.Equal(t, int64(0), id)
}

func TestAdminIDFromHeader_Empty(t *testing.T) {
	id := AdminIDFromHeader(func(key string) string { return "" })
	assert.Equal(t, int64(0), id)
}

func TestUserIDFromHeader_WithCtx(t *testing.T) {
	// ctx 当前未参与计算, 只传 nil 不 panic 即可
	id := UserIDFromHeader(context.Background(), func(key string) string {
		if key == "X-User-Id" {
			return "999"
		}
		return ""
	})
	assert.Equal(t, int64(999), id)
}

func TestJWT_Parse_WrongAlgorithm(t *testing.T) {
	// 用 RS256 签发 token, 让 HMAC 验证失败
	claims := Claims{UserId: 1}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	j := NewJWT(testSecret)
	_, err = j.Parse(signed)
	assert.Error(t, err)
	assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
}

func TestJWT_GenerateWithRole_CustomRole(t *testing.T) {
	j := NewJWT(testSecret)
	tok, err := j.GenerateWithRole(7, "moderator")
	require.NoError(t, err)
	c, err := j.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, "moderator", c.Role)
	assert.False(t, c.IsAdmin)
}

func TestClaimsFromContext_Present(t *testing.T) {
	claims := &Claims{UserId: 5, Role: RoleUser}
	ctx := context.WithValue(context.Background(), claimsKey{}, claims)
	got, ok := ClaimsFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, int64(5), got.UserId)
}

func TestClaimsFromContext_Absent(t *testing.T) {
	got, ok := ClaimsFromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, got)
}

// --- GRPC interceptor ---

type fakeHandler struct {
	called bool
	gotCtx context.Context
}

func (f *fakeHandler) handler(ctx context.Context, req interface{}) (interface{}, error) {
	f.called = true
	f.gotCtx = ctx
	// 从 ctx 取 claims 验证
	c, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, errors.New("claims missing")
	}
	return c.UserId, nil
}

func TestGRPCAuthInterceptor_Success(t *testing.T) {
	j := NewJWT(testSecret)
	tok, err := j.Generate(123)
	require.NoError(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer " + tok})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	fh := &fakeHandler{}
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	resp, err := j.GRPCAuthInterceptor()(ctx, nil, info, fh.handler)
	require.NoError(t, err)
	assert.True(t, fh.called)
	assert.Equal(t, int64(123), resp)
}

func TestGRPCAuthInterceptor_MissingMetadata(t *testing.T) {
	j := NewJWT(testSecret)
	ctx := context.Background() // 无 metadata
	_, err := j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, interface{}) (interface{}, error) { return nil, nil })
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPCAuthInterceptor_MissingAuthHeader(t *testing.T) {
	j := NewJWT(testSecret)
	md := metadata.New(map[string]string{}) // 无 authorization
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, interface{}) (interface{}, error) { return nil, nil })
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPCAuthInterceptor_InvalidToken(t *testing.T) {
	j := NewJWT(testSecret)
	md := metadata.New(map[string]string{"authorization": "Bearer bad.token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, interface{}) (interface{}, error) { return nil, nil })
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGRPCAuthInterceptor_NonBearer(t *testing.T) {
	// 不带 Bearer 前缀, 直接用 token 串作为值 (grpc 服务间常见)
	j := NewJWT(testSecret)
	tok, err := j.Generate(42)
	require.NoError(t, err)
	md := metadata.New(map[string]string{"authorization": tok})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err = j.GRPCAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, _ interface{}) (interface{}, error) {
		c, _ := ClaimsFromContext(ctx)
		if c == nil {
			return nil, errors.New("nil")
		}
		return c.UserId, nil
	})
	require.NoError(t, err)
}

// --- IdentityMiddleware 边角 ---

func TestIdentityMiddleware_AdminHeadersViaXUserId(t *testing.T) {
	// X-User-Id 直通: 当 Traefik 注入 X-User-Id 时, Role/Admin 不变 (admin 走 X-Admin-Id)
	j := NewJWT(mwSecret)
	mw := IdentityMiddleware(j)
	var gotRole, gotAdmin string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRole = r.Header.Get("X-User-Role")
		gotAdmin = r.Header.Get("X-Admin-Id")
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-User-Id", "11")
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", gotRole)
	assert.Equal(t, "", gotAdmin)
}

func TestIdentityMiddleware_BearerInvalidSignature(t *testing.T) {
	// 用错 secret 签的 token 应不注入任何身份头
	j := NewJWT(mwSecret)
	other := NewJWT("wrong")
	tok, _ := other.Generate(1)
	mw := IdentityMiddleware(j)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.Header.Get("X-User-Id"))
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mw(inner).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}