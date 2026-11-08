package article

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// toggleInteraction：匿名用户被 RequireUser 拒绝后直接 return（lines 500-502）
func TestToggleLike_UnauthorizedReturns(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/like", 0, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// addCoin：未登录被拒（lines 535-537）
func TestAddCoin_UnauthorizedReturns(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/coin", 0, `{}`)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleComments GET：根评论行 scan 错误时 continue（把 user_id 列设为 []byte）
func TestComments_RootScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "content", "like_count", "created_at", "nickname", "avatar"}).
			AddRow(1, 9, []byte("bad"), 0, "root", 0, nowTime, "小明", "/a.png"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []commentNode `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Empty(t, data.List)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleComments GET：viewer 点赞了某条回复（非根）时把该回复 MyLiked 置 true（lines 702-704）
func TestComments_ReplyMyLikedTrue(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "content", "like_count", "created_at", "nickname", "avatar"}).
			AddRow(1, 9, 4, 0, "root", 0, nowTime, "小明", "/a.png"))
	mock.ExpectQuery(`WHERE c.status=0 AND c.parent_id = ANY\(\$1\)`).
		WithArgs([]int64{1}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "reply_to_user_id", "content", "like_count", "created_at", "nickname", "avatar", "reply_to_name"}).
			AddRow(2, 9, 8, 1, 4, "re", 0, nowTime, "柯洁", "/b.png", "小明"))
	mock.ExpectQuery(`FROM article_comment_likes l JOIN article_comments c ON c.id = l.comment_id WHERE l.user_id=\$1 AND c.article_id=\$2`).
		WithArgs(int64(5), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_id"}).AddRow(2)) // 点赞的是回复 2
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 5, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []commentNode `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Len(t, data.List, 1)
	require.Len(t, data.List[0].Replies, 1)
	assertMyLiked := data.List[0].Replies[0].MyLiked
	require.True(t, assertMyLiked)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleFolders POST：scan-error（RETURNING 列不匹配）触发 500（lines 896-897 是 continue，这里走 exec 错误）
// 实际 896-897 是 GET 中 rows.Scan 出错 continue：塞一个 bad visible 列
func TestFolders_GetScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM favorite_folders f WHERE f.user_id=\$1 ORDER BY f.created_at ASC`).
		WithArgs(int64(7), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "visible", "cnt", "checked"}).
			AddRow(1, "x", []byte("bad"), 0, false))
	w := do(t, h, h.handleFolders, http.MethodGet, "/api/v1/article/folders", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []struct {
		ID int64 `json:"id"`
	}
	mustJSON(t, decode(t, w), &list)
	require.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleMentionSearch：行 scan 出错 continue（lines 1016-1017）
func TestMentionSearch_ScanErrorContinues(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM users WHERE \(nickname ILIKE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "avatar"}).
			AddRow([]byte("bad"), "柯洁", "/b.png"))
	w := do(t, h, h.handleMentionSearch, http.MethodGet, "/api/v1/article/mention-search?keyword=x", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []struct {
		UID int64 `json:"uid"`
	}
	mustJSON(t, decode(t, w), &list)
	require.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

// notifyArticleCommentLike：查询评论失败直接 return（lines 1071-1073）
func TestNotifyArticleCommentLike_QueryErrorReturns(t *testing.T) {
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
		WillReturnError(errDBDown())
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(1))
	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// notifyArticleMentions：用户查询出错 continue（lines 1098-1099）
func TestNotifyArticleMentions_UserQueryErrorSkips(t *testing.T) {
	h, mock := newTestHandler(t)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@柯洁 hi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(19, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WillReturnError(errDBDown())
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@柯洁 hi"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
