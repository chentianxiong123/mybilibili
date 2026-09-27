package comment

import (
	"sync"
	"time"
)

// commentRateLimiter 简单内存频控：单用户每窗口期最多 N 次动作。
// 对齐旧版 isRateLimited/recordAction/getRemainingCount 的防刷语义。
//
// 窗口与上限都可以在运行时通过 configure 改（后台"安全设置"页可配），
// 改完立即对新请求生效，无需重启服务。
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

// configure 热更新窗口与上限。非法的入参会被忽略，保留原配置。
// 收紧上限时顺带按新窗口裁剪已有记录，避免旧窗口里的历史计数被误算成超限。
func (l *commentRateLimiter) configure(window time.Duration, maxCount int, now time.Time) {
	if window <= 0 || maxCount <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.window == window && l.maxCount == maxCount {
		return
	}
	l.window = window
	l.maxCount = maxCount
	cutoff := now.Add(-window)
	for uid, times := range l.counts {
		kept := times[:0]
		for _, t := range times {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(l.counts, uid)
			continue
		}
		l.counts[uid] = kept
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
	// 窗口内没有记录时直接删 key，避免 map 随用户数无限增长
	if len(kept) == 0 {
		delete(l.counts, userID)
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
	if len(kept) == 0 {
		delete(l.counts, userID)
	} else {
		l.counts[userID] = kept
	}
	r := l.maxCount - len(kept)
	if r < 0 {
		r = 0
	}
	return r
}
