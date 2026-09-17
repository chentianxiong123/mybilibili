package support

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockSupport(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewHandler(NewService(NewRepository(db))), mock
}

func TestHandleCreate_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/operation/tickets", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHandleCreate_EmptyTitle(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/operation/tickets", strings.NewReader(`{"content":"x"}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "title required")
}

func TestHandleCreate_Success(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(int64(1), time.Now().Truncate(time.Second)))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/operation/tickets", strings.NewReader(`{"title":"无法播放"}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"ticket_no"`)
	assert.Contains(t, rec.Body.String(), `"title":"无法播放"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleList_EmptyResult(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/support/admin/tickets", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list":[]`)
	assert.Contains(t, rec.Body.String(), `"total":0`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTicketByID_NotFound(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`SELECT id, ticket_no FROM support_tickets`).
		WillReturnError(sqlmock.ErrCancelled)

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/support/admin/tickets/42", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":404`)
}

func TestHandleCustomerSession_FallbackUserID(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(int64(7), time.Now().Truncate(time.Second)))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	// 不传 userId，回退 X-User-Id
	r := httptest.NewRequest(http.MethodPost, "/api/v1/operation/internal/tickets/customer-session",
		strings.NewReader(`{"title":"咨询"}`))
	r.Header.Set("X-User-Id", "8")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"sessionId"`)
	assert.Contains(t, rec.Body.String(), `"ticketNo"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCreate_ServiceError(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnError(sqlmock.ErrCancelled)

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/operation/tickets", strings.NewReader(`{"title":"test"}`))
	r.Header.Set("X-User-Id", "1")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleCustomerSession_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/operation/internal/tickets/customer-session", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":405`)
}

func TestHandleCustomerSession_ServiceError(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnError(sqlmock.ErrCancelled)

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/operation/internal/tickets/customer-session",
		strings.NewReader(`{"userId":9,"title":"err"}`))
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":400`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTicketByID_Process_Success(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectExec(`UPDATE support_tickets SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/support/admin/tickets/1/process",
		strings.NewReader(`{"adminReply":"已处理"}`))
	r.Header.Set("X-Admin-Id", "10")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTicketByID_Delete_Success(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectExec(`DELETE FROM support_tickets`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/support/admin/tickets/5", nil)
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSessionProcess_MethodNotAllowed(t *testing.T) {
	h, _ := newMockSupport(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/operation/internal/tickets/session/1", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":405`)
}

func TestHandleSessionProcess_Success(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectExec(`UPDATE support_tickets SET admin_reply`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/operation/internal/tickets/session/99",
		strings.NewReader(`{"adminReply":"已回复"}`))
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSessionProcess_DBError(t *testing.T) {
	h, mock := newMockSupport(t)
	mock.ExpectExec(`UPDATE support_tickets SET admin_reply`).
		WillReturnError(sqlmock.ErrCancelled)

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/operation/internal/tickets/session/1",
		strings.NewReader(`{"adminReply":"fail"}`))
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":500`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_ProcessedAt(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"processed_at", "created_at",
		}).AddRow(1, "TK001", 10, 0, "web", "general", "normal",
			"PROCESSED", "title", "content", "", "reply", 5, now, now))

	ticket, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, ticket.ProcessedAt)
	assert.Equal(t, "PROCESSED", ticket.Status)
}

func TestRepository_GetByID_NullProcessedAt(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"processed_at", "created_at",
		}).AddRow(1, "TK002", 10, 0, "web", "general", "normal",
			"PENDING", "title", "content", "", "", 0, nil, now))

	ticket, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, ticket.ProcessedAt)
}

func TestRepository_Process_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE support_tickets SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Process(context.Background(), 1, 10, "已处理")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Process_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	mock.ExpectExec(`UPDATE support_tickets SET status`).
		WillReturnError(sqlmock.ErrCancelled)

	err = repo.Process(context.Background(), 1, 10, "fail")
	assert.Error(t, err)
}

func TestService_Process(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := NewService(NewRepository(db))
	mock.ExpectExec(`UPDATE support_tickets SET status`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.Process(context.Background(), 1, 10, "ok")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WillReturnError(sqlmock.ErrCancelled)

	ticket, err := repo.Create(context.Background(), 1, "web", "general", "normal", "t", "c")
	assert.Error(t, err)
	assert.Nil(t, ticket)
}

func TestRepository_List_WithStatus(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets WHERE status`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).
			AddRow(1, "TK001", 1, 0, "web", "general", "normal", "PENDING", "t1", "c1", "", "", 0, now, nil).
			AddRow(2, "TK002", 2, 0, "app", "bug", "high", "PROCESSED", "t2", "c2", "", "reply", 5, now, now))

	list, total, err := repo.List(context.Background(), "PENDING", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
	assert.Equal(t, "PENDING", list[0].Status)
}

func TestRepository_List_NoStatus(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).AddRow(3, "TK003", 3, 0, "web", "general", "normal", "PENDING", "t3", "c3", "", "", 0, now, nil))

	list, total, err := repo.List(context.Background(), "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestHandleList_WithStatus(t *testing.T) {
	h, mock := newMockSupport(t)

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets WHERE status`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, ticket_no`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).AddRow(10, "TK10", 1, 0, "web", "general", "normal", "PENDING", "t", "c", "", "", 0, now, nil))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/support/admin/tickets?status=PENDING", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"total":1`)
	require.NoError(t, mock.ExpectationsWereMet())
}
