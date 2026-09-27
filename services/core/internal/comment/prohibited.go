package comment

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"
)

// 违禁词匹配方式。
const (
	matchTypeContains = "CONTAINS" // 内容包含即命中
	matchTypeExact    = "EXACT"    // 去掉首尾空白后整条相等才算命中
)

// statusPendingReview 命中违禁词后的待审核状态。
// 评论用 comments.status=1（列表只查 status=0），回复用 replies.status='PENDING'
// （列表只查 replyStatusNormal），两者都只对管理员可见。
const statusPendingReview = "PENDING"

// replyStatusNormal 回复正常状态，公开列表只查这一种。
const replyStatusNormal = "NORMAL"

// ProhibitedWord 一条生效中的违禁词。
type ProhibitedWord struct {
	Word      string
	MatchType string
}

// prohibitedWordStore 把违禁词表缓存在进程内，避免每次发评论/回复都全表扫库。
//
// 原实现是 `SELECT COUNT(*) FROM prohibited_words WHERE $1 ILIKE '%'||word||'%'`，
// 前导通配符用不上任何索引，等于每条内容都全表扫一次；而且漏了 is_enabled 过滤、
// 也无视 match_type，导致后台"禁用"某个词完全不起作用、EXACT 词被当包含匹配误伤。
type prohibitedWordStore struct {
	mu     sync.RWMutex
	words  []ProhibitedWord
	db     *sql.DB
	ttl    time.Duration
	loaded bool
}

// newProhibitedWordStore 创建缓存。ttl <= 0 时退化为 5 分钟。
func newProhibitedWordStore(db *sql.DB, ttl time.Duration) *prohibitedWordStore {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &prohibitedWordStore{db: db, ttl: ttl}
}

// match 命中返回违禁词本身，未命中返回 ""。
// 查不到数据（db 为 nil 或首次加载前）一律放行，不阻塞正常评论。
func (s *prohibitedWordStore) match(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	words := s.snapshot()
	if len(words) == 0 {
		return ""
	}
	trimmed := strings.TrimSpace(content)
	lowerContent := strings.ToLower(content)
	lowerTrimmed := strings.ToLower(trimmed)

	for _, w := range words {
		word := strings.TrimSpace(w.Word)
		if word == "" {
			continue
		}
		if strings.EqualFold(w.MatchType, matchTypeExact) {
			if lowerTrimmed == strings.ToLower(word) {
				return word
			}
			continue
		}
		// CONTAINS 是默认语义，也覆盖 match_type 为空/未知的历史数据
		if strings.Contains(lowerContent, strings.ToLower(word)) {
			return word
		}
	}
	return ""
}

func (s *prohibitedWordStore) snapshot() []ProhibitedWord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.words
}

// size 返回当前缓存的词条数，供测试与监控使用。
func (s *prohibitedWordStore) size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.words)
}

// setTTL 更新刷新间隔，下一次 refresh 生效。
func (s *prohibitedWordStore) setTTL(ttl time.Duration) {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	s.mu.Lock()
	s.ttl = ttl
	s.mu.Unlock()
}

// refresh 从库里重载启用中的违禁词。加载失败保留旧缓存，避免线上抖动导致漏拦。
func (s *prohibitedWordStore) refresh(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT word, match_type FROM prohibited_words WHERE is_enabled = 1`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var words []ProhibitedWord
	for rows.Next() {
		var w ProhibitedWord
		var matchType sql.NullString
		if err := rows.Scan(&w.Word, &matchType); err != nil {
			return err
		}
		w.MatchType = matchType.String
		if w.MatchType == "" {
			w.MatchType = matchTypeContains
		}
		words = append(words, w)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	s.words = words
	s.loaded = true
	s.mu.Unlock()
	return nil
}

// StartRefresher 按当前 ttl 周期后台刷新，直到 ctx 结束。
// 每次刷新后会重新读一次 ttl，这样后台改了"缓存刷新间隔"能自动生效。
func (s *prohibitedWordStore) StartRefresher(ctx context.Context, ttlProvider func() time.Duration) {
	if ttlProvider == nil {
		ttlProvider = func() time.Duration { return s.currentTTL() }
	}
	go func() {
		if err := s.refresh(ctx); err != nil {
			log.Printf("comment: 违禁词缓存首次加载失败: %v", err)
		}
		for {
			wait := ttlProvider()
			if wait <= 0 {
				wait = 5 * time.Minute
			}
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				if err := s.refresh(ctx); err != nil {
					// 保留旧缓存继续用，只记日志
					log.Printf("comment: 违禁词缓存刷新失败，沿用旧数据: %v", err)
				}
			}
		}
	}()
}

func (s *prohibitedWordStore) currentTTL() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ttl
}
