-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    url VARCHAR(255),
    method VARCHAR(20),
    parent_id INT,
    description VARCHAR(255),
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES permissions(id)
);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT,
    permission_id INT,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id),
    FOREIGN KEY (permission_id) REFERENCES permissions(id)
);

-- 创建管理员用户角色关联表
CREATE TABLE IF NOT EXISTS admin_user_roles (
    admin_user_id INT,
    role_id INT,
    PRIMARY KEY (admin_user_id, role_id),
    FOREIGN KEY (admin_user_id) REFERENCES admin_users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- 插入默认角色
INSERT INTO roles (name, description) VALUES 
('超级管理员', '拥有所有权限'),
('普通管理员', '拥有基本管理权限');

-- 插入默认权限
INSERT INTO permissions (name, code, url, method, parent_id, description) VALUES 
('用户管理', 'user:manage', '/users', 'GET', NULL, '用户管理权限'),
('视频管理', 'video:manage', '/videos', 'GET', NULL, '视频管理权限'),
('评论管理', 'comment:manage', '/comments', 'GET', NULL, '评论管理权限'),
('分类管理', 'category:manage', '/categories', 'GET', NULL, '分类管理权限'),
('标签管理', 'tag:manage', '/tags', 'GET', NULL, '标签管理权限'),
('内容审核', 'review:manage', '/review', 'GET', NULL, '内容审核权限'),
('统计分析', 'statistics:manage', '/statistics', 'GET', NULL, '统计分析权限'),
('角色管理', 'role:manage', '/roles', 'GET', NULL, '角色管理权限');

-- 为超级管理员分配所有权限
INSERT INTO role_permissions (role_id, permission_id) VALUES 
(1, 1),
(1, 2),
(1, 3),
(1, 4),
(1, 5),
(1, 6),
(1, 7),
(1, 8);

-- 为普通管理员分配基本权限
INSERT INTO role_permissions (role_id, permission_id) VALUES 
(2, 1),
(2, 2),
(2, 3),
(2, 6);