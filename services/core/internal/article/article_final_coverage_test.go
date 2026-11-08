package article

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

var nowTime = time.Now()

// handleSave 覆盖 featured_comment 非 0 → featured=1 分支
func TestHandleSave_FeaturedComment(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, review_status, publish_time\) VALUES`).
		WithArgs(int64(7), "t", "<p>abc</p>", "abc", "", "", "", int64(0), int64(1), int64(1), int64(1), int64(0), int64(0), int64(3), int64(0), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(45))
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"title":"t","content":"<p>abc</p>","featured_comment":1}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleSave 覆盖 visibility 非法值（=2）保持默认 1 分支
func TestHandleSave_InvalidVisibilityFallsBack(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, review_status, publish_time\) VALUES`).
		WithArgs(int64(7), "t", "<p>abc</p>", "abc", "", "", "", int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(3), int64(0), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(46))
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"title":"t","content":"<p>abc</p>","visibility":2}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleComments GET：reply 行 scan 错误时 continue（给 11 列里塞一个非 int 的 reply_to_user_id）
func TestComments_ReplyScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "content", "like_count", "created_at", "nickname", "avatar"}).
			AddRow(1, 9, 4, 0, "root", 0, nowTime, "小明", "/a.png"))
	// reply 行：第 5 列 reply_to_user_id 给一个无法转 int64 的值 → Scan 出错 → continue
	mock.ExpectQuery(`WHERE c.status=0 AND c.parent_id = ANY\(\$1\)`).
		WithArgs([]int64{1}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "reply_to_user_id", "content", "like_count", "created_at", "nickname", "avatar", "reply_to_name"}).
			AddRow(2, 9, 8, 1, "oops", "re", 0, nowTime, "柯洁", "/b.png", "小明"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []commentNode `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Len(t, data.List, 1)
	// reply 被跳过
	require.Empty(t, data.List[0].Replies)
	require.NoError(t, mock.ExpectationsWereMet())
}
