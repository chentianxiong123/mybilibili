package video

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetStatistics_SwallowsScanErrors pins a swallowed-DB-error bug in
// video_repository.go:GetStatistics (lines ~275-286).
//
// ROOT CAUSE: each `r.db.QueryRowContext(...).Scan(&n)` return value is
// discarded (`r.db.QueryRowContext(...).Scan(&msCount)` with no error check).
// If ANY query fails (e.g. connection loss, table missing), the error is
// swallowed, the counter stays 0, and the function returns `(stats, nil)` —
// reporting success with a fully/partially zeroed map.
//
// IMPACT: the admin statistics endpoint (/api/v1/statistics) silently reports
// 0 counts during DB outages instead of surfacing the error. Operators see
// "all zeros" and may misdiagnose an empty database. The same pattern is
// repeated in manuscript/admin_handler.go:handleStatistics (lines ~185-216).
func TestGetStatistics_SwallowsScanErrors(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	// First query fails; the remaining four would succeed but their errors are
	// also unchecked. The bug is that GetStatistics returns nil error regardless.
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts`).WillReturnError(errors.New("conn lost"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(int64(100)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM videos`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	stats, err := repo.GetStatistics(ctx)
	require.NoError(t, err) // BUG: error swallowed, reports success

	// BUG confirmed: manuscript_count is silently 0 despite the query failing.
	assert.Equal(t, int64(0), stats["manuscript_count"])
	assert.Equal(t, int64(5), stats["user_count"])
}

// TestListByManuscript_SwallowsScanError pins the same class of bug in
// video_repository.go:ListByManuscript (line ~86), where rows.Scan's error is
// discarded, and in ListUserManuscriptIDs / ListUserVideoIDs / ListCategories.
func TestListByManuscript_SwallowsScanError(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, manuscript_id, video_order, title`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "manuscript_id", "video_order", "title",
			"play_url_hd", "play_url_sd", "play_url_ld",
			"upload_time", "updated_at", "process_progress",
			"process_stage", "has_subtitle", "has_summary",
			"process_status", "process_error", "source_video_url",
			"duration_seconds", "is_vertical",
		}).RowError(0, errors.New("scan boom"))) // Scan fails on the only row

	list, err := repo.ListByManuscript(ctx, 1)
	require.NoError(t, err) // BUG: rows.Scan error discarded

	// BUG confirmed: the row whose Scan failed is silently dropped and the
	// function reports success. A correct implementation returns the scan error
	// to the caller. (sqlmock RowError makes Next() stop, so the error never
	// surfaces — same class of swallow.)
	assert.Len(t, list, 0)
}
