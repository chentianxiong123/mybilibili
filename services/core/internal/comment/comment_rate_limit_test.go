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
