package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func UpsertDailyMetric(ctx context.Context, db *sql.DB, manuscriptID, userID int64, field string, delta int32) error {
	query := fmt.Sprintf(
		`INSERT INTO manuscript_daily_metrics (metric_date, manuscript_id, user_id, %s)
		 VALUES (CURRENT_DATE, $1, $2, $3)
		 ON CONFLICT (metric_date, manuscript_id)
		 DO UPDATE SET %s = manuscript_daily_metrics.%s + $3, updated_at = NOW()`,
		field, field, field,
	)
	_, err := db.ExecContext(ctx, query, manuscriptID, userID, delta)
	return err
}