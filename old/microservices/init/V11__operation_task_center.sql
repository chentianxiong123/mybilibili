CREATE TABLE IF NOT EXISTS `operation_tasks` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_key` varchar(128) NOT NULL,
  `task_type` varchar(64) NOT NULL,
  `task_name` varchar(128) NOT NULL,
  `target_type` varchar(64) DEFAULT NULL,
  `target_id` varchar(128) DEFAULT NULL,
  `status` varchar(32) NOT NULL,
  `progress` int(11) NOT NULL DEFAULT 0,
  `stage` varchar(128) DEFAULT NULL,
  `message` varchar(255) DEFAULT NULL,
  `error_message` text NULL,
  `operator_id` int(11) DEFAULT NULL,
  `operator_name` varchar(64) DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_operation_task_key` (`task_key`),
  KEY `idx_operation_task_type` (`task_type`),
  KEY `idx_operation_task_status` (`status`),
  KEY `idx_operation_task_target` (`target_type`, `target_id`),
  KEY `idx_operation_task_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '任务中心管理', 'operation:manage', '/operation-tasks', 'GET', NULL, '后台统一任务中心查询权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'operation:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM permissions p
JOIN (
    SELECT 1 AS role_id
    UNION ALL
    SELECT 2 AS role_id
) rp
WHERE p.code = 'operation:manage'
AND NOT EXISTS (
    SELECT 1 FROM role_permissions existing
    WHERE existing.role_id = rp.role_id AND existing.permission_id = p.id
);
