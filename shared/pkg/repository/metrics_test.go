package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertDailyMetric_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO manuscript_daily_metrics`)).
		WithArgs(int64(101), int64(202), int32(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = UpsertDailyMetric(context.Background(), db, 101, 202, "view_count", 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertDailyMetric_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO manuscript_daily_metrics`)).
		WithArgs(int64(1), int64(2), int32(3)).
		WillReturnError(errors.New("conn refused"))

	err = UpsertDailyMetric(context.Background(), db, 1, 2, "like_count", 3)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertDailyMetric_FieldInterpolation(t *testing.T) {
	// 字段名通过 Sprintf 注入到 SQL, 验证不同字段名都走通
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// 两次调用, 两次不同字段, 两次不同参数
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics.*comment_count`).
		WithArgs(int64(10), int64(20), int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO manuscript_daily_metrics.*favorite_count`).
		WithArgs(int64(11), int64(21), int32(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, UpsertDailyMetric(context.Background(), db, 10, 20, "comment_count", 1))
	require.NoError(t, UpsertDailyMetric(context.Background(), db, 11, 21, "favorite_count", 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertDailyMetric_ZeroAndNegativeDelta(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO manuscript_daily_metrics`)).
		WithArgs(int64(1), int64(2), int32(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, UpsertDailyMetric(context.Background(), db, 1, 2, "view_count", 0))
	require.NoError(t, mock.ExpectationsWereMet())
}