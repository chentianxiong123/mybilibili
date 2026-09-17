package manuscript

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEventWriter(t *testing.T) (*sqlManuscriptEventWriter, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &sqlManuscriptEventWriter{db: db}, mock
}

func TestRecordStatusEvent(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO manuscript_status_events`).
		WithArgs(int64(1), int64(10), int32(0), int32(3), "APPROVE", "ADMIN", int64(99), "good").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := ew.RecordStatusEvent(context.Background(), 1, 10, 0, 3, "APPROVE", "ADMIN", 99, "good")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordVideoProcessEvent(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO video_process_events`).
		WithArgs(int64(5), int64(1), int32(0), int32(1), "TRANSCODING", 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := ew.RecordVideoProcessEvent(context.Background(), 5, 1, 0, 1, "TRANSCODING", 100)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordEditVersion(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO manuscript_edit_versions`).
		WithArgs(int64(1), int64(10), "before", "after", "title,status").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := ew.RecordEditVersion(context.Background(), 1, 10, "before", "after", "title,status")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
