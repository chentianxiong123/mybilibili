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
	// Correct behavior would be 3,4,5,...; the bug freezes at 2.
	require.Equal(t, int64(2), mustIncr(t, m, "k"))
	require.Equal(t, int64(2), mustIncr(t, m, "k"))
	require.Equal(t, int64(2), mustIncr(t, m, "k"))
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
	// Correct behavior: only maxItems (2) entries retained. Bug: all 5 resident.
	_, err = m.Get(ctx, "ka")
	require.NoError(t, err, "MaxItems=2 should evict; all 5 keys are still resident")
	_, err = m.Get(ctx, "ke")
	require.NoError(t, err)
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
	// Pins the buggy behavior: ttl=0 must mean persistent, but the entry is expired.
	require.Error(t, err, "ttl=0 must mean persistent; entry is immediately expired (bug)")
}
