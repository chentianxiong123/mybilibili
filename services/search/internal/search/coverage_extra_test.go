package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "mybilibili/pkg/pb"
	"mybilibili/search/internal/hot"
)

// ---------- gRPC server ----------

func TestNewGrpcServer(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)
	require.NotNil(t, s)
	assert.Equal(t, svc, s.svc)
}

func TestGrpcSearchVideos_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "测试标题", "desc", "http://cover", 5, 3,
				100, 10, 2, 1, "00:10:00", 3, "2026-09-01 10:00:00",
				5, "user1", "昵称1", "http://avatar", 4, 0))

	req := &pb.SearchVideosRequest{Keyword: "测试", Page: 2, PageSize: 10}
	resp, err := s.SearchVideos(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, resp.List, 1)
	assert.Equal(t, int64(1), resp.List[0].Id)
	assert.Equal(t, "测试标题", resp.List[0].Title)
	assert.Equal(t, "http://cover", resp.List[0].Cover)
	assert.Equal(t, int32(2), resp.Page)
	assert.Equal(t, int32(10), resp.PageSize)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrpcSearchVideos_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)

	mock.ExpectQuery(`SELECT m.id, m.title`).WillReturnError(sql.ErrConnDone)

	req := &pb.SearchVideosRequest{Keyword: "x", Page: 1, PageSize: 10}
	resp, err := s.SearchVideos(context.Background(), req)
	require.ErrorIs(t, err, sql.ErrConnDone)
	assert.Nil(t, resp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrpcGetHotSearch_Success(t *testing.T) {
	repo, _ := newMockRepo(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	svc := NewService(repo, hot.NewRepository(rdb))
	s := NewGrpcServer(svc)

	mr.ZAdd("hot_search:rank", 80, "golang")
	mr.ZAdd("hot_search:rank", 60, "rust")

	resp, err := s.GetHotSearch(context.Background(), &pb.GetHotSearchRequest{})
	require.NoError(t, err)
	require.Len(t, resp.List, 2)
	assert.Equal(t, "golang", resp.List[0].Keyword)
	assert.Equal(t, "rust", resp.List[1].Keyword)
}

func TestGrpcGetHotSearch_Empty(t *testing.T) {
	repo, _ := newMockRepo(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	svc := NewService(repo, hot.NewRepository(rdb))
	s := NewGrpcServer(svc)

	// 空榜单：Hot 返回空列表且无错误
	resp, err := s.GetHotSearch(context.Background(), &pb.GetHotSearchRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.List)
}

func TestGrpcGetHotSearch_Error(t *testing.T) {
	// 使用指向关闭端口的 redis 客户端使 Top 返回错误，从而覆盖 GetHotSearch 的错误分支。
	repo, _ := newMockRepo(t)
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { rdb.Close() })
	svc := NewService(repo, hot.NewRepository(rdb))
	s := NewGrpcServer(svc)

	resp, err := s.GetHotSearch(context.Background(), &pb.GetHotSearchRequest{})
	require.Error(t, err) // FINDING-005 修复：错误已传播
	assert.Nil(t, resp)
}

func TestGrpcGetSuggestions_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)

	mock.ExpectQuery(`SELECT title FROM manuscripts`).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).
			AddRow("视频一").AddRow("视频二"))

	resp, err := s.GetSuggestions(context.Background(), &pb.GetSuggestionsRequest{Keyword: "视", Limit: 5})
	require.NoError(t, err)
	assert.Equal(t, []string{"视频一", "视频二"}, resp.Suggestions)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrpcGetSuggestions_EmptyKeyword(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)

	resp, err := s.GetSuggestions(context.Background(), &pb.GetSuggestionsRequest{Keyword: "", Limit: 5})
	require.NoError(t, err)
	assert.Empty(t, resp.Suggestions)
}

func TestGrpcGetSuggestions_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)
	s := NewGrpcServer(svc)

	mock.ExpectQuery(`SELECT title FROM manuscripts`).WillReturnError(sql.ErrConnDone)

	resp, err := s.GetSuggestions(context.Background(), &pb.GetSuggestionsRequest{Keyword: "x", Limit: 5})
	require.Error(t, err) // FINDING-005 修复：错误已传播
	assert.Nil(t, resp)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Handler: health / count / suggest / validate ----------

