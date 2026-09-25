package favorite

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doReq(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func muxForFav(h *FavoriteHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestHandleFavorites_List_QueryError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT f.id, f.name`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFavorites_CreateFolder_InvalidBody(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites", `bad`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleFavorites_CreateFolder_Error(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`INSERT INTO favorite_folders`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites", `{"name":"技术"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleByID_EmptyPath(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleByID_Put(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectExec(`UPDATE favorite_folders SET name = \$1, updated_at = NOW\(\) WHERE id = \$2 AND user_id = \$3`).
		WithArgs("newname", int64(5), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5", `{"name":"newname"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_Put_EmptyName(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5", `{"name":"  "}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleByID_Put_Error(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectExec(`UPDATE favorite_folders SET name`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5", `{"name":"x"}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleByID_Delete(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(5), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForFav(h), "DELETE", "/api/v1/favorites/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_Delete_Error(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "DELETE", "/api/v1/favorites/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleByID_MethodNotAllowed(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleFolderManuscripts_NotFound(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleFolderManuscripts_InvalidFolderID(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/abc/manuscripts", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleFolderManuscripts_NoUser(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{}`, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleFolderManuscripts_FolderNotOwned(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{"manuscript_id":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleFolderManuscripts_Delete(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id = \$1 AND manuscript_id = \$2`).
		WithArgs(int64(5), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForFav(h), "DELETE", "/api/v1/favorites/5/manuscripts/9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFolderManuscripts_Delete_InvalidMSID(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForFav(h), "DELETE", "/api/v1/favorites/5/manuscripts/abc", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleFolderManuscripts_MethodNotAllowed(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5/manuscripts", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleFolderManuscripts_Post_MissingMSID(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleFolderManuscripts_Post_InsertError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{"manuscript_id":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFolderManuscripts_Post_Success(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).WithArgs(int64(5), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/manuscripts", `{"manuscript_id":9}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFolderVideos_NotFound(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5/videos/extra", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleFolderVideos_InvalidFolderID(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/abc/videos", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleFolderVideos_NoUser(t *testing.T) {
	h, _ := newMockFavorite(t)
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5/videos", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleFolderVideos_FolderNotOwned(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5/videos", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleFolderVideos_Get(t *testing.T) {
	h, mock := newMockFavorite(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT ffv.manuscript_id`).
		WithArgs(int64(5), int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"manuscript_id", "created_at",
			"title", "description", "cover_url", "status", "review_status",
			"duration", "duration_seconds", "view_count", "like_count",
			"coin_count", "collect_count", "comment_count", "share_count",
			"category_id", "upload_time",
			"user_id", "nickname", "avatar",
		}).AddRow(
			int64(9), now,
			"title", "", "", int32(3), int32(1),
			"", int64(0), int64(0), int64(0),
			int64(0), int64(0), int64(0), int64(0),
			int64(0), now,
			int64(1), "u", "",
		))
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5/videos", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFolderVideos_Get_QueryError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT ffv.manuscript_id, m.title`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/5/videos", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFolderVideos_Put(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos \(folder_id, manuscript_id\)`).WithArgs(int64(5), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos \(folder_id, manuscript_id\)`).WithArgs(int64(5), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5/videos", `{"manuscript_ids":[9,10]}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFolderVideos_Put_BeginError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin().WillReturnError(errors.New("tx fail"))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5/videos", `{"manuscript_ids":[9]}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFolderVideos_Put_DeleteError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id = \$1`).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5/videos", `{"manuscript_ids":[9]}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFolderVideos_Put_InsertError(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos \(folder_id, manuscript_id\)`).WithArgs(int64(5), int64(9)).WillReturnError(errors.New("boom"))
	w := doReq(muxForFav(h), "PUT", "/api/v1/favorites/5/videos", `{"manuscript_ids":[9]}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleFolderVideos_MethodNotAllowed(t *testing.T) {
	h, mock := newMockFavorite(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders`).WithArgs(int64(5), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForFav(h), "POST", "/api/v1/favorites/5/videos", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleCheck(t *testing.T) {
	h, mock := newMockFavorite(t)

	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/check", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/check", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folder_videos ffv`).WithArgs(int64(1), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/check?manuscript_id=9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folder_videos ffv`).WithArgs(int64(1), int64(9)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/check?manuscriptId=9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFlatList(t *testing.T) {
	h, mock := newMockFavorite(t)
	now := time.Now()

	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/list", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mock.ExpectQuery(`SELECT ffv.manuscript_id, ff.name, ffv.created_at`).
		WithArgs(int64(1), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "name", "created_at"}).AddRow(9, "folder", now))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT ffv.manuscript_id, ff.name`).WillReturnError(errors.New("boom"))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleManuscriptFolders(t *testing.T) {
	h, mock := newMockFavorite(t)

	w := doReq(muxForFav(h), "GET", "/api/v1/favorites/manuscript/abc", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectQuery(`SELECT ff.id, ff.name FROM favorite_folders ff`).
		WithArgs(int64(0), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "folder"))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/manuscript/9", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`SELECT ff.id, ff.name FROM favorite_folders ff`).WillReturnError(errors.New("boom"))
	w = doReq(muxForFav(h), "GET", "/api/v1/favorites/manuscript/9", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
