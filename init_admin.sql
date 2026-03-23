-- 插入默认管理员账号，密码为 admin123（已加密）
INSERT INTO admin_users (username, password, admin_level) VALUES 
('admin', '$2a$10$E3QpBz9V9m8q7X0eQ4e0/e7e8e9e8e7e8e7e8e7e8e7e8e7e8e7e8e7e', 2);