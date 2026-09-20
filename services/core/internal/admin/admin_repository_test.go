package admin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func hexSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func expectRolesForAdmin(mock sqlmock.Sqlmock, adminID int64, roles ...Role) {
	rows := sqlmock.NewRows([]string{"id", "name", "description", "create_time"})
	for _, r := range roles {
		rows.AddRow(r.ID, r.Name, r.Description, r.CreateTime)
	}
	mock.ExpectQuery(`SELECT r.id, r.name`).WithArgs(adminID).WillReturnRows(rows)
}

func TestAdminLogin_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()
	pwd := "secret123"
	hash := hexSHA256(pwd)

	mock.ExpectQuery(`SELECT id, username, password, COALESCE\(admin_level,1\) FROM admin_users WHERE username = \$1 AND password = \$2`).
		WithArgs("admin", hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level"}).
			AddRow(1, "admin", hash, 1))
	expectRolesForAdmin(mock, 1, Role{ID: 1, Name: "super", Description: "Super Admin", CreateTime: "2025-01-01"})

	u, err := repo.AdminLogin(ctx, "admin", pwd)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, "admin", u.Username)
	assert.Equal(t, int32(1), u.AdminLevel)
	require.Len(t, u.Roles, 1)
	assert.Equal(t, "super", u.Roles[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminLogin_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, username, password, COALESCE\(admin_level,1\) FROM admin_users WHERE username = \$1 AND password = \$2`).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.AdminLogin(context.Background(), "nobody", "bad")
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestCreateAdmin_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	pwd := "mypassword"
	hash := hexSHA256(pwd)

	mock.ExpectQuery(`INSERT INTO admin_users`).
		WithArgs("newadmin", hash, int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	id, err := repo.CreateAdmin(context.Background(), "newadmin", pwd, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(42), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAdmins_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()
	hash := hexSHA256("pwd")

	mock.ExpectQuery(`SELECT id, username, password, COALESCE\(admin_level,1\), COALESCE\(created_at::text`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level", "created_at"}).
			AddRow(1, "a", hash, 1, "2025-01-01").
			AddRow(2, "b", hash, 2, "2025-02-01"))
	expectRolesForAdmin(mock, 1, Role{ID: 1, Name: "role1", Description: "desc1"})
	expectRolesForAdmin(mock, 2)

	list, err := repo.ListAdmins(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, "b", list[1].Username)
	require.Len(t, list[0].Roles, 1)
	assert.Empty(t, list[1].Roles)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAdmins_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, username, password, COALESCE\(admin_level,1\), COALESCE\(created_at::text`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "admin_level", "created_at"}))

	list, err := repo.ListAdmins(context.Background())
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestGetAdminByID_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, username, COALESCE\(admin_level,1\) FROM admin_users WHERE id=\$1`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "admin_level"}).
			AddRow(5, "testuser", 1))
	expectRolesForAdmin(mock, 5, Role{ID: 3, Name: "editor", Description: "Editor Role", CreateTime: "2025-01-01"})

	u, err := repo.GetAdminByID(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), u.ID)
	assert.Equal(t, "testuser", u.Username)
	require.Len(t, u.Roles, 1)
	assert.Equal(t, "editor", u.Roles[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAdminByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, username, COALESCE\(admin_level,1\) FROM admin_users WHERE id=\$1`).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetAdminByID(context.Background(), 999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestListRoles_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, name, COALESCE\(description,''`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "create_time"}).
			AddRow(1, "admin", "Administrator", "2025-01-01").
			AddRow(2, "editor", "Content Editor", "2025-01-02"))

	list, err := repo.ListRoles(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "admin", list[0].Name)
	assert.Equal(t, "editor", list[1].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateRole_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs("moderator", "Can moderate comments").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	id, err := repo.CreateRole(context.Background(), "moderator", "Can moderate comments")
	require.NoError(t, err)
	assert.Equal(t, int64(10), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateRole_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`UPDATE roles SET name=\$1, description=\$2`).
		WithArgs("newname", "newdesc", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateRole(context.Background(), 1, "newname", "newdesc"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRole_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`DELETE FROM roles WHERE id = \$1`).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteRole(context.Background(), 3))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListPermissions_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, name, code, COALESCE\(url,''`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "url", "method", "parent_id", "description"}).
			AddRow(1, "User Manage", "user:manage", "/api/users", "GET", 0, "manage users").
			AddRow(2, "Video Manage", "video:manage", "/api/videos", "POST", 0, "manage videos"))

	list, err := repo.ListPermissions(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "user:manage", list[0].Code)
	assert.Equal(t, "video:manage", list[1].Code)
	assert.Equal(t, "GET", list[0].Method)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRolePermissions_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT permission_id FROM role_permissions WHERE role_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"permission_id"}).
			AddRow(10).AddRow(20).AddRow(30))

	ids, err := repo.GetRolePermissions(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20, 30}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAdminPermissionCodes_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT DISTINCT p.code`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"code"}).
			AddRow("user:manage").AddRow("video:manage"))

	codes, err := repo.GetAdminPermissionCodes(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"user:manage", "video:manage"}, codes)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAuditLog_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`INSERT INTO audit_logs`).
		WithArgs(int64(1), "admin", "super", "user", "delete", "user", "42",
			"DELETE", "/api/users/42", "127.0.0.1", "curl/7.68",
			int32(1), "deleted", "detail").
		WillReturnResult(sqlmock.NewResult(0, 1))

	log := &AuditLog{
		OperatorID: 1, OperatorName: "admin", OperatorRole: "super",
		Module: "user", Action: "delete", TargetType: "user", TargetID: "42",
		RequestMethod: "DELETE", RequestURI: "/api/users/42",
		ClientIP: "127.0.0.1", UserAgent: "curl/7.68",
		Result: 1, Message: "deleted", Detail: "detail",
	}
	require.NoError(t, repo.CreateAuditLog(context.Background(), log))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListAuditLogs_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, operator_id, operator_name, operator_role, module, action, target_type, target_id`).
		WithArgs(int32(10), int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "operator_id", "operator_name", "operator_role", "module", "action",
			"target_type", "target_id", "request_method", "request_uri", "client_ip",
			"user_agent", "result", "message", "detail", "created_at",
		}).
			AddRow(1, 1, "admin", "super", "user", "delete", "user", "42",
				"DELETE", "/api/users/42", "127.0.0.1", "curl", int32(1), "ok", "", now))

	list, err := repo.ListAuditLogs(context.Background(), 2, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, "admin", list[0].OperatorName)
	assert.Equal(t, int32(1), list[0].Result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAuditLogByID_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, operator_id, operator_name, operator_role, module, action, target_type, target_id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "operator_id", "operator_name", "operator_role", "module", "action",
			"target_type", "target_id", "request_method", "request_uri", "client_ip",
			"user_agent", "result", "message", "detail", "created_at",
		}).
			AddRow(7, 2, "ops", "editor", "video", "approve", "video", "99",
				"PUT", "/api/videos/99/approve", "10.0.0.1", "curl", int32(0), "approved", "", now))

	l, err := repo.GetAuditLogByID(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, int64(7), l.ID)
	assert.Equal(t, "ops", l.OperatorName)
	assert.Equal(t, "approved", l.Message)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAuditLogByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT id, operator_id, operator_name, operator_role, module, action, target_type, target_id`).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetAuditLogByID(context.Background(), 999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestRecordLoginLog_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`INSERT INTO login_logs`).
		WithArgs(int64(1), "192.168.1.1", "Mozilla/5.0", int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.RecordLoginLog(context.Background(), 1, "192.168.1.1", "Mozilla/5.0", 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListLoginLogs_WithUserID(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := "2025-06-01 10:00:00"

	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs WHERE user_id = \$1`).
		WithArgs(int64(1), int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 1, "10.0.0.1", "curl", int64(1), now))

	list, err := repo.ListLoginLogs(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0]["id"])
	assert.Equal(t, "10.0.0.1", list[0]["ip"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListLoginLogs_AllUsers(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := "2025-06-01 10:00:00"

	mock.ExpectQuery(`SELECT id, user_id, ip, user_agent, status, login_time FROM login_logs ORDER BY login_time DESC`).
		WithArgs(int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "ip", "user_agent", "status", "login_time"}).
			AddRow(1, 1, "10.0.0.1", "curl", int64(1), now).
			AddRow(2, 2, "10.0.0.2", "chrome", int64(0), now))

	list, err := repo.ListLoginLogs(context.Background(), 0, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0]["id"])
	assert.Equal(t, int64(2), list[1]["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListScheduledTasks_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, task_key, task_name, description, cron_expr, task_type, task_config`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "task_key", "task_name", "description", "cron_expr", "task_type", "task_config",
			"enabled", "last_run_at", "last_run_result", "last_run_message",
			"next_run_at", "run_count", "max_retries", "retry_count", "timeout_seconds",
			"created_at", "updated_at",
		}).AddRow(1, "cleanup", "Cleanup Task", "Remove stale data", "0 2 * * *", "shell", "{}",
			int32(1), now, "success", "", now, int32(5), int32(3), int32(0), int32(300), now, now))

	list, err := repo.ListScheduledTasks(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "cleanup", list[0].TaskKey)
	assert.Equal(t, "shell", list[0].TaskType)
	assert.Equal(t, int32(1), list[0].Enabled)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateScheduledTask_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`INSERT INTO scheduled_tasks`).
		WithArgs("sync", "Sync Data", "Sync external data", "*/30 * * * *", "http", `{"url":"http://x"}`,
			int32(1), int32(3), int32(60)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	task := &ScheduledTask{
		TaskKey: "sync", TaskName: "Sync Data", Description: "Sync external data",
		CronExpr: "*/30 * * * *", TaskType: "http", TaskConfig: `{"url":"http://x"}`,
		Enabled: 1, MaxRetries: 3, TimeoutSecs: 60,
	}
	err := repo.CreateScheduledTask(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, int64(7), task.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteScheduledTask_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`DELETE FROM scheduled_tasks WHERE id=\$1`).
		WithArgs(int64(4)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.DeleteScheduledTask(context.Background(), 4))
	require.NoError(t, mock.ExpectationsWereMet())
}
