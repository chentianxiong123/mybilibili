package social

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/abstraction"
	"mybilibili/pkg/auth"
)

type fakeCacheStore struct {
	data map[string][]byte
}

func newFakeCache() *fakeCacheStore {
	return &fakeCacheStore{data: map[string][]byte{}}
}

func (c *fakeCacheStore) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := c.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return v, nil
}
func (c *fakeCacheStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.data[key] = value
	return nil
}
func (c *fakeCacheStore) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}
func (c *fakeCacheStore) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (c *fakeCacheStore) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return false, nil
}
func (c *fakeCacheStore) Unlock(ctx context.Context, key string) error { return nil }
func (c *fakeCacheStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return 0, nil
}
func (c *fakeCacheStore) Close() error { return nil }

func newMockSearchHistoryHandler(cache abstraction.CacheStore) *SearchHistoryHandler {
	return NewSearchHistoryHandler(cache, auth.NewJWT("test-secret"))
}

func muxForSearch(h *SearchHistoryHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestSearchHistoryKey(t *testing.T) {
	assert.Equal(t, "search:history:100", searchHistoryKey(100))
}

func TestSearchHistoryHandler_handle_Unauthorized(t *testing.T) {
	h := newMockSearchHistoryHandler(newFakeCache())
	w := doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchHistoryHandler_handle_NoCache(t *testing.T) {
	h := newMockSearchHistoryHandler(nil)
	w := doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchHistoryHandler_handle_Get(t *testing.T) {
	cache := newFakeCache()
	h := newMockSearchHistoryHandler(cache)

	w := doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	cache.data["search:history:1"] = []byte(`["golang","docker"]`)
	w = doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	cache.data["search:history:1"] = []byte(`bad-json`)
	w = doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchHistoryHandler_handle_Post(t *testing.T) {
	cache := newFakeCache()
	h := newMockSearchHistoryHandler(cache)

	w := doReq(muxForSearch(h), "POST", "/api/v1/search/history", ``, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(muxForSearch(h), "POST", "/api/v1/search/history", `{"keyword":"go"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []byte(`["go"]`), cache.data["search:history:1"])

	// dedup + prepend
	w = doReq(muxForSearch(h), "POST", "/api/v1/search/history", `{"keyword":"go"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []byte(`["go"]`), cache.data["search:history:1"])

	w = doReq(muxForSearch(h), "POST", "/api/v1/search/history", `{"keyword":"rust"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []byte(`["rust","go"]`), cache.data["search:history:1"])
}

func TestSearchHistoryHandler_handle_DeleteAndMethod(t *testing.T) {
	cache := newFakeCache()
	h := newMockSearchHistoryHandler(cache)
	cache.data["search:history:1"] = []byte(`["go"]`)

	w := doReq(muxForSearch(h), "DELETE", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	_, ok := cache.data["search:history:1"]
	assert.False(t, ok)

	w = doReq(muxForSearch(h), "PUT", "/api/v1/search/history", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestSearchHistoryHandler_GetUserID_WithJWT(t *testing.T) {
	h := newMockSearchHistoryHandler(newFakeCache())
	token, err := h.jwt.Generate(42)
	require.NoError(t, err)

	w := doReq(muxForSearch(h), "GET", "/api/v1/search/history", "", map[string]string{"Authorization": "Bearer " + token})
	assert.Equal(t, http.StatusOK, w.Code)
}
