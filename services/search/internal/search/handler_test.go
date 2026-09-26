package search

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mybilibili/search/internal/hot"
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

// ==================== helper ====================

func newTestHandlerWithHotRepo(t *testing.T) (*Handler, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	hotRepo := hot.NewRepository(rdb)

	svc := NewService(repo, hotRepo)
	h := NewHandler(svc)
	t.Cleanup(func() { db.Close(); rdb.Close() })
	return h, mock, mr
}

func newTestHandlerNoHot(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewRepository(db)
	svc := NewService(repo, nil)
	h := NewHandler(svc)
	t.Cleanup(func() { db.Close() })
	return h, mock
}

// ==================== 用户端 ====================

func TestHandleRelated_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(2, "相关视频", "desc", "http://cv", 1, 1,
			500, 50, 20, 10,
			"00:12:00", 3, "2026-09-10 10:00:00",
			1, "user1", "昵称", "http://av", 3, 0))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/related/42?size=10", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "相关视频", resp.Data[0]["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRelated_DefaultSize(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "cover_url", "user_id", "category_id",
			"view_count", "like_count", "comment_count", "danmaku_count",
			"duration", "status", "upload_time",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/related/1", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleForYou_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.user_id, m.title, m.cover_url`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "cover_url", "view_count", "like_count",
			"comment_count", "created_at", "duration_seconds",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(10, 5, "为你推荐", "http://cv", 1000, 100,
			50, "2026-09-01 10:00:00", 3600,
			5, "user1", "昵称", "http://av", 4, 0))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/for-you", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "为你推荐", resp.Data[0]["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleForYou_EmptyResult(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.user_id, m.title, m.cover_url`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "cover_url", "view_count", "like_count",
			"comment_count", "created_at", "duration_seconds",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/for-you", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleHotRecommend_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.user_id, m.title, m.cover_url`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "cover_url", "view_count", "like_count",
			"comment_count", "created_at", "duration_seconds",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}).AddRow(20, 3, "热门推荐", "http://cv2", 2000, 200,
			100, "2026-09-05 10:00:00", 7200,
			3, "user2", "昵称2", "http://av2", 5, 1))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/hot?categoryId=1&size=5", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "热门推荐", resp.Data[0]["title"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleHotRecommend_DefaultParams(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT m.id, m.user_id, m.title, m.cover_url`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "cover_url", "view_count", "like_count",
			"comment_count", "created_at", "duration_seconds",
			"uid", "uname", "unick", "uavatar", "ulevel", "isVertical",
		}))

	rr := do(t, h, http.MethodGet, "/api/v1/recommend/hot", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                      `json:"code"`
		Data []map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== 管理端 hot 操作 ====================

func TestReadKeyword(t *testing.T) {
	newReq := func(contentType, body string) *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/search/hot/increment", strings.NewReader(body))
		req.Header.Set("Content-Type", contentType)
		return req
	}

	assert.Equal(t, "golang", readKeyword(newReq("application/json", `{"keyword":"golang"}`)))
	assert.Equal(t, "测试", readKeyword(newReq("application/json", `{"keyword":"测试"}`)))
	assert.Equal(t, "", readKeyword(newReq("application/json", `{}`)))
	assert.Equal(t, "", readKeyword(newReq("application/json", `not json at all`)))

	assert.Equal(t, "golang", readKeyword(newReq("application/x-www-form-urlencoded", "keyword=golang")))
	// urlencoded 路径自带百分号解码
	assert.Equal(t, "测试", readKeyword(newReq("application/x-www-form-urlencoded", "keyword=%E6%B5%8B%E8%AF%95")))
	assert.Equal(t, "", readKeyword(newReq("application/x-www-form-urlencoded", "")))

	// web 端 /search/word/add 走的就是 multipart，原来这条路径静默丢弃
	const mp = "multipart/form-data; boundary=xxx"
	assert.Equal(t, "测试", readKeyword(newReq(mp,
		"--xxx\r\nContent-Disposition: form-data; name=\"keyword\"\r\n\r\n测试\r\n--xxx--\r\n")))
	assert.Equal(t, "", readKeyword(newReq(mp,
		"--xxx\r\nContent-Disposition: form-data; name=\"other\"\r\n\r\nx\r\n--xxx--\r\n")))
	assert.Equal(t, "", readKeyword(newReq(mp, "")))
}

