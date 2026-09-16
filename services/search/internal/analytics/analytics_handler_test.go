package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)
	return h, mock
}

func TestHandleOverview(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(view_count\\)").WillReturnRows(
		sqlmock.NewRows([]string{"totalViews", "totalLikes", "totalCoins", "totalCollections", "totalShares", "totalComments", "totalManuscripts"}).
			AddRow(100, 50, 30, 20, 10, 40, 5),
	)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(danmaku_count\\)").WillReturnRows(
		sqlmock.NewRows([]string{"totalDanmaku"}).AddRow(15),
	)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM follows").WillReturnRows(
		sqlmock.NewRows([]string{"totalFollowers"}).AddRow(1000),
	)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(view_count\\),0\\) FROM manuscript_daily_metrics").WillReturnRows(
		sqlmock.NewRows([]string{"viewsIncrease"}).AddRow(200),
	)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(CASE WHEN").WillReturnRows(
		sqlmock.NewRows([]string{"likesIncrease", "coinsIncrease", "collectionsIncrease", "sharesIncrease"}).
			AddRow(30, 20, 10, 5),
	)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM comments").WillReturnRows(
		sqlmock.NewRows([]string{"commentsIncrease"}).AddRow(25),
	)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM follows WHERE following_id = \\$1 AND created_at").WillReturnRows(
		sqlmock.NewRows([]string{"followersIncrease"}).AddRow(15),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/overview", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleOverview(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(100), data["totalViews"])
	assert.Equal(t, float64(50), data["totalLikes"])
	assert.Equal(t, float64(30), data["totalCoins"])
	assert.Equal(t, float64(20), data["totalCollections"])
	assert.Equal(t, float64(10), data["totalShares"])
	assert.Equal(t, float64(40), data["totalComments"])
	assert.Equal(t, float64(5), data["totalManuscripts"])
	assert.Equal(t, float64(15), data["totalDanmaku"])
	assert.Equal(t, float64(1000), data["totalFollowers"])
	assert.Equal(t, float64(200), data["viewsIncrease"])
	assert.Equal(t, float64(30), data["likesIncrease"])
	assert.Equal(t, float64(20), data["coinsIncrease"])
	assert.Equal(t, float64(10), data["collectionsIncrease"])
	assert.Equal(t, float64(5), data["sharesIncrease"])
	assert.Equal(t, float64(25), data["commentsIncrease"])
	assert.Equal(t, float64(15), data["followersIncrease"])
	assert.Equal(t, float64(0), data["danmakuIncrease"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleOverviewUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/overview", nil)
	rec := httptest.NewRecorder()

	h.handleOverview(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(401), resp["code"])
}

func TestHandleTrend(t *testing.T) {
	h, mock := newTestHandler(t)

	start7 := time.Now().AddDate(0, 0, -6).Format("2006-01-02")

	mock.ExpectQuery("SELECT metric_date::text").WillReturnRows(
		sqlmock.NewRows([]string{"date", "views"}).
			AddRow(start7, 100),
	)
	mock.ExpectQuery("SELECT DATE\\(i.created_at\\)::text").WillReturnRows(
		sqlmock.NewRows([]string{"date", "likes", "coins", "collects", "shares"}).
			AddRow(start7, 10, 5, 3, 2),
	)
	mock.ExpectQuery("SELECT DATE\\(c.created_at\\)::text, COUNT\\(\\*\\) FROM comments").WillReturnRows(
		sqlmock.NewRows([]string{"date", "count"}).AddRow(start7, 8),
	)
	mock.ExpectQuery("SELECT DATE\\(created_at\\)::text, COUNT\\(\\*\\) FROM follows").WillReturnRows(
		sqlmock.NewRows([]string{"date", "count"}).AddRow(start7, 3),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/trend?days=7", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleTrend(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	dates := data["dates"].([]interface{})
	assert.GreaterOrEqual(t, len(dates), 6)
	views := data["views"].([]interface{})
	assert.Equal(t, len(dates), len(views))
	comments := data["comments"].([]interface{})
	assert.Equal(t, len(dates), len(comments))
	followers := data["followers"].([]interface{})
	assert.Equal(t, len(dates), len(followers))
	danmaku := data["danmaku"].([]interface{})
	assert.Equal(t, len(dates), len(danmaku))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleTrendUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/trend", nil)
	rec := httptest.NewRecorder()

	h.handleTrend(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleRanking(t *testing.T) {
	h, mock := newTestHandler(t)

	uploadTime := time.Now().Add(-24 * time.Hour)
	mock.ExpectQuery("SELECT id, title, cover_url").WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "like_count", "comment_count", "danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}).
			AddRow(1, "Test Video", sql.NullString{String: "http://cover.jpg", Valid: true}, 1000, 200, 50, 30, 40, 20, 10, sql.NullTime{Time: uploadTime, Valid: true}).
			AddRow(2, "Another Video", sql.NullString{String: "", Valid: false}, 500, 100, 25, 15, 20, 10, 5, sql.NullTime{Time: uploadTime, Valid: true}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/ranking?sortBy=view_count&limit=10", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleRanking(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	assert.Equal(t, 2, len(list))
	assert.Equal(t, float64(2), data["total"])

	item1 := list[0].(map[string]interface{})
	assert.Equal(t, float64(1), item1["id"])
	assert.Equal(t, "Test Video", item1["title"])
	assert.Equal(t, float64(1000), item1["viewCount"])
	assert.Equal(t, float64(200), item1["likeCount"])
	assert.Greater(t, item1["interactionRate"], float64(0))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleRankingUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/ranking", nil)
	rec := httptest.NewRecorder()

	h.handleRanking(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleLatestComments(t *testing.T) {
	h, mock := newTestHandler(t)

	createdAt := time.Now().Add(-1 * time.Hour)
	mock.ExpectQuery("SELECT c.id, c.content").WillReturnRows(
		sqlmock.NewRows([]string{"id", "content", "manuscript_id", "title", "username", "avatar", "created_at"}).
			AddRow(1, "Great video!", 10, "My Video", "user1", sql.NullString{String: "http://avatar.jpg", Valid: true}, createdAt).
			AddRow(2, "Nice work!", 10, "My Video", "user2", sql.NullString{String: "", Valid: false}, createdAt),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/latest-comments?limit=10", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleLatestComments(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].([]interface{})
	assert.Equal(t, 2, len(data))

	comment1 := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), comment1["id"])
	assert.Equal(t, "Great video!", comment1["content"])
	assert.Equal(t, float64(10), comment1["manuscriptId"])
	assert.Equal(t, "My Video", comment1["manuscriptTitle"])
	assert.Equal(t, "user1", comment1["username"])
	assert.Equal(t, "http://avatar.jpg", comment1["avatar"])
	assert.NotEmpty(t, comment1["time"])
	assert.NotEmpty(t, comment1["createTime"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLatestCommentsUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/latest-comments", nil)
	rec := httptest.NewRecorder()

	h.handleLatestComments(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleFansTrend(t *testing.T) {
	h, mock := newTestHandler(t)

	start30 := time.Now().AddDate(0, 0, -29).Format("2006-01-02")

	mock.ExpectQuery("SELECT DATE\\(created_at\\)::text").WillReturnRows(
		sqlmock.NewRows([]string{"date", "newFollowers", "unfollows"}).
			AddRow(start30, 10, 2),
	)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM follows WHERE following_id = \\$1").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(500),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-trend?days=30", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleFansTrend(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	dates := data["dates"].([]interface{})
	assert.GreaterOrEqual(t, len(dates), 29)
	newFollowers := data["newFollowers"].([]interface{})
	assert.Equal(t, len(dates), len(newFollowers))
	unfollows := data["unfollows"].([]interface{})
	assert.Equal(t, len(dates), len(unfollows))
	totalFollowers := data["totalFollowers"].([]interface{})
	assert.Equal(t, len(dates), len(totalFollowers))
	assert.Equal(t, float64(500), data["currentFollowers"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFansTrendUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-trend", nil)
	rec := httptest.NewRecorder()

	h.handleFansTrend(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleFansRankingInteraction(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery("SELECT u.id, u.username, u.avatar, COUNT\\(\\*\\) AS interactionCount").WillReturnRows(
		sqlmock.NewRows([]string{"id", "username", "avatar", "interactionCount"}).
			AddRow(1, "fan1", sql.NullString{String: "http://a1.jpg", Valid: true}, 50).
			AddRow(2, "fan2", sql.NullString{String: "", Valid: false}, 30),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-ranking?type=interaction&limit=10", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleFansRanking(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].([]interface{})
	assert.Equal(t, 2, len(data))

	fan1 := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), fan1["id"])
	assert.Equal(t, "fan1", fan1["username"])
	assert.Equal(t, "http://a1.jpg", fan1["avatar"])
	assert.Equal(t, float64(50), fan1["interactionCount"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFansRankingComment(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery("SELECT u.id, u.username, u.avatar, COUNT\\(DISTINCT c.id\\)").WillReturnRows(
		sqlmock.NewRows([]string{"id", "username", "avatar", "commentCount"}).
			AddRow(3, "commenter1", sql.NullString{String: "http://a3.jpg", Valid: true}, 20),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-ranking?type=comment&limit=10", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleFansRanking(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].([]interface{})
	assert.Equal(t, 1, len(data))

	fan := data[0].(map[string]interface{})
	assert.Equal(t, float64(3), fan["id"])
	assert.Equal(t, "commenter1", fan["username"])
	assert.Equal(t, float64(20), fan["interactionCount"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFansRankingUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-ranking", nil)
	rec := httptest.NewRecorder()

	h.handleFansRanking(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandleManuscriptTrend(t *testing.T) {
	h, mock := newTestHandler(t)

	uploadTime := time.Now().Add(-48 * time.Hour)
	mock.ExpectQuery("SELECT id, title, COALESCE\\(view_count,0\\)").WillReturnRows(
		sqlmock.NewRows([]string{"id", "title", "views", "upload_time"}).
			AddRow(1, "Video A", 1000, sql.NullTime{Time: uploadTime, Valid: true}).
			AddRow(2, "Video B", 500, sql.NullTime{Time: uploadTime.Add(24 * time.Hour), Valid: true}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/manuscript-trend?days=7", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleManuscriptTrend(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	dates := data["dates"].([]interface{})
	assert.Equal(t, 2, len(dates))
	views := data["views"].([]interface{})
	assert.Equal(t, float64(1000), views[0])
	assert.Equal(t, float64(1500), views[1])
	titles := data["titles"].([]interface{})
	assert.Equal(t, "Video A", titles[0])
	assert.Equal(t, "Video B", titles[1])
	danmaku := data["danmaku"].([]interface{})
	assert.Equal(t, float64(0), danmaku[0])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleManuscriptTrendUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/manuscript-trend", nil)
	rec := httptest.NewRecorder()

	h.handleManuscriptTrend(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRegister(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	routes := []string{
		"/api/v1/creator/stats/overview",
		"/api/v1/creator/stats/trend",
		"/api/v1/creator/stats/ranking",
		"/api/v1/creator/stats/latest-comments",
		"/api/v1/creator/stats/fans-trend",
		"/api/v1/creator/stats/fans-ranking",
		"/api/v1/creator/stats/manuscript-trend",
	}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "route %s should return 401 without user", route)
	}
}

func TestServiceTrendBoundary(t *testing.T) {
	t.Run("days < 1 defaults to 7", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo2 := NewRepository(db)
		svc2 := NewService(repo2)

		mock.ExpectQuery("SELECT metric_date::text").WillReturnRows(sqlmock.NewRows([]string{"d", "v"}))
		mock.ExpectQuery("SELECT DATE\\(i.created_at\\)::text").WillReturnRows(sqlmock.NewRows([]string{"d", "l", "c", "cl", "s"}))
		mock.ExpectQuery("SELECT DATE\\(c.created_at\\)::text").WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))
		mock.ExpectQuery("SELECT DATE\\(created_at\\)::text, COUNT").WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))
		svc2.Trend(context.Background(), 1, -1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("days > 365 defaults to 7", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		repo2 := NewRepository(db)
		svc2 := NewService(repo2)

		mock.ExpectQuery("SELECT metric_date::text").WillReturnRows(sqlmock.NewRows([]string{"d", "v"}))
		mock.ExpectQuery("SELECT DATE\\(i.created_at\\)::text").WillReturnRows(sqlmock.NewRows([]string{"d", "l", "c", "cl", "s"}))
		mock.ExpectQuery("SELECT DATE\\(c.created_at\\)::text").WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))
		mock.ExpectQuery("SELECT DATE\\(created_at\\)::text, COUNT").WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))
		svc2.Trend(context.Background(), 1, 999)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceRankingBoundary(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewRepository(db)
	svc := NewService(repo)

	mock.ExpectQuery("SELECT id, title").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "cover_url", "view_count", "like_count", "comment_count", "danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}))
	svc.Ranking(context.Background(), 1, "view_count", -1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceLatestCommentsBoundary(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewRepository(db)
	svc := NewService(repo)

	mock.ExpectQuery("SELECT c.id, c.content").WillReturnRows(sqlmock.NewRows([]string{"id", "content", "manuscript_id", "title", "username", "avatar", "created_at"}))
	svc.LatestComments(context.Background(), 1, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceFansTrendBoundary(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewRepository(db)
	svc := NewService(repo)

	mock.ExpectQuery("SELECT DATE\\(created_at\\)::text").WillReturnRows(sqlmock.NewRows([]string{"d", "nf", "uf"}))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM follows WHERE following_id = \\$1").WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(0))
	svc.FansTrend(context.Background(), 1, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceFansRankingBoundary(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewRepository(db)
	svc := NewService(repo)

	mock.ExpectQuery("SELECT u.id, u.username, u.avatar, COUNT\\(DISTINCT").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "avatar", "commentCount"}))
	svc.FansRanking(context.Background(), 1, "comment", -1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		contains string
	}{
		{"just now", time.Now().Add(-30 * time.Second), "刚刚"},
		{"minutes ago", time.Now().Add(-5 * time.Minute), "分钟前"},
		{"hours ago", time.Now().Add(-3 * time.Hour), "小时前"},
		{"days ago", time.Now().Add(-10 * 24 * time.Hour), "天前"},
		{"months ago", time.Now().Add(-60 * 24 * time.Hour), "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}
