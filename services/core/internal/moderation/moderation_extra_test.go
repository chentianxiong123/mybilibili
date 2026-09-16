package moderation

import (
	"context"
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

func muxForHandler(h *Handler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestRepo_UpdateWord(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`UPDATE prohibited_words SET word=\$1, match_type=\$2, category=\$3, is_enabled=\$4 WHERE id=\$5`).
		WithArgs("bad", "exact", "chat", int32(0), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateWord(context.Background(), 1, "bad", "exact", "chat", 0))
}

func TestRepo_BatchImportWords(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("a", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("b", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	n, err := repo.BatchImportWords(context.Background(), []*ProhibitedWord{{Word: "a"}, {Word: ""}, {Word: "b"}})
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	mock.ExpectExec(`INSERT INTO prohibited_words`).WillReturnError(errors.New("boom"))
	n, err = repo.BatchImportWords(context.Background(), []*ProhibitedWord{{Word: "x"}})
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestRepo_ContainsProhibited(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM prohibited_words`).WithArgs("bad word").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	ok, err := repo.ContainsProhibited(context.Background(), "bad word")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRepo_ListReports(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, reporter_id, target_type, target_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "reporter_id", "target_type", "target_id", "manuscript_id", "reason", "description", "status", "admin_remark", "processed_at", "ai_verdict", "ai_risk_level", "created_at"}).
			AddRow(1, 2, "COMMENT", 3, 4, "spam", "desc", "PENDING", "remark", nil, "verdict", "HIGH", now))
	list, err := repo.ListReports(context.Background(), 1, 20, "")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "COMMENT", list[0].TargetType)

	mock.ExpectQuery(`WHERE status = \$1 ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs("RESOLVED", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "reporter_id", "target_type", "target_id", "manuscript_id", "reason", "description", "status", "admin_remark", "processed_at", "ai_verdict", "ai_risk_level", "created_at"}).
			AddRow(2, 3, "REPLY", 4, 0, "r", "d", "RESOLVED", "rm", now, "v", "LOW", now))
	list, err = repo.ListReports(context.Background(), 1, 20, "RESOLVED")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.NotNil(t, list[0].ProcessedAt)

	mock.ExpectQuery(`SELECT id, reporter_id, target_type`).WillReturnError(errors.New("boom"))
	_, err = repo.ListReports(context.Background(), 1, 20, "")
	assert.Error(t, err)
}

func TestRepo_ProcessReport(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`UPDATE reports SET status=\$1, admin_remark=\$2, processed_at=NOW\(\) WHERE id=\$3`).
		WithArgs("RESOLVED", "ok", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.ProcessReport(context.Background(), 1, "resolve", "ok"))

	mock.ExpectExec(`UPDATE reports SET status=\$1, admin_remark=\$2, processed_at=NOW\(\) WHERE id=\$3`).
		WithArgs("REJECTED", "no", int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.ProcessReport(context.Background(), 2, "reject", "no"))
}

func TestRepo_UpdateAIRegReview(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`UPDATE reports SET ai_verdict=\$1, ai_risk_level=\$2 WHERE id=\$3`).
		WithArgs("PASS", "LOW", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateAIRegReview(context.Background(), 1, "PASS", "LOW"))
}

func TestService_PassThrough(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	svc := NewService(repo)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, word, match_type`).WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
		AddRow(1, "w", "exact", "chat", 1, now, now))
	w, err := svc.GetWord(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "w", w.Word)

	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("w", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.CreateWord(ctx, "w", "", ""))

	mock.ExpectExec(`DELETE FROM prohibited_words`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.DeleteWord(ctx, 1))

	mock.ExpectExec(`UPDATE prohibited_words SET word=\$1`).WithArgs("n", "", "", int32(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.UpdateWord(ctx, 1, "n", "", "", 1))

	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("a", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	n, err := svc.BatchImportWords(ctx, []*ProhibitedWord{{Word: "a"}})
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM prohibited_words`).WithArgs("x").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	ok, err := svc.ContainsProhibited(ctx, "x")
	require.NoError(t, err)
	assert.False(t, ok)

	mock.ExpectQuery(`INSERT INTO reports`).WithArgs(int64(1), "COMMENT", int64(2), int64(3), "spam", "d").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	require.NoError(t, svc.SubmitReport(ctx, 1, "COMMENT", 2, 3, "spam", "d"))

	mock.ExpectQuery(`SELECT id, reporter_id, target_type`).WillReturnRows(sqlmock.NewRows([]string{"id", "reporter_id", "target_type", "target_id", "manuscript_id", "reason", "description", "status", "admin_remark", "processed_at", "ai_verdict", "ai_risk_level", "created_at"}))
	list, err := svc.ListReports(ctx, 1, 20, "")
	require.NoError(t, err)
	assert.Len(t, list, 0)

	mock.ExpectExec(`UPDATE reports SET status=\$1`).WithArgs("RESOLVED", "rm", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.ProcessReport(ctx, 1, "resolve", "rm"))

	mock.ExpectExec(`UPDATE reports SET ai_verdict=\$1`).WithArgs("PASS", "LOW", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.UpdateAIRegReview(ctx, 1, "PASS", "LOW"))
}