func TestHandleHealth_200(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)
	rr := do(t, h, http.MethodGet, "/api/v1/search/health", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"ok"}`, rr.Body.String())
}

func TestHandleSearchCount_200(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users u`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	rr := do(t, h, http.MethodGet, "/api/v1/search/count?keyword=go", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Code int   `json:"code"`
		Data []int `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, []int{12, 7}, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSearchCount_400(t *testing.T) {
	h, _ := newTestHandlerNoHot(t)

	// 空关键词触发 CountByKeyword 的 "keyword required" 错误 → 400
	rr := do(t, h, http.MethodGet, "/api/v1/search/count", "")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "keyword required")
}

func TestHandleSuggest_DefaultSize(t *testing.T) {
	h, mock := newTestHandler(t)

	// 无 size 参数 → 默认 10
	mock.ExpectQuery(`SELECT title FROM manuscripts`).
		WithArgs("%test%", int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("test video"))

	rr := do(t, h, http.MethodGet, "/api/v1/search/suggest?keyword=test", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp struct {
		Data []string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, []string{"test video"}, resp.Data)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleSuggest_SizeZero(t *testing.T) {
	h, mock := newTestHandler(t)

	// size=0 → 默认 10
	mock.ExpectQuery(`SELECT title FROM manuscripts`).
		WithArgs("%go%", int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"title"}))

	rr := do(t, h, http.MethodGet, "/api/v1/search/suggest?keyword=go&size=0", "")
	assert.Equal(t, http.StatusOK, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIndexValidate_500(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)

	rr := do(t, h, http.MethodPost, "/api/v1/search/admin/index/validate", "")
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// failingReader 让 io.ReadAll 失败，覆盖 readKeyword 的读 body 错误分支。
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read error") }
func (failingReader) Close() error             { return nil }

func TestReadKeyword_BodyReadError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/hot/increment", nil)
	req.Body = failingReader{}
	assert.Equal(t, "", readKeyword(req))
}

func TestReadKeyword_BadJSONThenForm(t *testing.T) {
	// JSON 解析失败后走 multipart/form 分支，仍失败再走 ParseForm 分支。
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/hot/increment", nil)
	req.Body = io.NopCloser(strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/octet-stream")
	assert.Equal(t, "", readKeyword(req))
}

func TestReadKeyword_AllParseFail(t *testing.T) {
	// JSON 不是合法 JSON，Content-Type 非法（boundary 缺 `=`）→ ParseMultipartForm 与
	// ParseForm 均报错 → 返回 ""（readKeyword 的兜底分支）。
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/hot/increment", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary") // 非法参数
	assert.Equal(t, "", readKeyword(req))
}

// ---------- Repository / Service: count ----------

func TestCountByKeyword_EmptyKeyword(t *testing.T) {
	repo, _ := newMockRepo(t)
	_, _, err := repo.CountByKeyword(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "keyword required")
}

func TestCountByKeyword_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users u`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	vc, uc, err := repo.CountByKeyword(context.Background(), "go")
	require.NoError(t, err)
	assert.Equal(t, int64(5), vc)
	assert.Equal(t, int64(3), uc)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByKeyword_VideoError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts m`).WillReturnError(sql.ErrConnDone)

	_, _, err := repo.CountByKeyword(context.Background(), "go")
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountByKeyword_UserError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users u`).WillReturnError(sql.ErrConnDone)

	_, _, err := repo.CountByKeyword(context.Background(), "go")
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceCount_Delegates(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts m`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(9))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users u`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	vc, uc, err := svc.Count(context.Background(), "rust")
	require.NoError(t, err)
	assert.Equal(t, int64(9), vc)
	assert.Equal(t, int64(4), uc)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository / Service: search & users error paths ----------

func TestSearchManuscripts_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT m.id, m.title`).WillReturnError(sql.ErrConnDone)

	_, err := repo.SearchManuscripts(context.Background(), "", "", 0, 1, 10)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchUsers_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT u.id, u.username`).WillReturnError(sql.ErrConnDone)

	_, err := repo.SearchUsers(context.Background(), "", 1, 10)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchUsers_ScanError(t *testing.T) {
	repo, mock := newMockRepo(t)

	// id 列给一个无法转换为 int64 的值 → Scan 失败
	columns := []string{"id", "username", "nickname", "avatar", "signature", "level", "follower_count", "manuscript_count"}
	mock.ExpectQuery(`SELECT u.id, u.username`).
		WillReturnRows(sqlmock.NewRows(columns).AddRow("notanint", "alice", "Alice", "a.jpg", "sig", 5, 100, 10))

	_, err := repo.SearchUsers(context.Background(), "alice", 1, 10)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Service.Search with hotRepo (Increment path) ----------

func TestServiceSearch_IncrementsHot(t *testing.T) {
	repo, mock := newMockRepo(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	svc := NewService(repo, hot.NewRepository(rdb))

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	mock.ExpectQuery(`SELECT m.id, m.title`).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, "Go", "", "", 1, 1, 1, 1, 1, 1, "1:00", 3, "2026-01-01", 1, "u", "n", "", 1, 0))

	list, err := svc.Search(context.Background(), "go", "", 0, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	// 热度自增已写入 miniredis
	assert.True(t, mr.Exists("hot_search:rank"))
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Service.Hot with hotRepo ----------

func TestServiceHot_WithData(t *testing.T) {
	repo, _ := newMockRepo(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	svc := NewService(repo, hot.NewRepository(rdb))

	mr.ZAdd("hot_search:rank", 90, "golang")

	list, err := svc.Hot(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "golang", list[0]["keyword"])
	assert.Equal(t, 1, list[0]["rank"])
}

func TestServiceHot_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)

	list, err := svc.Hot(context.Background())
	require.NoError(t, err)
	assert.Empty(t, list)
}

// ---------- Service.Suggest error paths ----------

func TestSuggest_EmptyKeyword(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)

	out, err := svc.Suggest(context.Background(), "", 10)
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestSuggest_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	mock.ExpectQuery(`SELECT title FROM manuscripts`).WillReturnError(sql.ErrConnDone)

	_, err := svc.Suggest(context.Background(), "x", 10)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- HotRecommend: query error & duration formatting ----------

func TestHotRecommend_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	mock.ExpectQuery(`SELECT m.id, m.user_id`).WillReturnError(sql.ErrConnDone)

	_, err := svc.HotRecommend(context.Background(), 1, 10)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHotRecommend_DurationWithHours(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	columns := []string{"id", "user_id", "title", "cover_url", "view_count", "like_count",
		"comment_count", "created_at", "duration_seconds",
		"uid", "uname", "unick", "uavatar", "ulevel", "isVertical"}
	mock.ExpectQuery(`SELECT m.id, m.user_id`).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(1, 5, "标题", "http://cv", 100, 10,
				5, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC), 3661,
				5, "user1", "昵称", "http://av", 4, 0))

	list, err := svc.HotRecommend(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "1:01:01", list[0]["duration"])
	assert.Equal(t, int64(3661), list[0]["duration_seconds"])
	assert.Equal(t, "2026-09-01 10:00:00", list[0]["created_at"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHotRecommend_DurationNoHours(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	columns := []string{"id", "user_id", "title", "cover_url", "view_count", "like_count",
		"comment_count", "created_at", "duration_seconds",
		"uid", "uname", "unick", "uavatar", "ulevel", "isVertical"}
	mock.ExpectQuery(`SELECT m.id, m.user_id`).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(2, 5, "标题2", "http://cv", 200, 20,
				10, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC), 61,
				5, "user1", "昵称", "http://av", 4, 0))

	list, err := svc.HotRecommend(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "1:01", list[0]["duration"])
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- hot service methods with hotRepo ----------

func newHotService(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	repo, _ := newMockRepo(t)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewService(repo, hot.NewRepository(rdb)), mr
}

func TestIncrementHotSearch_EmptyWhitespace(t *testing.T) {
	svc, _ := newHotService(t)
	// 空白关键词不写 redis
	require.NoError(t, svc.IncrementHotSearch(context.Background(), "   "))
}

func TestIncrementHotSearch_NonEmpty(t *testing.T) {
	svc, mr := newHotService(t)
	require.NoError(t, svc.IncrementHotSearch(context.Background(), "golang"))
	assert.True(t, mr.Exists("hot_search:rank"))
}

func TestIncrementHotSearch_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.IncrementHotSearch(context.Background(), "golang"))
}

