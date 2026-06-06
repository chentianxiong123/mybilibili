-- 会议模块数据库表
CREATE TABLE IF NOT EXISTS `meeting_room` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `room_name` VARCHAR(100) NOT NULL COMMENT '会议室名称',
  `room_code` VARCHAR(10) NOT NULL UNIQUE COMMENT '邀请码(6位数字)',
  `creator_id` BIGINT NOT NULL COMMENT '创建者用户ID',
  `creator_name` VARCHAR(50) NOT NULL COMMENT '创建者用户名',
  `max_participants` INT DEFAULT 5 COMMENT '最大参与人数',
  `status` TINYINT DEFAULT 0 COMMENT '状态: 0=未开始 1=进行中 2=已结束',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_room_code` (`room_code`),
  INDEX `idx_creator_id` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会议室表';

CREATE TABLE IF NOT EXISTS `meeting_participant` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `room_id` BIGINT NOT NULL COMMENT '会议室ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `user_name` VARCHAR(50) NOT NULL COMMENT '用户名',
  `user_avatar` VARCHAR(255) DEFAULT NULL COMMENT '用户头像',
  `role` TINYINT DEFAULT 0 COMMENT '角色: 0=参与者 1=主持人',
  `audio_enabled` TINYINT DEFAULT 0 COMMENT '音频是否开启: 0=关闭 1=开启',
  `video_enabled` TINYINT DEFAULT 0 COMMENT '视频是否开启: 0=关闭 1=开启',
  `screen_share_enabled` TINYINT DEFAULT 0 COMMENT '屏幕共享是否开启: 0=关闭 1=开启',
  `join_time` DATETIME DEFAULT NULL COMMENT '加入时间',
  `leave_time` DATETIME DEFAULT NULL COMMENT '离开时间',
  INDEX `idx_room_id` (`room_id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会议参与者表';