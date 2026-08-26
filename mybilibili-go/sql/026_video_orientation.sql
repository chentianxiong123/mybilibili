-- 026_video_orientation.sql
-- 为视频增加横竖屏方向字段。
--   0 = 横屏 (landscape, width >= height)
--   1 = 竖屏 (portrait,  height > width)
-- 该字段由 work-service 转码时用 ffprobe 探测源文件宽高自动写入，
-- 播放器据此决定竖屏/横屏的展示比例。

ALTER TABLE videos ADD COLUMN IF NOT EXISTS is_vertical SMALLINT NOT NULL DEFAULT 0;

-- 存量数据回填：基于已转码 HLS 产物实测出的宽高（高>宽为竖屏）
UPDATE videos SET is_vertical = 1 WHERE id IN (28, 30, 34, 36);
