package manuscript

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUploadComplete_VideoInsertFailure_LeavesOrphanManuscript pins a
// transaction-atomicity bug in manuscript_http_handler.go:handleUploadComplete
// (lines ~510-525).
//
// ROOT CAUSE: the "create manuscript" and "create video" steps are NOT wrapped
// in a transaction. `INSERT INTO manuscripts ... RETURNING id` executes and
// autocommits first; only then does `INSERT INTO videos ... RETURNING id` run.
// If the video insert fails, the handler returns 500 but the manuscript row is
// already durable — an orphaned manuscript with no video, never cleaned up.
//
// IMPACT: partial write on error. Repeated failures create orphaned manuscripts
// (drafts the author cannot see as complete, cluttering stats and DB). The same
// non-transactional pattern also affects handleUploadCompleteWeb (lines 1689-1756)
// and handleUpdateManuscript (lines 921-951).
func TestUploadComplete_VideoInsertFailure_LeavesOrphanManuscript(t *testing.T) {
	h, mock := newHTTPHandler(t)

	// Session owner check.
	mock.ExpectQuery(`SELECT user_id FROM upload_sessions WHERE id=\$1`).WithArgs("sess1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(10)))
	// Load session metadata (playable URL present).
	mock.ExpectQuery(`SELECT title, description, category_id, tags, videos FROM upload_sessions WHERE id = \$1`).
		WithArgs("sess1").
		WillReturnRows(sqlmock.NewRows([]string{"title", "description", "category_id", "tags", "videos"}).
			AddRow("t", "d", int64(1), `[]`, `[{"url":"/uploads/a.mp4"}]`))
	// Category exists.
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1\)`).WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	// Manuscript INSERT succeeds and autocommits (no BEGIN expectation anywhere).
	mock.ExpectQuery(`INSERT INTO manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	// Video INSERT FAILS.
	mock.ExpectQuery(`INSERT INTO videos`).WillReturnError(errors.New("video insert failed"))

	req := newReq("POST", "/api/v1/manuscript/upload-session/sess1/complete", "")
	w := httptest.NewRecorder()
	h.handleRouter(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// BUG confirmed: manuscript 42 was INSERTed (no rollback). A correct
	// implementation wraps manuscript+video in one tx with BEGIN/COMMIT and a
	// ROLLBACK on failure — sqlmock would have required ExpectBegin here, and
	// the manuscript INSERT would not survive the video error.
	require.NoError(t, mock.ExpectationsWereMet())
}
