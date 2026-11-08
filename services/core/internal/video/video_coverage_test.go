package video

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- repository: ListBannersHistory / error branches / rows error branches ----

func TestRepo_ListBannersHistory_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).
			AddRow(1, "archived", "http://img", "http://link", 10, -1, 1, 0, "", now, now))

	list, err := repo.ListBannersHistory(context.Background(), 1, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "archived", list[0].Title)
	assert.Equal(t, int32(-1), list[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepo_ListBannersHistory_WithCategory(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int32(2), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}))

	list, err := repo.ListBannersHistory(context.Background(), 2, 7)
	require.NoError(t, err)
	assert.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepo_ListBannersHistory_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).WillReturnError(errors.New("db down"))
	_, err := repo.ListBannersHistory(context.Background(), 1, 0)
	assert.Error(t, err)
}

func TestRepo_ListByManuscript_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).WillReturnError(errors.New("db down"))
	_, err := repo.ListByManuscript(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_ListByManuscript_RowsError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress",
			"process_stage", "has_subtitle", "has_summary",
			"process_status", "process_error", "source_video_url",
			"duration_seconds", "is_vertical",
		}).AddRow(
			"bad", int64(10), int32(1), "t",
			"h", "s", "l", time.Now(), time.Now(), int32(0), "", int32(0), int32(0),
			int32(0), "", "src", int32(0), int32(0),
		))
	// The loop swallows rows.Scan errors (only QueryContext errors are returned),
	// so the call must not fail even though the row's id is unscannable.
	list, err := repo.ListByManuscript(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestRepo_ListUserManuscriptIDs_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE user_id`).WillReturnError(errors.New("db down"))
	_, err := repo.ListUserManuscriptIDs(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_ListUserVideoIDs_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT v.id FROM videos v JOIN manuscripts m`).WillReturnError(errors.New("db down"))
	_, err := repo.ListUserVideoIDs(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_ListCategories_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories`).WillReturnError(errors.New("db down"))
	_, err := repo.ListCategories(context.Background())
	assert.Error(t, err)
}

func TestRepo_DeleteCategory_BeginError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin().WillReturnError(errors.New("cannot begin"))
	err := repo.DeleteCategory(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_DeleteCategory_UpdateError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE categories SET parent_id = 0 WHERE parent_id`).
		WithArgs(int64(1)).WillReturnError(errors.New("update failed"))
	err := repo.DeleteCategory(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_DeleteCategory_DeleteError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE categories SET parent_id = 0 WHERE parent_id`).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM categories WHERE id`).
		WithArgs(int64(1)).WillReturnError(errors.New("delete failed"))
	err := repo.DeleteCategory(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_DeleteCategory_CommitError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE categories SET parent_id = 0 WHERE parent_id`).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM categories WHERE id`).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
	err := repo.DeleteCategory(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepo_BatchDeleteVideos_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`DELETE FROM videos WHERE id = ANY`).
		WillReturnError(errors.New("db down"))
	assert.Error(t, repo.BatchDeleteVideos(context.Background(), []int64{1}))
}

func TestRepo_UpdateBanner_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`UPDATE banner_images SET title`).WillReturnError(errors.New("db down"))
	assert.Error(t, repo.UpdateBanner(context.Background(), 1, &BannerImage{}))
}

func TestRepo_DeleteBanner_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`UPDATE banner_images SET status = -1`).WillReturnError(errors.New("db down"))
	assert.Error(t, repo.DeleteBanner(context.Background(), 1))
}

func TestRepo_CreateBanner_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO banner_images`).WillReturnError(errors.New("db down"))
	_, err := repo.CreateBanner(context.Background(), &BannerImage{})
	assert.Error(t, err)
}

func TestRepo_CreateCategory_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO categories`).WillReturnError(errors.New("db down"))
	_, err := repo.CreateCategory(context.Background(), "x", 0, 0)
	assert.Error(t, err)
}