func TestHandleHotIncrement_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPost, "/api/v1/search/hot/increment", `{"keyword":"golang"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleHotIncrement_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/increment", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHotKeyword_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPost, "/api/v1/search/hot/keyword", `{"keyword":"golang","score":80,"rank":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleHotKeyword_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/keyword", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHotRank_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPut, "/api/v1/search/hot/rank", `{"keyword":"golang","rank":1}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleHotRank_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/rank", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHotScore_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPut, "/api/v1/search/hot/score", `{"keyword":"golang","score":95}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleHotScore_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/score", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleCleanExpired_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPost, "/api/v1/search/hot/clean-expired", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleCleanExpired_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/clean-expired", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHotDelete_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodDelete, "/api/v1/search/hot/delete", `{"keyword":"golang"}`)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleHotDelete_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/delete", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleHotGet_200(t *testing.T) {
	h, mock, mr := newTestHandlerWithHotRepo(t)
	_ = mock

	mr.ZAdd("hot_search:rank", 80, "golang")

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/get?keyword=golang", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "golang", resp.Data["keyword"])
	assert.Equal(t, float64(80), resp.Data["score"])
}

func TestHandleHotGet_400(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/get", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleHotGet_404(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/get?keyword=nonexistent", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleHotScoreGet_200(t *testing.T) {
	h, mock, mr := newTestHandlerWithHotRepo(t)
	_ = mock

	mr.ZAdd("hot_search:rank", 95, "python")

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/score-get?keyword=python", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "python", resp.Data["keyword"])
	assert.Equal(t, float64(95), resp.Data["score"])
}

func TestHandleHotScoreGet_400(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/score-get", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandleHotScoreGet_404(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/hot/score-get?keyword=nonexistent", "")
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ==================== 管理端 index 操作 ====================

func TestHandleIndexStatus_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{
		"total", "published", "indexed", "null_count",
	}).AddRow(100, 80, 75, 5))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/status", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexStatus_500(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/status", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexValidate_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{
		"total", "published", "indexed", "null_count",
	}).AddRow(50, 40, 40, 0))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/validate", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexValidate_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/validate", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleIndexRebuild_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`UPDATE manuscripts`).WillReturnResult(sqlmock.NewResult(0, 10))

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/rebuild", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 200, resp.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexRebuild_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/rebuild", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleIndexRebuild_500(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`UPDATE manuscripts`).WillReturnError(sql.ErrConnDone)

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/rebuild", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexRefresh_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/refresh", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleIndexRefresh_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/refresh", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleIndexNotNeeded_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/bulk", "")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleIndexNotNeeded_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/index/bulk", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// ==================== 管理端 recommend-config ====================

func TestHandleRecommendConfig_GET_200(t *testing.T) {
	h, mock := newTestHandler(t)

	cfg := `{"refresh_interval":300,"for_you_size":20,"related_size":10,"hot_size":10,"personalized":true}`
	mock.ExpectQuery(`SELECT config_json FROM recommend_configs`).
		WillReturnRows(sqlmock.NewRows([]string{"config_json"}).AddRow(cfg))

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/recommend-config", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(300), resp.Data["refresh_interval"])
	assert.Equal(t, true, resp.Data["personalized"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfig_GET_Empty(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT config_json FROM recommend_configs`).
		WillReturnError(sql.ErrNoRows)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/recommend-config", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfig_PUT_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"refresh_interval":600,"for_you_size":30}`
	rr := do(t, h, http.MethodPut, "/api/v1/search/admin/recommend-config", body)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(600), resp.Data["refresh_interval"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfig_PUT_WithUsername(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"for_you_size":15}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/search/admin/recommend-config", strings.NewReader(body))
	req.Header.Set("X-Username", "alice")
	rr := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfig_PUT_Error(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnError(sql.ErrConnDone)

	body := `{"for_you_size":15}`
	rr := do(t, h, http.MethodPut, "/api/v1/search/admin/recommend-config", body)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfig_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodDelete, "/api/v1/search/admin/recommend-config", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRecommendConfigReset_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnResult(sqlmock.NewResult(0, 1))

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/recommend-config/reset", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, float64(300), resp.Data["refresh_interval"])
	assert.Equal(t, float64(20), resp.Data["for_you_size"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRecommendConfigReset_405(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	rr := do(t, h, http.MethodGet, "/api/v1/search/admin/recommend-config/reset", "")
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleRecommendConfigReset_Error(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec(`INSERT INTO recommend_configs`).WillReturnError(sql.ErrConnDone)

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/recommend-config/reset", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
