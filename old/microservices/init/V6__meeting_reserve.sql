ALTER TABLE `meeting_room`
  ADD COLUMN `scheduled_start` DATETIME DEFAULT NULL COMMENT '预约开始时间' AFTER `end_time`,
  ADD COLUMN `scheduled_end` DATETIME DEFAULT NULL COMMENT '预约结束时间' AFTER `scheduled_start`,
  ADD COLUMN `scheduled_reason` VARCHAR(500) DEFAULT NULL COMMENT '预约事由' AFTER `scheduled_end`;
