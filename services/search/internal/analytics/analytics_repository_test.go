package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func TestOverview_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	// 1st query: aggregate sums
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\)`).WillReturnRows(
		sqlmock.NewRows([]string{"views", "likes", "coins", "collections", "shares", "comments", "manuscripts"}).
			AddRow(1000, 200, 50, 30, 10, 100, 5))

	// 2nd query: danmaku sum
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(danmaku_count\)`).WillReturnRows(
		sqlmock.NewRows([]string{"danmaku"}).AddRow(80))

	// 3rd query: follower count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(500))

	// 4th query: 7-day views increase
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscript_daily_metrics`).WillReturnRows(
		sqlmock.NewRows([]string{"views_increase"}).AddRow(150))

	// 5th query: interaction increases
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(CASE WHEN i\.interaction_type`).WillReturnRows(
		sqlmock.NewRows([]string{"likes", "coins", "collections", "shares"}).
			AddRow(30, 10, 5, 2))

	// 6th query: comments increase
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(20))

	// 7th query: followers increase
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id.*created_at`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(15))

	result, err := repo.Overview(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), result["totalViews"])
	assert.Equal(t, int64(200), result["totalLikes"])
	assert.Equal(t, int64(50), result["totalCoins"])
	assert.Equal(t, int64(30), result["totalCollections"])
	assert.Equal(t, int64(10), result["totalShares"])
	assert.Equal(t, int64(100), result["totalComments"])
	assert.Equal(t, int64(5), result["totalManuscripts"])
	assert.Equal(t, int64(80), result["totalDanmaku"])
	assert.Equal(t, int64(500), result["totalFollowers"])
	assert.Equal(t, int64(150), result["viewsIncrease"])
	assert.Equal(t, int64(30), result["likesIncrease"])
	assert.Equal(t, int64(10), result["coinsIncrease"])
	assert.Equal(t, int64(5), result["collectionsIncrease"])
	assert.Equal(t, int64(2), result["sharesIncrease"])
	assert.Equal(t, int64(20), result["commentsIncrease"])
	assert.Equal(t, int64(15), result["followersIncrease"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRanking_SortByViews(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Video A", "a.jpg", 1000, 100, 20, 30, 10, 5, 3, nil).
		AddRow(2, "Video B", "b.jpg", 500, 50, 10, 15, 5, 3, 1, nil)

	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id.*view_count DESC`).WillReturnRows(rows)

	result, err := repo.Ranking(ctx, 1, "views", 10)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	assert.Equal(t, 2, result["total"])
	assert.Equal(t, "Video A", list[0]["title"])
	assert.Equal(t, int64(1000), list[0]["viewCount"])
	assert.Equal(t, int64(500), list[1]["viewCount"])
}

func TestRanking_SortByLikes(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Popular", "p.jpg", 200, 200, 50, 100, 30, 20, 10, nil)

	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id.*like_count DESC`).WillReturnRows(rows)

	result, err := repo.Ranking(ctx, 1, "likes", 10)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, "Popular", list[0]["title"])
}

func TestRanking_SortByComments(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	rows := sqlmock.NewRows(columns)

	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id.*comment_count DESC`).WillReturnRows(rows)

	result, err := repo.Ranking(ctx, 1, "comments", 10)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	assert.Empty(t, list)
}

func TestRanking_SortByTime(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	rows := sqlmock.NewRows(columns).
		AddRow(5, "Latest", "l.jpg", 10, 1, 0, 0, 0, 0, 0, nil)

	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id.*upload_time DESC`).WillReturnRows(rows)

	result, err := repo.Ranking(ctx, 1, "time", 5)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, "Latest", list[0]["title"])
}

func TestRanking_EmptyResult(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	rows := sqlmock.NewRows(columns)

	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id`).WillReturnRows(rows)

	result, err := repo.Ranking(ctx, 999, "views", 10)
	require.NoError(t, err)
	assert.Equal(t, 0, result["total"])
	assert.Empty(t, result["list"])
}

func TestLatestComments_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "content", "manuscript_id", "title", "username", "avatar", "created_at"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "Great video!", 10, "My Video", "viewer1", "av1.jpg", time.Date(2025, 1, 15, 10, 30, 0, 0, time.Local))

	mock.ExpectQuery(`SELECT c\.id, c\.content, c\.manuscript_id`).WillReturnRows(rows)

	result, err := repo.LatestComments(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0]["id"])
	assert.Equal(t, "Great video!", result[0]["content"])
	assert.Equal(t, int64(10), result[0]["manuscriptId"])
	assert.Equal(t, "My Video", result[0]["manuscriptTitle"])
	assert.Equal(t, "viewer1", result[0]["username"])
	assert.Equal(t, "av1.jpg", result[0]["avatar"])
}

func TestLatestComments_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "content", "manuscript_id", "title", "username", "avatar", "created_at"}
	rows := sqlmock.NewRows(columns)

	mock.ExpectQuery(`SELECT c\.id, c\.content, c\.manuscript_id`).WillReturnRows(rows)

	result, err := repo.LatestComments(ctx, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFansTrend_7Days(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"date", "new_followers", "unfollows"}
	rows := sqlmock.NewRows(columns).
		AddRow("2025-01-13", 5, 1).
		AddRow("2025-01-15", 3, 0)

	mock.ExpectQuery(`SELECT DATE\(created_at\).*interaction_type='FOLLOW'.*UNFOLLOW`).WillReturnRows(rows)

	// current followers count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(500))

	result, err := repo.FansTrend(ctx, 1, 7)
	require.NoError(t, err)
	assert.Equal(t, int64(500), result["currentFollowers"])
	dates := result["dates"].([]string)
	assert.GreaterOrEqual(t, len(dates), 6)
	assert.LessOrEqual(t, len(dates), 7)
	newF := result["newFollowers"].([]int64)
	assert.Equal(t, len(dates), len(newF))
	unfollows := result["unfollows"].([]int64)
	assert.Equal(t, len(dates), len(unfollows))
	totalF := result["totalFollowers"].([]int64)
	assert.Equal(t, len(dates), len(totalF))
	assert.NotNil(t, result["growthRate"])
	assert.NotNil(t, result["newFollowersToday"])
	assert.NotNil(t, result["unfollowsToday"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestManuscriptTrend_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	columns := []string{"id", "title", "views", "upload_time"}
	rows := sqlmock.NewRows(columns).
		AddRow(1, "First", 100, time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local)).
		AddRow(2, "Second", 200, time.Date(2025, 1, 5, 14, 0, 0, 0, time.Local))

	mock.ExpectQuery(`SELECT id, title, COALESCE\(view_count,0\)`).WillReturnRows(rows)

	result, err := repo.ManuscriptTrend(ctx, 1)
	require.NoError(t, err)
	dates := result["dates"].([]string)
	views := result["views"].([]int64)
	titles := result["titles"].([]string)
	danmaku := result["danmaku"].([]int64)

	require.Len(t, dates, 2)
	assert.Equal(t, "2025-01-01", dates[0])
	assert.Equal(t, "2025-01-05", dates[1])
	// cumulative views
	assert.Equal(t, int64(100), views[0])
	assert.Equal(t, int64(300), views[1])
	assert.Equal(t, "First", titles[0])
	assert.Equal(t, "Second", titles[1])
	assert.Len(t, danmaku, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
