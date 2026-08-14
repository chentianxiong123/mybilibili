package core

import (
	"sync"
	"time"
)

// commentRateLimiter 简单内存频控：单用户每窗口期最多 N 次评论/回复。
// 对齐旧版 isRateLimited/recordAction/getRemainingCount 的防刷语义。
type commentRateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxCount int
	counts   map[int64][]time.Time
}

func newCommentRateLimiter(window time.Duration, maxCount int) *commentRateLimiter {
	return &commentRateLimiter{
		window:   window,
		maxCount: maxCount,
		counts:   make(map[int64][]time.Time),
	}
}

// record 记录一次动作，返回是否超出频控上限。
func (l *commentRateLimiter) record(userID int64, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	times := l.counts[userID]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.maxCount {
		l.counts[userID] = kept
		return true
	}
	l.counts[userID] = append(kept, now)
	return false
}

// remaining 返回剩余可用次数。
func (l *commentRateLimiter) remaining(userID int64, now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	times := l.counts[userID]
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.counts[userID] = kept
	r := l.maxCount - len(kept)
	if r < 0 {
		r = 0
	}
	return r
}