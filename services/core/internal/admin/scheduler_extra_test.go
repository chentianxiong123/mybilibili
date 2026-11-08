package admin

import (
	"context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// prefixMatcher 断言字符串参数以指定前缀开头（用于动态错误消息）。
type prefixMatcher struct{ prefix string }

func (p prefixMatcher) Match(v driver.Value) bool {
	s, ok := v.(string)
	return ok && strings.HasPrefix(s, p.prefix)
}

// waitExpectationsMet 轮询等待所有 sqlmock 期望被消费。
// checkTasks 会在后台 goroutine 里调用 runTask，必须等它落地后再收尾。
func waitExpectationsMet(t *testing.T, mock sqlmock.Sqlmock, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := mock.ExpectationsWereMet(); err == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func scheduledRow(taskKey, taskType string, enabled int32, nextRunAt *time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "task_key", "task_name", "description", "cron_expr", "task_type", "task_config",
		"enabled", "last_run_at", "last_run_result", "last_run_message",
		"next_run_at", "run_count", "max_retries", "retry_count", "timeout_seconds",
		"created_at", "updated_at",
	}).AddRow(1, taskKey, "task", "desc", "0 * * * *", taskType, "{}",
		enabled, nil, "", "", nextRunAt, 0, 3, 0, 60, time.Now(), time.Now())
}

// ---- checkTasks ----

func TestCheckTasks_ListError(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).WillReturnError(context.DeadlineExceeded)

	s.checkTasks(context.Background()) // 出错即返回，不 panic
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckTasks_DisabledTask(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).
		WillReturnRows(scheduledRow("k1", "unknown_type", 0, nil))

	s.checkTasks(context.Background()) // Enabled==0 -> continue，不触发 runTask
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckTasks_FutureNextRun(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	future := time.Now().Add(time.Hour)
	mock.ExpectQuery(`SELECT id, task_key, task_name`).
		WillReturnRows(scheduledRow("k1", "unknown_type", 1, &future))

	s.checkTasks(context.Background()) // NextRunAt 在未来 -> continue
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckTasks_DueTask(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).
		WillReturnRows(scheduledRow("k1", "unknown_type", 1, nil))
	// runTask 在后台 goroutine 里执行：未知类型 -> skipped -> 更新任务
	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown task type: unknown_type", sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.checkTasks(context.Background())
	waitExpectationsMet(t, mock, 2*time.Second)
}

// ---- runTask ----

func TestRunTask_UnknownType(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown task type: weird", sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.runTask(context.Background(), &ScheduledTask{TaskKey: "k1", TaskType: "weird"})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunTask_HotSearch_HTTPError(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	// 起一个 server 再关掉，保证连接被拒。
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := strings.TrimPrefix(ts.URL, "http://")
	ts.Close()
	t.Setenv("SEARCH_SERVICE_ADDR", addr)

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("failed", prefixMatcher{"cleanup request failed: "}, sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.runTask(context.Background(), &ScheduledTask{TaskKey: "k1", TaskType: "hot_search_cleanup"})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunTask_HotSearch_Non200(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	t.Setenv("SEARCH_SERVICE_ADDR", strings.TrimPrefix(ts.URL, "http://"))

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("failed", "cleanup returned status 500", sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.runTask(context.Background(), &ScheduledTask{TaskKey: "k1", TaskType: "hot_search_cleanup"})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunTask_HotSearch_200(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	t.Setenv("SEARCH_SERVICE_ADDR", strings.TrimPrefix(ts.URL, "http://"))

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("success", "hot search expired entries cleaned", sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.runTask(context.Background(), &ScheduledTask{TaskKey: "k1", TaskType: "hot_search_cleanup"})
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunTask_UpdateError(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown task type: weird", sqlmock.AnyArg(), "k1").
		WillReturnError(context.DeadlineExceeded)

	s.runTask(context.Background(), &ScheduledTask{TaskKey: "k1", TaskType: "weird"})
	require.NoError(t, mock.ExpectationsWereMet()) // 仅 log.Printf，不 panic
}

// ---- TriggerTaskNow ----

func TestTriggerTaskNow_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).WithArgs("missing").
		WillReturnError(context.DeadlineExceeded)

	err := s.TriggerTaskNow(context.Background(), "missing")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTriggerTaskNow_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).WithArgs("k1").
		WillReturnRows(scheduledRow("k1", "unknown_type", 1, nil))
	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown task type: unknown_type", sqlmock.AnyArg(), "k1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.TriggerTaskNow(context.Background(), "k1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TriggerTaskNow 同步跑 runTask（成功路径），断言结果正确。
func TestTriggerTaskNow_ResultApplied(t *testing.T) {
	repo, mock := newMockRepo(t)
	s := NewScheduler(NewService(repo))

	mock.ExpectQuery(`SELECT id, task_key, task_name`).WithArgs("cleanup").
		WillReturnRows(scheduledRow("cleanup", "weird", 1, nil))
	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown task type: weird", sqlmock.AnyArg(), "cleanup").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, s.TriggerTaskNow(context.Background(), "cleanup"))
	require.NoError(t, mock.ExpectationsWereMet())
}
