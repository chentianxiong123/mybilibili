package manuscript

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/pkg/pb"
)

// failingFlusher 实现 http.Flusher 但 Write 恒失败，用于覆盖 SSE 写失败路径。
type failingFlusher struct{}

func (failingFlusher) Header() http.Header         { return http.Header{} }
func (failingFlusher) WriteHeader(int)             {}
func (failingFlusher) Write([]byte) (int, error)   { return 0, fmt.Errorf("boom") }
func (failingFlusher) Flush()                      {}

// ---- handleStream: SSE 循环（事件 + ctx 取消） ----

func TestXC_HandleStream_SSELoop(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"video_id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
		}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/stream", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.handleStream(w, req)
		close(done)
	}()

	// 推送一个事件，等待循环处理后再取消 ctx
	h.Hub().Broadcast(ProgressEvt{
		VideoID: 1, ManuscriptID: 2, Title: "t", Stage: "stage", StageText: "txt",
		Progress: 50, Status: 1,
	})
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleStream 未在 ctx 取消后退出")
	}
	assert.Contains(t, w.Body.String(), "event: snapshot")
}

// ---- handleStream: 快照写失败立即返回 ----

func TestXC_HandleStream_WriteFailure(t *testing.T) {
	h, mock := newVideoProcessHandler(t)
	mock.ExpectQuery(`SELECT .+ FROM videos v JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"video_id", "manuscript_id", "title", "stage_text", "process_status", "process_progress",
		}))

	req := httptest.NewRequest("GET", "/api/v1/video/process/admin/stream", nil)
	h.handleStream(failingFlusher{}, req)
	// 无 panic、无挂起即可
}

// ---- handleVideoFavoriteFolderByID: 成功删除 + 非法请求 ----

func TestXC_HandleVideoFavoriteFolderByID_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM favorite_folder_videos ffv USING favorite_folders ff`).
		WithArgs(int64(3), int64(10), int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/5/favorite/folders/3", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "5")
	req.SetPathValue("folderId", "3")
	w := httptest.NewRecorder()
	h.handleVideoFavoriteFolderByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestXC_HandleVideoFavoriteFolderByID_Invalid(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/0/favorite/folders/0", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "0")
	req.SetPathValue("folderId", "0")
	w := httptest.NewRecorder()
	h.handleVideoFavoriteFolderByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestXC_HandleVideoFavoriteFolderByID_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM favorite_folder_videos ffv USING favorite_folders ff`).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/5/favorite/folders/3", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "5")
	req.SetPathValue("folderId", "3")
	w := httptest.NewRecorder()
	h.handleVideoFavoriteFolderByID(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUpdateManuscript: JSON 成功路径（分类 + 标签 + 描述） ----

func TestXC_HandleUpdateManuscript_JSONFull(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE manuscripts SET title`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE manuscripts SET description`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`UPDATE manuscripts SET category_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM video_tags WHERE video_id IN`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"title":"new","description":"nd","category_id":7,"tags":["go"]}`
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestXC_HandleUpdateManuscript_InvalidBody(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", strings.NewReader(`{bad json`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestXC_HandleUpdateManuscript_NotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM manuscripts WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("PUT", "/api/v1/manuscript/1", strings.NewReader(`{"title":"new"}`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleUpdateManuscript(w, req)
	assert.Equal(t, 404, w.Code)
}

// ---- handleVideoFavoriteFolders: PUT 全量设置成功 ----

func TestXC_HandleVideoFavoriteFolders_PutSuccess(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM favorite_folder_videos ffv USING favorite_folders ff`).
		WithArgs(int64(10), int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO favorite_folder_videos`).
		WithArgs(int64(3), int64(5), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	body := `{"folderIds":[3]}`
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/5/favorite/folders", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "5")
	w := httptest.NewRecorder()
	h.handleVideoFavoriteFolders(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleFavoriteFolderByID: PUT / DELETE 成功 ----

func TestXC_HandleFavoriteFolderByID_PutSuccess(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`UPDATE favorite_folders SET name`).WithArgs("n", int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("PUT", "/api/v1/manuscript/favorite/folders/1", strings.NewReader(`{"name":"n"}`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestXC_HandleFavoriteFolderByID_PutEmptyName(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/favorite/folders/1", strings.NewReader(`{"name":""}`))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderByID(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestXC_HandleFavoriteFolderByID_DeleteSuccess(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/favorite/folders/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestXC_HandleFavoriteFolderByID_DeleteError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`DELETE FROM favorite_folder_videos WHERE folder_id`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM favorite_folders WHERE id`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/favorite/folders/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleFavoriteFolderByID(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleCollect: POST 成功路径 ----

func TestXC_HandleCollect_PostSuccess(t *testing.T) {
	m := newHTTPHandlerWithInteract(t)

	m.interact.ExpectQuery(`SELECT COUNT\(\*\) FROM user_interactions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	m.interact.ExpectExec(`INSERT INTO user_interactions`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`UPDATE manuscripts SET collect_count`).WillReturnResult(sqlmock.NewResult(0, 1))
	m.interact.ExpectExec(`INSERT INTO manuscript_daily_metrics`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"folderId":1}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/1/collect", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	m.h.handleCollect(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleManuscriptDetail: DB 错误分支 ----

func TestXC_HandleManuscriptDetail_DBError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectQuery(`SELECT id FROM manuscripts WHERE id = \$1`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/detail/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleManuscriptDetail(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleDeleteManuscript: 服务层错误（未找到） ----

func TestXC_HandleDeleteManuscript_ServiceError(t *testing.T) {
	h, mock := newHTTPHandler(t)
	mock.ExpectExec(`DELETE FROM manuscripts`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/1", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.handleDeleteManuscript(w, req)
	assert.Equal(t, 403, w.Code)
}

// ---- handleUploadComplete: 视频插入失败 → continue 后仍成功 ----

func TestXC_HandleUploadComplete_VideoInsertError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-vx").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-vx").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), "[]", `[{"url":"http://v.mp4"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-vx/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-vx")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- 服务层: PublishManuscript / UnpublishManuscript 错误分支 ----

func TestXC_Service_Publish_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrNoRows)

	_, err := svc.PublishManuscript(t.Context(), &pb.PublishManuscriptRequest{Id: 1, UserId: 10})
	require.Error(t, err)
}

func TestXC_Service_Publish_Internal(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrConnDone)

	_, err := svc.PublishManuscript(t.Context(), &pb.PublishManuscriptRequest{Id: 1, UserId: 10})
	require.Error(t, err)
}

func TestXC_Service_Unpublish_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrNoRows)

	_, err := svc.UnpublishManuscript(t.Context(), &pb.UnpublishManuscriptRequest{Id: 1, UserId: 10})
	require.Error(t, err)
}

func TestXC_Service_Unpublish_Internal(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectExec(`UPDATE manuscripts SET status`).WillReturnError(sql.ErrConnDone)

	_, err := svc.UnpublishManuscript(t.Context(), &pb.UnpublishManuscriptRequest{Id: 1, UserId: 10})
	require.Error(t, err)
}

// ---- 服务层: GetManuscriptWithVideos 未找到 ----

func TestXC_Service_GetManuscriptWithVideos_NotFound(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	_, err := svc.GetManuscriptWithVideos(t.Context(), &pb.GetManuscriptWithVideosRequest{Id: 999})
	require.Error(t, err)
}

// ---- 服务层: GetManuscript 数据库错误 ----

func TestXC_Service_GetManuscript_DBError(t *testing.T) {
	svc, msMock, _ := newService(t)
	msMock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE id`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	_, err := svc.GetManuscript(t.Context(), &pb.GetManuscriptRequest{Id: 1})
	require.Error(t, err)
}

// ---- 仓库: ListByCategory 子分区展开 + 错误分支 ----

func TestXC_Repo_ListByCategory_ChildExpansion(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	// expandCategoryIDs: 返回子分区
	mock.ExpectQuery(`SELECT id FROM categories WHERE parent_id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)).AddRow(int64(3)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	list, total, err := repo.ListByCategory(ctx, 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, list, 0)
}

func TestXC_Repo_ListByCategory_CountError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id FROM categories WHERE parent_id = \$1`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = ANY`).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.ListByCategory(ctx, 1, 1, 20)
	require.Error(t, err)
}

func TestXC_Repo_ListByCategory_ExpandError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id FROM categories WHERE parent_id = \$1`).WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT .+ FROM manuscripts WHERE category_id = ANY`).
		WillReturnRows(sqlmock.NewRows(manuscriptCols))

	_, total, err := repo.ListByCategory(ctx, 1, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

// ---- 仓库: FindVideosByManuscriptID / FindTagsByVideoID 错误分支 ----

func TestXC_Repo_FindVideos_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	mock.ExpectQuery(`SELECT .+ FROM videos WHERE manuscript_id`).WillReturnError(sql.ErrConnDone)

	_, err := repo.FindVideosByManuscriptID(ctx, 1)
	require.Error(t, err)
}

func TestXC_Repo_FindTags_QueryError(t *testing.T) {
	repo, mock := newRepo(t)
	ctx := context.Background()
	mock.ExpectQuery(`SELECT t.name FROM tags t`).WillReturnError(sql.ErrConnDone)

	_, err := repo.FindTagsByVideoID(ctx, 1)
	require.Error(t, err)
}
