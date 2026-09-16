package ai

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepoAndMock(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

var errDB = errors.New("db error")

const (
	configCols = "id,name,type,base_url,api_key,model,max_tokens,temperature,enabled,extra_config,created_at"
	skillCols  = "id,name,description,system_prompt,few_shot_examples,type,enabled,created_at"
)

func configRow(id int64, name string) []driver.Value {
	return []driver.Value{id, name, "LLM", "http://localhost", "sk-x", "gpt-4o", 4096, 0.7, 1, "", time.Now()}
}

func skillRow(id int64, name string, enabled int32) []driver.Value {
	return []driver.Value{id, name, "desc", "prompt", "", "CUSTOMER_SERVICE", enabled, time.Now()}
}

func TestRepository_ListConfigs_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT id, name, COALESCE\(type,'LLM'\)`).WillReturnError(errDB)
	list, err := repo.ListConfigs(context.Background())
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListConfigsByType_Ok(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_api_configs WHERE COALESCE\(type,'LLM'\)`).
		WithArgs("LLM").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "base_url", "api_key", "model", "max_tokens", "temperature", "enabled", "extra_config", "created_at",
		}).AddRow(configRow(1, "cfg1")...).AddRow(configRow(2, "cfg2")...))
	list, err := repo.ListConfigsByType(context.Background(), "LLM")
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "cfg1", list[0].Name)
	assert.Equal(t, "cfg2", list[1].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListConfigsByType_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_api_configs WHERE COALESCE\(type,'LLM'\)`).WillReturnError(errDB)
	list, err := repo.ListConfigsByType(context.Background(), "LLM")
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_ListConfigsByType(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_api_configs WHERE COALESCE\(type,'LLM'\)`).
		WithArgs("ASR").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "base_url", "api_key", "model", "max_tokens", "temperature", "enabled", "extra_config", "created_at",
		}).AddRow(configRow(1, "asr1")...))
	svc := NewService(repo)
	list, err := svc.ListConfigsByType(context.Background(), "ASR")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "asr1", list[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetConfig_NotFound(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_api_configs WHERE id`).WithArgs(999).WillReturnError(sql.ErrNoRows)
	c, err := repo.GetConfig(context.Background(), 999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, c)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetConfig_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_api_configs WHERE id`).WithArgs(1).WillReturnError(errDB)
	c, err := repo.GetConfig(context.Background(), 1)
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, c)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListAllBindings_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT feature, api_config_id FROM ai_bindings`).
		WillReturnRows(sqlmock.NewRows([]string{"feature", "api_config_id"}).
			AddRow("chat", "not-a-number"))
	m, err := repo.ListAllBindings(context.Background())
	assert.Error(t, err)
	assert.Nil(t, m)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListSkills_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT id, name, description, system_prompt`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
		}).AddRow("bad", "x", "y", "z", "", "T", 1, time.Now()))
	list, err := repo.ListSkills(context.Background())
	assert.Error(t, err)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListSkillsByType_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).WithArgs("MODERATION").WillReturnError(errDB)
	list, err := repo.ListSkillsByType(context.Background(), "MODERATION")
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSkill_NotFound(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE id`).WithArgs(99).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetSkill(context.Background(), 99)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateSkill_Ok(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`INSERT INTO ai_skills`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	svc := NewService(repo)
	err := svc.CreateSkill(context.Background(), &Skill{Name: "新技能", Type: "LLM"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateSkill_Error(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`INSERT INTO ai_skills`).WillReturnError(errDB)
	svc := NewService(repo)
	err := svc.CreateSkill(context.Background(), &Skill{Name: "新技能"})
	assert.ErrorIs(t, err, errDB)
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_avgDuration_Zero(t *testing.T) {
	assert.Equal(t, float64(0), avgDuration(0, 0))
	assert.Equal(t, float64(123.5), avgDuration(247, 2))
}

func TestRepository_UsageOverview_QueryErrors(t *testing.T) {
	queries := []struct {
		re  string
		row []string
	}{
		{`SELECT COUNT\(\*\) FROM ai_usage_logs`, []string{"c"}},
		{`SELECT COALESCE\(SUM\(token_count\),0\) FROM ai_usage_logs`, []string{"s"}},
		{`SELECT COALESCE\(SUM\(duration_ms\),0\) FROM ai_usage_logs`, []string{"s"}},
		{`SELECT COUNT\(DISTINCT feature\) FROM ai_usage_logs`, []string{"c"}},
		{`SELECT COUNT\(DISTINCT user_id\) FROM ai_usage_logs WHERE user_id IS NOT NULL`, []string{"c"}},
	}
	for i, _ := range queries {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			repo, mock := newRepoAndMock(t)
			for j := 0; j <= i; j++ {
				expect := mock.ExpectQuery(queries[j].re)
				if j < i {
					expect.WillReturnRows(sqlmock.NewRows(queries[j].row).AddRow(1))
				} else {
					expect.WillReturnError(errDB)
				}
			}
			_, err := repo.UsageOverview(context.Background())
			assert.ErrorIs(t, err, errDB)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepository_UsageByFeature_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT feature, COUNT\(\*\)`).WillReturnError(errDB)
	out, err := repo.UsageByFeature(context.Background())
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UsageByFeature_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT feature, COUNT\(\*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"feature", "calls", "tokens", "avg_duration_ms"}).
			AddRow("chat", "bad", 1, 1.0))
	out, err := repo.UsageByFeature(context.Background())
	assert.Error(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UsageDaily_QueryError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT TO_CHAR\(created_at`).WithArgs("2026-01-01").WillReturnError(errDB)
	out, err := repo.UsageDaily(context.Background(), "2026-01-01")
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UsageDaily_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT TO_CHAR\(created_at`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "calls", "tokens"}).
			AddRow("2026-01-01", "bad", 1))
	out, err := repo.UsageDaily(context.Background(), "2026-01-01")
	assert.Error(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListPendingSessions_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_conversations WHERE status = 0`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "status", "created_at"}).
			AddRow("bad", 1, "t", 0, time.Now()))
	out, err := repo.ListPendingSessions(context.Background())
	assert.Error(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetSessionMessages_ScanError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_chat_messages WHERE conversation_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "content", "token_count", "created_at"}).
			AddRow("bad", "user", "hi", 1, time.Now()))
	out, err := repo.GetSessionMessages(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_SendSessionReply_InsertError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectExec(`INSERT INTO ai_chat_messages`).WillReturnError(errDB)
	err := repo.SendSessionReply(context.Background(), 1, 2, "hi")
	assert.ErrorIs(t, err, errDB)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_SendSessionReply_UpdateError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectExec(`INSERT INTO ai_chat_messages`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE ai_conversations SET updated_at`).WillReturnError(errDB)
	err := repo.SendSessionReply(context.Background(), 1, 2, "hi")
	assert.ErrorIs(t, err, errDB)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateMissingCustomerServiceDefaults_AllExisting(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	for i := 0; i < 6; i++ {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_skills WHERE name`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	}
	svc := NewService(repo)
	created, err := svc.CreateMissingCustomerServiceDefaults(context.Background())
	require.NoError(t, err)
	assert.Zero(t, created)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateMissingCustomerServiceDefaults_CountError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_skills WHERE name`).WillReturnError(errDB)
	svc := NewService(repo)
	created, err := svc.CreateMissingCustomerServiceDefaults(context.Background())
	require.NoError(t, err)
	assert.Zero(t, created)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreateMissingCustomerServiceDefaults_CreateError(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM ai_skills WHERE name`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`INSERT INTO ai_skills`).WillReturnError(errDB)
	svc := NewService(repo)
	created, err := svc.CreateMissingCustomerServiceDefaults(context.Background())
	require.NoError(t, err)
	assert.Zero(t, created)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_MatchCustomerServiceSkill_Error(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).WithArgs("CUSTOMER_SERVICE").WillReturnError(errDB)
	svc := NewService(repo)
	out, err := svc.MatchCustomerServiceSkill(context.Background(), "账号")
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RouteSkills_Error(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).WithArgs("CUSTOMER_SERVICE").WillReturnError(errDB)
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "账号", 3)
	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func skillsRows(rows ...[]driver.Value) *sqlmock.Rows {
	rr := sqlmock.NewRows([]string{
		"id", "name", "description", "system_prompt", "few_shot_examples", "type", "enabled", "created_at",
	})
	for _, r := range rows {
		rr.AddRow(r...)
	}
	return rr
}

func TestService_RouteSkills_ExactMatchAndLimit(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).
		WithArgs("CUSTOMER_SERVICE").
		WillReturnRows(skillsRows(
			skillRow(1, "账号问题", 1),
			skillRow(2, "直播问题", 1),
			skillRow(3, "支付问题", 1),
		))
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "账号问题", 1)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, int64(1), out[0]["id"])
	// 完全匹配 100 + 单关键词命中 10 = 110
	assert.Equal(t, 110, out[0]["score"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RouteSkills_WordScoringAndDisabled(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).
		WithArgs("CUSTOMER_SERVICE").
		WillReturnRows(skillsRows(
			skillRow(1, "篮球训练", 0),
			skillRow(2, "篮球场技巧", 1),
			skillRow(3, "完全无关技能", 1),
		))
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "我要 篮球 训练", 3)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, int64(2), out[0]["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RouteSkills_DefaultLimitAndNoMatch(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).
		WithArgs("CUSTOMER_SERVICE").
		WillReturnRows(skillsRows(
			skillRow(1, "完全无关的技能", 1),
		))
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "xyz", 0)
	require.NoError(t, err)
	assert.Empty(t, out)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RouteSkills_SortByScore(t *testing.T) {
	repo, mock := newRepoAndMock(t)
	mock.ExpectQuery(`FROM ai_skills WHERE type`).
		WithArgs("CUSTOMER_SERVICE").
		WillReturnRows(skillsRows(
			skillRow(1, "科技评测", 1),
			skillRow(2, "评测技巧", 1),
		))
	svc := NewService(repo)
	out, err := svc.RouteSkills(context.Background(), "科技评测 技巧", 3)
	require.NoError(t, err)
	require.Len(t, out, 2)
	// 得分高的技能排在首位
	assert.Equal(t, int64(1), out[0]["id"])
	assert.Equal(t, int64(2), out[1]["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}
