-- 028_banner_time_slot.sql
-- 为首页背景图(banner_images.type=3)扩展为 3 张时段图（morning/afternoon/evening）。
-- 不影响 home(type=1)轮播、category(type=2)、user-profile(type=4)。
-- 前端根据 time_slot 把 background 接口返回的列表按层渲染（CSS 用 --percentage 驱动）。

ALTER TABLE banner_images ADD COLUMN IF NOT EXISTS time_slot VARCHAR(16) NOT NULL DEFAULT 'default';

CREATE INDEX IF NOT EXISTS idx_banners_time_slot ON banner_images(time_slot);

-- 修复脏数据：id=5 / id=6 是历史手工 INSERT 的孤儿记录（type=0 status=0，
-- 本地文件已随 /tmp 清空、MinIO 里也没有对象）。
DELETE FROM banner_images WHERE id IN (5, 6);