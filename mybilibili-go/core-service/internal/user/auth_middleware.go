package user

import (
	"net/http"
	"strconv"
	"strings"
)

// AuthMiddleware 解析 Authorization: Bearer <token>，校验 JWT 签名后
// 将 userId 写入 X-User-Id 请求头，下流 handler 无感兼容。
// 如果请求已有 X-User-Id（Flutter 双发），优先保留。
// 如果 Bearer token 无效，不清除已有 X-User-Id（Flutter 兜底），
// 但如无 X-User-Id 且 token 无效，不设头（handler 自行判 401）。
func AuthMiddleware(j *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				tokenStr := strings.TrimPrefix(auth, "Bearer ")
				userID, err := j.Parse(tokenStr)
				if err == nil && userID > 0 {
					// 只在没有 X-User-Id 时写入（Flutter 双发时保留原值）
					if r.Header.Get("X-User-Id") == "" {
						r.Header.Set("X-User-Id", strconv.FormatInt(userID, 10))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}