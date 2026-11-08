package profile

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"mybilibili/pkg/abstraction"
)

// handlerErrStore wraps a fakeDocumentStore so that specific operations
// can be forced to fail for error-path handler testing.
type handlerErrStore struct {
	fake *fakeDocumentStore

	queryErr  error
	insertErr error
	updateErr error
}

func (h *handlerErrStore) Insert(ctx context.Context, collection string, doc any) (string, error) {
	if h.insertErr != nil {
		return "", h.insertErr
	}
	return h.fake.Insert(ctx, collection, doc)
}

func (h *handlerErrStore) FindByID(ctx context.Context, collection, id string, result any) error {
	return h.fake.FindByID(ctx, collection, id, result)
}

func (h *handlerErrStore) Update(ctx context.Context, collection, id string, doc any) error {
	if h.updateErr != nil {
		return h.updateErr
	}
	return h.fake.Update(ctx, collection, id, doc)
}

func (h *handlerErrStore) Delete(ctx context.Context, collection, id string) error {
	return h.fake.Delete(ctx, collection, id)
}

func (h *handlerErrStore) Query(ctx context.Context, collection string, filter abstraction.QueryFilter, result any) error {
	if h.queryErr != nil {
		return h.queryErr
	}
	return h.fake.Query(ctx, collection, filter, result)
}

var _ abstraction.DocumentStore = (*handlerErrStore)(nil)

func TestHandler_InitServiceError(t *testing.T) {
	fake := newFakeDocumentStore()
	store := &handlerErrStore{fake: fake, queryErr: errors.New("init query down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	// GET profile -> GetOrCreate -> GetByUser fails -> 500
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/42", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 500, rr.Code)
	assert.Contains(t, rr.Body.String(), "init query down")

	// POST /init -> GetByUser fails -> FINDING-019 修复：错误传播，handler 返回 400
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/profile/42/init",
		bytes.NewBufferString(`{"tags":["a"]}`))
	req2.Header.Set("X-User-Id", "42")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	assert.Equal(t, 400, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "init query down")
}

func TestHandler_InitServiceUpdateError(t *testing.T) {
	fake := newFakeDocumentStore()
	fake.profiles[42] = &Profile{ID: "existing-42", UserID: 42, Tags: []string{"old"}}
	store := &handlerErrStore{fake: fake, updateErr: errors.New("init update down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/42/init",
		bytes.NewBufferString(`{"tags":["new"]}`))
	req.Header.Set("X-User-Id", "42")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Contains(t, rr.Body.String(), "init update down")
}

func TestHandler_GetOrCreateServiceError(t *testing.T) {
	fake := newFakeDocumentStore()
	store := &handlerErrStore{fake: fake, queryErr: errors.New("get-or-create down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/123", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 500, rr.Code)
	assert.Contains(t, rr.Body.String(), "get-or-create down")
}

func TestHandler_GetOrCreateInsertError(t *testing.T) {
	// Query succeeds (no profile) but Insert fails -> 500.
	fake := newFakeDocumentStore()
	store := &handlerErrStore{fake: fake, insertErr: errors.New("insert down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/77", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 500, rr.Code)
	assert.Contains(t, rr.Body.String(), "insert down")
}

func TestHandler_RecordServiceError(t *testing.T) {
	// Pre-create profile so RecordWatch reaches the final Upsert which fails.
	fake := newFakeDocumentStore()
	fake.profiles[1] = &Profile{ID: "id-1", UserID: 1}
	store := &handlerErrStore{fake: fake, updateErr: errors.New("record upsert down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/watch",
		bytes.NewBufferString(`{"categoryId":3,"tags":["t"],"durationSeconds":5}`))
	req.Header.Set("X-User-Id", "1")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
	assert.Contains(t, rr.Body.String(), "record upsert down")
}

func TestHandler_RecordServiceErrorGetOrCreate(t *testing.T) {
	// No pre-populated profile; Query fails so GetOrCreate errors in all
	// three record actions.
	fake := newFakeDocumentStore()
	store := &handlerErrStore{fake: fake, queryErr: errors.New("record get-or-create down")}
	h := NewHandler(NewService(NewRepository(store)))
	mux := http.NewServeMux()
	h.Register(mux)

	for _, action := range []string{"watch", "like", "collect"} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/"+action,
			bytes.NewBufferString(`{}`))
		req.Header.Set("X-User-Id", "2")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		assert.Equal(t, 400, rr.Code, "action %s", action)
		assert.Contains(t, rr.Body.String(), "record get-or-create down")
	}
}

func TestHandler_RecordMissingUserHeader(t *testing.T) {
	// No X-User-Id header -> RequireUser 401（FINDING-018 修复：不再为 user 0 建 profile）。
	fake := newFakeDocumentStore()
	h := NewHandler(NewService(NewRepository(fake)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/like",
		bytes.NewBufferString(`{}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
}

func TestHandler_RecordInvalidUserHeader(t *testing.T) {
	// Invalid X-User-Id header value -> RequireUser 401（FINDING-018 修复）。
	fake := newFakeDocumentStore()
	h := NewHandler(NewService(NewRepository(fake)))
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/record/collect",
		bytes.NewBufferString(`{}`))
	req.Header.Set("X-User-Id", "not-a-number")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, 401, rr.Code)
}
