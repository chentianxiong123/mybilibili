package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"mybilibili/pkg/auth"
)

// TestJWTSecret 测试用固定 JWT 密钥。生产请勿复用。
const TestJWTSecret = "test-secret-do-not-use-in-prod"

// NewTestJWT 返回一个 24h 有效期的 JWT 实例（用于本地/单测）。
func NewTestJWT() *auth.JWT {
	return auth.NewJWT(TestJWTSecret)
}

// NewTestJWTShort 返回一个 1ms 短有效期 JWT 实例（用于过期用例）。
func NewTestJWTShort() *auth.JWT {
	return auth.NewJWTWithDuration(TestJWTSecret, 1)
}

// RandomString 生成 n 字节随机字符串（hex 编码，长度 = 2n）。
// 用于：测试用户名、昵称、邮箱等避免冲突。
func RandomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RandomUsername 返回 `test_<hex8>` 用户名（避免与真实用户冲突）。
func RandomUsername() string {
	return "test_" + RandomString(4)
}

// AssertGRPCError 断言 err 是 gRPC 风格 error 且 code 匹配。
// 简化 gRPC status 检查。
func AssertGRPCError(t *testing.T, err error, codeWant string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected gRPC error code %s, got nil", codeWant)
	}
	got := err.Error()
	if !contains(got, codeWant) {
		t.Fatalf("expected error to contain %q, got %q", codeWant, got)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}