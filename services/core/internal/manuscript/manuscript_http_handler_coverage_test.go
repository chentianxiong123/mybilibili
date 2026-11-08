package manuscript

import (
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ---- handleUploadCompleteWeb: early-return branches ----

func TestHandleUploadCompleteWeb_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-complete-web", nil)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHandleUploadCompleteWeb_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete-web", nil)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHandleUploadCompleteWeb_MissingUploadID(t *testing.T) {
	h, _ := newHTTPHandler(t)

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": ""}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete-web", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHandleUploadCompleteWeb_SessionNotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-x").
		WillReturnError(sql.ErrNoRows)

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-x"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete-web", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHandleUploadCompleteWeb_Forbidden(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-x").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-x"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete-web", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestHandleUploadCompleteWeb_NoParts(t *testing.T) {
	t.Setenv("MYBILIBILI_UPLOAD_DIR", t.TempDir())
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-np").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-np").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(0), "[]", "[]"))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-np"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete-web", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadComplete (session-complete route): error branches ----

func TestHandleUploadComplete_NotOwner(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-a").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-a/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-a")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 403, w.Code)
}

func TestHandleUploadComplete_NoPlayableSource(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-b").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-b").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(0), "[]", "[]")) // no url -> error

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-b/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-b")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleLike / handleCoin / handleCollect: error branches ----

func TestHandleLike_ServiceError(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// LIKE: HasInteraction check fails -> treated as not-liked, then AddInteraction fails
	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/like", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleLike(w, req)
	// errors are swallowed by the service; still 200
	assert.Equal(t, 200, w.Code)
}

func TestHandleCoin_InsufficientCoins(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// CoinManuscript fails: user has no coins
	m.interact.ExpectQuery(`SELECT coin_count FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/coin", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleCoin(w, req)
	// ErrResourceExhausted maps to 500 in this codebase's HTTP error mapping
	assert.Equal(t, 500, w.Code)
}

func TestHandleCollect_DeleteMethod(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// UncollectManuscript: RemoveInteraction + decrement paths
	m.interact.ExpectExec(`DELETE FROM user_interactions`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET collect_count`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/1/collect", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleCollect(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadChunk: not-POST + no-file branches ----

func TestHandleUploadChunk_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-chunk", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHandleUploadChunk_SessionNotOwner(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-9").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(42)))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{
		"uploadId": "up-9", "partIndex": "0", "chunkIndex": "0",
	}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 403, w.Code)
}

// ---- handleMySubtitles: db error branch ----

func TestHandleMySubtitles_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/me/subtitles", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleMySubtitles(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleFavoriteFolderVideos: sortOrder branches (earliest/play/publish) ----

func TestHandleFavoriteFolderVideos_SortOrderBranches(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// owner check
	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))

	// earliest
	mock.ExpectQuery(`SELECT ffv.manuscript_id, ffv.created_at`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "collect_time"}))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/favorite/folders/1/videos?sortOrder=earliest", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderVideos(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUpdateManuscript: multipart branch ----

func TestHandleUpdateManuscript_Multipart(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET title`).WillReturnResult(sqlmock.NewResult(0, 1))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{
		"title": "new title", "description": "new desc",
	}, "f", "x")
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleDeleteManuscript / handlePublishManuscript / handleUnpublishManuscript ----

func TestHandleDeleteManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHandlePublishManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/1/publish", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handlePublishManuscript(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHandleUnpublishManuscript_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/1/unpublish", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUnpublishManuscript(w, req)
	assert.Equal(t, 405, w.Code)
}

// helper to write files for disk-based upload tests
func touchParts(t *testing.T, dir string, part int, chunks []string) {
	t.Helper()
	for _, c := range chunks {
		p := filepath.Join(dir, fmt.Sprintf("%d", part), c)
		assert.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		assert.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}
}
