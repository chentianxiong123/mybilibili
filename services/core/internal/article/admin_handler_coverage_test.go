package article

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doAdmin(t *testing.T, fn http.HandlerFunc, method, target, adminID string, uid int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	if uid != 0 {
		r.Header.Set("X-User-Id", strconv.FormatInt(uid, 10))
	}
	if adminID != "" {
		r.Header.Set("X-Admin-Id", adminID)
	}
	w := httptest.NewRecorder()
	fn(w, r)
	return w
}

func TestAdminHandlerRegister(t *testing.T) {
	h, _ := newTestHandler(t)
	ah := NewAdminHandler(h.db)
	ah.SetPermChecker(&fakePerm{adminID: 1, perms: map[string]bool{"review:manage": true}})
	mux := http.NewServeMux()
	ah.Register(mux)
	// 触发一次 statistics，证明 Register 生效
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/article/admin/statistics", nil))
}

func TestAdminRequirePerm_NilChecker(t *testing.T) {
	h, mock := newTestHandler(t)
	ah := NewAdminHandler(h.db) // 无 perm checker
	fn := ah.requirePerm("review:manage", ah.handlePending)
	w := doAdmin(t, fn, http.MethodGet, "/api/v1/article/admin/pending", "", 0, "")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "权限校验未配置", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRequirePerm_DeniedWithAdminID(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 5, perms: map[string]bool{}})
	fn := ah.requirePerm("review:manage", ah.handlePending)
	w := doAdmin(t, fn, http.MethodGet, "/api/v1/article/admin/pending", "5", 0, "")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "forbidden", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRequirePerm_DeniedAnonymous(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 0, perms: map[string]bool{}})
	fn := ah.requirePerm("review:manage", ah.handlePending)
	w := doAdmin(t, fn, http.MethodGet, "/api/v1/article/admin/pending", "", 0, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "unauthorized", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminPending_MethodNotAllowed(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := doAdmin(t, ah.handlePending, http.MethodPost, "/api/v1/article/admin/pending", "9", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminAll_MethodNotAllowed(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := doAdmin(t, ah.handleAll, http.MethodPost, "/api/v1/article/admin/all", "9", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminAll_ListOK(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id\s+WHERE a.status <> -1 ORDER BY`).
		WithArgs(int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status <> -1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleAll, http.MethodGet, "/api/v1/article/admin/all", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List  []adminArticle `json:"list"`
		Total int            `json:"total"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Empty(t, data.List)
	assert.Equal(t, 0, data.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminAll_StatusFilterAndKeyword(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id\s+WHERE a.status = \$1 AND a.title ILIKE \$2 ORDER BY`).
		WithArgs(int64(1), "%专栏%", int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status = \$1 AND a.title ILIKE \$2`).
		WithArgs(int64(1), "%专栏%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleAll, http.MethodGet, "/api/v1/article/admin/all?status=1&keyword=专栏", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminAll_InvalidStatusIgnoresFilter(t *testing.T) {
	// 非法 status 值：Atoi 失败 → where 保持 a.status <> -1
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id\s+WHERE a.status <> -1 ORDER BY`).
		WithArgs(int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status <> -1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleAll, http.MethodGet, "/api/v1/article/admin/all?status=abc", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminAll_SizeClamp(t *testing.T) {
	// size 越界时回落到默认 20；页码保留
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id.*WHERE a.status <> -1 ORDER BY`).
		WithArgs(int64(20), int64(20)). // page=2, size clamp to 20 → offset=20
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status <> -1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleAll, http.MethodGet, "/api/v1/article/admin/all?page=2&size=500", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Page int `json:"page"`
		Size int `json:"size"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, 2, data.Page)
	assert.Equal(t, 20, data.Size)
	require.NoError(t, mock.ExpectationsWereMet())
}

// 直接测 list() 的参数归一化：size<=0 与 size>100 都回落 20（覆盖 clamp 分支 84-86）
// 说明：size=0 因 ParsePageParams 已保证 >=1 而不可达（死代码），size>50 被夹回 20。
func TestAdminList_SizeNormalization(t *testing.T) {
	cases := []struct {
		name   string
		target string
		want   int
	}{
		{"default", "/api/v1/article/admin/all", 20},
		{"size=101", "/api/v1/article/admin/all?size=101", 20},
		{"size=50", "/api/v1/article/admin/all?size=50", 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
			mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id.*WHERE a.status <> -1 ORDER BY`).
				WithArgs(int64(tc.want), int64(0)).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
					"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
				}))
			mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE a.status <> -1`).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			w := doAdmin(t, ah.handleAll, http.MethodGet, tc.target, "9", 0, "")
			require.Equal(t, http.StatusOK, w.Code)
			var data struct {
				Size int `json:"size"`
			}
			mustJSON(t, decode(t, w), &data)
			assert.Equal(t, tc.want, data.Size)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAdminList_QueryError(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id`).
		WillReturnError(errDBDown())
	w := doAdmin(t, ah.handlePending, http.MethodGet, "/api/v1/article/admin/pending", "9", 0, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_MethodNotAllowed(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := doAdmin(t, ah.handleReview, http.MethodGet, "/api/v1/article/admin/review/8", "9", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_InvalidID(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := doAdmin(t, ah.handleReview, http.MethodPost, "/api/v1/article/admin/review/abc", "9", 0, `{"approved":true}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "invalid id", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_ArticleNotFound(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleReview, http.MethodPost, "/api/v1/article/admin/review/8", "9", 0, `{"approved":true}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "专栏不存在", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_ExecError(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`UPDATE articles SET status = 1, review_status = 1`).
		WillReturnError(errDBDown())
	w := doAdmin(t, ah.handleReview, http.MethodPost, "/api/v1/article/admin/review/8", "9", 0, `{"approved":true}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "操作失败", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_RejectReasonTruncated(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`UPDATE articles SET status = 2, review_status = 2, review_reason = \$1`).
		WithArgs(strings.Repeat("字", 500), int64(9), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w := doAdmin(t, ah.handleReview, http.MethodPost, "/api/v1/article/admin/review/8", "9", 0,
		`{"approved":false,"reason":"`+strings.Repeat("字", 600)+`"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminReview_ApproveOK(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles WHERE id = \$1`).
		WithArgs(int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`UPDATE articles SET status = 1, review_status = 1`).
		WithArgs(int64(9), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w := doAdmin(t, ah.handleReview, http.MethodPost, "/api/v1/article/admin/review/8", "9", 0, `{"approved":true}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, "审核通过", data["message"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminStatistics_MethodNotAllowed(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	w := doAdmin(t, ah.handleStatistics, http.MethodPost, "/api/v1/article/admin/statistics", "9", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminStatistics_QueryErrorSwallowed(t *testing.T) {
	// statistics 的查询错误被静默吞掉（_ =）：返回 0 值
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	for i := 0; i < 5; i++ {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles`).
			WillReturnError(errDBDown())
	}
	w := doAdmin(t, ah.handleStatistics, http.MethodGet, "/api/v1/article/admin/statistics", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, float64(0), data["total"])
	assert.Equal(t, float64(0), data["pending"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func errDBDown() error { return errors.New("db down") }
