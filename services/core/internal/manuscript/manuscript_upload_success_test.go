package manuscript

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUploadDir 在临时目录下创建 uploadID/part/chunk 文件并返回该临时目录。
func setupUploadDir(t *testing.T, uploadID string, parts map[int][]string) string {
	t.Helper()
	dir := t.TempDir()
	for part, chunks := range parts {
		for _, c := range chunks {
			p := filepath.Join(dir, uploadID, itoa(part), c)
			require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
			require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
		}
	}
	return dir
}

// ---- handleUploadCompleteWeb: 完整成功路径（多分片 + 多标签 + 分类校验） ----

func TestHandleUploadCompleteWeb_Success(t *testing.T) {
	dir := setupUploadDir(t, "up-w", map[int][]string{0: {"0", "1"}, 1: {"0"}})
	t.Setenv("MYBILIBILI_UPLOAD_DIR", dir)

	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-w").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-w").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), `["a","b"]`, `[{"title":"T1"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	// part 0
	mock.ExpectQuery(`INSERT INTO videos`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))

	// part 1
	mock.ExpectQuery(`INSERT INTO videos`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(101)))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`UPDATE upload_sessions SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users WHERE id = \$1`).WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(int32(1), int64(0)))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-w"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadCompleteWeb: 分类缺失（category_id<=0）→ 400 ----

func TestHandleUploadCompleteWeb_CategoryMissing(t *testing.T) {
	dir := setupUploadDir(t, "up-c", map[int][]string{0: {"0"}})
	t.Setenv("MYBILIBILI_UPLOAD_DIR", dir)

	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-c").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-c").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(0), "[]", `[{"title":"T1"}]`))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-c"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadCompleteWeb: 会话加载失败 → 500 ----

func TestHandleUploadCompleteWeb_LoadSessionError(t *testing.T) {
	t.Setenv("MYBILIBILI_UPLOAD_DIR", t.TempDir())
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-e").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-e").WillReturnError(sql.ErrConnDone)

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-e"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUploadCompleteWeb: 创建稿件失败 → 500 ----

func TestHandleUploadCompleteWeb_CreateManuscriptError(t *testing.T) {
	dir := setupUploadDir(t, "up-m", map[int][]string{0: {"0"}})
	t.Setenv("MYBILIBILI_UPLOAD_DIR", dir)

	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-m").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-m").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), "[]", `[{"title":"T1"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).WillReturnError(sql.ErrConnDone)

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-m"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUploadCompleteWeb: 视频插入失败 → 跳过该分片但整体成功 ----

func TestHandleUploadCompleteWeb_CreateVideoError(t *testing.T) {
	dir := setupUploadDir(t, "up-v", map[int][]string{0: {"0"}})
	t.Setenv("MYBILIBILI_UPLOAD_DIR", dir)

	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-v").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-v").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), "[]", `[{"title":"T1"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnError(sql.ErrConnDone)

	mock.ExpectExec(`UPDATE upload_sessions SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users WHERE id = \$1`).WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(int32(1), int64(0)))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-v"}, "f", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-complete", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadCompleteWeb(w, req)
	// 视频插入失败被 continue 跳过，整体仍返回成功
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadComplete: 逗号分隔标签路径 + video_url 字段 + 完整成功 ----

func TestHandleUploadComplete_CSVTagsAndVideoURL(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-csv").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-csv").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), "a, b, c", `[{"video_url":"http://v.mp4"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectQuery(`INSERT INTO videos`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO tags`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(3)))
	mock.ExpectExec(`INSERT INTO video_tags`).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`UPDATE upload_sessions SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT level, experience FROM users WHERE id = \$1`).WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"level", "experience"}).AddRow(int32(1), int64(0)))
	mock.ExpectExec(`UPDATE users SET experience`).WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-csv/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-csv")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadComplete: 创建稿件失败 → 500 ----

func TestHandleUploadComplete_CreateManuscriptError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-1").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), "[]", `[{"url":"http://v.mp4"}]`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`INSERT INTO manuscripts`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-1/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-1")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUploadComplete: 会话加载失败 → 500 ----

func TestHandleUploadComplete_LoadSessionError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("up-2").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("up-2").WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session/up-2/complete", nil)
	req.Header.Set("X-User-Id", "10")
	req.SetPathValue("id", "up-2")
	w := httptest.NewRecorder()
	h.handleUploadComplete(w, req)
	assert.Equal(t, 500, w.Code)
}

// ---- handleUploadChunk: 完整成功（含 partIndex/chunkIndex 缺省回退） ----

func TestHandleUploadChunk_Success_DefaultIndexes(t *testing.T) {
	t.Setenv("MYBILIBILI_UPLOAD_DIR", t.TempDir())
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-d").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	mock.ExpectExec(`UPDATE upload_sessions SET uploaded_chunks`).WillReturnResult(sqlmock.NewResult(0, 1))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-d"}, "chunk", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadChunk: 分片写入失败（os.Create 不可写）→ 500 ----

func TestHandleUploadChunk_CreateFileError(t *testing.T) {
	// 用一个"文件"作为上传根目录路径，使 os.MkdirAll 失败
	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	require.NoError(t, os.WriteFile(blocked, []byte("x"), 0o644))
	t.Setenv("MYBILIBILI_UPLOAD_DIR", blocked)

	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id = \$1`).WithArgs("up-ro").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))

	boundary := "b"
	body := buildMultipartFile(boundary, map[string]string{"uploadId": "up-ro", "partIndex": "0", "chunkIndex": "0"}, "chunk", "data")
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-chunk", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	w := httptest.NewRecorder()
	h.handleUploadChunk(w, req)
	// os.MkdirAll 失败 → 500
	assert.Equal(t, 500, w.Code)
}

