INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '管理员管理', 'admin:manage', '/admin', 'GET', NULL, '管理员账号管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'admin:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '轮播图管理', 'banner:manage', '/banner-images', 'GET', NULL, '轮播图和背景图管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'banner:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '安全设置', 'security:manage', '/security-settings', 'GET', NULL, '安全设置和登录日志管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'security:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '直播管理', 'live:manage', '/live', 'GET', NULL, '直播间管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'live:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '会议管理', 'meeting:manage', '/meeting', 'GET', NULL, '会议室管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'meeting:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT 'AI管理', 'ai:manage', '/ai', 'GET', NULL, 'AI配置、技能和处理任务管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'ai:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '存储管理', 'storage:manage', '/storage', 'POST', NULL, '对象存储迁移管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'storage:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '消息管理', 'message:manage', '/message', 'POST', NULL, '系统消息广播管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'message:manage');

INSERT INTO permissions (`name`, `code`, `url`, `method`, `parent_id`, `description`, `create_time`, `update_time`)
SELECT '搜索索引管理', 'search:manage', '/search/admin/index', 'GET', NULL, '搜索索引管理权限', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE code = 'search:manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id
FROM permissions p
WHERE p.code IN (
    'admin:manage',
    'banner:manage',
    'security:manage',
    'live:manage',
    'meeting:manage',
    'ai:manage',
    'storage:manage',
    'message:manage',
    'search:manage'
)
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 1 AND rp.permission_id = p.id
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, p.id
FROM permissions p
WHERE p.code = 'statistics:manage'
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = 2 AND rp.permission_id = p.id
);
