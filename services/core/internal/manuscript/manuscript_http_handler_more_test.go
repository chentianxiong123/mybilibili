package manuscript

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ---- handleMyCollections / handleMyLikes / handleMyCoins ----

func TestHandleMyCollections_Success(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// interaction service repo: user_interactions query for COLLECT ids
	m.interact.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(int64(5)))

	// manuscriptsByIDs -> GetManuscript per id
	m.db.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(5, "t5", "d5", "c5.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))
	m.db.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(int64(1), "cat", int64(0), int32(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/collections", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	m.h.handleMyCollections(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleMyLikes_Success(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}).AddRow(int64(7)))

	m.db.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(7, "t7", "d7", "c7.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))
	m.db.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/likes", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	m.h.handleMyLikes(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleMyCoins_Success(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"target_id"}))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/coins", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	m.h.handleMyCoins(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleMyCollections_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/collections", nil)
	w := httptest.NewRecorder()
	h.handleMyCollections(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHandleMyLikes_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/likes", nil)
	w := httptest.NewRecorder()
	h.handleMyLikes(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestHandleMyCoins_Unauthorized(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/coins", nil)
	w := httptest.NewRecorder()
	h.handleMyCoins(w, req)
	assert.Equal(t, 401, w.Code)
}

// ---- handleHot: seed clamping branches ----

func TestHandleHot_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptColsWithExposure))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/hot", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleHot(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleHot_SeedClamp(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptColsWithExposure))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/hot?seed=5.5&offset=30", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleHot(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleHot_SeedNegativeClamp(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptColsWithExposure))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/hot?seed=-5.5", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleHot(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleHot_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/hot", nil)
	w := httptest.NewRecorder()
	h.handleHot(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleRecommended: refresh_time parse branches ----

func TestHandleRecommended_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptColsWithExposure).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local", int64(0)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/recommended?refresh_time=12345", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleRecommended(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleRecommended_InvalidRefreshTime(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows(manuscriptColsWithExposure))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/recommended?refresh_time=abc", nil)
	w := httptest.NewRecorder()
	h.handleRecommended(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleRecommended_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`WITH _seed AS .+ SELECT .+ FROM manuscripts`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/recommended", nil)
	w := httptest.NewRecorder()
	h.handleRecommended(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleManuscriptDetail: id<=0 branch + video fallback ----

func TestHandleManuscriptDetail_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/detail/0", nil)
	w := httptest.NewRecorder()
	h.handleManuscriptDetail(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHandleManuscriptDetail_VideoFallback(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// manuscript lookup fails -> try videos table
	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE id = \$1`).WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT manuscript_id FROM videos WHERE id = \$1`).WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id"}).AddRow(int64(1)))

	// GetManuscriptWithVideos(1)
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))
	mock.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(int64(1), "cat", int64(0), int32(0)))
	mock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress", "process_stage",
			"has_subtitle", "has_summary", "process_status", "process_error",
			"source_video_url", "duration_seconds", "is_vertical",
		}))
	mock.ExpectExec(`UPDATE manuscripts SET view_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/detail/42", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "42")
	w := httptest.NewRecorder()
	h.handleManuscriptDetail(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleManuscriptDetail_VideoFallback_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE id = \$1`).WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT manuscript_id FROM videos WHERE id = \$1`).WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"manuscript_id"}))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/detail/42", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "42")
	w := httptest.NewRecorder()
	h.handleManuscriptDetail(w, req)
	assert.Equal(t, 404, w.Code)
}

// ---- handleCategory / handleUserManuscripts / handleUserSearch: service error branches ----

func TestHandleCategory_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/category/1", nil)
	w := httptest.NewRecorder()
	h.handleCategory(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleCategory_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = ANY`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/category/1", nil)
	w := httptest.NewRecorder()
	h.handleCategory(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHandleUserManuscripts_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/10/list", nil)
	w := httptest.NewRecorder()
	h.handleUserManuscripts(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUserManuscripts_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/10/list", nil)
	w := httptest.NewRecorder()
	h.handleUserManuscripts(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHandleUserSearch_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/10/search?keyword=go&sort=latest", nil)
	w := httptest.NewRecorder()
	h.handleUserSearch(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUserSearch_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/10/search", nil)
	w := httptest.NewRecorder()
	h.handleUserSearch(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUserManuscriptStats: id<=0 branch ----

func TestHandleUserManuscriptStats_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/0/stats", nil)
	w := httptest.NewRecorder()
	h.handleUserManuscriptStats(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- requireCategory: err / not-exist / zero-id branches ----

func TestRequireCategory_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(1)).WillReturnError(sql.ErrConnDone)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/list", nil)
	w := httptest.NewRecorder()
	ok := h.requireCategory(req.Context(), w, 1)
	assert.False(t, ok)
	assert.Equal(t, 400, w.Code)
}

func TestRequireCategory_NotExist(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	req := httptest.NewRequest("GET", "/api/v1/manuscript/list", nil)
	w := httptest.NewRecorder()
	ok := h.requireCategory(req.Context(), w, 1)
	assert.False(t, ok)
	assert.Equal(t, 400, w.Code)
}

func TestRequireCategory_ZeroID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/list", nil)
	w := httptest.NewRecorder()
	ok := h.requireCategory(req.Context(), w, 0)
	assert.False(t, ok)
	assert.Equal(t, 400, w.Code)
}

// ---- handleShare: success + method-not-allowed ----

func TestHandleShare_Success(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// ShareManuscript upserts into user_interactions (interaction db)
	m.interact.ExpectExec(`INSERT INTO user_interactions`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/share", strings.NewReader(`{"channel":"wechat"}`))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	m.h.handleShare(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleShare_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/1/share", nil)
	w := httptest.NewRecorder()
	h.handleShare(w, req)
	assert.Equal(t, 405, w.Code)
}

// ---- handleInteractionStatus: success + invalid id ----

func TestHandleInteractionStatus_Success(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	// GetInteractionStatus: 3 HasInteraction checks (LIKE, COLLECT, SHARE)
	for i := 0; i < 3; i++ {
		m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	}
	// GetManuscript on main db
	m.db.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))
	m.db.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(int64(1), "cat", int64(0), int32(0)))
	// HasInteraction for COIN (uid>0)
	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// Set path values the way handleManuscriptRoute would
	req := httptest.NewRequest("GET", "/api/v1/manuscript/1/status", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleInteractionStatus(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleInteractionStatus_InvalidID(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/0/status", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "0")
	w := httptest.NewRecorder()
	h.handleInteractionStatus(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadSessionByID: DELETE + 405 branches ----

func TestHandleUploadSessionByID_Delete(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`DELETE FROM upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/upload-session/abc", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUploadSessionByID_DeleteError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`DELETE FROM upload_sessions`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/upload-session/abc", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHandleUploadSessionByID_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/upload-session/abc", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 405, w.Code)
}

// ---- handleFixDurations ----

func TestHandleFixDurations_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`UPDATE manuscripts m SET duration_seconds`).WillReturnResult(sqlmock.NewResult(0, 10))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/fix-durations", nil)
	w := httptest.NewRecorder()
	h.handleFixDurations(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleFixDurations_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/fix-durations", nil)
	w := httptest.NewRecorder()
	h.handleFixDurations(w, req)
	assert.Equal(t, 405, w.Code)
}

// ---- handleInternalByPath: take-down + 404 ----

func TestHandleInternalByPath_Takedown(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`UPDATE manuscripts SET status = -1`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("PUT", "/api/v1/manuscript/internal/5/take-down", nil)
	w := httptest.NewRecorder()
	h.handleInternalByPath(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleInternalByPath_NotFound(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/internal/5/other", nil)
	w := httptest.NewRecorder()
	h.handleInternalByPath(w, req)
	assert.Equal(t, 404, w.Code)
}

// ---- handleUploadSession: client_id path + insert error + category path ----

func TestHandleUploadSession_ClientID_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(`{"client_id":"cid-1"}`))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "cid-1", resp["upload_id"])
}

func TestHandleUploadSession_InsertError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(`{"client_id":"cid-2"}`))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHandleUploadSession_WithCategoryID(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session",
		strings.NewReader(`{"client_id":"cid-3","title":"t","category_id":2}`))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleManuscriptList / meStats: success ----

func TestHandleManuscriptList_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE user_id`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols).
			AddRow(1, "t", "d", "c.jpg", int64(10), int64(1),
				int64(0), int64(0), int64(0), int64(0), int64(0),
				int64(0), int64(0), int32(3), int32(1), "",
				sql.NullTime{}, int64(0), sql.NullTime{}, sql.NullTime{}, "", int32(0), "local"))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/me/list", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleManuscriptList(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleMyManuscriptStats_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	for _, q := range []string{
		`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 0`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 1`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 3`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = 4`,
		`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id = \$1 AND status = -1`,
	} {
		mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	}

	req := httptest.NewRequest("GET", "/api/v1/manuscript/me/stats", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleMyManuscriptStats(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadChunk: missing uploadId + full multipart success ----

func TestHandleUploadChunk_MissingUploadID(t *testing.T) {
	h, _ := newHTTPHandler(t)

	boundary := "testBoundary"
	body := buildMultipart(boundary, map[string]string{"uploadId": ""})
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestHandleUploadChunk_Success(t *testing.T) {
	t.Setenv("MYBILIBILI_UPLOAD_DIR", t.TempDir())
	h, mock := newHTTPHandler(t)

	// upload session owner check
	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE upload_sessions SET uploaded_chunks`).WillReturnResult(sqlmock.NewResult(0, 1))

	boundary := "testBoundary"
	body := buildMultipartFile(boundary, map[string]string{
		"uploadId": "up-1", "partIndex": "0", "chunkIndex": "0",
	}, "chunk-0", "hello-bytes")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUploadChunk_NotOwner(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	boundary := "testBoundary"
	body := buildMultipartFile(boundary, map[string]string{
		"uploadId": "up-1",
	}, "chunk-0", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 403, w.Code)
}

// ---- multipart helpers ----

func buildMultipart(boundary string, fields map[string]string) string {
	var b strings.Builder
	for k, v := range fields {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString(`Content-Disposition: form-data; name="` + k + `"` + "\r\n\r\n")
		b.WriteString(v + "\r\n")
	}
	b.WriteString("--" + boundary + "--\r\n")
	return b.String()
}

func buildMultipartFile(boundary string, fields map[string]string, filename, fileContent string) string {
	var b strings.Builder
	for k, v := range fields {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString(`Content-Disposition: form-data; name="` + k + `"` + "\r\n\r\n")
		b.WriteString(v + "\r\n")
	}
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString(`Content-Disposition: form-data; name="file"; filename="` + filename + `"` + "\r\n")
	b.WriteString("Content-Type: application/octet-stream\r\n\r\n")
	b.WriteString(fileContent + "\r\n")
	b.WriteString("--" + boundary + "--\r\n")
	return b.String()
}