// ---- handleUploadSession: 完整成功（client_id 提供 + 分类校验） ----

func TestHandleUploadSession_Success(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO upload_sessions`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"client_id":"abc","title":"t","description":"d","category_id":1,"tags":["a"],"videos":[{"total_chunks":3}],"total_chunks":2}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 200, w.Code)
}

// ---- handleUploadSession: 分区不存在 → 400 ----

func TestHandleUploadSession_CategoryNotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	body := `{"client_id":"abc","title":"t","category_id":5}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadSession: 分区查询错误 → 400 ----

func TestHandleUploadSession_CategoryError(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(5)).
		WillReturnError(sql.ErrConnDone)

	body := `{"client_id":"abc","title":"t","category_id":5}`
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(body))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadSession: 非法请求体（空 body）→ 400 ----

func TestHandleUploadSession_InvalidBody(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("POST", "/api/v1/manuscript/upload-session", strings.NewReader(`{}`))
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 400, w.Code)
}

// ---- handleUploadSessionByID: GET / DELETE / 未找到 / 方法不允许 ----

func TestHandleUploadSessionByID_Get(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id, user_id, title`).WithArgs("up-g", int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "category_id", "uploaded_chunks", "total_chunks", "status"}).
			AddRow("up-g", int64(10), "t", int64(1), 2, 3, "pending"))

	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-session/up-g", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUploadSessionByID_GetNotFound(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectQuery(`SELECT id, user_id, title`).WithArgs("up-x", int64(10)).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-session/up-x", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestHandleUploadSessionByID_Delete2(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM upload_sessions WHERE id`).WithArgs("up-d", int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/upload-session/up-d", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestHandleUploadSessionByID_DeleteError2(t *testing.T) {
	h, mock := newHTTPHandler(t)

	mock.ExpectExec(`DELETE FROM upload_sessions WHERE id`).WithArgs("up-d", int64(10)).
		WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest("DELETE", "/api/v1/manuscript/upload-session/up-d", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestHandleUploadSessionByID_MethodNotAllowed2(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("PUT", "/api/v1/manuscript/upload-session/up-d", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestHandleUploadSessionByID_EmptyPath(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-session/", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSessionByID(w, req)
	assert.Equal(t, 404, w.Code)
}

// ---- handleUploadSession 方法不允许 ----

func TestHandleUploadSession_MethodNotAllowed(t *testing.T) {
	h, _ := newHTTPHandler(t)
	req := httptest.NewRequest("GET", "/api/v1/manuscript/upload-session", nil)
	req.Header.Set("X-User-Id", "10")
	w := httptest.NewRecorder()
	h.handleUploadSession(w, req)
	assert.Equal(t, 405, w.Code)
}
