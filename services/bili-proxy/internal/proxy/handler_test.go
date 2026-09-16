package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth_200(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestHandleResolve_400_InvalidID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleResolve_404_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	// loadVideo: SELECT v.id ... WHERE v.id = $1 → no rows
	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "cid"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/resolve/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}

func TestHandleStream_400_InvalidID(t *testing.T) {
	h := NewHandler(nil, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/abc", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
}

func TestHandleStream_404_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	h := NewHandler(db, nil)
	mux := http.NewServeMux()
	h.Register(mux)

	// loadVideo: video doesn't exist → 404
	mock.ExpectQuery(`SELECT v.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "cid"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bili/stream/1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "body=%s", rec.Body.String())
}
