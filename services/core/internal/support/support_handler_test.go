package support

import (
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
