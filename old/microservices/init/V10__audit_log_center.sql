CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `operator_id` int(11) DEFAULT NULL,
  `operator_name` varchar(64) DEFAULT NULL,
  `operator_role` varchar(64) DEFAULT NULL,
  `module` varchar(64) NOT NULL,
  `action` varchar(64) NOT NULL,
  `target_type` varchar(64) DEFAULT NULL,
  `target_id` varchar(128) DEFAULT NULL,
  `request_method` varchar(16) DEFAULT NULL,
  `request_uri` varchar(255) DEFAULT NULL,
  `client_ip` varchar(64) DEFAULT NULL,
  `user_agent` varchar(512) DEFAULT NULL,
  `result` tinyint(4) NOT NULL COMMENT '1成功 0失败',
  `message` varchar(255) DEFAULT NULL,
  `detail` text NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_audit_operator` (`operator_id`),
  KEY `idx_audit_module_action` (`module`, `action`),
  KEY `idx_audit_result` (`result`),
  KEY `idx_audit_target` (`target_type`, `target_id`),
  KEY `idx_audit_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '审计日志管理', 'audit:manage', '/audit-logs', 'GET', NULL, '后台操作审计日志查询权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'audit:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM permissions p
JOIN (
    SELECT 1 AS role_id
    UNION ALL
    SELECT 2 AS role_id
) rp
WHERE p.code = 'audit:manage'
AND NOT EXISTS (
    SELECT 1 FROM role_permissions existing
    WHERE existing.role_id = rp.role_id AND existing.permission_id = p.id
);
