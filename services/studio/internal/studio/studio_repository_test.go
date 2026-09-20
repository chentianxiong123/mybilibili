package studio

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

func TestCreateTask_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO operation_tasks.*RETURNING id`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	task, err := repo.CreateTask(ctx, 1, "proj_123")
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, itoa(42), task.ID)
	assert.Equal(t, int64(1), task.UserID)
	assert.Equal(t, "proj_123", task.ProjectID)
	assert.Equal(t, "PENDING", task.Status)
	assert.Equal(t, int32(0), task.Progress)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTask_DBError(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO operation_tasks`).WillReturnError(assert.AnError)

	task, err := repo.CreateTask(ctx, 1, "proj_123")
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestGetTask_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.Local)
	updatedAt := time.Date(2025, 1, 1, 10, 5, 0, 0, time.Local)

	rows := sqlmock.NewRows([]string{"id", "task_key", "task_name", "status", "progress", "message", "error_message", "created_at", "updated_at"}).
		AddRow(42, "studio_proj_123_1700000000", "Studio Export", "RUNNING", 50, "processing", "", createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, task_key, task_name, status, progress, message, error_message, created_at, updated_at`).
		WillReturnRows(rows)

	task, err := repo.GetTask(ctx, itoa(42))
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, itoa(42), task.ID)
	assert.Equal(t, "studio_proj_123_1700000000", task.TaskKey)
	assert.Equal(t, "Studio Export", task.TaskName)
	assert.Equal(t, "RUNNING", task.Status)
	assert.Equal(t, int32(50), task.Progress)
	assert.Equal(t, "processing", task.Message)
	assert.Equal(t, createdAt, task.CreatedAt)
	assert.Equal(t, updatedAt, task.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTask_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, task_key, task_name, status, progress, message, error_message, created_at, updated_at`).
		WillReturnError(assert.AnError)

	task, err := repo.GetTask(ctx, "999")
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestCancelTask_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE operation_tasks SET status = 'CANCELLED' WHERE id = \$1 AND status = 'PENDING'`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CancelTask(ctx, itoa(42))
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancelTask_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE operation_tasks SET status = 'CANCELLED' WHERE id = \$1 AND status = 'PENDING'`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.CancelTask(ctx, "999")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{999999, "999999"},
		{10, "10"},
		{100, "100"},
		{123456789, "123456789"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itoa(tt.input))
		})
	}
}
