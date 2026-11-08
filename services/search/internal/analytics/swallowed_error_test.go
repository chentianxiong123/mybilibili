package analytics

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FINDING C1 (swallowed error): analytics handlers ignore the error returned by
// the repository (analytics_handler.go:39,50,62,...: `data, _ := h.svc.Overview(...)`).
// When the backing DB query fails, Overview still returns (zeros, nil), so the
// handler responds 200 with an all-zero payload instead of 500. This pins that
// a DB error surfaces as HTTP 200.
func TestCorrectness_Overview_DBError_Returns200(t *testing.T) {
	h, mock := newTestHandler(t)

	// First (aggregate) query fails -> repository swallows the error.
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\)`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/overview?days=7", nil)
	req.Header.Set("X-User-Id", "42")
	rr := httptest.NewRecorder()

	h.handleOverview(rr, req)

	// A correct handler surfaces the DB error as 500. Observed: 200 => swallowed.
	assert.Equal(t, http.StatusOK, rr.Code, "DB error in Overview must NOT return 200")
}
