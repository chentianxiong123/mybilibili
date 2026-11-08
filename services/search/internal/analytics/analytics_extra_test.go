package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Handler: handleFansByManuscript ----------

func TestHandleFansByManuscript(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "cnt"}).
			AddRow(1, 5).
			AddRow(2, 3),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-by-manuscript", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleFansByManuscript(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	require.Len(t, list, 2)
	assert.Equal(t, float64(1), list[0].(map[string]interface{})["manuscriptId"])
	assert.Equal(t, float64(5), list[0].(map[string]interface{})["fansGained"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleFansByManuscriptUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/fans-by-manuscript", nil)
	rec := httptest.NewRecorder()
	h.handleFansByManuscript(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------- Handler: handleManuscriptInsights ----------

func TestHandleManuscriptInsights(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM watch_history w`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "logged_viewers", "fan_viewers", "view_count", "avg_progress"}).
			AddRow(1, 10, 4, 100, 0.5),
	)
	mock.ExpectQuery(`COUNT\(DISTINCT w\.user_id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"date", "fans"}).AddRow(time.Now().Format("2006-01-02"), 3),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/manuscript-insights?days=7", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleManuscriptInsights(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	list := data["list"].([]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, float64(1), list[0].(map[string]interface{})["manuscriptId"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleManuscriptInsightsUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/manuscript-insights", nil)
	rec := httptest.NewRecorder()
	h.handleManuscriptInsights(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------- Handler: handleAccountDiagnosis ----------

func TestHandleAccountDiagnosis(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT m\.id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).AddRow(3, 1000, 50, 5))
	mock.ExpectQuery(`WHERE m\.user_id <> \$1`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).
			AddRow(2, 800, 10, 3).
			AddRow(1, 200, 2, 1))
	mock.ExpectQuery(`FROM watch_history wh`).WillReturnRows(
		sqlmock.NewRows([]string{"guests", "fans"}).AddRow(90, 10))
	mock.ExpectQuery(`SELECT wh\.client_type, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"client_type", "cnt"}).
			AddRow("mobile", 60).
			AddRow("pc", 40))
	mock.ExpectQuery(`SELECT MAX\(metric_date\)`).WillReturnRows(
		sqlmock.NewRows([]string{"d"}).AddRow("2026-10-12"))
	mock.ExpectQuery(`FROM manuscript_daily_metrics dm`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "title", "view_count"}).
			AddRow(29, "乒乓球训练", 79).
			AddRow(13, "清理C盘", 9))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/account-diagnosis", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleAccountDiagnosis(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	metrics := data["metrics"].(map[string]interface{})
	assert.Equal(t, float64(3), metrics["uploads"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleAccountDiagnosisUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/account-diagnosis", nil)
	rec := httptest.NewRecorder()
	h.handleAccountDiagnosis(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------- Handler: handleIncome ----------

func TestHandleIncome(t *testing.T) {
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND DATE\(created_at\)`).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(12345))
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND type = \$2`).WillReturnRows(
		sqlmock.NewRows([]string{"charge"}).AddRow(500))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/income?days=30", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleIncome(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(200), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.InDelta(t, 123.45, data["totalAmount"], 0.001)
	assert.InDelta(t, 5.0, data["chargeAmount"], 0.001)
	assert.Equal(t, float64(30), data["periodDays"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIncomeDaysDefault(t *testing.T) {
	// days 不为 7/30 时应钳制为 30（handler 内分支）。
	h, mock := newTestHandler(t)
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND DATE\(created_at\)`).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(0))
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND type = \$2`).WillReturnRows(
		sqlmock.NewRows([]string{"charge"}).AddRow(0))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/income?days=15", nil)
	req.Header.Set("X-User-Id", "123")
	rec := httptest.NewRecorder()

	h.handleIncome(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(30), data["periodDays"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleIncomeUnauthorized(t *testing.T) {
	h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/creator/stats/income", nil)
	rec := httptest.NewRecorder()
	h.handleIncome(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------- Repository: Overview default days branch ----------

func mockOverviewQueries(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\)`).WillReturnRows(
		sqlmock.NewRows([]string{"v", "l", "c", "cl", "s", "cm", "m"}).AddRow(1, 2, 3, 4, 5, 6, 7))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(danmaku_count\)`).WillReturnRows(
		sqlmock.NewRows([]string{"d"}).AddRow(8))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(9))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscript_daily_metrics`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(10))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(view_count\),0\) FROM manuscript_daily_metrics`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(11))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(CASE WHEN UPPER\(i\.interaction_type`).WillReturnRows(
		sqlmock.NewRows([]string{"l", "c", "cl", "s"}).AddRow(12, 13, 14, 15))
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(CASE WHEN UPPER\(i\.interaction_type`).WillReturnRows(
		sqlmock.NewRows([]string{"l", "c", "cl", "s"}).AddRow(16, 17, 18, 19))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(20))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comments c JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(21))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku d JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(22))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM danmaku d JOIN manuscripts m`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(23))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id.*created_at`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(24))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id.*created_at`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(25))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM space_visits`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(26))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM space_visits`).WillReturnRows(
		sqlmock.NewRows([]string{"c"}).AddRow(27))
}

