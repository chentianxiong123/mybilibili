package comment

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// SecuritySettings 后台"违禁词 - 频率限制与缓存设置"卡片对应的配置。
// 字段名与 admin 前端 ProhibitedWordsView 提交的保持一致。
type SecuritySettings struct {
	CommentMaxCount             int `json:"commentMaxCount"`
	CommentWindowSeconds        int `json:"commentWindowSeconds"`
	ReplyMaxCount               int `json:"replyMaxCount"`
	ReplyWindowSeconds          int `json:"replyWindowSeconds"`
	CacheRefreshIntervalSeconds int `json:"cacheRefreshIntervalSeconds"`
}

const (
	defaultCommentMaxCount   = 20
	defaultCommentWindowSecs = 600
	defaultReplyMaxCount     = 20
	defaultReplyWindowSecs   = 600
	defaultCacheRefreshSecs  = 300

	securitySettingsKey = "security_settings"
	// 配置读取比违禁词表轻，但也没必要每次发评论都查，30s 足够跟随后台改动。
	securitySettingsTTL = 30 * time.Second
)

func defaultSecuritySettings() SecuritySettings {
	return SecuritySettings{
		CommentMaxCount:             defaultCommentMaxCount,
		CommentWindowSeconds:        defaultCommentWindowSecs,
		ReplyMaxCount:               defaultReplyMaxCount,
		ReplyWindowSeconds:          defaultReplyWindowSecs,
		CacheRefreshIntervalSeconds: defaultCacheRefreshSecs,
	}
}

// securityConfigProvider 带 TTL 缓存地读 system_configs.security_settings。
// 读失败或字段缺失一律回落到默认值，绝不因为配置问题让评论发不出去。
type securityConfigProvider struct {
	mu     sync.RWMutex
	db     *sql.DB
	ttl    time.Duration
	cached SecuritySettings
	at     time.Time
	valid  bool
}

func newSecurityConfigProvider(db *sql.DB) *securityConfigProvider {
	return &securityConfigProvider{db: db, ttl: securitySettingsTTL}
}

// get 返回当前生效配置，必要时回源。
func (p *securityConfigProvider) get(ctx context.Context) SecuritySettings {
	p.mu.RLock()
	if p.valid && time.Since(p.at) < p.ttl {
		s := p.cached
		p.mu.RUnlock()
		return s
	}
	p.mu.RUnlock()

	s := p.load(ctx)

	p.mu.Lock()
	p.cached = s
	p.at = time.Now()
	p.valid = true
	p.mu.Unlock()
	return s
}

// invalidate 让下次 get 强制回源，后台保存配置后可调用。
func (p *securityConfigProvider) invalidate() {
	p.mu.Lock()
	p.valid = false
	p.mu.Unlock()
}

func (p *securityConfigProvider) load(ctx context.Context) SecuritySettings {
	out := defaultSecuritySettings()
	if p.db == nil {
		return out
	}
	var raw string
	if err := p.db.QueryRowContext(ctx,
		`SELECT config_value FROM system_configs WHERE config_key=$1`, securitySettingsKey).Scan(&raw); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("comment: 读取安全设置失败，用默认值: %v", err)
		}
		return out
	}
	if raw == "" {
		return out
	}
	var parsed SecuritySettings
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		log.Printf("comment: 安全设置不是合法 JSON，用默认值: %v", err)
		return out
	}
	// 逐字段兜底：后台可能只存了其中一项，或者存了 0/负数
	if parsed.CommentMaxCount > 0 {
		out.CommentMaxCount = parsed.CommentMaxCount
	}
	if parsed.CommentWindowSeconds > 0 {
		out.CommentWindowSeconds = parsed.CommentWindowSeconds
	}
	if parsed.ReplyMaxCount > 0 {
		out.ReplyMaxCount = parsed.ReplyMaxCount
	}
	if parsed.ReplyWindowSeconds > 0 {
		out.ReplyWindowSeconds = parsed.ReplyWindowSeconds
	}
	if parsed.CacheRefreshIntervalSeconds > 0 {
		out.CacheRefreshIntervalSeconds = parsed.CacheRefreshIntervalSeconds
	}
	return out
}

// apply 把配置下发到评论/回复两个独立的限流桶。
// 评论与回复分开计数（原先共用一个桶，后台分开配没有意义）。
func (s *CommentService) apply(cfg SecuritySettings) {
	now := time.Now()
	s.commentLimiter.configure(time.Duration(cfg.CommentWindowSeconds)*time.Second, cfg.CommentMaxCount, now)
	s.replyLimiter.configure(time.Duration(cfg.ReplyWindowSeconds)*time.Second, cfg.ReplyMaxCount, now)
	if s.prohibited != nil {
		s.prohibited.setTTL(time.Duration(cfg.CacheRefreshIntervalSeconds) * time.Second)
	}
}

// refreshSecurityConfig 拉一次配置并下发，供后台协程周期调用。
func (s *CommentService) refreshSecurityConfig(ctx context.Context) {
	if s.securityCfg == nil {
		return
	}
	s.apply(s.securityCfg.get(ctx))
}
