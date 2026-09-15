package manuscript

import (
	"context"
	"database/sql"
)

// ManuscriptEventPublisher 稿件域跨服务事件发布接口。
// 旧版行为：稿件上架/下架发索引事件（ManuscriptIndex 主题）；播放/互动发分析事件（Analytics 主题）；
// 视频处理推进发处理事件（VideoProcess 主题）。由 core.EventPublisher 适配实现，main 注入。
type ManuscriptEventPublisher interface {
	PublishManuscriptIndex(ctx context.Context, manuscriptID int64, operation, trigger string) error
	PublishAnalytics(ctx context.Context, manuscriptID, userID int64, eventType, metricType string, delta int64) error
	PublishVideoProcess(ctx context.Context, manuscriptID, videoID int64, processType, sourceURL string, uploaderID int64) error
}

// ManuscriptEventWriter 稿件域事件/流水落库接口。
// 旧版行为：审核通过/拒绝、发布/下架/重试等状态流转必须写 manuscript_status_events；
// 视频处理阶段推进必须写 video_process_events；审核写 manuscript_edit_versions。
type ManuscriptEventWriter interface {
	RecordStatusEvent(ctx context.Context, manuscriptID, userID int64, fromStatus, toStatus int32, action, operatorType string, operatorID int64, reason string) error
	RecordVideoProcessEvent(ctx context.Context, videoID, manuscriptID int64, fromStatus, toStatus int32, stage string, progress int) error
	RecordEditVersion(ctx context.Context, manuscriptID, userID int64, before, after, changedFields string) error
}

// sqlManuscriptEventWriter 基于真 SQL 的实现。
type sqlManuscriptEventWriter struct {
	db *sql.DB
}

func NewManuscriptEventWriter(db *sql.DB) ManuscriptEventWriter {
	return &sqlManuscriptEventWriter{db: db}
}

func (w *sqlManuscriptEventWriter) RecordStatusEvent(ctx context.Context, manuscriptID, userID int64, fromStatus, toStatus int32, action, operatorType string, operatorID int64, reason string) error {
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO manuscript_status_events
		 (manuscript_id, user_id, from_status, to_status, action, operator_type, operator_id, reason)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		manuscriptID, userID, fromStatus, toStatus, action, operatorType, operatorID, reason)
	return err
}

func (w *sqlManuscriptEventWriter) RecordVideoProcessEvent(ctx context.Context, videoID, manuscriptID int64, fromStatus, toStatus int32, stage string, progress int) error {
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO video_process_events
		 (video_id, manuscript_id, from_status, to_status, stage, progress)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		videoID, manuscriptID, fromStatus, toStatus, stage, progress)
	return err
}

func (w *sqlManuscriptEventWriter) RecordEditVersion(ctx context.Context, manuscriptID, userID int64, before, after, changedFields string) error {
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO manuscript_edit_versions
		 (manuscript_id, user_id, before_snapshot, after_snapshot, changed_fields)
		 VALUES ($1,$2,$3,$4,$5)`,
		manuscriptID, userID, before, after, changedFields)
	return err
}
