CREATE TABLE IF NOT EXISTS `manuscript_status_events` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `manuscript_id` int(11) NOT NULL,
  `user_id` int(11) DEFAULT NULL,
  `from_status` int(11) DEFAULT NULL,
  `to_status` int(11) NOT NULL,
  `action` varchar(64) NOT NULL,
  `operator_type` varchar(32) NOT NULL DEFAULT 'SYSTEM',
  `operator_id` int(11) DEFAULT NULL,
  `reason` varchar(500) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_mse_manuscript_created` (`manuscript_id`, `created_at`),
  KEY `idx_mse_user_created` (`user_id`, `created_at`),
  KEY `idx_mse_status_created` (`to_status`, `created_at`),
  KEY `idx_mse_action_created` (`action`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `manuscript_daily_metrics` (
  `metric_date` date NOT NULL,
  `manuscript_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `view_count` int(11) NOT NULL DEFAULT 0,
  `like_count` int(11) NOT NULL DEFAULT 0,
  `coin_count` int(11) NOT NULL DEFAULT 0,
  `collect_count` int(11) NOT NULL DEFAULT 0,
  `share_count` int(11) NOT NULL DEFAULT 0,
  `comment_count` int(11) NOT NULL DEFAULT 0,
  `danmaku_count` int(11) NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`metric_date`, `manuscript_id`),
  KEY `idx_mdm_user_date` (`user_id`, `metric_date`),
  KEY `idx_mdm_manuscript_date` (`manuscript_id`, `metric_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `video_process_events` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `video_id` int(11) NOT NULL,
  `manuscript_id` int(11) DEFAULT NULL,
  `from_status` int(11) DEFAULT NULL,
  `to_status` int(11) NOT NULL,
  `stage` varchar(64) DEFAULT NULL,
  `progress` int(11) DEFAULT NULL,
  `error_message` varchar(500) DEFAULT NULL,
  `operator_type` varchar(32) NOT NULL DEFAULT 'SYSTEM',
  `operator_id` int(11) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_vpe_video_created` (`video_id`, `created_at`),
  KEY `idx_vpe_manuscript_created` (`manuscript_id`, `created_at`),
  KEY `idx_vpe_status_created` (`to_status`, `created_at`),
  KEY `idx_vpe_stage_created` (`stage`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
