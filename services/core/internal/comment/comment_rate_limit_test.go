package comment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLimiter(max int) *commentRateLimiter {
	return newCommentRateLimiter(time.Minute, max)
}

func TestCommentRateLimiter_RecordWithinLimit(t *testing.T) {
	l := newTestLimiter(3)
	now := time.Now()

	assert.False(t, l.record(1, now))
	assert.False(t, l.record(1, now.Add(time.Second)))
	assert.False(t, l.record(1, now.Add(2*time.Second)))
}

func TestCommentRateLimiter_RecordExceedsLimit(t *testing.T) {
	l := newTestLimiter(2)
	now := time.Now()

	assert.False(t, l.record(1, now))
	assert.False(t, l.record(1, now.Add(time.Second)))
	assert.True(t, l.record(1, now.Add(2*time.Second)))
}

func TestCommentRateLimiter_UsersIsolated(t *testing.T) {
	l := newTestLimiter(1)
	now := time.Now()

	assert.False(t, l.record(1, now))
	assert.False(t, l.record(2, now))
	assert.True(t, l.record(1, now.Add(time.Second)))
}

func TestCommentRateLimiter_WindowExpiry(t *testing.T) {
	l := newTestLimiter(1)
	now := time.Now()

	assert.False(t, l.record(1, now))
	// 窗口过期后再次允许
	assert.False(t, l.record(1, now.Add(61*time.Second)))
}

func TestCommentRateLimiter_Remaining(t *testing.T) {
	l := newTestLimiter(3)
	now := time.Now()

	assert.Equal(t, 3, l.remaining(1, now))
	assert.False(t, l.record(1, now))
	assert.Equal(t, 2, l.remaining(1, now.Add(time.Second)))
	assert.False(t, l.record(1, now.Add(time.Second)))
	assert.False(t, l.record(1, now.Add(2*time.Second)))
	assert.Equal(t, 0, l.remaining(1, now.Add(3*time.Second)))
}

func TestCommentRateLimiter_RemainingNeverNegative(t *testing.T) {
	l := newTestLimiter(1)
	now := time.Now()
	assert.False(t, l.record(1, now))
	assert.True(t, l.record(1, now.Add(time.Second)))
	assert.Equal(t, 0, l.remaining(1, now.Add(2*time.Second)))
}

func TestCommentRateLimiter_RemainingClearsExpired(t *testing.T) {
	l := newTestLimiter(2)
	now := time.Now()
	assert.False(t, l.record(1, now))

	// 过期后 remaining 恢复满
	assert.Equal(t, 2, l.remaining(1, now.Add(2*time.Minute)))
}

func TestCommentRateLimiter_EdgeCases(t *testing.T) {
	l := newCommentRateLimiter(0, 0)
	assert.True(t, l.record(1, time.Now()))
	assert.Equal(t, 0, l.remaining(1, time.Now()))
	require.NotNil(t, l)
}

// ==================== 运行时重配 ====================

// 后台把上限从 20 改成 3 后，无需重启即生效。
func TestCommentRateLimiter_ConfigureLowersLimit(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 20)
	now := time.Now()
	for i := 0; i < 5; i++ {
		require.False(t, l.record(1, now))
	}
	assert.Equal(t, 15, l.remaining(1, now))

	l.configure(time.Minute, 3, now)
	// 收紧后立刻生效：已有 5 条 > 新上限 3
	assert.True(t, l.record(1, now))
}

// 放宽上限后历史计数仍然保留，不会把用户立刻"洗白"。
func TestCommentRateLimiter_ConfigureRaisesLimit(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 2)
	now := time.Now()
	assert.False(t, l.record(1, now))
	assert.False(t, l.record(1, now))
	assert.True(t, l.record(1, now))

	l.configure(time.Minute, 10, now)
	assert.False(t, l.record(1, now))
	assert.Equal(t, 7, l.remaining(1, now))
}

// 配置没变时不触发裁剪，行为等价于noop。
func TestCommentRateLimiter_ConfigureSameValueIsNoop(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 3)
	now := time.Now()
	for i := 0; i < 3; i++ {
		require.False(t, l.record(1, now))
	}
	l.configure(time.Minute, 3, now)
	assert.True(t, l.record(1, now))
}

// 非法配置必须被忽略，保留原阈值，绝不能把限流关掉。
func TestCommentRateLimiter_ConfigureRejectsInvalid(t *testing.T) {
	l := newCommentRateLimiter(time.Minute, 3)
	now := time.Now()

	l.configure(0, 100, now)
	l.configure(-time.Minute, 100, now)
	l.configure(time.Minute, 0, now)
	l.configure(time.Minute, -1, now)

	// 原阈值仍是 3
	for i := 0; i < 3; i++ {
		require.False(t, l.record(1, now), i)
	}
	assert.True(t, l.record(1, now))
}

// 窗口缩短时按新窗口裁剪历史记录。
func TestCommentRateLimiter_ConfigureShrinksWindow(t *testing.T) {
	l := newCommentRateLimiter(time.Hour, 10)
	now := time.Now()
	// 50 分钟前记过一次
	require.False(t, l.record(1, now.Add(-50*time.Minute)))
	// 现在窗口缩到 10 分钟，那次记录已过期
	l.configure(10*time.Minute, 10, now)
	assert.Equal(t, 10, l.remaining(1, now))
	// map 里的过期项被清掉，不泄漏内存
	l.mu.Lock()
	assert.Empty(t, l.counts)
	l.mu.Unlock()
}
