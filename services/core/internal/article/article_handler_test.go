package article

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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

// testConverter 让 sqlmock 接受 []int64（生产环境 lib/pq 会把切片转成数组字面量，
// 用于评论列表的 parent_id = ANY($1)）。
type testConverter struct{}

func (testConverter) ConvertValue(v any) (driver.Value, error) {
	if _, ok := v.([]int64); ok {
		return v, nil
	}
	return driver.DefaultParameterConverter.ConvertValue(v)
}

func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
		sqlmock.ValueConverterOption(testConverter{}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewHandler(db), mock
}

var articleRowCols = []string{
	"id", "user_id", "title", "content", "summary", "cover_url", "topics",
	"collection_name", "status", "visibility", "comment_enabled", "featured_comment",
	"is_original", "ai_declaration", "word_count", "view_count", "like_count",
	"favorite_count", "comment_count", "coin_count", "share_count", "scheduled_at",
	"publish_time", "created_at", "updated_at", "nickname", "avatar",
}

func articleRows() *sqlmock.Rows { return sqlmock.NewRows(articleRowCols) }

func addArticleRow(rows *sqlmock.Rows, id, uid int64, status, visibility int) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, uid, "标题", "<p>正文</p>", "正文", "", "", "",
		status, visibility, 1, 0, 1, 0, 2, 10, 3, 1, 0, 0, 0,
		nil, now, now, now, "小明", "/a.png",
	)
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func do(t *testing.T, h *Handler, fn http.HandlerFunc, method, target string, uid int64, body string) *httptest.ResponseRecorder {
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
	w := httptest.NewRecorder()
	fn(w, r)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e
}

func mustJSON(t *testing.T, e envelope, dst any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(e.Data, dst))
}

func TestHelpers(t *testing.T) {
	assert.Equal(t, "hello world", htmlToText("<p>hello</p> <b>world</b>"))
	assert.Equal(t, "a & b", htmlToText("a &amp; b"))
	assert.Equal(t, 5, runeCount("<p>hello</p>"))
	assert.Equal(t, "abc", excerpt("abcdef", 3))
	assert.Equal(t, "short", excerpt("short", 100))
	assert.Equal(t, 1, boolInt(true))
	assert.Equal(t, 0, boolInt(false))
}

