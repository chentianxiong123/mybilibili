package search

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FINDING C2 (swallowed error): handleHotScore / handleHotRank / handleHotDelete
// ignore the error from json.NewDecoder(...).Decode (search_handler.go:159,173,195).
// A malformed body leaves the request struct at zero values, the handler then
// reports HTTP 200 "success" instead of 400 — the caller is told the mutation
// succeeded when nothing (meaningful) was applied.
func TestCorrectness_HotScore_MalformedBody_Returns200(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/search/hot/score", strings.NewReader(`{not-json`))
	h.handleHotScore(rr, req)

	// A handler that validates its input returns 400. Observed: 200 => decode error swallowed.
	assert.Equal(t, http.StatusOK, rr.Code, "malformed JSON body to /hot/score must NOT return 200")
}

func TestCorrectness_HotDelete_MalformedBody_Returns200(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/search/hot/delete", strings.NewReader(`{`))
	h.handleHotDelete(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "malformed JSON body to /hot/delete must NOT return 200")
}

// FINDING C3 (incorrect validation): handleRelated (search_handler.go:240-244)
// only defaults size when it is exactly 0. A NEGATIVE ?size is passed straight
// through as LIMIT $N, which Postgres rejects ("LIMIT must not be negative").
// The repository then returns an error that the handler swallows -> 200 empty
// list instead of a 400. This pins that a negative size reaches the SQL args.
func TestCorrectness_Related_NegativeSizeReachesSQL(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT m.id, m.title`).
		WithArgs(int64(-5), int64(0)). // LIMIT -5 OFFSET 0 => invalid LIMIT in Postgres
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/related/42?size=-5", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet(), "negative size must be clamped before reaching the SQL LIMIT")
}

// FINDING C4 (swallowed error): SearchManuscripts ignores the error from
// rows.Scan (search_repository.go:74-77). A row whose non-nullable column is
// NULL makes Scan fail, yet the partial (all-zero) row is still appended and
// the repository returns nil error — the API reports 200 with a garbage row.
func TestCorrectness_SearchManuscripts_SwalledScanError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT m.id, m.title`).WillReturnRows(
		sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(nil, "title-x", "desc", "cover", 5, 3, 1, 1, 0, 0, "00:01", 3, "2026-01-01", 5, "u", "n", "a", 1, 0),
	)

	list, err := repo.SearchManuscripts(t.Context(), "", "", 0, 1, 10)
	// BUG pinned: the rows.Scan error is swallowed, so err is nil (it should be
	// a non-nil scan error) and the failed row is emitted as a zero-valued entry.
	assert.NoError(t, err, "rows.Scan error is swallowed: SearchManuscripts returns nil error")
	assert.Len(t, list, 1, "a row that failed to scan must not be emitted")
	if len(list) == 1 {
		assert.Equal(t, int64(0), list[0]["id"], "a row that failed to scan must not be emitted")
	}
}
