CREATE TABLE IF NOT EXISTS `ai_sessions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `skill_id` bigint DEFAULT NULL COMMENT '命中的AI技能ID',
  `type` varchar(40) NOT NULL DEFAULT 'CUSTOMER_SERVICE' COMMENT '会话类型',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=AI服务中,1=等待人工,2=已处理',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_ai_sessions_user_type` (`user_id`, `type`),
  KEY `idx_ai_sessions_type_status` (`type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI客服会话表';

CREATE TABLE IF NOT EXISTS `ai_chat_messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `session_id` bigint DEFAULT NULL COMMENT '所属AI客服会话ID',
  `role` varchar(20) NOT NULL COMMENT 'USER/ASSISTANT/HUMAN/SYSTEM',
  `content` text NOT NULL COMMENT '消息内容',
  `token_count` int NOT NULL DEFAULT 0 COMMENT '消耗Token数',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_ai_chat_messages_session_id` (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI客服消息表';

SET @has_session_id := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'ai_chat_messages' AND column_name = 'session_id'
);
SET @sql := IF(@has_session_id = 0,
  'ALTER TABLE `ai_chat_messages` ADD COLUMN `session_id` bigint DEFAULT NULL COMMENT ''所属AI客服会话ID'' AFTER `id`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_conversation_id := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE() AND table_name = 'ai_chat_messages' AND column_name = 'conversation_id'
);
SET @sql := IF(@has_conversation_id > 0,
  'ALTER TABLE `ai_chat_messages` MODIFY COLUMN `conversation_id` bigint NULL COMMENT ''旧AI会话ID，已由session_id替代''',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_session_idx := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE() AND table_name = 'ai_chat_messages' AND index_name = 'idx_ai_chat_messages_session_id'
);
SET @sql := IF(@has_session_idx = 0,
  'ALTER TABLE `ai_chat_messages` ADD KEY `idx_ai_chat_messages_session_id` (`session_id`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `support_tickets` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `ticket_no` varchar(40) NOT NULL COMMENT '工单编号',
  `user_id` bigint DEFAULT NULL COMMENT '提交用户ID',
  `session_id` bigint DEFAULT NULL COMMENT '关联客服会话ID',
  `source` varchar(40) NOT NULL DEFAULT 'USER_FEEDBACK' COMMENT 'USER_FEEDBACK/AI_CUSTOMER_SERVICE/HUMAN_CUSTOMER_SERVICE/ADMIN',
  `category` varchar(40) NOT NULL DEFAULT 'GENERAL' COMMENT 'ACCOUNT/CONTENT_REVIEW/VIDEO_PLAYBACK/UPLOAD/AI_FEATURE/COMPLAINT/GENERAL',
  `priority` varchar(20) NOT NULL DEFAULT 'NORMAL' COMMENT 'LOW/NORMAL/HIGH/URGENT',
  `status` varchar(20) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING/PROCESSING/PROCESSED/CLOSED',
  `title` varchar(200) NOT NULL COMMENT '工单标题',
  `content` text NOT NULL COMMENT '用户诉求或投诉内容',
  `entry_reply` text DEFAULT NULL COMMENT '入口侧回复或AI回复摘要',
  `admin_reply` text DEFAULT NULL COMMENT '处理结果',
  `assignee_admin_id` bigint DEFAULT NULL COMMENT '处理人管理员ID',
  `processed_at` datetime DEFAULT NULL COMMENT '处理时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_support_tickets_ticket_no` (`ticket_no`),
  KEY `idx_support_tickets_status` (`status`),
  KEY `idx_support_tickets_user` (`user_id`),
  KEY `idx_support_tickets_session` (`session_id`),
  KEY `idx_support_tickets_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台反馈工单表';
