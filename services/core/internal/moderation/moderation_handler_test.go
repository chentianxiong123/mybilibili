package moderation

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

func newMockModerationHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewHandler(NewService(NewRepository(db))), mock
}

func TestHandleWords_GET_List(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectQuery(`SELECT id, word, match_type FROM prohibited_words ORDER BY`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word", "match_type", "category", "is_enabled", "created_at", "updated_at"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM prohibited_words`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/moderation/admin/prohibited-words", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list":[]`)
	assert.Contains(t, rec.Body.String(), `"total":0`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleWords_POST_Create(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	mock.ExpectExec(`INSERT INTO prohibited_words`).WillReturnResult(sqlmock.NewResult(1, 1))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/moderation/admin/prohibited-words",
		strings.NewReader(`{"word":"badword","matchType":"exact","category":"spam"}`))
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleReports_Submit(t *testing.T) {
	h, mock := newMockModerationHandler(t)
	// 先查稿件 user_id，再 INSERT
	mock.ExpectQuery(`SELECT user_id FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(5)))
	mock.ExpectQuery(`INSERT INTO reports`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))

	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	body := `{"targetType":"MANUSCRIPT","targetId":7,"reason":"版权","description":"侵权"}`
	r := httptest.NewRequest(http.MethodPost, "/api/v1/report/submit", strings.NewReader(body))
	r.Header.Set("X-User-Id", "3")
	mux.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

var _ = time.Now
