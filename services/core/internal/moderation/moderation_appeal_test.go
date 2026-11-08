package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ appeal pure helper functions ============

func TestAppealStatusGroup(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"processing", []string{"PENDING", "PROCESSING"}},
		{"PENDING", []string{"PENDING", "PROCESSING"}},
		{"doing", []string{"PENDING", "PROCESSING"}},
		{"  done  ", []string{"APPROVED", "REJECTED"}},
		{"FINISHED", []string{"APPROVED", "REJECTED"}},
		{"completed", []string{"APPROVED", "REJECTED"}},
		{"", nil},
		{"bogus", nil},
		{"all", nil},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := appealStatusGroup(c.in)
			if c.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, c.want, got)
		})
	}
}

func TestAppealStatusText(t *testing.T) {
	cases := map[string]string{
		"pending":     "申诉处理中",
		"processing":  "申诉处理中",
		"  PENDING  ": "申诉处理中",
		"approved":    "申诉完成",
		"rejected":    "申诉完成",
		"weird":       "申诉处理中",
		"":            "申诉处理中",
	}
	for in, want := range cases {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			assert.Equal(t, want, appealStatusText(in))
		})
	}
}

func TestManuscriptAppealStatusText(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "审核中"},
		{1, "处理中"},
		{3, "已通过"},
		{4, "未通过"},
		{-1, "已下架"},
		{99, "处理中"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d", c.in), func(t *testing.T) {
			assert.Equal(t, c.want, manuscriptAppealStatusText(c.in))
		})
	}
}

// ============ handleAppealList ============

func TestHandleAppealList_MethodNotAllowed(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "POST", "/api/v1/appeal/me/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAppealList_Unauthorized(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleAppealList_SuccessEmpty(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	// count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// list query
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at",
		}))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"list":[]`)
	assert.Contains(t, w.Body.String(), `"total":0`)
	assert.Contains(t, w.Body.String(), `"totalPages":0`)
}

func TestHandleAppealList_WithRows(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	now := time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at",
		}).AddRow(int64(7), int64(3), "My Video", int64(3), "cover.jpg", "why", "desc", "APPROVED", "done", now))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"manuscriptId":3`)
	assert.Contains(t, w.Body.String(), `"appealTime":"2024-05-01 10:00:00"`)
	assert.Contains(t, w.Body.String(), `"totalPages":1`)
}

func TestHandleAppealList_StatusFilter(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	// statuses from "done" -> APPROVED, REJECTED
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at",
		}))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list?status=done", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleAppealList_PagingClamp(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	// page=0 -> 1, size=500 -> 10
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at",
		}))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list?page=0&size=500", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleAppealList_QueryError(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnError(errors.New("boom"))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}

// ============ handleAppealStats ============

func TestHandleAppealStats_MethodNotAllowed(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "POST", "/api/v1/appeal/me/stats", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAppealStats_Unauthorized(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/stats", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleAppealStats_Success(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`PENDING.*PROCESSING`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`APPROVED.*REJECTED`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/stats", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":7`)
	assert.Contains(t, w.Body.String(), `"processing":2`)
	assert.Contains(t, w.Body.String(), `"done":5`)
}

// ============ handleAppealByID ============

func TestHandleAppealByID_MethodNotAllowed(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "DELETE", "/api/v1/appeal/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAppealByID_Unauthorized(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/1", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleAppealByID_InvalidID(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/abc", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400`)
}

func TestHandleAppealByID_ByID_ZeroID(t *testing.T) {
	h, _ := newMockModerationHandler(t)
	// ParseInt("0")==0 -> rejected
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/0", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleAppealByID_NotFound(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WithArgs(int64(1), int64(1)).
		WillReturnError(sql.ErrNoRows)
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/1", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"code":404`)
}

func TestHandleAppealByID_Success_Unprocessed(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WithArgs(int64(9), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at", "processed_at",
		}).AddRow(int64(5), "T", int64(0), "", "why", "", "PENDING", "", now, nil))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"appealNo":9`)
	assert.Contains(t, w.Body.String(), `"appealTime":"2024-01-02 03:04:05"`)
	// processedAt invalid -> key should be absent
	assert.NotContains(t, w.Body.String(), `"processedTime"`)
}

func TestHandleAppealByID_Success_Processed(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	processed := time.Date(2024, 6, 7, 8, 9, 10, 0, time.UTC)
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WithArgs(int64(9), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at", "processed_at",
		}).AddRow(int64(5), "T", int64(3), "c.jpg", "why", "d", "APPROVED", "ok", created, processed))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/9", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"processedTime":"2024-06-07 08:09:10"`)
}

