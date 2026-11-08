package article

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// collect 的 scan-error continue：给 author nickname 列（倒数第2）塞一个无法转 string 的值。
// articleRowCols 共 27 列；构造 27 个值，将 nickname（第25列, target string）设为 []byte 无法扫描为 string？
// 实际 sqlmock 默认转换 []byte→string 是支持的。改为把 id（int64 target）换成 []byte 使 Scan 报错。
func TestCollect_ScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	// 27 列；把第 1 列 id 设为 []byte("x")，无法转 int64 → Scan 报错 → continue
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.status = 1 AND a.review_status = 1 AND a.visibility = 1 ORDER BY`).
		WithArgs(int64(20), int64(0)).
		WillReturnRows(articleRows().AddRow(
			[]byte("x"), int64(7), "标题", "<p>正文</p>", "正文", "", "", "",
			1, 1, 1, 0, 1, 0, 2, 10, 3, 1, 0, 0, 0,
			nil, nowTime, nowTime, nowTime, "小明", "/a.png"))
	w := do(t, h, h.handleFeed, http.MethodGet, "/api/v1/article/feed", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []articleRow `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Empty(t, data.List)
	require.NoError(t, mock.ExpectationsWereMet())
}

// collectReview 的 scan-error continue（多两列 review_status/review_reason）
func TestCollectReview_ScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1 AND a.status >= 0 ORDER BY`).
		WithArgs(int64(7)).
		WillReturnRows(articleReviewRows().AddRow(
			[]byte("x"), int64(7), "标题", "<p>正文</p>", "正文", "", "", "",
			1, 1, 1, 0, 1, 0, 2, 10, 3, 1, 0, 0, 0,
			nil, nowTime, nowTime, nowTime, "小明", "/a.png",
			0, ""))
	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []articleRow
	mustJSON(t, decode(t, w), &list)
	require.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

// admin list 的 scan-error continue：把 id 列设为无法转 int64 的 []byte
func TestAdminList_ScanErrorContinues(t *testing.T) {
	ah, mock := newAdminTestHandler(t, &fakePerm{adminID: 9, perms: map[string]bool{"review:manage": true}})
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "user_id", "nickname", "avatar", "cover_url", "summary", "word_count",
			"status", "review_status", "review_reason", "publish_time", "created_at", "updated_at",
		}).AddRow([]byte("x"), "t", int64(4), "小明", "/a.png", "", "s", 10, 1, 0, "", "", "", ""))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM articles a WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := doAdmin(t, ah.handleAll, http.MethodGet, "/api/v1/article/admin/all", "9", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []adminArticle `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Empty(t, data.List)
	require.NoError(t, mock.ExpectationsWereMet())
}
