package favorite

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockFavorite(t *testing.T) (*FavoriteHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewFavoriteHandler(db), mock
}

func TestHandleFavorites_401_NoUser(t *testing.T) {
	h, _ := newMockFavorite(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":401`)
}

func TestHandleFavorites_ListEmpty(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT f.id, f.name`).
		WithArgs(int64(1), false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "visible", "video_count", "cover"}))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":200`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFavorites_ListWithFolder(t *testing.T) {
	h, mock := newMockFavorite(t)
	createdAt := time.Now().UTC().Truncate(time.Second)
	mock.ExpectQuery(`SELECT f.id, f.name`).
		WithArgs(int64(1), false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "visible", "video_count", "cover"}).
			AddRow(int64(5), "我的收藏", createdAt, createdAt, int32(1), int64(12), "/covers/1/cover.webp"))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"我的收藏"`)
	assert.Contains(t, rec.Body.String(), `"visible":1`)
	assert.Contains(t, rec.Body.String(), `"video_count":12`)
	assert.Contains(t, rec.Body.String(), `"/covers/1/cover.webp"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 查看他人空间：仅返回公开收藏夹（visible=1）
func TestHandleFavorites_ListOthersVisibleOnly(t *testing.T) {
	h, mock := newMockFavorite(t)
	createdAt := time.Now().UTC().Truncate(time.Second)
	mock.ExpectQuery(`SELECT f.id, f.name`).
		WithArgs(int64(2), true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "visible", "video_count", "cover"}).
			AddRow(int64(5), "我的收藏", createdAt, createdAt, int32(1), int64(12), "/covers/1/cover.webp"))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/favorites?uid=2", nil)
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"visible":1`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFavorites_CreateFolder_EmptyName(t *testing.T) {
	h, _ := newMockFavorite(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", strings.NewReader(`{"name":"   "}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "name required")
}

func TestHandleFavorites_CreateFolder_Success(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`INSERT INTO favorite_folders`).
		WithArgs(int64(1), "技术", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", strings.NewReader(`{"name":"技术"}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":9`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFavorites_CreateFolder_Private(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`INSERT INTO favorite_folders`).
		WithArgs(int64(1), "私密夹", 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/favorites", strings.NewReader(`{"name":"私密夹","visible":0}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":10`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFavorites_MethodNotAllowed(t *testing.T) {
	h, _ := newMockFavorite(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/favorites", nil)
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleByID_InvalidID(t *testing.T) {
	h, _ := newMockFavorite(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/favorites/abc", strings.NewReader(`{"name":"x"}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid id")
}

func TestHandleByID_NoUser(t *testing.T) {
	h, _ := newMockFavorite(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/favorites/5", nil)
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
