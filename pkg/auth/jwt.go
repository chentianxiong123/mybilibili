// Package auth 提供统一 JWT 签发/验证能力。
// 设计目标：整个系统（普通用户 + 管理员）使用同一套 HS256 secret，
// 通过 claims 中的 Role/IsAdmin 区分身份。Traefik 网关全权验签后注入身份头，
// 各服务只信任头（HTTP）或通过 gRPC 拦截器验签（服务间调用）。
package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Role 常量
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Claims 标准身份声明。UserId 对应用户表或管理员表的 ID，靠 Role 区分。
type Claims struct {
	UserId  int64  `json:"user_id"`
	Role    string `json:"role,omitempty"`
	IsAdmin bool   `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

// JWT 封装的签发/验证工具。
type JWT struct {
	secret   string
	duration time.Duration
}

func NewJWT(secret string) *JWT {
	return &JWT{secret: secret, duration: 24 * time.Hour}
}

// NewJWTWithDuration 允许自定义有效期（测试/短期 token 用）。
func NewJWTWithDuration(secret string, d time.Duration) *JWT {
	return &JWT{secret: secret, duration: d}
}

// Generate 签发普通用户 token。
func (j *JWT) Generate(userID int64) (string, error) {
	return j.GenerateWithRole(userID, RoleUser)
}

// GenerateAdmin 签发管理员 token（Role=admin, IsAdmin=true）。
func (j *JWT) GenerateAdmin(userID int64) (string, error) {
	return j.GenerateWithRole(userID, RoleAdmin)
}

// GenerateWithRole 按 role 签发 token。
func (j *JWT) GenerateWithRole(userID int64, role string) (string, error) {
	claims := Claims{
		UserId:  userID,
		Role:    role,
		IsAdmin: role == RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// GenerateRefresh 签发 7 天刷新 token（沿用旧行为）。
func (j *JWT) GenerateRefresh(userID int64) (string, error) {
	claims := Claims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// Parse 验签并返回 claims。
func (j *JWT) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// ParseUserID 验签并返回 user_id（普通身份）。
func (j *JWT) ParseUserID(tokenStr string) (int64, error) {
	claims, err := j.Parse(tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.UserId, nil
}

// ---- HTTP 辅助 ----

// BearerToken 从 Authorization 头提取 Bearer token。
func BearerToken(authHeader string) (string, bool) {
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), true
	}
	return "", false
}

// ---- gRPC 拦截器 ----

// GRPCAuthInterceptor 从 gRPC metadata 的 authorization 取 JWT 验签，
// 将身份注入 context（key claimsKey）。服务间调用不走 Traefik，必须自验。
func (j *JWT) GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		auths := md.Get("authorization")
		if len(auths) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}
		tokenStr, ok := BearerToken(auths[0])
		if !ok {
			tokenStr = auths[0]
		}
		claims, err := j.Parse(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(context.WithValue(ctx, claimsKey{}, claims), req)
	}
}

type claimsKey struct{}

// ClaimsFromContext 从 context 取已验签的 claims。
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	return c, ok
}

// ---- 头注入（Traefik 全权验签后，后端读头） ----

// IdentityHeaders 从请求头读取由 Traefik 注入的身份。traefik 上游只注入
// X-User-Id / X-User-Role / X-Admin-Id，后端服务全信这些头（零信任第 1 层）。
func UserIDFromHeader(ctx context.Context, headers func(key string) string) int64 {
	s := headers("X-User-Id")
	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		return id
	}
	return 0
}

// AdminIDFromHeader 读 Traefik 注入的管理员 ID。
func AdminIDFromHeader(getHeader func(key string) string) int64 {
	s := getHeader("X-Admin-Id")
	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		return id
	}
	return 0
}