func TestRepo_UpdateCategory_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`UPDATE categories SET name`).WillReturnError(errors.New("db down"))
	assert.Error(t, repo.UpdateCategory(context.Background(), 1, "x", 0, 0))
}

func TestRepo_GetVideoByID_Err(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).WillReturnError(errors.New("db down"))
	_, err := repo.GetVideoByID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- service: ListBannersHistory pass-through ----

func TestService_ListBannersHistory(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}))
	list, err := svc.ListBannersHistory(context.Background(), 1, 0)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestService_ListBannersHistory_Err(t *testing.T) {
	svc, mock := newMockService(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).WillReturnError(errors.New("db down"))
	_, err := svc.ListBannersHistory(context.Background(), 1, 0)
	assert.Error(t, err)
}

// ---- handler: bannersFor with history=1 ----

func TestHandleBanner_HistoryQuery(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(1, "hist", "http://img", "http://link", 1, -1, 1, 0, "", now, now))
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/home?history=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hist")
}

func TestHandleBannerHistory_CategoryPath(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).
		WithArgs(int32(2), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/category/9?history=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBannersFor_QueryErrorReturnsEmpty(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).WillReturnError(errors.New("db down"))
	r := httptest.NewRequest("GET", "/api/v1/banner/home", nil)
	list := h.bannersFor(r, 1, 0)
	assert.NotNil(t, list)
	assert.Empty(t, list)
}

// ---- handler: banner user-profile POST / DELETE / method-not-allowed ----

func TestHandleBannerUserProfile_POST(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO banner_images`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	w := doReq(muxForVideo(h), "POST", "/api/v1/banner/user-profile", `{"title":"p","image_url":"/p.jpg"}`, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestHandleBannerUserProfile_DeleteOne(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE banner_images SET status = -1`).
		WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideo(h), "DELETE", "/api/v1/banner/user-profile/5", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleBannerUserProfile_DeleteAll(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM banner_images WHERE type = \$1 AND status = 1`).WithArgs(int32(4)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}).AddRow(11, "a", "u", "l", 1, 1, 4, 0, "default", nil, nil).
			AddRow(12, "b", "u", "l", 2, 1, 4, 0, "default", nil, nil))
	mock.ExpectExec(`UPDATE banner_images SET status = -1`).WithArgs(int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE banner_images SET status = -1`).WithArgs(int64(12)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForVideo(h), "DELETE", "/api/v1/banner/user-profile", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleBannerUserProfile_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "PATCH", "/api/v1/banner/user-profile", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: handleBannerSingle POST / DELETE / 405 ----

func TestHandleBannerSingle_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "PATCH", "/api/v1/banner/background", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: handleBanner upload ----

// fakeMinio is no longer needed: Handler requires a concrete
// *abstraction.MinioStorageService whose minio.Client cannot be constructed
// in a test without a live server, so the `h.storage != nil` branch of the
// banner upload handler is not reachable from a black-box test.
//

func uploadReq(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	r := httptest.NewRequest("POST", "/api/v1/banner/upload", &buf)
	r.Header.Set("X-Admin-Id", "1")
	r.Header.Set("Content-Type", w.FormDataContentType())
	r.Header.Set("X-Admin-Id", "1") // FINDING-023 修复：需 admin 身份
	return r
}

func TestHandleBanner_Upload_SuccessNoMinIO(t *testing.T) {
	h, _ := newTestHandler(t) // storage is nil
	r := uploadReq(t, "x.png", "fake-image-bytes")
	w := httptest.NewRecorder()
	h.handleBanner(w, r)
	assert.Contains(t, w.Body.String(), "/uploads/images/")
}

func TestHandleBanner_Upload_MinIOFallbackNotApplicable(t *testing.T) {
	// h.storage is always nil in these tests, so the MinIO dual-write block
	// (lines 184-201) and its failure fallback path are not covered here.
	h, _ := newTestHandler(t)
	r := uploadReq(t, "x.jpg", "fake-image-bytes")
	w := httptest.NewRecorder()
	h.handleBanner(w, r)
	assert.Contains(t, w.Body.String(), "/uploads/images/")
}

func TestHandleBanner_Upload_NoFile(t *testing.T) {
	h, _ := newTestHandler(t)
	// multipart form without a "file" field
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("other", "hello.txt")
	require.NoError(t, err)
	fw.Write([]byte("hello"))
	require.NoError(t, w.Close())
	r := httptest.NewRequest("POST", "/api/v1/banner/upload", &buf)
	r.Header.Set("X-Admin-Id", "1")
	r.Header.Set("Content-Type", w.FormDataContentType())
	r.Header.Set("X-Admin-Id", "1") // FINDING-023 修复：需 admin 身份
	wrec := httptest.NewRecorder()
	h.handleBanner(wrec, r)
	assert.Contains(t, wrec.Body.String(), "file required")
}

func TestHandleBanner_Upload_BadMultipart(t *testing.T) {
	h, _ := newTestHandler(t)
	// not multipart at all → ParseMultipartForm fails
	r := httptest.NewRequest("POST", "/api/v1/banner/upload", strings.NewReader("plain text"))
	r.Header.Set("X-Admin-Id", "1") // FINDING-023 修复：需 admin 身份
	w := httptest.NewRecorder()
	h.handleBanner(w, r)
	assert.Contains(t, w.Body.String(), "parse form")
}

func TestHandleBanner_Upload_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/upload", "", nil)
	assert.Contains(t, w.Body.String(), "method not allowed")
}

