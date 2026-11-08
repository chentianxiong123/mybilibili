package abstraction

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMemoryCache_IncrStuckAtTwo pins FINDING: memoryCache.Incr is not a real counter.
//
// Incr stores the counter as a single byte []byte{byte(n)} and later recovers the
// current value via len(e.value). Once the value is a one-byte slice (len==1) the
// "current" value is always 1, so after reaching 2 the counter is permanently stuck:
// every subsequent Incr re-reads len==1 and stores 2 again. Incrementing beyond 255
// also truncates (byte(n)). This is the in-process fallback used when Redis is down,
// so rate-limit / dedup counters silently freeze at 2.
func TestMemoryCache_IncrStuckAtTwo(t *testing.T) {
	m, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)

	require.Equal(t, int64(1), mustIncr(t, m, "k"))
	require.Equal(t, int64(2), mustIncr(t, m, "k"))
	// FINDING-077 修复：计数器不再冻结在 2，应继续递增。
	require.Equal(t, int64(3), mustIncr(t, m, "k"))
	require.Equal(t, int64(4), mustIncr(t, m, "k"))
	require.Equal(t, int64(5), mustIncr(t, m, "k"))
}

func mustIncr(t *testing.T, m CacheStore, key string) int64 {
	t.Helper()
	n, err := m.Incr(context.Background(), key, 0)
	require.NoError(t, err)
	return n
}

// TestMemoryCache_MaxItemsNotEnforced pins FINDING: memoryCache.maxItems is assigned in
// newMemoryCache but never used — Set performs no eviction, so the in-process cache
// grows without bound (unbounded memory growth) despite a configured cap.
func TestMemoryCache_MaxItemsNotEnforced(t *testing.T) {
	m, err := newMemoryCache(CacheStoreConfig{MaxItems: 2})
	require.NoError(t, err)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, m.Set(ctx, "k"+string(rune('a'+i)), []byte("v"), time.Hour))
	}
	// FINDING-078 修复：MaxItems=2 → 仅保留最近 2 个，最旧被驱逐。
	_, err = m.Get(ctx, "ka")
	require.Error(t, err, "FINDING-078 修复: ka 应被驱逐")
	_, err = m.Get(ctx, "ke")
	require.NoError(t, err, "ke 应保留")
}

// TestMemoryCache_TTLZeroIsImmediatelyExpired pins FINDING: memoryCache.Set with ttl=0
// sets exp=time.Now(), so the entry is instantly "expired" on the next Get. Redis treats
// ttl=0 as "no expiry" (persistent), so the in-process fallback has opposite TTL
// semantics: a caller passing 0 to mean "persist" gets an unusable entry.
func TestMemoryCache_TTLZeroIsImmediatelyExpired(t *testing.T) {
	m, err := newMemoryCache(CacheStoreConfig{})
	require.NoError(t, err)
	ctx := context.Background()

	require.NoError(t, m.Set(ctx, "k", []byte("v"), 0))
	_, err = m.Get(ctx, "k")
	// FINDING-079 修复：ttl=0 为持久（不再即时过期）。
	require.NoError(t, err, "FINDING-079 修复: ttl=0 应持久")
}
