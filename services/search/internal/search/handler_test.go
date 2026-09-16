package search

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHandler 构造 handler + sqlmock DB，返回 handler、sqlmock 与 mux。
func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo, nil)
	h := NewHandler(svc)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

// do 发起一次请求并返回 response recorder。
func do(t *testing.T, h *Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	h.Register(mux)
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func decodeList(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	return resp.Data
}

func TestHandleSearchVideos_200(t *testing.T) {
	h, mock := newTestHandler(t)

	// SearchManuscripts: 19 列
	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(1, "测试视频", "desc", "http://cover", 5, 3,
			100, 10, 2, 1,
			"00:10:00", 3, "2026-09-01 10:00:00",
			5, "user1", "昵称1", "http://avatar", 4, 0))

	rr := do(t, h, http.MethodGet, "/api/v1/search/videos?keyword=测试&page=1&pageSize=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	list, ok := resp["list"].([]interface{})
	require.True(t, ok)
	require.Len(t, list, 1)
	m := list[0].(map[string]interface{})
	assert.Equal(t, "测试视频", m["title"])
	// 搜索关键词会触发热搜自增，但 hotRepo 为 nil 时应被跳过，不产生 DB/Redis 调用
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearchVideos_Empty(t *testing.T) {
	h, mock := newTestHandler(t)

	// 空关键词：不带 search_vector 条件，只有 LIMIT/OFFSET
	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}))

	rr := do(t, h, http.MethodGet, "/api/v1/search/videos", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	// 空结果列表为 null（handler 未初始化为空 slice）
	list, ok := resp["list"].([]interface{})
	if ok {
		assert.Empty(t, list)
	} else {
		assert.Nil(t, resp["list"])
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearchUsers_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT u.id, u.username`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "nickname", "avatar", "signature", "level", "follower_count", "manuscript_count",
		}).AddRow(5, "user1", "昵称1", "http://avatar", "签名", 4, 100, 3))

	rr := do(t, h, http.MethodGet, "/api/v1/search/users?keyword=user", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	list, ok := resp["list"].([]interface{})
	require.True(t, ok)
	require.Len(t, list, 1)
	m := list[0].(map[string]interface{})
	assert.Equal(t, float64(5), m["mid"]) // JSON 数字反序列化为 float64
	assert.Equal(t, "昵称1", m["name"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSuggest_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT title FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).
			AddRow("测试视频一").AddRow("测试视频二"))

	rr := do(t, h, http.MethodGet, "/api/v1/search/suggest?keyword=测试&size=5", "")
	assert.Equal(t, http.StatusOK, rr.Code)

	// suggest 的 data 是数组，不能复用 decodeList（其期望 data 为对象）
	var resp struct {
		Data []string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "测试视频一", resp.Data[0])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleHot_200(t *testing.T) {
	// hotRepo 为 nil 时 Hot() 返回空列表且不报错
	h, mock := newTestHandler(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearch_Keyword(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(1, "Go语言入门", "Go教程", "http://cover", 5, 3,
			200, 20, 5, 3,
			"00:15:00", 3, "2026-09-01 10:00:00",
			5, "user1", "昵称1", "http://avatar", 4, 0))

	rr := do(t, h, http.MethodGet, "/api/v1/search/videos?keyword=Go语言&page=1&pageSize=10", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	list := resp["list"].([]interface{})
	require.Len(t, list, 1)
	m := list[0].(map[string]interface{})
	assert.Equal(t, "Go语言入门", m["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearch_EmptyKeyword(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(7, "无关键词视频", "", "http://cv", 2, 1,
			50, 5, 1, 0,
			"00:05:00", 3, "2026-09-10 10:00:00",
			2, "user2", "昵称2", "http://av", 3, 0))

	rr := do(t, h, http.MethodGet, "/api/v1/search/videos?keyword=&page=1&pageSize=20", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	list := resp["list"].([]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, "无关键词视频", list[0].(map[string]interface{})["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearchByType(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(10, "分类筛选视频", "desc", "http://cv", 1, 5,
			300, 30, 10, 5,
			"00:20:00", 3, "2026-09-15 10:00:00",
			1, "user1", "昵称1", "http://av", 5, 1))

	rr := do(t, h, http.MethodGet, "/api/v1/search/videos?keyword=视频&category_id=5&page=1&pageSize=10", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	resp := decodeList(t, rr)
	list := resp["list"].([]interface{})
	require.Len(t, list, 1)
	m := list[0].(map[string]interface{})
	assert.Equal(t, "分类筛选视频", m["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}