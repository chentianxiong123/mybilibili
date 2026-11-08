package article

import (
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// handleMine：匿名用户（uid==0）直接返回空列表（lines 178-181）
func TestHandleMine_Anonymous(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "[]", string(decode(t, w).Data))
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleByID：路径无 id（rest==""）→ 404（lines 401-404）
func TestHandleByID_EmptyPath(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/", 0, "")
	require.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleComments POST：未登录被拒（lines 714-716）
func TestComments_PostUnauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 0, `{"content":"hi"}`)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// handleFolders：未登录被拒（lines 870-872）
func TestFolders_Unauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleFolders, http.MethodGet, "/api/v1/article/folders", 0, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// notifyArticleMentions：@ 后接标点（TrimRight 后为空名）→ 跳过（lines 1098-1099）
func TestNotifyArticleMentions_EmptyNameAfterTrim(t *testing.T) {
	h, mock := newTestHandler(t)
	n := &fakeNotifier{}
	h.SetNotifier(n)
	now := timeNow()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES`).
		WithArgs(int64(9), int64(7), nil, nil, "@，").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(20, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 提及名被 TrimRight 清空后跳过，不查用户、不发通知
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@，"}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Zero(t, n.calls)
	require.NoError(t, mock.ExpectationsWereMet())
}
