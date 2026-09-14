-- 数据修复：让 MySQL 数据与 Go 代码兼容

-- 1. 重置用户密码为 SHA256（Go 代码用 SHA256 验证密码）
-- 密码统一为：admin123（SHA256: 240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9）
UPDATE users SET password = '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9';
UPDATE admin_users SET password = '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9';

-- 2. 迁移 favorite_manuscripts 数据到 favorite_folder_videos
INSERT INTO favorite_folder_videos (folder_id, manuscript_id, created_at)
SELECT folder_id, manuscript_id, created_at
FROM favorite_manuscripts
ON CONFLICT DO NOTHING;

-- 3. 添加默认推荐配置
INSERT INTO recommend_configs (config_key, config_json, updated_by)
VALUES ('default', '{"strategy":"popular","limit":20}', 'system')
ON CONFLICT (config_key) DO NOTHING;