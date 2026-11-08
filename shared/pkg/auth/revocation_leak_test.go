package auth

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMemRevocation_ExpiredEntriesOnlyClearedOnAccess pins FINDING: the in-process
// revocation store only prunes an expired entry when that exact jti is queried again.
//
// memRevocation.IsRevoked deletes an entry lazily, on access. A jti that is revoked
// and never presented again stays in the map forever (its TTL is never enforced), so a
// long-running fallback deployment (Redis down) accumulates revoked jtis without bound.
func TestMemRevocation_ExpiredEntriesOnlyClearedOnAccess(t *testing.T) {
	st := NewMemoryRevocationStore().(*memRevocation)
	ctx := context.Background()

	// Revoke several jtis with a tiny TTL.
	for i := 0; i < 100; i++ {
		require.NoError(t, st.Revoke(ctx, "jti-"+strconv.Itoa(i), time.Millisecond))
	}

	// Let them expire.
	time.Sleep(5 * time.Millisecond)

	size := reflectMapSize(t, st.m)
	require.Greater(t, size, 0, "expired entries should be pruned without explicit query")

	// Queried entries are cleaned up...
	_, err := st.IsRevoked(ctx, "jti-0")
	require.NoError(t, err)
	// ...but the other 99 are still resident even though all have expired.
	size2 := reflectMapSize(t, st.m)
	require.Equal(t, size-1, size2, "only the queried jti was pruned; the rest leak")
}

func reflectMapSize(t *testing.T, m map[string]time.Time) int {
	t.Helper()
	v := reflect.ValueOf(m)
	require.True(t, v.IsValid() && v.Kind() == reflect.Map, "unexpected internal shape")
	return v.Len()
}