func TestSetKeyword_WithRepo(t *testing.T) {
	svc, mr := newHotService(t)
	require.NoError(t, svc.SetKeyword(context.Background(), "golang", 80, 1))
	assert.True(t, mr.Exists("hot_search:rank"))
}

func TestSetKeyword_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.SetKeyword(context.Background(), "golang", 80, 1))
}

func TestSetRank_WithRepo(t *testing.T) {
	svc, mr := newHotService(t)
	require.NoError(t, svc.SetRank(context.Background(), "golang", 2))
	assert.True(t, mr.Exists("hot_search:rank"))
}

func TestSetRank_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.SetRank(context.Background(), "golang", 2))
}

func TestSetScore_WithRepo(t *testing.T) {
	svc, mr := newHotService(t)
	require.NoError(t, svc.SetScore(context.Background(), "golang", 95))
	assert.True(t, mr.Exists("hot_search:rank"))
}

func TestSetScore_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.SetScore(context.Background(), "golang", 95))
}

func TestGetKeyword_NotFound(t *testing.T) {
	svc, _ := newHotService(t)
	_, err := svc.GetKeyword(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestGetKeyword_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	_, err := svc.GetKeyword(context.Background(), "golang")
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetScore_NotFound(t *testing.T) {
	svc, _ := newHotService(t)
	_, err := svc.GetScore(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestGetScore_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	_, err := svc.GetScore(context.Background(), "golang")
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetScore_Success(t *testing.T) {
	svc, mr := newHotService(t)
	mr.ZAdd("hot_search:rank", 95, "python")
	score, err := svc.GetScore(context.Background(), "python")
	require.NoError(t, err)
	assert.Equal(t, int64(95), score)
}

func TestCleanExpiredHotSearch_WithRepo(t *testing.T) {
	svc, _ := newHotService(t)
	require.NoError(t, svc.CleanExpiredHotSearch(context.Background()))
}

func TestCleanExpiredHotSearch_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.CleanExpiredHotSearch(context.Background()))
}

func TestDeleteOne_WithRepo(t *testing.T) {
	svc, mr := newHotService(t)
	mr.ZAdd("hot_search:rank", 80, "golang")
	require.NoError(t, svc.DeleteOne(context.Background(), "golang"))
	assert.False(t, mr.Exists("hot_search:rank"))
}

func TestDeleteOne_NilRepo(t *testing.T) {
	repo, _ := newMockRepo(t)
	svc := NewService(repo, nil)
	require.NoError(t, svc.DeleteOne(context.Background(), "golang"))
}

// ---------- GetIndexStatus warning branch ----------

func TestGetIndexStatus_Warning(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"total", "published", "indexed", "null_count"}).
			AddRow(100, 80, 75, 5))
	// GIN 索引缺失 → health = warning
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	st, err := svc.GetIndexStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "warning", st.Health)
	assert.False(t, st.GINIndex)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetIndexStatus_Active(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"total", "published", "indexed", "null_count"}).
			AddRow(100, 90, 90, 0))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	st, err := svc.GetIndexStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "active", st.Health)
	assert.True(t, st.GINIndex)
	assert.True(t, st.Trigger)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------- concurrency smoke test ----------

func TestSearchConcurrent(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo, nil)

	columns := []string{"id", "title", "description", "cover_url", "user_id", "category_id",
		"view_count", "like_count", "comment_count", "danmaku_count",
		"duration", "status", "upload_time",
		"id", "username", "nickname", "avatar", "level", "is_vertical"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "并发视频", "", "", 1, 1, 1, 1, 1, 1, "1:00", 3, "2026-01-01", 1, "u", "n", "", 1, 0)
	mock.ExpectQuery(`SELECT m.id, m.title`).WillReturnRows(rows)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.Search(context.Background(), "", "", 0, 1, 10)
		}()
	}
	wg.Wait()
	require.NoError(t, mock.ExpectationsWereMet())
}