func TestHandleWordByID_Get(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(1, "bad", "exact", "chat", 1, now, now))
	w := doReq(muxForHandler(h), "GET", "/api/v1/moderation/admin/prohibited-words/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleWordByID_Put(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(1, "bad", "exact", "chat", 1, now, now))
	mock.ExpectExec(`UPDATE prohibited_words SET word=\$1`).WithArgs("new", "exact", "chat", int32(1), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/prohibited-words/1", `{"word":"new","match_type":"exact","category":"chat","is_enabled":0}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleWordByID_Put_NotFound(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(9)).WillReturnError(errors.New("no rows"))
	w := doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/prohibited-words/9", `{"word":"x"}`, nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleWordByID_Delete(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(1, "bad", "exact", "chat", 1, now, now))
	mock.ExpectExec(`DELETE FROM prohibited_words WHERE id = \$1`).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForHandler(h), "DELETE", "/api/v1/moderation/admin/prohibited-words/1", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleWordByID_Delete_NotFound(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(9)).WillReturnError(errors.New("no rows"))
	w := doReq(muxForHandler(h), "DELETE", "/api/v1/moderation/admin/prohibited-words/9", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleWordByID_MethodNotAllowed(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "POST", "/api/v1/moderation/admin/prohibited-words/1", `{}`, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleBatchImport_Error(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "GET", "/api/v1/moderation/admin/prohibited-words/batch-import", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	mock.ExpectExec(`INSERT INTO prohibited_words`).WillReturnError(errors.New("boom"))
	w = doReq(muxForHandler(h), "POST", "/api/v1/moderation/admin/prohibited-words/batch-import", `{"words":[{"word":"x"}]}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSubmitReport_Error(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`INSERT INTO reports`).WillReturnError(errors.New("boom"))
	w := doReq(muxForHandler(h), "POST", "/api/v1/report/submit", `{"target_type":"COMMENT","target_id":1}`, map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleReports(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, reporter_id, target_type, target_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "reporter_id", "target_type", "target_id", "manuscript_id", "reason", "description", "status", "admin_remark", "processed_at", "ai_verdict", "ai_risk_level", "created_at"}).
			AddRow(1, 2, "COMMENT", 3, 0, "spam", "d", "PENDING", "", nil, "", "", now))
	w := doReq(muxForHandler(h), "GET", "/api/v1/moderation/admin/report/list?status=PENDING", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE reports SET ai_verdict=\$1`).WithArgs("PASS", "LOW", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/report/ai-review-result", `{"reportId":1,"verdict":"PASS","riskLevel":"LOW"}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE reports SET status=\$1, admin_remark=\$2, processed_at=NOW\(\) WHERE id=\$3`).
		WithArgs("RESOLVED", "ok", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/report/process/1", `{"action":"resolve","admin_remark":"ok"}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(muxForHandler(h), "GET", "/api/v1/moderation/admin/report/unknown", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- AdminHandler ----

func muxForAdmin(h *AdminHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func newAdminHandler(t *testing.T) (*AdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewAdminHandler(db), mock
}

func TestAdminHandler_Pending(t *testing.T) {
	h, mock := newAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM content_reviews WHERE status = 'pending'`).
		WithArgs("", int32(1), int32(20)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "user_id", "content", "status", "reviewed_at"}).
			AddRow(1, "comment", int64(2), "bad", "pending", now))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/pending", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_All_WithFilters(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`AND type = \$1`).
		WithArgs("comment", "pending", int32(1), int32(20)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "user_id", "content", "status", "reviewed_at"}).
			AddRow(1, "comment", nil, "bad", nil, nil))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/all?contentType=comment&status=pending", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_QueryError(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectQuery(`FROM content_reviews WHERE status = 'pending'`).WillReturnError(errors.New("boom"))
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/pending", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_Restore(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'active'`).WithArgs("comment", "5").WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "PUT", "/api/v1/moderation/admin/restore/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Delete(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'deleted'`).WithArgs("comment", "5").WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForAdmin(h), "DELETE", "/api/v1/moderation/admin/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_Batch(t *testing.T) {
	h, mock := newAdminHandler(t)

	w := doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch", `bad`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectExec(`UPDATE content_reviews SET status = 'deleted'`).WithArgs("comment", int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE content_reviews SET status = 'deleted'`).WithArgs("comment", int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch", `{"action":"delete","items":[{"type":"comment","id":1},{"type":"comment","id":2}]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE content_reviews SET status = 'active'`).WithArgs("comment", int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForAdmin(h), "POST", "/api/v1/moderation/admin/batch", `{"action":"restore","items":[{"type":"comment","id":3}]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminHandler_NotFound(t *testing.T) {
	h, _ := newAdminHandler(t)
	w := doReq(muxForAdmin(h), "GET", "/api/v1/moderation/admin/unknown", "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_exec_Error(t *testing.T) {
	h, mock := newAdminHandler(t)
	mock.ExpectExec(`UPDATE content_reviews SET status = 'active'`).WillReturnError(errors.New("boom"))
	w := doReq(muxForAdmin(h), "PUT", "/api/v1/moderation/admin/restore/comment/5", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- AdminContentHandler ----

func muxForContentAdmin(h *AdminContentHandler) http.Handler {
	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func newContentAdminHandler(t *testing.T) (*AdminContentHandler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewAdminContentHandler(db), mock
}

func TestAdminContentHandler_CommentList(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()

	w := doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/comments", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	mock.ExpectQuery(`FROM comments c JOIN users u`).
		WithArgs(int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "manuscript_id", "user_id", "nickname", "avatar", "manuscript_title", "parent_content", "content", "like_count", "status", "created_at"}).
			AddRow(1, nil, int64(9), int64(2), "nick", "a", "title", nil, "bad", 3, "0", now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentList_ReplyType(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM replies r JOIN users u`).
		WithArgs("bad", "REMOVED", int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "manuscript_id", "user_id", "nickname", "avatar", "manuscript_title", "parent_content", "content", "like_count", "status", "created_at"}).
			AddRow(1, int64(10), int64(9), int64(2), "nick", "a", "title", "parent", "bad", 3, "REMOVED", now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM replies`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments?type=reply&keyword=bad&status=REMOVED", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_CommentList_Error(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectQuery(`FROM comments c JOIN users u`).WillReturnError(errors.New("boom"))
	w := doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/comments", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAppendFilter(t *testing.T) {
	q, args := appendFilter("SELECT 1 WHERE 1=1", nil, "v", "c.content")
	assert.Equal(t, "SELECT 1 WHERE 1=1 AND c.content = $1", q)
	assert.Equal(t, []interface{}{"v"}, args)

	q, args = appendFilter("SELECT 1", nil, "", "c.content")
	assert.Equal(t, "SELECT 1", q)
	assert.Nil(t, args)
}

func TestAdminContentHandler_CommentByID(t *testing.T) {
	h, mock := newContentAdminHandler(t)

	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/abc", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectExec(`UPDATE comments SET status = 1 WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE comments SET status = 0 WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE replies SET status = 'REMOVED' WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/reply/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectExec(`UPDATE replies SET status = 'NORMAL' WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/comments/reply/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_DanmakuList(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	now := time.Now()

	w := doReq(muxForContentAdmin(h), "POST", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	mock.ExpectQuery(`FROM danmaku d JOIN users u`).
		WithArgs("%bad%", int32(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "manuscript_title", "user_id", "nickname", "avatar", "content", "time", "mode", "created_at"}).
			AddRow(1, 2, 3, "title", 4, "nick", "a", "bad", 1.5, 1, now))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku WHERE content ILIKE \$1`).WithArgs("%bad%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w = doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku?keyword=bad", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`FROM danmaku d JOIN users u`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "manuscript_id", "manuscript_title", "user_id", "nickname", "avatar", "content", "time", "mode", "created_at"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku$`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w = doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	mock.ExpectQuery(`FROM danmaku d JOIN users u`).WillReturnError(errors.New("boom"))
	w = doReq(muxForContentAdmin(h), "GET", "/api/v1/moderation/admin/danmaku", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_DanmakuByID(t *testing.T) {
	h, mock := newContentAdminHandler(t)

	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/danmaku/abc", "", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mock.ExpectExec(`DELETE FROM danmaku WHERE id = \$1`).WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	w = doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/danmaku/5", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	w = doReq(muxForContentAdmin(h), "PUT", "/api/v1/moderation/admin/danmaku/5", "", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminContentHandler_exec_Error(t *testing.T) {
	h, mock := newContentAdminHandler(t)
	mock.ExpectExec(`UPDATE comments SET status = 1 WHERE id = \$1`).WillReturnError(errors.New("boom"))
	w := doReq(muxForContentAdmin(h), "DELETE", "/api/v1/moderation/admin/comments/comment/5", "", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
