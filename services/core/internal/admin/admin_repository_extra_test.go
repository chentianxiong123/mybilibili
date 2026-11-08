package admin

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// InitPermissions 依次插入 21 个权限码，再查询全部权限并为角色 1 授予。
const initPermsCount = 21

// ---- InitPermissions ----

func TestInitPermissions_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	// 21 个 upsert
	for i := 0; i < initPermsCount; i++ {
		mock.ExpectExec(`INSERT INTO permissions`).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	// 查询全部权限 id
	mock.ExpectQuery(`SELECT id FROM permissions`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
	// 每个 id 都授予角色 1
	for i := 0; i < 3; i++ {
		mock.ExpectExec(`INSERT INTO role_permissions`).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	require.NoError(t, repo.InitPermissions(context.Background()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInitPermissions_InsertError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`INSERT INTO permissions`).WillReturnError(sql.ErrConnDone)

	err := repo.InitPermissions(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInitPermissions_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)

	for i := 0; i < initPermsCount; i++ {
		mock.ExpectExec(`INSERT INTO permissions`).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectQuery(`SELECT id FROM permissions`).WillReturnError(sql.ErrConnDone)

	err := repo.InitPermissions(context.Background())
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- GetAdminRoles (repo) ----

func TestGetAdminRoles_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT role_id FROM admin_user_roles WHERE admin_user_id = \$1`).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}).AddRow(1).AddRow(2).AddRow(7))

	ids, err := repo.GetAdminRoles(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 7}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAdminRoles_QueryError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT role_id FROM admin_user_roles WHERE admin_user_id = \$1`).
		WillReturnError(sql.ErrConnDone)

	ids, err := repo.GetAdminRoles(context.Background(), 9)
	require.Error(t, err)
	assert.Nil(t, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- SetAdminRoles (repo) ----

func TestSetAdminRoles_Success(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM admin_user_roles WHERE admin_user_id = \$1`).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO admin_user_roles`).
		WithArgs(int64(3), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO admin_user_roles`).
		WithArgs(int64(3), int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.SetAdminRoles(context.Background(), 3, []int64{2, 5}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetAdminRoles_BeginError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	err := repo.SetAdminRoles(context.Background(), 3, []int64{2})
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- RecordLoginLog (repo, service wrapper) ----

func TestService_RecordLoginLog(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)

	mock.ExpectExec(`INSERT INTO login_logs`).
		WithArgs(int64(7), "10.0.0.2", "curl/8", int32(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.RecordLoginLog(context.Background(), 7, "10.0.0.2", "curl/8", 0))
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- UpdateScheduledTaskRun (repo + service wrapper) ----

func TestUpdateScheduledTaskRun_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	next := time.Now().Add(24 * time.Hour)

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("success", "done", next, "cleanup").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateScheduledTaskRun(context.Background(), "cleanup", "success", "done", &next))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateScheduledTaskRun_Error(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("failed", "boom", sqlmock.AnyArg(), "cleanup").
		WillReturnError(sql.ErrConnDone)

	err := repo.UpdateScheduledTaskRun(context.Background(), "cleanup", "failed", "boom", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateScheduledTaskRun(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)

	mock.ExpectExec(`UPDATE scheduled_tasks SET last_run_at=NOW\(\)`).
		WithArgs("skipped", "unknown", sqlmock.AnyArg(), "key").
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, svc.UpdateScheduledTaskRun(context.Background(), "key", "skipped", "unknown", nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- Service wrappers: SetAdminRoles / GetAdminRoles ----

func TestService_SetAdminRoles(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM admin_user_roles WHERE admin_user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO admin_user_roles`).
		WithArgs(int64(1), int64(4)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, svc.SetAdminRoles(context.Background(), 1, []int64{4}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAdminRoles(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)

	mock.ExpectQuery(`SELECT role_id FROM admin_user_roles WHERE admin_user_id = \$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}).AddRow(3).AddRow(6))

	ids, err := svc.GetAdminRoles(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, []int64{3, 6}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}
