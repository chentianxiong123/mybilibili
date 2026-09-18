package manuscript

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordStatusEvent_Error(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO manuscript_status_events`).
		WillReturnError(sql.ErrConnDone)

	err := ew.RecordStatusEvent(context.Background(), 1, 10, 0, 3, "APPROVE", "ADMIN", 99, "good")
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordVideoProcessEvent_Error(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO video_process_events`).
		WillReturnError(sql.ErrConnDone)

	err := ew.RecordVideoProcessEvent(context.Background(), 5, 1, 0, 1, "TRANSCODING", 100)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordEditVersion_Error(t *testing.T) {
	ew, mock := newEventWriter(t)
	mock.ExpectExec(`INSERT INTO manuscript_edit_versions`).
		WillReturnError(sql.ErrConnDone)

	err := ew.RecordEditVersion(context.Background(), 1, 10, "before", "after", "title,status")
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
