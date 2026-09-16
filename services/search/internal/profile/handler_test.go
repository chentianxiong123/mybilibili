package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	return NewHandler(newTestService(t))
}

func newMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestHandleProfileByPath_GetOrCreate(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/42", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	var p Profile
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &p))
	assert.Equal(t, int64(42), p.UserID)
}

func TestHandleProfileByPath_Init(t *testing.T) {
	h := newTestHandler(t)
	body := bytes.NewBufferString(`{"tags":["游戏","科技"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/42/init", body)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	var p Profile
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &p))
	assert.Equal(t, []string{"游戏", "科技"}, p.Tags)
}

func TestHandleProfileByPath_MissingUserID(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 400, rr.Code)
}

func TestHandleProfileByPath_InvalidUserID(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/abc", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 400, rr.Code)
}

func TestHandleProfileByPath_InitWrongMethod(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/42/init", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 405, rr.Code)
}

func TestHandleProfileByPath_WrongMethod(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/profile/42", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 405, rr.Code)
}

func TestHandleRecord_Watch(t *testing.T) {
	h := newTestHandler(t)
	body := bytes.NewBufferString(`{"categoryId":5,"tags":["电竞"],"durationSeconds":120}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/watch", body)
	req.Header.Set("X-User-Id", "42")
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 200, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")

	p, err := h.svc.Get(req.Context(), 42)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.WatchCount)
}

func TestHandleRecord_LikeAndCollect(t *testing.T) {
	h := newTestHandler(t)
	for _, action := range []string{"like", "collect"} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/"+action,
			bytes.NewBufferString(`{"categoryId":3,"tags":["音乐"]}`))
		req.Header.Set("X-User-Id", "7")
		rr := httptest.NewRecorder()
		newMux(h).ServeHTTP(rr, req)
		assert.Equal(t, 200, rr.Code)
	}

	p, err := h.svc.Get(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.LikeCount)
	assert.Equal(t, int64(1), p.CollectCount)
}

func TestHandleRecord_UnknownAction(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/frobnicate", nil)
	req.Header.Set("X-User-Id", "1")
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 400, rr.Code)
}

func TestHandleRecord_WrongMethod(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/record/watch", nil)
	rr := httptest.NewRecorder()
	newMux(h).ServeHTTP(rr, req)

	assert.Equal(t, 405, rr.Code)
}