func TestOverview_DefaultDays(t *testing.T) {
	// days 既非 7 也非 30 → 钳制为 7（repo 内分支）。
	repo, mock := newMockRepo(t)
	mockOverviewQueries(mock)

	result, err := repo.Overview(context.Background(), 1, 15)
	require.NoError(t, err)
	assert.Equal(t, 7, result["periodDays"])
	assert.Equal(t, int64(1), result["totalViews"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: Trend default days branch ----------

func TestTrend_RepoDefaultDays(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT metric_date::text`).WillReturnRows(sqlmock.NewRows([]string{"d", "v"}))
	mock.ExpectQuery(`SELECT DATE\(i\.created_at\)::text`).WillReturnRows(sqlmock.NewRows([]string{"d", "l", "c", "cl", "s"}))
	mock.ExpectQuery(`SELECT DATE\(c\.created_at\)::text`).WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))
	mock.ExpectQuery(`SELECT DATE\(created_at\)::text, COUNT\(\*\) FROM follows`).WillReturnRows(sqlmock.NewRows([]string{"d", "c"}))

	result, err := repo.Trend(context.Background(), 1, 0)
	require.NoError(t, err)
	dates := result["dates"].([]string)
	assert.GreaterOrEqual(t, len(dates), 6)
	assert.LessOrEqual(t, len(dates), 7)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: Ranking coins sort + query error ----------

func TestRanking_SortByCoins(t *testing.T) {
	repo, mock := newMockRepo(t)
	columns := []string{"id", "title", "cover_url", "view_count", "like_count", "comment_count",
		"danmaku_count", "coin_count", "collect_count", "share_count", "upload_time"}
	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id.*coin_count DESC`).WillReturnRows(
		sqlmock.NewRows(columns).AddRow(1, "CoinKing", "c.jpg", 100, 10, 5, 3, 50, 2, 1, nil))

	result, err := repo.Ranking(context.Background(), 1, "coins", 10)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, "CoinKing", list[0]["title"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRanking_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT.*FROM manuscripts WHERE user_id`).WillReturnError(errors.New("db boom"))

	_, err := repo.Ranking(context.Background(), 1, "views", 10)
	require.Error(t, err)
	assert.EqualError(t, err, "db boom")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: LatestComments query error ----------

func TestLatestComments_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT c\.id, c\.content`).WillReturnError(errors.New("db boom"))

	_, err := repo.LatestComments(context.Background(), 1, 10)
	require.Error(t, err)
	assert.EqualError(t, err, "db boom")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: FansTrend fallback + running<0 ----------

func TestFansTrend_FallbackQuery(t *testing.T) {
	// 第一条 user_interactions 查询失败 → rows==nil → 回退到 follows 查询。
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT DATE\(created_at\)::text.*FROM user_interactions WHERE target_id = \$1 AND target_type = 'USER'`).
		WillReturnError(errors.New("interactions unavailable"))
	mock.ExpectQuery(`SELECT DATE\(created_at\)::text, COUNT\(\*\) FROM follows WHERE following_id = \$1 AND created_at >= \$2 GROUP BY`).
		WillReturnRows(sqlmock.NewRows([]string{"d", "c"}).AddRow(time.Now().Format("2006-01-02"), 5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(100))

	result, err := repo.FansTrend(context.Background(), 1, 7)
	require.NoError(t, err)
	assert.Equal(t, int64(100), result["currentFollowers"])
	dates := result["dates"].([]string)
	assert.GreaterOrEqual(t, len(dates), 6)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansTrend_RunningNegative(t *testing.T) {
	// currentFollowers=0，totalNet=5 → running 初始为 -5，空档天触发 running<0 钳制。
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT DATE\(created_at\)::text.*FROM user_interactions WHERE target_id = \$1 AND target_type = 'USER'`).
		WillReturnRows(sqlmock.NewRows([]string{"d", "nf", "uf"}).
			AddRow(time.Now().Format("2006-01-02"), 5, 0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(0))

	result, err := repo.FansTrend(context.Background(), 1, 7)
	require.NoError(t, err)
	totalF := result["totalFollowers"].([]int64)
	for _, v := range totalF {
		assert.GreaterOrEqual(t, v, int64(0))
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansTrend_RepoDefaultDays(t *testing.T) {
	// repo 层：days < 1 → 钳制为 30。
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT DATE\(created_at\)::text.*FROM user_interactions WHERE target_id = \$1 AND target_type = 'USER'`).
		WillReturnRows(sqlmock.NewRows([]string{"d", "nf", "uf"}))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(10))

	result, err := repo.FansTrend(context.Background(), 1, 0)
	require.NoError(t, err)
	dates := result["dates"].([]string)
	assert.GreaterOrEqual(t, len(dates), 29)
	assert.LessOrEqual(t, len(dates), 30)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: FansRanking error branches ----------

func TestFansRanking_InteractionError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT u\.id, u\.username, u\.avatar, COUNT\(\*\) AS interactionCount`).
		WillReturnError(errors.New("db boom"))

	_, err := repo.FansRanking(context.Background(), 1, "interaction", 10)
	require.Error(t, err)
	assert.EqualError(t, err, "db boom")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansRanking_CommentError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT u\.id, u\.username, u\.avatar, COUNT\(DISTINCT c\.id\)`).
		WillReturnError(errors.New("db boom"))

	_, err := repo.FansRanking(context.Background(), 1, "comment", 10)
	require.Error(t, err)
	assert.EqualError(t, err, "db boom")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: ManuscriptTrend query error ----------

func TestManuscriptTrend_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, title, COALESCE\(view_count,0\)`).WillReturnError(errors.New("db boom"))

	_, err := repo.ManuscriptTrend(context.Background(), 1)
	require.Error(t, err)
	assert.EqualError(t, err, "db boom")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: ManuscriptInsights edge paths ----------

func TestManuscriptInsights_ScanErrorAndDenomClamp(t *testing.T) {
	repo, mock := newMockRepo(t)
	// 第一行：mid 无法解析为 int64 → Scan 错误 → continue。
	// 第二行：logged(10) > views(5) → denom 钳制为 logged，guestViews=0。
	mock.ExpectQuery(`FROM watch_history w`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "logged_viewers", "fan_viewers", "view_count", "avg_progress"}).
			AddRow("bad", 1, 0, 1, 0.5).
			AddRow(2, 10, 2, 5, 0.5))
	mock.ExpectQuery(`COUNT\(DISTINCT w\.user_id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"date", "fans"}))

	result, err := repo.ManuscriptInsights(context.Background(), 1, 7)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 1)
	assert.Equal(t, int64(2), list[0]["manuscriptId"])
	assert.InDelta(t, 0.2, list[0]["fanRate"], 0.001)
	assert.InDelta(t, 0.0, list[0]["guestRate"], 0.001)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: FansByManuscript ----------

func TestFansByManuscript_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "cnt"}).
			AddRow(1, 5).
			AddRow(2, 3))

	result, err := repo.FansByManuscript(context.Background(), 1)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0]["manuscriptId"])
	assert.Equal(t, int64(5), list[0]["fansGained"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansByManuscript_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "cnt"}))

	result, err := repo.FansByManuscript(context.Background(), 1)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	assert.Empty(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansByManuscript_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnError(errors.New("db boom"))

	result, err := repo.FansByManuscript(context.Background(), 1)
	require.NoError(t, err) // 出错时返回空列表，不向上抛错
	list := result["list"].([]map[string]interface{})
	assert.Empty(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFansByManuscript_ScanError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "cnt"}).
			AddRow("bad", 1).
			AddRow(3, 7))

	result, err := repo.FansByManuscript(context.Background(), 1)
	require.NoError(t, err)
	list := result["list"].([]map[string]interface{})
	require.Len(t, list, 1) // 坏行被跳过
	assert.Equal(t, int64(3), list[0]["manuscriptId"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Repository: Income ----------

func TestIncome_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND DATE\(created_at\)`).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(12345))
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND type = \$2`).WillReturnRows(
		sqlmock.NewRows([]string{"charge"}).AddRow(500))

	result, err := repo.Income(context.Background(), 1, 30)
	require.NoError(t, err)
	assert.InDelta(t, 123.45, result["totalAmount"], 0.001)
	assert.InDelta(t, 5.0, result["chargeAmount"], 0.001)
	assert.Equal(t, 30, result["periodDays"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIncome_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND DATE\(created_at\)`).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(0))
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND type = \$2`).WillReturnRows(
		sqlmock.NewRows([]string{"charge"}).AddRow(0))

	result, err := repo.Income(context.Background(), 99, 7)
	require.NoError(t, err)
	assert.Equal(t, float64(0), result["totalAmount"])
	assert.Equal(t, float64(0), result["chargeAmount"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Service wrapper coverage ----------

func TestServiceFansByManuscript(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "cnt"}).AddRow(1, 5))

	result, err := svc.FansByManuscript(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, result["list"].([]map[string]interface{}), 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceManuscriptInsights_DefaultDays(t *testing.T) {
	// days=15 → 钳制为 7（Service 内分支）。
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	mock.ExpectQuery(`FROM watch_history w`).WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "logged_viewers", "fan_viewers", "view_count", "avg_progress"}))
	mock.ExpectQuery(`COUNT\(DISTINCT w\.user_id\)`).WillReturnRows(sqlmock.NewRows([]string{"date", "fans"}))

	result, err := svc.ManuscriptInsights(context.Background(), 1, 15)
	require.NoError(t, err)
	ft := result["fanTrend"].(map[string]interface{})
	assert.Len(t, ft["dates"], 7)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceManuscriptInsights_Days30(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	mock.ExpectQuery(`FROM watch_history w`).WillReturnRows(sqlmock.NewRows([]string{"manuscript_id", "logged_viewers", "fan_viewers", "view_count", "avg_progress"}))
	mock.ExpectQuery(`COUNT\(DISTINCT w\.user_id\)`).WillReturnRows(sqlmock.NewRows([]string{"date", "fans"}))

	result, err := svc.ManuscriptInsights(context.Background(), 1, 30)
	require.NoError(t, err)
	ft := result["fanTrend"].(map[string]interface{})
	assert.Len(t, ft["dates"], 30)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceIncome(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND DATE\(created_at\)`).WillReturnRows(
		sqlmock.NewRows([]string{"total"}).AddRow(100))
	mock.ExpectQuery(`FROM income_records WHERE user_id = \$1 AND type = \$2`).WillReturnRows(
		sqlmock.NewRows([]string{"charge"}).AddRow(50))

	result, err := svc.Income(context.Background(), 1, 30)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, result["totalAmount"], 0.001)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceAccountDiagnosis(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT m\.id\)`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}).AddRow(1, 100, 10, 1))
	mock.ExpectQuery(`WHERE m\.user_id <> \$1`).WillReturnRows(
		sqlmock.NewRows([]string{"uploads", "views", "likes", "newFollowers"}))
	mock.ExpectQuery(`FROM watch_history wh`).WillReturnRows(
		sqlmock.NewRows([]string{"guests", "fans"}).AddRow(0, 0))
	mock.ExpectQuery(`SELECT wh\.client_type, COUNT\(\*\)`).WillReturnRows(
		sqlmock.NewRows([]string{"client_type", "cnt"}))
	mock.ExpectQuery(`SELECT MAX\(metric_date\)`).WillReturnRows(
		sqlmock.NewRows([]string{"d"}).AddRow(nil))
	mock.ExpectQuery(`FROM manuscript_daily_metrics dm`).WillReturnRows(
		sqlmock.NewRows([]string{"manuscript_id", "title", "view_count"}))

	result, err := svc.AccountDiagnosis(context.Background(), 1)
	require.NoError(t, err)
	metrics := result["metrics"].(map[string]interface{})
	assert.Equal(t, int64(1), metrics["uploads"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------- Context cancellation propagation ----------

func TestRepository_ContextCanceled(t *testing.T) {
	repo, mock := newMockRepo(t)
	// 模拟底层驱动返回 context.Canceled（FansByManuscript 吞掉错误并返回空列表）。
	mock.ExpectQuery(`SELECT f\.source_manuscript_id, COUNT\(\*\)`).WillReturnError(context.Canceled)

	result, err := repo.FansByManuscript(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, result["list"].([]map[string]interface{}))
	assert.NoError(t, mock.ExpectationsWereMet())
}
