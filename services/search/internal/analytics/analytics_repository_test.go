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

	// 4th query: views increase (current period)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscript_daily_metrics`).WillReturnRows(
		sqlmock.NewRows([]string{"views_increase"}).AddRow(150))

	// 5th query: views previous period
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscript_daily_metrics`).WillReturnRows(
		sqlmock.NewRows([]string{"prev_views"}).AddRow(100))

	// 6th query: interaction increases (current period)
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(CASE WHEN UPPER\(i\.interaction_type`).WillReturnRows(
		sqlmock.NewRows([]string{"likes", "coins", "collections", "shares"}).
			AddRow(30, 10, 5, 2))

	// 7th query: interaction previous period
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(CASE WHEN UPPER\(i\.interaction_type`).WillReturnRows(
		sqlmock.NewRows([]string{"likes", "coins", "collections", "shares"}).
			AddRow(20, 8, 4, 1))

	// 8th query: comments increase (current period)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(20))

	// 9th query: comments previous period
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(12))

	// 10th query: danmaku increase (current period)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku d JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(7))

	// 11th query: danmaku previous period
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku d JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(3))

	// 12th query: followers increase (current period)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id.*created_at`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(15))

	// 13th query: followers previous period
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id.*created_at`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(9))

	// 14th query: space visits (current period)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM space_visits`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(25))

	// 15th query: space visits previous period
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM space_visits`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(18))

	result, err := repo.Overview(ctx, 1, 7)
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
	assert.Equal(t, int64(7), result["danmakuIncrease"])
	assert.Equal(t, int64(100), result["prevViews"])
	assert.Equal(t, int64(20), result["prevLikes"])
	assert.Equal(t, int64(8), result["prevCoins"])
	assert.Equal(t, int64(4), result["prevCollections"])
	assert.Equal(t, int64(1), result["prevShares"])
	assert.Equal(t, int64(12), result["prevComments"])
	assert.Equal(t, int64(3), result["prevDanmaku"])
	assert.Equal(t, int64(9), result["prevFollowers"])
	assert.Equal(t, int64(25), result["spaceVisits"])
	assert.Equal(t, int64(18), result["prevSpaceVisits"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestManuscriptInsights_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	cols := []string{"manuscript_id", "logged_viewers", "fan_viewers", "view_count", "avg_progress"}
	rows := sqlmock.NewRows(cols).
		AddRow(1, 10, 4, 100, 0.85).
		AddRow(2, 0, 0, 0, nil)
	mock.ExpectQuery(`FROM watch_history w`).WillReturnRows(rows)

	fcols := []string{"date", "fans"}
	frows := sqlmock.NewRows(fcols).AddRow("2025-01-14", 2)
	mock.ExpectQuery(`COUNT\(DISTINCT w\.user_id\)`).WillReturnRows(frows)

	result, err := repo.ManuscriptInsights(ctx, 1, 7)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0]["manuscriptId"])
	assert.InDelta(t, 0.4, list[0]["fanRate"], 0.001)
	assert.InDelta(t, 0.9, list[0]["guestRate"], 0.001)
	assert.InDelta(t, 0.85, list[0]["avgProgress"], 0.001)
	_, hasProgress := list[1]["avgProgress"]
	assert.False(t, hasProgress)
	ft := result["fanTrend"].(map[string]interface{})
	assert.Len(t, ft["dates"], 7)
	assert.Len(t, ft["fans"], 7)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountDiagnosis_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	// 1st query: my metrics (uploads, views, likes, newFollowers)
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT m\.id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).AddRow(3, 6263, 27, 5))

	// 2nd query: peer metrics (grouped by other users)
	mock.ExpectQuery(`WHERE m\.user_id <> \$1`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).
			AddRow(2, 800, 10, 3).
			AddRow(1, 200, 2, 1))

	// 3rd query: person distribution (guests, fans)
	mock.ExpectQuery(`FROM watch_history wh`).WillReturnRows(
		sqlmock.NewRows([]string{"guests", "fans"}).AddRow(99, 1))

	// 4th query: terminal distribution
	mock.ExpectQuery(`SELECT wh\.client_type, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"client_type", "cnt"}).
			AddRow("mobile", 60).
			AddRow("pc", 30))

	// 5th query: last metric date
	mock.ExpectQuery(`SELECT MAX\(metric_date\)`).WillReturnRows(
		sqlmock.NewRows([]string{"d"}).AddRow("2026-10-12"))

	// 6th query: sources (recently-viewed manuscript shares)
	mock.ExpectQuery(`FROM manuscript_daily_metrics dm`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "title", "view_count"}).
			AddRow(29, "乒乓球训练", 79).
			AddRow(13, "清理C盘", 9))

	result, err := repo.AccountDiagnosis(ctx, 1)
	require.NoError(t, err)

	metrics := result["metrics"].(map[string]interface{})
	assert.Equal(t, int64(3), metrics["uploads"])
	assert.Equal(t, int64(6263), metrics["views"])
	assert.Equal(t, int64(27), metrics["likes"])
	assert.Equal(t, int64(5), metrics["newFollowers"])
	assert.InDelta(t, 1.5, metrics["peerUploads"], 0.001)
	assert.InDelta(t, 500.0, metrics["peerViews"], 0.001)

	person := result["personDistribution"].(map[string]interface{})
	assert.InDelta(t, 0.99, person["guestPercent"], 0.001)
	assert.InDelta(t, 0.01, person["fanPercent"], 0.001)

	terminals := result["terminals"].([]terminalItem)
	require.Len(t, terminals, 5)
	assert.Equal(t, "mobile", terminals[0].Key)
	assert.InDelta(t, 0.666, terminals[0].Percent, 0.001)
	assert.Equal(t, "pc", terminals[1].Key)
	assert.InDelta(t, 0.333, terminals[1].Percent, 0.001)

	sources := result["sources"].([]map[string]interface{})
	require.Len(t, sources, 2)
	assert.InDelta(t, 79.0/88.0, sources[0]["percent"], 0.001)
	assert.Equal(t, "2026-10-12", result["sourceDate"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountsDiagnosis_EmptySources(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(DISTINCT m\.id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).AddRow(0, 0, 0, 0))
	mock.ExpectQuery(`WHERE m\.user_id <> \$1`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}))
	mock.ExpectQuery(`FROM watch_history wh`).WillReturnRows(
		sqlmock.NewRows([]string{"guests", "fans"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT wh\.client_type, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"client_type", "cnt"}))
	mock.ExpectQuery(`SELECT MAX\(metric_date\)`).WillReturnRows(
		sqlmock.NewRows([]string{"d"}).AddRow(nil))

	result, err := repo.AccountDiagnosis(ctx, 99)
	require.NoError(t, err)
	person := result["personDistribution"].(map[string]interface{})
	assert.InDelta(t, 1.0, person["guestPercent"], 0.001)
	sources := result["sources"].([]map[string]interface{})
	assert.Len(t, sources, 0)
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
