package admin

import (
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// doRawReq 发送原始请求体（用于触发 JSON 解码错误分支）。
func doRawReq(t *testing.T, h http.Handler, method, path, rawBody string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ---- handleLogout ----

func TestHandleLogout_200(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doReq(t, mux, "POST", "/api/v1/admin/logout", nil, nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleLogout_405_Method(t *testing.T) {
	h, _ := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := doReq(t, mux, "GET", "/api/v1/admin/logout", nil, nil)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// ---- handleListAdmins: 内部 403 分支（绕过 requirePerm 直接调用）----

func TestHandleListAdmins_Internal403(t *testing.T) {
	h, mock := newTestHandler(t)

	// 无 X-Admin-Id、无 token -> requirePermission 直接返回 (0,false)，不触库。
	req := httptest.NewRequest("GET", "/api/v1/admin/list", nil)
	rec := httptest.NewRecorder()
	h.handleListAdmins(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleRoles: 内部 403 分支 ----

func TestHandleRoles_Internal403(t *testing.T) {
	h, mock := newTestHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/admin/roles", nil)
	rec := httptest.NewRecorder()
	h.handleRoles(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleLoginLogs: 过滤/错误分支 ----

func TestHandleLoginLogs_Filters(t *testing.T) {
	cases := []struct {
		name        string
		query       string
		selectRegex string
		selectArgs  []driver.Value
		countArgs   []driver.Value
	}{
		{
			name:        "no_filter",
			query:       "",
			selectRegex: `FROM login_logs WHERE 1=1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int32(20), int32(0)},
			countArgs:   []driver.Value{},
		},
		{
			name:        "user_id",
			query:       "?user_id=10",
			selectRegex: `FROM login_logs WHERE 1=1 AND user_id = \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int64(10), int32(20), int32(0)},
			countArgs:   []driver.Value{int64(10)},
		},
		{
			name:        "user_id_camel",
			query:       "?userId=11",
			selectRegex: `FROM login_logs WHERE 1=1 AND user_id = \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int64(11), int32(20), int32(0)},
			countArgs:   []driver.Value{int64(11)},
		},
		{
			name:        "user_id_invalid",
			query:       "?user_id=abc",
			selectRegex: `FROM login_logs WHERE 1=1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int32(20), int32(0)},
			countArgs:   []driver.Value{},
		},
		{
			name:        "user_id_zero",
			query:       "?user_id=0",
			selectRegex: `FROM login_logs WHERE 1=1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int32(20), int32(0)},
			countArgs:   []driver.Value{},
		},
		{
			name:        "ip",
			query:       "?ip=127",
			selectRegex: `FROM login_logs WHERE 1=1 AND ip ILIKE \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{"%127%", int32(20), int32(0)},
			countArgs:   []driver.Value{"%127%"},
		},
		{
			name:        "status",
			query:       "?status=1",
			selectRegex: `FROM login_logs WHERE 1=1 AND status = \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int32(1), int32(20), int32(0)},
			countArgs:   []driver.Value{int32(1)},
		},
		{
			name:        "status_invalid",
			query:       "?status=abc",
			selectRegex: `FROM login_logs WHERE 1=1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int32(20), int32(0)},
			countArgs:   []driver.Value{},
		},
		{
			name:        "start_time_camel",
			query:       "?startTime=2025-01-01",
			selectRegex: `FROM login_logs WHERE 1=1 AND login_time >= \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{"2025-01-01", int32(20), int32(0)},
			countArgs:   []driver.Value{"2025-01-01"},
		},
		{
			name:        "start_time_snake",
			query:       "?start_time=2025-01-02",
			selectRegex: `FROM login_logs WHERE 1=1 AND login_time >= \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{"2025-01-02", int32(20), int32(0)},
			countArgs:   []driver.Value{"2025-01-02"},
		},
		{
			name:        "end_time_camel",
			query:       "?endTime=2025-02-01",
			selectRegex: `FROM login_logs WHERE 1=1 AND login_time <= \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{"2025-02-01", int32(20), int32(0)},
			countArgs:   []driver.Value{"2025-02-01"},
		},
		{
			name:        "end_time_snake",
			query:       "?end_time=2025-02-02",
			selectRegex: `FROM login_logs WHERE 1=1 AND login_time <= \$1 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{"2025-02-02", int32(20), int32(0)},
			countArgs:   []driver.Value{"2025-02-02"},
		},
		{
			name:        "user_id_and_status",
			query:       "?user_id=10&status=1",
			selectRegex: `FROM login_logs WHERE 1=1 AND user_id = \$1 AND status = \$2 ORDER BY login_time DESC`,
			selectArgs:  []driver.Value{int64(10), int32(1), int32(20), int32(0)},
			countArgs:   []driver.Value{int64(10), int32(1)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, mock := newTestHandler(t)
			mux := http.NewServeMux()
			h.Register(mux)

			expectPermQuery(mock, 1, []string{"security:manage"})
			mock.ExpectQuery(tc.selectRegex).WithArgs(tc.selectArgs...).
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
					AddRow(1, 10, "127.0.0.1", "Mozilla/5.0", 1, "2025-01-01"))
			mock.ExpectQuery(`SELECT COUNT`).WithArgs(tc.countArgs...).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

			rec := doReq(t, mux, "GET", "/api/v1/admin/login-logs"+tc.query, nil,
				map[string]string{"X-Admin-Id": "1"})
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"list"`)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandleLoginLogs_QueryError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"security:manage"})
	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs`).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	rec := doReq(t, mux, "GET", "/api/v1/admin/login-logs", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list":[]`)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- handleScheduledTasks: 错误/边界分支 ----

func TestHandleScheduledTasks_POST_InvalidBody(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doRawReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks", "{invalid",
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid body")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_POST_CreateError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectQuery(`INSERT INTO scheduled_tasks`).WillReturnError(sql.ErrConnDone)

	rec := doReq(t, mux, "POST", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{
			"task_key": "t", "task_name": "T", "cron_expr": "0 * * * *",
			"task_type": "cron", "task_config": "{}", "enabled": 1, "max_retries": 3, "timeout_seconds": 60,
		}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "操作失败")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_PUT_InvalidBody(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})

	rec := doRawReq(t, mux, "PUT", "/api/v1/admin/scheduled-tasks", "{invalid",
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid body")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_PUT_UpdateError(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectExec(`UPDATE scheduled_tasks SET task_name`).WillReturnError(sql.ErrConnDone)

	rec := doReq(t, mux, "PUT", "/api/v1/admin/scheduled-tasks",
		map[string]interface{}{
			"id": 1, "task_name": "U", "description": "d", "cron_expr": "0 * * * *",
			"task_type": "cron", "task_config": "{}", "enabled": 1, "max_retries": 3, "timeout_seconds": 60,
		}, map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "操作失败")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleScheduledTasks_GET_Empty(t *testing.T) {
	h, mock := newTestHandler(t)
	mux := http.NewServeMux()
	h.Register(mux)

	expectPermQuery(mock, 1, []string{"scheduled:manage"})
	mock.ExpectQuery(`SELECT id, task_key, task_name`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_key", "task_name", "description", "cron_expr", "task_type", "task_config",
			"enabled", "last_run_at", "last_run_result", "last_run_message",
			"next_run_at", "run_count", "max_retries", "retry_count", "timeout_seconds",
			"created_at", "updated_at",
		}))

	rec := doReq(t, mux, "GET", "/api/v1/admin/scheduled-tasks", nil,
		map[string]string{"X-Admin-Id": "1"})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"list":[]`)
	require.NoError(t, mock.ExpectationsWereMet())
}
