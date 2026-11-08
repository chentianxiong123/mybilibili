package manuscript

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/pb"
)

// ---- handleShare: full success path ----

func TestHandleShare_SuccessWithMock(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectExec(`INSERT INTO user_interactions`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET share_count`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/share", strings.NewReader(`{"channel":"wechat"}`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleShare(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleVideoFavoriteFolders: db error branch ----

func TestHandleVideoFavoriteFolders_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT .+ FROM favorite_folders`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/1/favorite/folders", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleVideoFavoriteFolders(w, req)
	// db error -> ErrInternal -> 500
	assert.Equal(t, 500, w.Code)
}

// ---- handleFavoriteFolderVideos: forbidden branch ----

func TestHandleFavoriteFolderVideos_Forbidden(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM favorite_folders WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/favorite/folders/1/videos", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderVideos(w, req)
	assert.Equal(t, 403, w.Code)
}

// ---- handleManuscriptList: service error branch ----

func TestHandleManuscriptList_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE user_id`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/me/list", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleManuscriptList(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleMySubtitles: success path with rows ----

func TestHandleMySubtitles_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"m_id", "m_title", "m_cover", "v_id", "v_order", "v_title", "v_dur", "v_has_sub", "v_stage", "v_status",
		}).AddRow(int64(1), "ms", "c.jpg", int64(7), int64(0), "P1", int64(60), int64(1), "DONE", int64(5)))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/me/subtitles", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleMySubtitles(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUnpublishManuscript: success path (service-level) ----

func TestHandleUnpublishManuscript_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// svc.UnpublishManuscript -> repo.UpdateStatus
	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/unpublish", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUnpublishManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUnpublishManuscript_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/unpublish", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUnpublishManuscript(w, req)
	assert.Equal(t, 404, w.Code)
}

// ---- handlePublishManuscript: success path (service-level) ----

func TestHandlePublishManuscript_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/publish", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handlePublishManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleDeleteManuscript: success path (service-level) ----

func TestHandleDeleteManuscript_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// svc.DeleteManuscript -> repo.Delete
	mock.ExpectExec(`DELETE FROM manuscripts`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUpdateManuscript: owner mismatch branch ----

func TestHandleUpdateManuscript_NotOwner(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(99)))

	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", strings.NewReader(`{"title":"new"}`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 403, w.Code)
}

// ---- handleMyCollections: service error branch ----

func TestHandleMyCollections_ServiceError(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT target_id FROM user_interactions`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/user/collections", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	m.h.handleMyCollections(w, req)
	// GetInteractionIDs error is swallowed -> empty list, 200
	assert.Equal(t, 200, w.Code)
}

// ---- service-level: ListCategories error ----

func TestService_ListCategories_Error(t *testing.T) {
	svc, msMock, _ := newService(t)

	msMock.ExpectQuery(`SELECT .+ FROM categories`).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.ListCategories(t.Context(), &pb.ListCategoriesRequest{})
	require.Error(t, err)
}

// ---- service-level: DeleteManuscript error path ----

func TestService_DeleteManuscript_Error(t *testing.T) {
	svc, msMock, _ := newService(t)

	msMock.ExpectExec(`DELETE FROM manuscripts`).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.DeleteManuscript(t.Context(), &pb.DeleteManuscriptRequest{Id: 1, UserId: 10})
	require.Error(t, err)
}