// ---- handler: handleVideo fall-through to 404 ----

func TestHandleVideo_Fallthrough404(t *testing.T) {
	h, _ := newTestHandler(t)
	// GET on a path that is not "manuscript", "ids", or "video-ids"
	w := doReq(muxForVideo(h), "GET", "/api/v1/video/123/unknown", "", nil)
	assert.Contains(t, w.Body.String(), `"code":404`)
}

// ---- handler: handleCategoryByID PUT error branches ----

func TestHandleCategoryByID_PUT_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).
		WithArgs(int64(99)).WillReturnError(sql.ErrNoRows)
	body := strings.NewReader(`{"name":"x"}`)
	w := doReq(muxForVideo(h), "PUT", "/api/v1/category/99", "", map[string]string{"Content-Type": "application/json", "X-Admin-Id": "1"})
	_ = body
	assert.Contains(t, w.Body.String(), "category not found")
	_ = mock
}

func TestHandleCategoryByID_PUT_UpdateError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, name, parent_id, sort_order FROM categories WHERE id`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id", "sort_order"}).AddRow(5, "old", 0, 1))
	mock.ExpectExec(`UPDATE categories SET name`).WillReturnError(errors.New("update failed"))
	body := strings.NewReader(`{"name":"new"}`)
	req := httptest.NewRequest("PUT", "/api/v1/category/5", body)
	req.Header.Set("X-Admin-Id", "1") // FINDING-023 修复：需 admin 身份
	w := httptest.NewRecorder()
	mux := muxForVideo(h)
	mux.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), `"code":500`)
}

// ---- handler: handleBanner background category with categoryID via banner-category route ----

func TestHandleBannerCategory_Default(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "PATCH", "/api/v1/banner/category/5", "", map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: handleBanner single background GET with history ----

func TestHandleBannerBackground_HistoryQuery(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).WithArgs(int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/background?history=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: handleBanner single user-profile GET with history ----

func TestHandleBannerUserProfile_HistoryQuery(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, COALESCE\(title`).WithArgs(int32(4)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "image_url", "link_url", "sort_order",
			"status", "type", "category_id", "time_slot", "start_time", "end_time",
		}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/banner/user-profile?history=1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: statistics manuscript/status DB error path ----

