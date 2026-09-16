package live

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockLinkmicRepo(t *testing.T) (*LinkmicRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewLinkmicRepository(db), mock
}

func TestLinkmicConstants(t *testing.T) {
	assert.Equal(t, int32(0), LinkmicStatusApplying)
	assert.Equal(t, int32(1), LinkmicStatusConnected)
	assert.Equal(t, int32(2), LinkmicStatusDisconnected)
	assert.Equal(t, int32(3), LinkmicStatusRejected)
}

func TestLinkmicRepository_Apply(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	mock.ExpectQuery(`INSERT INTO live_linkmic`).
		WithArgs(int64(1), int64(2), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(int64(9), int64(1), int64(2), int64(3), int32(0), time.Now().Truncate(time.Second)))

	lm, err := repo.Apply(context.Background(), 1, 2, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(9), lm.ID)
	assert.Equal(t, LinkmicStatusApplying, lm.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicRepository_UpdateStatus(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	mock.ExpectExec(`UPDATE live_linkmic SET status`).
		WithArgs(int32(LinkmicStatusConnected), int32(LinkmicStatusConnected), int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateStatus(context.Background(), 9, LinkmicStatusConnected))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicRepository_ActiveByRoom(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	mock.ExpectQuery(`SELECT id, room_id, streamer_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "streamer_id", "viewer_id", "status", "created_at"}).
			AddRow(int64(1), int64(5), int64(1), int64(2), int32(1), time.Now().Truncate(time.Second)).
			AddRow(int64(2), int64(5), int64(1), int64(3), int32(0), time.Now().Truncate(time.Second)))

	list, err := repo.ActiveByRoom(context.Background(), 5)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, LinkmicStatusConnected, list[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicRepository_QueuePosition(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM live_linkmic`).
		WithArgs(int64(5), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	pos, err := repo.QueuePosition(context.Background(), 5, 7)
	require.NoError(t, err)
	assert.Equal(t, 3, pos, "position = count + 1")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicRepository_ToggleAudio(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	mock.ExpectQuery(`UPDATE live_linkmic SET audio_enabled`).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"audio_enabled"}).AddRow(int32(0)))

	enabled, err := repo.ToggleAudio(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, int32(0), enabled)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkmicService_Delegation(t *testing.T) {
	repo, mock := newMockLinkmicRepo(t)
	svc := NewLinkmicService(repo, nil, nil)

	mock.ExpectExec(`UPDATE live_linkmic SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Accept(context.Background(), 9))

	mock.ExpectExec(`UPDATE live_linkmic SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Reject(context.Background(), 9))

	mock.ExpectExec(`UPDATE live_linkmic SET status`).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, svc.Disconnect(context.Background(), 9))

	require.NoError(t, mock.ExpectationsWereMet())
}
