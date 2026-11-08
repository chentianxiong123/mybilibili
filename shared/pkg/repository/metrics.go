package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// dailyMetricFields 是 manuscript_daily_metrics 上可安全写入的列 allowlist，
// 防动态列名拼接（FINDING-062：当前调用方传硬编码字面量，但未来调用方若转发
// 用户输入 → 完整 SQLi）。
var dailyMetricFields = map[string]bool{
	"comment_count": true, "like_count": true, "coin_count": true,
	"collect_count": true, "share_count": true, "view_count": true,
}

func UpsertDailyMetric(ctx context.Context, db *sql.DB, manuscriptID, userID int64, field string, delta int32) error {
	if !dailyMetricFields[field] {
		return fmt.Errorf("unsupported daily metric field: %q", field)
	}
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