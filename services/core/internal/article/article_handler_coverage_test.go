package article

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// doWithAdmin 构造带 X-Admin-Id 头的请求（部分 admin 分支需要）。
func doWithAdmin(t *testing.T, fn http.HandlerFunc, method, target, adminID string, uid int64, body string) *httptest.ResponseRecorder {
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

// --- Register ---

func TestHandlerRegister(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)
	// 触发一次即可证明 Register 注册的 handler 生效（覆盖 Register 本身）。
	req := httptest.NewRequest(http.MethodGet, "/api/v1/article/mention-search", nil)
	mux.ServeHTTP(httptest.NewRecorder(), req)
}

// --- handleDrafts 查询失败 ---

func TestHandleDrafts_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1 AND a.status = 0`).
		WithArgs(int64(7), "").
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleDrafts, http.MethodGet, "/api/v1/article/drafts", 7, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleMine status=all / 非法 status 走默认分支 + 查询失败 ---

func TestHandleMine_AllStatus(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1 AND a.status >= 0 ORDER BY`).
		WithArgs(int64(7)).
		WillReturnRows(articleReviewRows())
	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine?status=all", 7, "")
	assert.Equal(t, http.StatusOK, w.Code)
	var list []articleRow
	mustJSON(t, decode(t, w), &list)
	assert.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMine_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine", 7, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMine_StatusZero(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`WHERE a.user_id = \$1 AND a.status = 0 ORDER BY`).
		WithArgs(int64(7)).
		WillReturnRows(articleReviewRows())
	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine?status=0", 7, "")
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleUserArticles uid<=0 与查询失败 ---

