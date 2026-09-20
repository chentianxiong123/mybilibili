package support

import (
	"context"
	"database/sql"
	"errors"
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

func TestCreate_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WithArgs(sqlmock.AnyArg(), int64(1), "USER_FEEDBACK", "GENERAL", "NORMAL", "title", "content").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	tk, err := repo.Create(context.Background(), 1, "USER_FEEDBACK", "GENERAL", "NORMAL", "title", "content")
	require.NoError(t, err)
	assert.Equal(t, int64(1), tk.ID)
	assert.Equal(t, now, tk.CreatedAt)
	assert.Equal(t, int64(1), tk.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`INSERT INTO support_tickets`).WillReturnError(errors.New("db down"))
	_, err := repo.Create(context.Background(), 1, "", "", "", "", "")
	assert.Error(t, err)
}

func TestGetByID_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()
	procTime := now.Add(time.Hour)

	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets WHERE id`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"processed_at", "created_at",
		}).AddRow(1, "TK001", 10, 0, "USER_FEEDBACK", "GENERAL", "NORMAL",
			"PENDING", "t", "c", "", "", 0, procTime, now))

	tk, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), tk.ID)
	assert.Equal(t, "TK001", tk.TicketNo)
	assert.Equal(t, "PENDING", tk.Status)
	require.NotNil(t, tk.ProcessedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets WHERE id`).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 99)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestList_AllStatus(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets ORDER BY`).
		WithArgs(int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).
			AddRow(1, "TK001", 10, 0, "USER_FEEDBACK", "GENERAL", "NORMAL", "PENDING", "a", "c1", "", "", 0, now, nil).
			AddRow(2, "TK002", 20, 0, "ADMIN", "BUG", "HIGH", "PROCESSED", "b", "c2", "", "r", 1, now, now))

	list, total, err := repo.List(context.Background(), "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, list, 2)
	assert.Equal(t, "TK001", list[0].TicketNo)
	assert.Equal(t, "TK002", list[1].TicketNo)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_FilteredByStatus(t *testing.T) {
	repo, mock := newMockRepo(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets WHERE status`).
		WithArgs("PENDING").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets WHERE status.*ORDER BY`).
		WithArgs("PENDING", int32(5), int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).AddRow(3, "TK003", 30, 0, "USER_FEEDBACK", "GENERAL", "NORMAL", "PENDING", "x", "y", "", "", 0, now, nil))

	list, total, err := repo.List(context.Background(), "PENDING", 2, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	assert.Equal(t, "TK003", list[0].TicketNo)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_Empty(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets ORDER BY`).
		WithArgs(int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}))

	list, total, err := repo.List(context.Background(), "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProcess_Success(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`UPDATE support_tickets SET status = 'PROCESSED'`).
		WithArgs("fixed", int64(99), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Process(context.Background(), 1, 99, "fixed"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Create(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO support_tickets`).
		WithArgs(sqlmock.AnyArg(), int64(5), "USER_FEEDBACK", "GENERAL", "NORMAL", "hi", "body").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(10, now))

	tk, err := svc.Create(context.Background(), 5, "hi", "body")
	require.NoError(t, err)
	assert.Equal(t, int64(10), tk.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestService_List(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	now := time.Now()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM support_tickets`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets ORDER BY`).
		WithArgs(int32(10), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"created_at", "processed_at",
		}).AddRow(1, "TK001", 1, 0, "USER_FEEDBACK", "GENERAL", "NORMAL", "PENDING", "t", "c", "", "", 0, now, nil))

	list, total, err := svc.List(context.Background(), "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
}

func TestService_GetByID(t *testing.T) {
	repo, mock := newMockRepo(t)
	svc := NewService(repo)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, ticket_no.*FROM support_tickets WHERE id`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "ticket_no", "user_id", "session_id", "source", "category", "priority",
			"status", "title", "content", "entry_reply", "admin_reply", "assignee_admin_id",
			"processed_at", "created_at",
		}).AddRow(7, "TK007", 7, 0, "USER_FEEDBACK", "GENERAL", "NORMAL", "PENDING", "x", "y", "", "", 0, nil, now))

	tk, err := svc.GetByID(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, int64(7), tk.ID)
	assert.Equal(t, "TK007", tk.TicketNo)
}
