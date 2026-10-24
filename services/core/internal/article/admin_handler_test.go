package article

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePerm struct {
	adminID int64
	perms   map[string]bool
}

func (f *fakePerm) CheckPermission(_ *http.Request, perm string) (int64, bool) {
	return f.adminID, f.perms[perm]
}

func newAdminTestHandler(t *testing.T, perm *fakePerm) (*AdminHandler, sqlmock.Sqlmock) {
	t.Helper()
	h, mock := newTestHandler(t)
	ah := NewAdminHandler(h.db)
	ah.SetPermChecker(perm)
	return ah, mock
}

func TestArticleAdmin_PendingList(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 1, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id\s+WHERE a.status = 1 AND a.review_status = 0 ORDER BY`).
		WithArgs(int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}).AddRow(8, "待审专栏", 4, "小明", "/a.png", "", "摘要", 120, 1, 0, "", "", "2026-10-22T10:00:00Z", "2026-10-22T10:00:00Z"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status = 1 AND a.review_status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := do(t, nil, func(w http.ResponseWriter, r *http.Request) { ah.handlePending(w, r) },
		http.MethodGet, "/api/v1/article/admin/pending", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List  []adminArticle `json:"list"`
		Total int            `json:"total"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Len(t, data.List, 1)
	assert.Equal(t, "待审专栏", data.List[0].Title)
	assert.Equal(t, 1, data.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleAdmin_NoPermission(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 2, perms: map[string]bool{}})
	w := do(t, nil, ah.requirePerm("review:manage", ah.handlePending),
		http.MethodGet, "/api/v1/article/admin/pending", 0, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleAdmin_Approve(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`UPDATE articles SET status = 1, review_status = 1, review_reason = '', review_time = NOW\(\),\s+reviewer_id = \$1, publish_time = COALESCE\(publish_time, NOW\(\)\), updated_at = NOW\(\)\s+WHERE id = \$2`).
		WithArgs(int64(9), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := do(t, nil, func(w http.ResponseWriter, r *http.Request) { ah.handleReview(w, r) },
		http.MethodPost, "/api/v1/article/admin/review/8", 0, `{"approved":true}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, float64(8), data["article_id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleAdmin_RejectRequiresReason(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := do(t, nil, func(w http.ResponseWriter, r *http.Request) { ah.handleReview(w, r) },
		http.MethodPost, "/api/v1/article/admin/review/8", 0, `{"approved":false}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleAdmin_Reject(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`UPDATE articles SET status = 2, review_status = 2, review_reason = \$1, review_time = NOW\(\),\s+reviewer_id = \$2, updated_at = NOW\(\)\s+WHERE id = \$3`).
		WithArgs("内容不合规", int64(9), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := do(t, nil, func(w http.ResponseWriter, r *http.Request) { ah.handleReview(w, r) },
		http.MethodPost, "/api/v1/article/admin/review/8", 0, `{"approved":false,"reason":"内容不合规"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleAdmin_Statistics(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	for _, n := range []int{10, 3, 6, 1, 2} {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
	}
	w := do(t, nil, func(w http.ResponseWriter, r *http.Request) { ah.handleStatistics(w, r) },
		http.MethodGet, "/api/v1/article/admin/statistics", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, float64(3), data["pending"])
	assert.Equal(t, float64(1), data["rejected"])
	require.NoError(t, mock.ExpectationsWereMet())
}
