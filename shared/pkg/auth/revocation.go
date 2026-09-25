package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Redis key 前缀
const (
	revokedKeyPrefix = "auth:revoked:" // 访问令牌吊销名单
	liveKeyPrefix    = "auth:live:"    // 尚未消费的一次性刷新令牌
)

// RevocationStore 访问令牌吊销名单。
//
// 为什么需要它：JWT 是无状态的，签出去就收不回，登出、改密、封号在令牌
// 自然过期前都是"纸面生效"。这里用一份带 TTL 的名单把它补上。
//
// 实现要求：IsRevoked 出错时必须返回 err，由调用方**放行**——
// Redis 抖动不应该把全站登录用户踢下线，吊销是尽力而为的安全增强，
// 可用性优先。
type RevocationStore interface {
	// Revoke 让 jti 在 ttl 内失效（ttl 取令牌剩余寿命即可，到期自动清理）。
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	// IsRevoked 查询 jti 是否已吊销。
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

// TokenStore 刷新令牌防重放（轮换）。
//
// 用"首次消费标记"而不是"取出即删"：上线当天存量刷新令牌从未被登记过，
// 取出即删会把它们全判成重放、把所有在线用户踢下线。
// 首次消费时 SETNX 打标，第二次见到即为重放——既兼容存量，又能抓到窃取重用。
type TokenStore interface {
	// Consume 标记为已使用；返回 false 表示此前已用过（重放）。
	// ttl 取令牌剩余寿命，到期自动清理。
	Consume(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// RevokeRequestSession 把当前请求携带的登录态立刻作废：
// 访问令牌进吊销名单（等它自然过期为止），刷新令牌打上已消费标记。
// 供登出接口调用——只清浏览器 cookie 的话，被复制走的令牌照样能用到过期。
func RevokeRequestSession(r *http.Request, j *JWT) {
	if r == nil || j == nil {
		return
	}
	ctx := r.Context()

	if st := CurrentRevocationStore(); st != nil {
		if tok := TokenFromRequest(r); tok != "" {
			if c, err := j.Parse(tok); err == nil && c != nil && c.Jti != "" {
				_ = st.Revoke(ctx, c.Jti, remaining(c.ExpiresAt))
			}
		}
	}
	if ts := CurrentTokenStore(); ts != nil {
		for _, name := range []string{UserRefreshTokenCookie, AdminRefreshTokenCookie} {
			if cookie, err := r.Cookie(name); err == nil && cookie.Value != "" {
				if c, err := j.Parse(cookie.Value); err == nil && c != nil && c.Jti != "" {
					_, _ = ts.Consume(ctx, c.Jti, remaining(c.ExpiresAt))
				}
			}
		}
	}
}

// RevokeRefreshToken 单独作废一张刷新令牌（刷新轮换时消费旧的那张）。
func RevokeRefreshToken(ctx context.Context, j *JWT, tokenStr string) {
	if j == nil || tokenStr == "" {
		return
	}
	c, err := j.Parse(tokenStr)
	if err != nil || c == nil || c.Jti == "" {
		return
	}
	if ts := CurrentTokenStore(); ts != nil {
		_, _ = ts.Consume(ctx, c.Jti, remaining(c.ExpiresAt))
	}
}

// ConsumeRefreshOnce 刷新令牌防重放：首次消费返回 true，重放返回 false。
// 无 Jti 的存量令牌直接放行（无法追踪，宁可兼容也不把在线用户踢下线）；
// 名单读取出错同样放行，理由与 isRevoked 一致——可用性优先。
func ConsumeRefreshOnce(ctx context.Context, j *JWT, tokenStr string) bool {
	if j == nil {
		return true
	}
	c, err := j.Parse(tokenStr)
	if err != nil || c == nil || c.Jti == "" {
		return true
	}
	ts := CurrentTokenStore()
	if ts == nil {
		return true
	}
	ok, err := ts.Consume(ctx, c.Jti, remaining(c.ExpiresAt))
	if err != nil {
		return true
	}
	return ok
}

// remaining 距离过期还有多久，异常时给个兜底值。
func remaining(exp *jwt.NumericDate) time.Duration {
	if exp == nil {
		return time.Hour
	}
	d := time.Until(exp.Time)
	if d <= 0 {
		return time.Minute
	}
	return d
}

// ---- Redis 实现 ----

type redisRevocation struct{ c *redis.Client }

func (r *redisRevocation) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	return r.c.Set(ctx, revokedKeyPrefix+jti, "1", ttl).Err()
}

func (r *redisRevocation) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := r.c.Exists(ctx, revokedKeyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

type redisTokenStore struct{ c *redis.Client }

// Consume 用 SETNX 原子打标：并发重放同一个刷新令牌时只有一个能成功。
func (r *redisTokenStore) Consume(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, nil
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return r.c.SetNX(ctx, liveKeyPrefix+key, "1", ttl).Result()
}

// ---- 进程内实现（Redis 不可用 / 测试用）----

type memRevocation struct {
	mu sync.Mutex
	m  map[string]time.Time
}

func (m *memRevocation) Revoke(_ context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = map[string]time.Time{}
	}
	m.m[jti] = time.Now().Add(ttl)
	return nil
}

func (m *memRevocation) IsRevoked(_ context.Context, jti string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.m[jti]
	if !ok {
		return false, nil
	}
	if time.Now().After(exp) {
		delete(m.m, jti)
		return false, nil
	}
	return true, nil
}

type memTokenStore struct {
	mu sync.Mutex
	m  map[string]time.Time
}

func (m *memTokenStore) Consume(_ context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = map[string]time.Time{}
	}
	if exp, ok := m.m[key]; ok && time.Now().Before(exp) {
		return false, nil // 已消费过 = 重放
	}
	m.m[key] = time.Now().Add(ttl)
	return true, nil
}

// NewMemoryRevocationStore 进程内吊销名单（测试 / Redis 缺席时的降级）。
func NewMemoryRevocationStore() RevocationStore { return &memRevocation{} }

// NewMemoryTokenStore 进程内一次性令牌存储。
func NewMemoryTokenStore() TokenStore { return &memTokenStore{} }

// ---- 全局注册 ----
//
// core 的 main 里 auth.NewJWT 被调用了 8 次（每个 handler 一个实例），
// 逐实例注入既啰嗦又容易漏，因此用包级存储 + 请求时读取。

var (
	regMu         sync.RWMutex
	revocation    RevocationStore
	tokenRegistry TokenStore
)

// SetRevocationStore 注册吊销名单（nil 表示关闭吊销功能）。
func SetRevocationStore(s RevocationStore) {
	regMu.Lock()
	revocation = s
	regMu.Unlock()
}

// CurrentRevocationStore 取当前吊销名单，未注册时返回 nil。
func CurrentRevocationStore() RevocationStore {
	regMu.RLock()
	defer regMu.RUnlock()
	return revocation
}

// SetTokenStore 注册一次性令牌存储。
func SetTokenStore(s TokenStore) {
	regMu.Lock()
	tokenRegistry = s
	regMu.Unlock()
}

// CurrentTokenStore 取当前一次性令牌存储，未注册时返回 nil。
func CurrentTokenStore() TokenStore {
	regMu.RLock()
	defer regMu.RUnlock()
	return tokenRegistry
}

// SetupStoresFromEnv 用 REDIS_ADDR 连接 Redis 并注册吊销 / 一次性令牌存储。
//
// 连不上时返回错误，但**仍然注册进程内实现**：吊销功能宁可退化成
// "单实例有效"，也不能因为 Redis 抖动就整站失效。调用方打个 warning 即可。
func SetupStoresFromEnv() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	c := redis.NewClient(&redis.Options{Addr: addr, DialTimeout: 2 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		SetRevocationStore(NewMemoryRevocationStore())
		SetTokenStore(NewMemoryTokenStore())
		return fmt.Errorf("redis %s 不可达，吊销名单降级为进程内: %w", addr, err)
	}
	SetRevocationStore(&redisRevocation{c: c})
	SetTokenStore(&redisTokenStore{c: c})
	return nil
}

// LogSetupWarning 是给各服务 main 用的统一日志出口，避免每处重复格式化。
func LogSetupWarning(err error) {
	if err != nil {
		log.Printf("auth: %s", strings.TrimSpace(err.Error()))
	}
}