func TestHandleFeed_OK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.status = 1 AND a.visibility = 1 ORDER BY`).
		WithArgs(int64(20), int64(0)).
		WillReturnRows(addArticleRow(articleRows(), 2, 4, 1, 1))

	w := do(t, h, h.handleFeed, http.MethodGet, "/api/v1/article/feed", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	e := decode(t, w)
	var data struct {
		List []articleRow `json:"list"`
		Page int          `json:"page"`
		Size int          `json:"size"`
	}
	mustJSON(t, e, &data)
	require.Len(t, data.List, 1)
	assert.Equal(t, int64(2), data.List[0].ID)
	assert.Equal(t, "小明", data.List[0].AuthorNickname)
	assert.Equal(t, 1, data.Page)
	assert.Equal(t, 20, data.Size)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFeed_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u`).WillReturnError(errors.New("db down"))

	w := do(t, h, h.handleFeed, http.MethodGet, "/api/v1/article/feed", 0, "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 500, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDrafts_Anonymous(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleDrafts, http.MethodGet, "/api/v1/article/drafts", 0, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", string(decode(t, w).Data))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDrafts_LoggedIn(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1 AND a.status = 0 AND \(\$2 = '' OR a.title ILIKE`).
		WithArgs(int64(7), "go").
		WillReturnRows(addArticleRow(articleRows(), 3, 7, 0, 1))

	w := do(t, h, h.handleDrafts, http.MethodGet, "/api/v1/article/drafts?keyword=go", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []articleRow
	mustJSON(t, decode(t, w), &list)
	require.Len(t, list, 1)
	assert.Equal(t, int64(3), list[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleMine_StatusFilter(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.user_id = \$1 AND a.status = 1 ORDER BY`).
		WithArgs(int64(7)).
		WillReturnRows(addArticleRow(articleRows(), 4, 7, 1, 1))

	w := do(t, h, h.handleMine, http.MethodGet, "/api/v1/article/mine?status=1", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []articleRow
	mustJSON(t, decode(t, w), &list)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserArticles_InvalidUID(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleUserArticles, http.MethodGet, "/api/v1/article/user/abc", 0, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 400, decode(t, w).Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserArticles_OK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`WHERE a.user_id = \$1 AND a.status = 1 AND \(a.visibility = 1 OR \$2 = a.user_id\)`).
		WithArgs(int64(9), int64(5), int64(20), int64(0)).
		WillReturnRows(addArticleRow(articleRows(), 5, 9, 1, 1))

	w := do(t, h, h.handleUserArticles, http.MethodGet, "/api/v1/article/user/9", 5, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List []articleRow `json:"list"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Len(t, data.List, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleSave, http.MethodGet, "/api/v1/article/save", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_Anonymous(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 0, `{"title":"t"}`)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_PublishValidation(t *testing.T) {
	t.Run("empty title", func(t *testing.T) {
		h, mock := newTestHandler(t)
		w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"publish":true,"content":"<p>x</p>"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "标题不能为空", decode(t, w).Message)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("empty content", func(t *testing.T) {
		h, mock := newTestHandler(t)
		w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"publish":true,"title":"t","content":"<p> </p>"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "正文不能为空", decode(t, w).Message)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("invalid body", func(t *testing.T) {
		h, mock := newTestHandler(t)
		w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHandleSave_CreateDraft(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, publish_time\) VALUES`).
		WithArgs(int64(7), "hi", "<p>hello</p>", "hello", "", "", "", int64(0), int64(1), int64(1), int64(0), int64(0), int64(0), int64(5), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"title":"  hi  ","content":"<p>hello</p>"}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID        int64 `json:"id"`
		Published bool  `json:"published"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(42), data.ID)
	assert.False(t, data.Published)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_CreatePublish(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO articles \(user_id, title, content, summary, cover_url, topics, collection_name, status, visibility, comment_enabled, featured_comment, is_original, ai_declaration, word_count, publish_time\) VALUES`).
		WithArgs(int64(7), "t", "<p>abc</p>", "abc", "", "", "", int64(1), int64(0), int64(0), int64(0), int64(1), int64(1), int64(3), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(43))

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7,
		`{"title":"t","content":"<p>abc</p>","publish":true,"visibility":0,"comment_enabled":0,"is_original":1,"ai_declaration":1}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID        int64 `json:"id"`
		Published bool  `json:"published"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(43), data.ID)
	assert.True(t, data.Published)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_UpdateNotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM articles WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrNoRows)

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"id":5,"title":"t"}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_UpdateForbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM articles WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"id":5,"title":"t"}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSave_UpdatePublish(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id FROM articles WHERE id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
	mock.ExpectExec(`UPDATE articles SET title=\$1, content=\$2, summary=\$3.*word_count=\$12, status = 1, publish_time = COALESCE\(publish_time, NOW\(\)\), updated_at=NOW\(\) WHERE id=\$13 AND user_id=\$14`).
		WithArgs("t", "<p>body</p>", "body", "", "", "", int64(1), int64(1), int64(0), int64(0), int64(0), int64(4), int64(5), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := do(t, h, h.handleSave, http.MethodPost, "/api/v1/article/save", 7, `{"id":5,"title":"t","content":"<p>body</p>","publish":true}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID        int64 `json:"id"`
		Published bool  `json:"published"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(5), data.ID)
	assert.True(t, data.Published)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_GetHiddenFromOther(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(addArticleRow(articleRows(), 5, 4, 0, 1))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/5", 7, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_GetOwnerDraft(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(addArticleRow(articleRows(), 5, 4, 0, 1))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type='like'\), EXISTS`).
		WithArgs(int64(4), int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"liked", "favorited"}).AddRow(true, false))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(count\),0\) FROM article_coins WHERE article_id=\$1 AND user_id=\$2`).
		WithArgs(int64(5), int64(4)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/5", 4, "")
	require.Equal(t, http.StatusOK, w.Code)
	var a articleRow
	mustJSON(t, decode(t, w), &a)
	assert.Equal(t, int64(5), a.ID)
	assert.True(t, a.MyLiked)
	assert.False(t, a.MyFavorited)
	assert.Equal(t, 2, a.MyCoin)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_GetAnonymousPublic(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM articles a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = \$1`).
		WithArgs(int64(5)).
		WillReturnRows(addArticleRow(articleRows(), 5, 4, 1, 1))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/5", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var a articleRow
	mustJSON(t, decode(t, w), &a)
	assert.Equal(t, "<p>正文</p>", a.Content)
	assert.False(t, a.MyLiked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_InvalidID(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/abc", 0, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_DeleteUnauthorized(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodDelete, "/api/v1/article/5", 0, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_DeleteNotOwner(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE articles SET status = -1, updated_at = NOW\(\) WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(5), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	w := do(t, h, h.handleByID, http.MethodDelete, "/api/v1/article/5", 7, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleArticle_DeleteOK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE articles SET status = -1, updated_at = NOW\(\) WHERE id = \$1 AND user_id = \$2`).
		WithArgs(int64(5), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := do(t, h, h.handleByID, http.MethodDelete, "/api/v1/article/5", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestToggleLike_Add(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type=\$3\)`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO user_interactions \(user_id, target_type, target_id, interaction_type\) VALUES \(\$1, 'article', \$2, \$3\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET like_count = like_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(4))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Active bool  `json:"active"`
		Count  int64 `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.True(t, data.Active)
	assert.Equal(t, int64(4), data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestToggleLike_Remove(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`DELETE FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type=\$3`).
		WithArgs(int64(7), int64(9), "like").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET like_count = GREATEST\(like_count - 1, 0\) WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT like_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/like", 7, "")
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

func TestToggleFavorite_UsesFavoriteCounter(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9), "favorite").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO user_interactions`).
		WithArgs(int64(7), int64(9), "favorite").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET favorite_count = favorite_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT favorite_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"favorite_count"}).AddRow(1))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/favorite", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCoin_ClampToTwo(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(count,0\) FROM article_coins WHERE article_id=\$1 AND user_id=\$2`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO article_coins \(article_id, user_id, count\) VALUES \(\$1,\$2,\$3\)`).
		WithArgs(int64(9), int64(7), 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET coin_count = coin_count \+ \$1 WHERE id=\$2`).
		WithArgs(2, int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT coin_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"coin_count"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/coin", 7, `{"count":5}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		MyCoin int   `json:"my_coin"`
		Count  int64 `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, 2, data.MyCoin)
	assert.Equal(t, int64(2), data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCoin_AlreadyTwo(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COALESCE\(count,0\) FROM article_coins`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/coin", 7, `{}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddCoin_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/coin", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAddShare(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectExec(`UPDATE articles SET share_count = share_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT share_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"share_count"}).AddRow(7))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/share", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Count int64 `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(7), data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInteraction_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT like_count, favorite_count, coin_count, comment_count, share_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrNoRows)

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/interaction", 0, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInteraction_Anonymous(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT like_count, favorite_count, coin_count, comment_count, share_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count", "favorite_count", "coin_count", "comment_count", "share_count"}).
			AddRow(1, 2, 3, 4, 5))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/interaction", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, float64(1), data["like_count"])
	assert.Equal(t, float64(5), data["share_count"])
	assert.Equal(t, false, data["liked"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetInteraction_Viewer(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT like_count, favorite_count, coin_count, comment_count, share_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count", "favorite_count", "coin_count", "comment_count", "share_count"}).
			AddRow(1, 2, 3, 4, 5))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"liked", "favorited"}).AddRow(true, true))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(count\),0\) FROM article_coins WHERE article_id=\$1 AND user_id=\$2`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/interaction", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data map[string]any
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, true, data["liked"])
	assert.Equal(t, float64(2), data["coin"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_ListWithReplies(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id WHERE c.article_id=\$1 AND c.status=0 AND c.parent_id IS NULL`).
		WithArgs(int64(9), int64(20), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "content", "like_count", "created_at", "nickname", "avatar"}).
			AddRow(1, 9, 4, 0, "root", 0, now, "小明", "/a.png"))
	mock.ExpectQuery(`FROM article_comments c LEFT JOIN users u ON u.id=c.user_id LEFT JOIN users ru ON ru.id=c.reply_to_user_id WHERE c.status=0 AND c.parent_id = ANY\(\$1\)`).
		WithArgs([]int64{1}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "user_id", "parent_id", "reply_to_user_id", "content", "like_count", "created_at", "nickname", "avatar", "reply_to_name"}).
			AddRow(2, 9, 8, 1, 4, "re", 0, now, "柯洁", "/b.png", "小明"))
	mock.ExpectQuery(`FROM article_comment_likes l JOIN article_comments c ON c.id = l.comment_id WHERE l.user_id=\$1 AND c.article_id=\$2`).
		WithArgs(int64(5), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_id"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	w := do(t, h, h.handleByID, http.MethodGet, "/api/v1/article/9/comments", 5, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		List  []commentNode `json:"list"`
		Total int           `json:"total"`
	}
	mustJSON(t, decode(t, w), &data)
	require.Len(t, data.List, 1)
	require.Len(t, data.List[0].Replies, 1)
	assert.Equal(t, int64(2), data.List[0].Replies[0].ID)
	assert.Equal(t, "小明", data.List[0].Replies[0].ReplyToName)
	assert.True(t, data.List[0].MyLiked)
	assert.Equal(t, 2, data.Total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostValidation(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		h, mock := newTestHandler(t)
		w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"   "}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("too long", func(t *testing.T) {
		h, mock := newTestHandler(t)
		w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7,
			`{"content":"`+strings.Repeat("字", 1001)+`"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("comment disabled", func(t *testing.T) {
		h, mock := newTestHandler(t)
		mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
			WithArgs(int64(9)).
			WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(0))
		w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"hi"}`)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "该专栏已关闭评论", decode(t, w).Message)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("article missing", func(t *testing.T) {
		h, mock := newTestHandler(t)
		mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
			WithArgs(int64(9)).
			WillReturnError(sql.ErrNoRows)
		w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"hi"}`)
		assert.Equal(t, http.StatusNotFound, w.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestComments_PostRootWithMention(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
	mock.ExpectQuery(`SELECT comment_enabled FROM articles WHERE id=\$1 AND status=1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_enabled"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO article_comments \(article_id, user_id, parent_id, reply_to_user_id, content\) VALUES \(\$1,\$2,\$3,\$4,\$5\) RETURNING id, created_at`).
		WithArgs(int64(9), int64(7), nil, nil, "@柯洁 hi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(11, now))
	mock.ExpectExec(`UPDATE articles SET comment_count = comment_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id FROM users WHERE nickname=\$1 OR username=\$1 LIMIT 1`).
		WithArgs("柯洁").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectExec(`INSERT INTO messages \(sender_id, receiver_id, content, message_type, target_id, comment_id, is_read, created_at\)`).
		WithArgs(int64(7), int64(8), "在评论中@了你：@柯洁 hi", int64(9), int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COALESCE\(nickname,''\), COALESCE\(avatar,''\) FROM users WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"nickname", "avatar"}).AddRow("小明", "/a.png"))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/comments", 7, `{"content":"@柯洁 hi"}`)
	require.Equal(t, http.StatusOK, w.Code)
	var c commentNode
	mustJSON(t, decode(t, w), &c)
	assert.Equal(t, int64(11), c.ID)
	assert.Equal(t, "小明", c.AuthorName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_PostReply(t *testing.T) {
	h, mock := newTestHandler(t)
	now := time.Now()
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
	var c commentNode
	mustJSON(t, decode(t, w), &c)
	assert.Equal(t, int64(12), c.ID)
	assert.Equal(t, int64(100), c.ParentID)
	assert.Equal(t, int64(300), c.ReplyToUserID)
	assert.Equal(t, "楼主", c.ReplyToName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentLike_Toggle(t *testing.T) {
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
	mock.ExpectQuery(`SELECT like_count FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"like_count"}).AddRow(3))

	w := do(t, h, h.handleCommentSub, http.MethodPost, "/api/v1/article/comment/11/like", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Active bool `json:"active"`
		Count  int  `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.True(t, data.Active)
	assert.Equal(t, 3, data.Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentDelete_Forbidden(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id, article_id FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "article_id"}).AddRow(9, 5))

	w := do(t, h, h.handleCommentSub, http.MethodDelete, "/api/v1/article/comment/11", 7, "")
	assert.Equal(t, http.StatusForbidden, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCommentDelete_OK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT user_id, article_id FROM article_comments WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "article_id"}).AddRow(7, 5))
	mock.ExpectExec(`UPDATE article_comments SET status=1, updated_at=NOW\(\) WHERE id=\$1`).
		WithArgs(int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET comment_count = \(SELECT COUNT\(\*\) FROM article_comments WHERE article_id=\$1 AND status=0\) WHERE id=\$1`).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT comment_count FROM articles WHERE id=\$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"comment_count"}).AddRow(0))

	w := do(t, h, h.handleCommentSub, http.MethodDelete, "/api/v1/article/comment/11", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, "ok", data.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_Get(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM favorite_folders f WHERE f.user_id=\$1 ORDER BY f.created_at ASC`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "visible", "cnt", "checked"}).
			AddRow(1, "默认收藏夹", 1, 2, true))

	w := do(t, h, h.handleFolders, http.MethodGet, "/api/v1/article/folders?article_id=9", 7, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Checked bool   `json:"checked"`
	}
	mustJSON(t, decode(t, w), &list)
	require.Len(t, list, 1)
	assert.Equal(t, "默认收藏夹", list[0].Name)
	assert.True(t, list[0].Checked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_PostValidation(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleFolders, http.MethodPost, "/api/v1/article/folders", 7, `{"name":"  "}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = do(t, h, h.handleFolders, http.MethodPost, "/api/v1/article/folders", 7,
		`{"name":"`+strings.Repeat("字", 21)+`"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFolders_PostOK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`INSERT INTO favorite_folders \(user_id, name, visible\) VALUES \(\$1,\$2,1\) RETURNING id`).
		WithArgs(int64(7), "新夹").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	w := do(t, h, h.handleFolders, http.MethodPost, "/api/v1/article/folders", 7, `{"name":"新夹"}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.Equal(t, int64(3), data.ID)
	assert.Equal(t, "新夹", data.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_Add(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT status FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folders WHERE id=\$1 AND user_id=\$2`).
		WithArgs(int64(1), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO favorite_folder_articles \(folder_id, article_id\) VALUES \(\$1,\$2\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(1), int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM favorite_folder_articles ffa`).
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM user_interactions WHERE user_id=\$1 AND target_type='article' AND target_id=\$2 AND interaction_type='favorite'\)`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`INSERT INTO user_interactions \(user_id, target_type, target_id, interaction_type\) VALUES \(\$1,'article',\$2,'favorite'\) ON CONFLICT DO NOTHING`).
		WithArgs(int64(7), int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE articles SET favorite_count = favorite_count \+ 1 WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT favorite_count FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"favorite_count"}).AddRow(1))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{"adds":[1]}`)
	require.Equal(t, http.StatusOK, w.Code)
	var data struct {
		Favorited bool  `json:"favorited"`
		Count     int64 `json:"count"`
		Folders   int   `json:"folders"`
	}
	mustJSON(t, decode(t, w), &data)
	assert.True(t, data.Favorited)
	assert.Equal(t, int64(1), data.Count)
	assert.Equal(t, 1, data.Folders)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectFolders_DraftNotFound(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT status FROM articles WHERE id=\$1`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(0))

	w := do(t, h, h.handleByID, http.MethodPost, "/api/v1/article/9/collect", 7, `{"adds":[1]}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMentionSearch_Empty(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleMentionSearch, http.MethodGet, "/api/v1/article/mention-search", 0, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", string(decode(t, w).Data))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMentionSearch_OK(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM users WHERE \(nickname ILIKE`).
		WithArgs("柯").
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "avatar"}).AddRow(8, "柯洁", "/b.png"))

	w := do(t, h, h.handleMentionSearch, http.MethodGet, "/api/v1/article/mention-search?keyword=%E6%9F%AF", 0, "")
	require.Equal(t, http.StatusOK, w.Code)
	var list []struct {
		UID      int64  `json:"uid"`
		Nickname string `json:"nickname"`
	}
	mustJSON(t, decode(t, w), &list)
	require.Len(t, list, 1)
	assert.Equal(t, int64(8), list[0].UID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComments_MethodNotAllowed(t *testing.T) {
	h, mock := newTestHandler(t)
	w := do(t, h, h.handleByID, http.MethodPut, "/api/v1/article/9/comments", 7, "")
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