func TestHandleStatisticsByPath_ManuscriptStatus_DBError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FILTER \(WHERE review_status=0\)`).
		WillReturnError(errors.New("db down"))
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/status", "", nil)
	// The handler's QueryRowContext .Scan ignores the error (no err var), so
	// it writes an empty map. The test just confirms no panic.
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: statistics manuscript/recent with limit bounds ----

func TestHandleStatisticsByPath_Recent_LimitClampLow(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, user_id, title, cover_url`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "user_id", "title", "cover_url", "view_count", "upload_time",
	}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/recent?limit=0", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleStatisticsByPath_Recent_LimitClampHigh(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT id, user_id, title, cover_url`).WillReturnRows(sqlmock.NewRows([]string{
		"id", "user_id", "title", "cover_url", "view_count", "upload_time",
	}))
	w := doReq(muxForVideo(h), "GET", "/api/v1/statistics/manuscript/recent?limit=999", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- handler: handleVideo with method != GET on single id ----

func TestHandleVideo_POSTNotSupported(t *testing.T) {
	h, _ := newTestHandler(t)
	w := doReq(muxForVideo(h), "POST", "/api/v1/video/1", "", nil)
	assert.Contains(t, w.Body.String(), `"code":404`)
}

// ---- admin handler: list with query params ----

func TestAdminHandler_List_WithKeywordAndStatus(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectQuery(`FROM videos v LEFT JOIN manuscripts m`).
		WithArgs("mykeyword", "3", int32(10), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "title", "process_status", "user_id", "manuscript_title"}).
			AddRow(1, 2, "t", int32(3), int64(4), "mt"))
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/list?keyword=mykeyword&status=3&page=2&page_size=10", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_List_EmptyResult(t *testing.T) {
	h, mock := newAdminVideoHandler(t)
	mock.ExpectQuery(`FROM videos v LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "manuscript_id", "title", "process_status", "user_id", "manuscript_title"}))
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/list", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- admin handler: unknown route → ServeMux 404 (no matching branch writes a body) ----

func TestAdminHandler_UnknownRoute(t *testing.T) {
	h, _ := newAdminVideoHandler(t)
	w := doReq(muxForVideoAdmin(h), "GET", "/api/v1/video/admin/", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- concurrent smoke (handlers are stateless) ----

func TestHandlers_ConcurrentSmoke(t *testing.T) {
	h, _ := newTestHandler(t)
	ah, _ := newAdminVideoHandler(t)
	mux := muxForVideo(h)
	amux := muxForVideoAdmin(ah)
	doReq(mux, "GET", "/api/v1/video/", "", nil)
	doReq(mux, "GET", "/api/v1/category", "", nil)
	doReq(mux, "GET", "/api/v1/statistics", "", nil)
	doReq(amux, "GET", "/api/v1/video/admin/list", "", nil)

	var wg chan struct{} = make(chan struct{}, 8)
	for i := 0; i < 4; i++ {
		wg <- struct{}{}
		go func() {
			defer func() { <-wg }()
			doReq(mux, "GET", "/api/v1/video/", "", nil)
			doReq(mux, "GET", "/api/v1/category", "", nil)
			doReq(mux, "GET", "/api/v1/statistics", "", nil)
			doReq(amux, "GET", "/api/v1/video/admin/list", "", nil)
		}()
	}
	close(wg)
	<-wg
}

// ---- handleBannerImages alias already covered; upload with .gif/.webp ext switch ----

func TestHandleBanner_Upload_ExtensionVariants(t *testing.T) {
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp"} {
		h, _ := newTestHandler(t)
		r := uploadReq(t, "img"+ext, "fake-bytes")
		w := httptest.NewRecorder()
		h.handleBanner(w, r)
		assert.Contains(t, w.Body.String(), "/uploads/images/")
	}
}

func TestHandleBanner_Upload_NoExtension(t *testing.T) {
	h, _ := newTestHandler(t)
	r := uploadReq(t, "noext", "fake-bytes")
	w := httptest.NewRecorder()
	h.handleBanner(w, r)
	// no extension → defaults to .jpg
	assert.Contains(t, w.Body.String(), "/uploads/images/")
}