// ============ handleBatchImport success (covers the encoded-response line) ============

func TestHandleBatchImport_Success(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("a", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO prohibited_words`).WithArgs("b", "", "").WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForHandler(h), "POST", "/api/v1/moderation/admin/prohibited-words/batch-import",
		`{"words":[{"word":"a"},{"word":"b"}]}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"imported":2`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ============ handleWordByID PUT enabled path ============

func TestHandleWordByID_Put_Enabled(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT id, word, match_type`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}).
			AddRow(1, "bad", "exact", "chat", 1, now, now))
	mock.ExpectExec(`UPDATE prohibited_words SET word=\$1`).
		WithArgs("new", "exact", "chat", int32(1), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// is_enabled=5 (non-zero) -> stays enabled
	w := doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/prohibited-words/1",
		`{"word":"new","match_type":"exact","category":"chat","is_enabled":5}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

// ============ handleReports error / branch paths ============

func TestHandleReports_Process_Reject(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectExec(`UPDATE reports SET status=\$1, admin_remark=\$2`).
		WithArgs("REJECTED", "no", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w := doReq(muxForHandler(h), "PUT", "/api/v1/moderation/admin/report/process/2",
		`{"action":"reject","admin_remark":"no"}`, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ============ ListWords row-scan error path ============
// NOTE: sqlmock only validates that the number of returned columns is >= the
// number of destination values in Scan. Providing fewer columns triggers a
// "wrong number of arguments" error on the NEXT Next() call (not during
// iteration), so this path is partially covered. The empty-list + error
// case is still reachable via an empty result set with correct arity.

func TestRepoListWords_QueryError(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectQuery(`SELECT id, word, match_type, COALESCE`).
		WillReturnError(errors.New("boom"))
	_, err := repo.ListWords(context.Background(), 1, 20)
	assert.Error(t, err)
}

// ============ context cancellation / timeout ============

func TestRepoListWords_ContextCancelled(t *testing.T) {
	repo, _ := newMockModerationRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := repo.ListWords(ctx, 1, 20)
	assert.Error(t, err)
}

func TestRepoContainsProhibited_Error(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM prohibited_words`).
		WithArgs("x").WillReturnError(errors.New("boom"))
	_, err := repo.ContainsProhibited(context.Background(), "x")
	assert.Error(t, err)
}

func TestRepoBatchImport_ContextCancelled(t *testing.T) {
	repo, _ := newMockModerationRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n, err := repo.BatchImportWords(ctx, []*ProhibitedWord{{Word: "x"}})
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

// ============ SubmitReport / ProcessReport error paths ============

func TestService_SubmitReport_Error(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectQuery(`INSERT INTO reports`).WillReturnError(errors.New("boom"))
	assert.Error(t, NewService(repo).SubmitReport(context.Background(), 1, "COMMENT", 2, 3, "spam", "d"))
}

func TestRepoProcessReport_Error(t *testing.T) {
	repo, mock := newMockModerationRepo(t)
	mock.ExpectExec(`UPDATE reports SET status=\$1`).WillReturnError(errors.New("boom"))
	assert.Error(t, repo.ProcessReport(context.Background(), 1, "resolve", "ok"))
}

// ============ handleAppealList Scan error inside loop ============
// Note: sqlmock validates only that returned column count >= Scan destination
// count. A row with fewer columns is silently consumed without error, so this
// `continue` branch on row-level scan error cannot be deterministically
// triggered through sqlmock. The list-level error path is covered by
// TestHandleAppealList_QueryError. This test instead verifies that a valid
// row flows through and the response reflects it.

func TestHandleAppealList_SuccessWithNullProcessedAt(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.MatchExpectationsInOrder(false)
	now := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM appeals`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM appeals a LEFT JOIN manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "title", "status", "cover_url", "reason", "description", "ap_status", "admin_remark", "created_at",
		}).AddRow(int64(2), int64(1), "V", int64(-1), "", "r", "", "REJECTED", "", now))
	w := doReq(muxForHandler(h), "GET", "/api/v1/appeal/me/list", "", map[string]string{"X-User-Id": "1"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"manuscriptStatus":-1`)
	assert.Contains(t, w.Body.String(), `"manuscriptStatusText":"已下架"`)
	assert.Contains(t, w.Body.String(), `"statusText":"申诉完成"`)
}