func TestHandleUserArticles_ZeroUID(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleUserArticles, http.MethodGet, "/api/v1/article/user/0", 0, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserArticles_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`WHERE a.user_id = \$1 AND a.status = 1 AND a.review_status = 1`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleUserArticles, http.MethodGet, "/api/v1/article/user/9", 0, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleByID 各子路由 ---

func TestHandleByID_View(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE articles SET view_count = view_count \+ 1 WHERE id = \$1 AND status = 1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/view", 0, "")
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	// PUT 不在方法表中，直接 405，不触发文章查询
	w := do(t, h, h.handleByID, http.MethodPut, "/api/v1/article/5", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, 405, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_UnknownSubPath_GET(t *testing.T) {
	// 未知子路径 + GET：未知 case 被忽略后按 GET 正常取详情
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(addArticleReviewRow(articleReviewRows(), 5, 4, 1, 1, 1, ""))
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/5/unknown", 0, "")
	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_GetNotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = \$1`).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrNoRows)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/5", 0, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleByID_DeleteExecError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE articles SET status = -1`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleByID, http.MethodDelete, "/api/v1/article/5", 7, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- toggleInteraction 分支 ---

func TestToggleFavorite_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/favorite", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestToggleLike_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/like", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestToggleFavorite_Remove(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9), "favorite").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`DELETE FROM user_interactions WHERE user_id=\$1`).
		WithArgs(int64(7), int64(9), "favorite").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET favorite_count = GREATEST\(favorite_count - 1, 0\)`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT favorite_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"favorite_count"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/favorite", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Active bool  `json:"active"`
		Count  int64 `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.False(t, data.Active)
	assert.Equal(t, int64(0), data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- addCoin ---

func TestAddCoin_DefaultCount(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(count,0\) FROM article_coins`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO article_coins \(article_id, user_id, count\) VALUES \(\$1,\$2,\$3\)`).
		WithArgs(int64(9), int64(7), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET coin_count = coin_count \+ \$1 WHERE id=\$2`).
		WithArgs(1, int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT coin_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(1))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/coin", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		MyCoin int `json:"my_coin"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, 1, data.MyCoin)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCoin_PartialRemaining(t *testing.T) {
	// 已有 1 枚，再投 5 枚 → 实际只加 1 枚（封顶 2）
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(count,0\) FROM article_coins`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO article_coins \(article_id, user_id, count\) VALUES \(\$1,\$2,\$3\)`).
		WithArgs(int64(9), int64(7), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET coin_count = coin_count \+ \$1 WHERE id=\$2`).
		WithArgs(1, int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT coin_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/coin", 7, `{"count":5}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		MyCoin int `json:"my_coin"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, 2, data.MyCoin)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleComments GET 空列表 / 查询失败 ---

func TestComments_ListEmpty(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "content", "like_count", "created_at", "nickname", "avatar"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List  []commentNode `json:"list"`
		Total int           `json:"total"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Empty(t, data.List)
	assert.Equal(t, 0, data.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_ListQueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 0, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleComments POST：超长回复父评论不存在 / 入库失败 / 自回复 / 通知 ---

func TestComments_PostReplyParentNotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`SELECT user_id FROM article_comments WHERE id=\$1 AND article_id=\$2 AND status=0`).
		WithArgs(int64(100), int64(9)).
		WillReturnError(sql.ErrNoRows)

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"re","parent_id":100}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "回复的评论不存在", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostInvalidBody(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostInsertError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "hi").
		WillReturnError(errors.New("db down"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"hi"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostSelfReplyNoNotify(t *testing.T) {
	// 回复自己的评论：replyTo == senderID 时不发送通知
	h, mock := newTestHandler(t)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`SELECT user_id FROM article_comments WHERE id=\$1 AND article_id=\$2 AND status=0`).
		WithArgs(int64(100), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), int64(100), int64(7), "re").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(12, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 自回复不触发 reply 通知（receiverID==senderID 直接 return），也不触发 mention 通知
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname"}).AddRow("小明"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"re","parent_id":100}`)
	require.Equal(t, http.StatusOK, w.Code)
	var c commentNode
	mustJSON(t, decode(t, w), &c)
	assert.Equal(t, int64(100), c.ParentID)
	assert.Equal(t, int64(7), c.ReplyToUserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostReplyExplicitReplyTo(t *testing.T) {
	// 显式指定 reply_to_user_id 且不同于父评论作者
	h, mock := newTestHandler(t)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`SELECT user_id FROM article_comments WHERE id=\$1 AND article_id=\$2 AND status=0`).
		WithArgs(int64(100), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(300))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), int64(100), int64(500), "re").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(13, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO messages \(sender_id, receiver_id, content, message_type, target_id, comment_id, is_read, created_at\)`).
		WithArgs(int64(7), int64(500), "回复了你的评论：re", int64(9), int64(13)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(500)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname"}).AddRow("路人"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7,
		`{"content":"re","parent_id":100,"reply_to_user_id":500}`)
	require.Equal(t, http.StatusOK, w.Code)
	var c commentNode
	mustJSON(t, decode(t, w), &c)
	assert.Equal(t, int64(500), c.ReplyToUserID)
	assert.Equal(t, "路人", c.ReplyToName)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleComments POST：通知走 notifier 的分支 ---

func TestComments_PostMentionUsesNotifier(t *testing.T) {
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@柯洁 hi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(14, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("柯洁").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@柯洁 hi"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, n.calls)
	assert.Equal(t, int64(8), n.receiver)
	assert.Equal(t, int32(3), n.typ)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleCommentSub 分支 ---

func TestCommentSub_InvalidID(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleCommentSub, http.MethodGet, "/api/v1/article/comment/abc", 0, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_DeleteUnauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleCommentSub, http.MethodDelete, "/api/v1/article/comment/11", 0, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_DeleteQueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id, article_id FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnError(sql.ErrNoRows)
	w := do(t, h, h.handleCommentSub, http.MethodDelete, "/api/v1/article/comment/11", 7, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "评论不存在", decode(t, w).Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_LikeUnauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 0, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_LikeMethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleCommentSub, http.MethodGet, "/api/v1/article/comment/11/like", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_LikeRemove(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM article_comment_likes WHERE comment_id=\$1 AND user_id=\$2\)`).
		WithArgs(int64(11), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`DELETE FROM article_comment_likes WHERE comment_id=\$1 AND user_id=\$2`).
		WithArgs(int64(11), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE article_comments SET like_count = GREATEST\(like_count - 1, 0\) WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(0))

	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Active bool `json:"active"`
		Count  int  `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.False(t, data.Active)
	assert.Equal(t, 0, data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentSub_LikeMethodNotAllowedForDeletePath(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleCommentSub, http.MethodGet, "/api/v1/article/comment/11", 0, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleFolders 分支 ---

func TestFolders_AnnotationQueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM favorite_folders f WHERE f.user_id=\$1 ORDER BY f.created_at ASC`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleFolders, http.MethodGet, "/api/v1/article/folders", 7, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_PostMethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleFolders, http.MethodDelete, "/api/v1/article/folders", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_PostInvalidJSON(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleFolders, http.MethodPost, "/api/v1/article/folders", 7, `{`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_PostExecError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO favorite_folders \(user_id, name, visible\) VALUES \(\$1,\$2,1\) RETURNING id`).
		WithArgs(int64(7), "新夹").
		WillReturnError(errors.New("unique violation"))
	w := do(t, h, h.handleFolders, http.MethodPost, "/api/v1/article/folders", 7, `{"name":"新夹"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- collectFolders 分支 ---

func TestCollectFolders_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/collect", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_Unauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 0, `{}`)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_InvalidBody(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_ArticleQueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT status FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrNoRows)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{"adds":[1]}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_RemoveFolder(t *testing.T) {
	// 有删除收藏夹：favorite_count 回退，user_interactions 删除
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT status FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))
	mock.ExpectExec(`DELETE FROM favorite_folder_articles WHERE folder_id=\$1 AND article_id=\$2.*`).
		WithArgs(int64(1), int64(9), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folder_articles ffa`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type='favorite'\)`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`DELETE FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type='favorite'`).
		WithArgs(int64(7), int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET favorite_count = GREATEST\(favorite_count - 1, 0\) WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT favorite_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"favorite_count"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{"removes":[1]}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Favorited bool  `json:"favorited"`
		Count     int64 `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.False(t, data.Favorited)
	assert.Equal(t, int64(0), data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_FolderNotOwned(t *testing.T) {
	// 收藏夹不属于该用户 → 跳过插入
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT status FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders WHERE id=\$1 AND user_id=\$2`).
		WithArgs(int64(1), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folder_articles ffa`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type='favorite'\)`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`SELECT favorite_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"favorite_count"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{"adds":[1]}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Favorited bool `json:"favorited"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.False(t, data.Favorited)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleMentionSearch 查询失败 ---

func TestMentionSearch_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM users WHERE \(nickname ILIKE`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleMentionSearch, http.MethodGet, "/api/v1/article/mention-search?keyword=x", 0, "")
	// 查询失败时退回空列表
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", string(decode(t, w).Data))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- handleSave 分支 ---

func TestHandleSave_UpdateExecError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM articles WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
	mock.ExpectExec(`UPDATE articles SET title=\$1`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"id":5,"title":"t"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_CreateInsertError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, review_status, publish_time\) VALUES`).
		WillReturnError(errors.New("db down"))
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"title":"t","content":"<p>body</p>"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_UpdateNoPublish(t *testing.T) {
	// 更新草稿（publish=false）：setPublish 为空串分支
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM articles WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
	mock.ExpectExec(`UPDATE articles SET title=\$1, content=\$2, summary=\$3, cover_url=\$4, topics=\$5,\s+collection_name=\$6, visibility=\$7, comment_enabled=\$8, featured_comment=\$9,\s+is_original=\$10, ai_declaration=\$11, word_count=\$12, updated_at=NOW\(\)\s+WHERE id=\$13 AND user_id=\$14`).
		WithArgs("t", "<p>body</p>", "body", "", "", "", int64(1), int64(1), int64(0), int64(0), int64(0), int64(4), int64(5), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"id":5,"title":"t","content":"<p>body</p>"}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID        int64 `json:"id"`
		Published bool  `json:"published"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(5), data.ID)
	assert.False(t, data.Published)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_WorldCountOverride(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, review_status, publish_time\) VALUES`).
		WithArgs(int64(7), "t", "<p>abc</p>", "abc", "", "", "", int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(99), int64(0), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(44))
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"title":"t","content":"<p>abc</p>","word_count":99}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(44), data.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- notify* 未覆盖分支 ---

func TestNotifyArticleLike_SelfNoNotify(t *testing.T) {
	h, mock := newTestHandler(t)
	// 自己赞自己的专栏：owner == senderID 直接 return，不发通知
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO user_interactions`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET like_count = like_count \+ 1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id, COALESCE\(title,''\) FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "title"}).AddRow(7, "我的"))
	mock.ExpectQuery(`SELECT like_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleLike_QueryErrorNoNotify(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO user_interactions`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET like_count = like_count \+ 1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// notifyArticleLike 查询失败时直接 return
	mock.ExpectQuery(`SELECT user_id, COALESCE\(title,''\) FROM articles`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT like_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleCommentLike_SelfNoNotify(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM article_comment_likes WHERE comment_id=\$1 AND user_id=\$2\)`).
		WithArgs(int64(11), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO article_comment_likes \(comment_id, user_id\) VALUES \(\$1,\$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(11), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE article_comments SET like_count = like_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 评论作者是自己 → 不通知
	mock.ExpectQuery(`SELECT user_id, article_id, COALESCE\(content,''\) FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "article_id", "content"}).AddRow(7, 9, "我的评论"))
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))

	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleCommentLike_OwnerZeroNoNotify(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM article_comment_likes WHERE comment_id=\$1 AND user_id=\$2\)`).
		WithArgs(int64(11), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO article_comment_likes \(comment_id, user_id\) VALUES \(\$1,\$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(11), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE article_comments SET like_count = like_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id, article_id, COALESCE\(content,''\) FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "article_id", "content"}).AddRow(int64(0), int64(9), "无主"))
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))

	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleCommentLike_UsesNotifier(t *testing.T) {
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM article_comment_likes WHERE comment_id=\$1 AND user_id=\$2\)`).
		WithArgs(int64(11), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO article_comment_likes \(comment_id, user_id\) VALUES \(\$1,\$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(11), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE article_comments SET like_count = like_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT user_id, article_id, COALESCE\(content,''\) FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "article_id", "content"}).AddRow(300, 9, "原评论"))
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))

	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, n.calls)
	assert.Equal(t, int64(300), n.receiver)
	assert.Equal(t, int32(6), n.typ)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleReply_UsesNotifier(t *testing.T) {
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`SELECT user_id FROM article_comments WHERE id=\$1 AND article_id=\$2 AND status=0`).
		WithArgs(int64(100), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(300))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), int64(100), int64(300), "re").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(12, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(300)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname"}).AddRow("楼主"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"re","parent_id":100}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, n.calls)
	assert.Equal(t, int32(2), n.typ)
	assert.Equal(t, int64(300), n.receiver)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleMentions_SelfMentionSkipped(t *testing.T) {
	// @自己 → 跳过
	h, mock := newTestHandler(t)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@小明 hi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(15, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("小明").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7)) // 自己
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@小明 hi"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleMentions_UnknownUserSkipped(t *testing.T) {
	// @不存在的人 → 跳过
	h, mock := newTestHandler(t)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@路人甲 hi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(16, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("路人甲").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@路人甲 hi"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleMentions_PunctuationTrim(t *testing.T) {
	// 尾随全角逗号被 TrimRight 剥掉（正则捕获 "柯洁，"，去尾得 "柯洁"）
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@柯洁，").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(17, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("柯洁").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@柯洁，"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, n.calls)
	assert.Equal(t, int64(8), n.receiver)
	assert.Equal(t, int32(3), n.typ)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotifyArticleMentions_DedupAndSelfSkip(t *testing.T) {
	// 多个提及：自提及跳过，重复提及去重（只通知一次）
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@小明 @柯洁 @柯洁 @路人").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(18, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 依次查：小明(自己,跳过)、柯洁、柯洁(重复,跳过)、路人(查不到,跳过)
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("小明").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("柯洁").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("柯洁").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("路人").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@小明 @柯洁 @柯洁 @路人"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, n.calls)
	assert.Equal(t, int64(8), n.receiver)
	assert.Equal(t, int32(3), n.typ)
	require.NoError(t, mock.ExpectationsWereMet())
}

// timeNow 封装 time.Now()，便于测试中作为 created_at 的 driver.Value。
func timeNow() time.Time {
	return time.Now()
}
