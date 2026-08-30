ALTER TABLE `live_rooms`
  ADD COLUMN `category` VARCHAR(50) DEFAULT NULL COMMENT '直播分类' AFTER `viewer_count`,
  ADD COLUMN `scheduled_at` DATETIME DEFAULT NULL COMMENT '定时开播时间' AFTER `category`;

ALTER TABLE `meeting_room`
  MODIFY COLUMN `status` TINYINT(4) NULL DEFAULT 0 COMMENT '状态: 0=未开始/已通过 1=进行中 2=已结束 3=待审批 4=已拒绝';

UPDATE `meeting_room`
SET `status` = 3
WHERE `scheduled_start` IS NOT NULL
  AND `status` = 0;
