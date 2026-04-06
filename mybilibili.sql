/*
 Navicat Premium Data Transfer

 Source Server         : localhost_3306
 Source Server Type    : MySQL
 Source Server Version : 50717
 Source Host           : localhost:3306
 Source Schema         : mybilibili

 Target Server Type    : MySQL
 Target Server Version : 50717
 File Encoding         : 65001

 Date: 23/03/2026 12:56:57
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for admin_user_roles
-- ----------------------------
DROP TABLE IF EXISTS `admin_user_roles`;
CREATE TABLE `admin_user_roles`  (
  `admin_user_id` int(11) NOT NULL,
  `role_id` int(11) NOT NULL,
  PRIMARY KEY (`admin_user_id`, `role_id`) USING BTREE,
  INDEX `role_id`(`role_id`) USING BTREE,
  CONSTRAINT `admin_user_roles_ibfk_1` FOREIGN KEY (`admin_user_id`) REFERENCES `admin_users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `admin_user_roles_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of admin_user_roles
-- ----------------------------
INSERT INTO `admin_user_roles` VALUES (1, 1);
INSERT INTO `admin_user_roles` VALUES (2, 2);
INSERT INTO `admin_user_roles` VALUES (3, 2);

-- ----------------------------
-- Table structure for admin_users
-- ----------------------------
DROP TABLE IF EXISTS `admin_users`;
CREATE TABLE `admin_users`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '管理员用户名',
  `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码',
  `admin_level` int(11) NOT NULL DEFAULT 1 COMMENT '管理员级别：1-普通管理员，2-超级管理员',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '管理员表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of admin_users
-- ----------------------------
INSERT INTO `admin_users` VALUES (1, 'admin', '$2a$10$LaQOY69/vt9IyHo1kwiKPeetblSK3Ka9lm1oLg4NL7OoV3tI3ClNe', 2, '2026-03-05 22:50:02', '2026-03-14 21:37:40');
INSERT INTO `admin_users` VALUES (2, 'string', '$2a$10$LaQOY69/vt9IyHo1kwiKPeetblSK3Ka9lm1oLg4NL7OoV3tI3ClNe', 1, '2026-03-05 23:06:44', '2026-03-05 23:06:44');
INSERT INTO `admin_users` VALUES (3, 'string1', '$2a$10$LaQOY69/vt9IyHo1kwiKPeetblSK3Ka9lm1oLg4NL7OoV3tI3ClNe', 1, '2026-03-05 23:23:15', '2026-03-14 21:37:36');

-- ----------------------------
-- Table structure for categories
-- ----------------------------
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 24 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of categories
-- ----------------------------
INSERT INTO `categories` VALUES (1, '人工智能', '2026-03-02 18:28:07', '2026-03-19 23:06:22');
INSERT INTO `categories` VALUES (2, '电子', '2026-03-02 18:28:07', '2026-03-19 22:24:53');
INSERT INTO `categories` VALUES (3, '数学', '2026-03-02 18:28:07', '2026-03-19 22:24:58');
INSERT INTO `categories` VALUES (4, '英语', '2026-03-02 18:28:07', '2026-03-19 22:25:10');
INSERT INTO `categories` VALUES (5, '运动', '2026-03-02 18:28:07', '2026-03-20 18:30:30');
INSERT INTO `categories` VALUES (6, '心理学', '2026-03-02 18:28:07', '2026-03-19 22:50:25');
INSERT INTO `categories` VALUES (7, '软件', '2026-03-02 18:28:07', '2026-03-19 22:50:31');
INSERT INTO `categories` VALUES (8, '硬件', '2026-03-02 18:28:07', '2026-03-19 22:50:35');
INSERT INTO `categories` VALUES (9, '物理', '2026-03-02 18:28:07', '2026-03-19 22:50:45');
INSERT INTO `categories` VALUES (10, '机械', '2026-03-02 18:28:07', '2026-03-19 23:07:22');
INSERT INTO `categories` VALUES (11, '科技', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `categories` VALUES (12, '政治', '2026-03-02 18:28:07', '2026-03-20 18:30:22');
INSERT INTO `categories` VALUES (13, '历史', '2026-03-02 18:28:07', '2026-03-20 18:30:43');
INSERT INTO `categories` VALUES (14, '经济', '2026-03-02 18:28:07', '2026-03-20 18:30:14');
INSERT INTO `categories` VALUES (15, '文学', '2026-03-02 18:28:07', '2026-03-20 18:30:50');
INSERT INTO `categories` VALUES (16, '哲学', '2026-03-02 18:28:07', '2026-03-20 18:31:37');
INSERT INTO `categories` VALUES (17, '教育学', '2026-03-02 18:28:07', '2026-03-20 18:31:47');
INSERT INTO `categories` VALUES (18, '医学', '2026-03-02 18:28:07', '2026-03-20 18:32:00');
INSERT INTO `categories` VALUES (19, '管理学', '2026-03-02 18:28:07', '2026-03-20 18:32:07');
INSERT INTO `categories` VALUES (20, '艺术', '2026-03-02 18:28:07', '2026-03-20 18:32:16');
INSERT INTO `categories` VALUES (21, '地理', '2026-03-02 18:28:07', '2026-03-20 18:32:57');
INSERT INTO `categories` VALUES (22, '语言', '2026-03-02 18:28:07', '2026-03-20 18:33:01');
INSERT INTO `categories` VALUES (23, '测试', '2026-03-19 23:07:36', '2026-03-19 23:07:36');

-- ----------------------------
-- Table structure for comments
-- ----------------------------
DROP TABLE IF EXISTS `comments`;
CREATE TABLE `comments`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `manuscript_id` int(11) NULL DEFAULT NULL,
  `user_id` int(11) NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `like_count` int(11) NULL DEFAULT 0,
  `reply_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  `target_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'VIDEO' COMMENT '目标类型：VIDEO/DYNAMIC',
  `target_id` int(11) NULL DEFAULT 0 COMMENT '目标ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_comments_video_id`(`manuscript_id`) USING BTREE,
  CONSTRAINT `comments_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of comments
-- ----------------------------
INSERT INTO `comments` VALUES (3, 10, 4, '你好', 0, 3, '2026-03-15 14:23:23', '2026-03-22 20:30:08', 0, 'VIDEO', 10);
INSERT INTO `comments` VALUES (6, NULL, 4, '测试', 1, 3, '2026-03-16 21:56:24', '2026-03-20 18:23:08', 0, 'DYNAMIC', 1);

-- ----------------------------
-- Table structure for conversations
-- ----------------------------
DROP TABLE IF EXISTS `conversations`;
CREATE TABLE `conversations`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '当前用户ID',
  `target_user_id` int(11) NOT NULL COMMENT '对方用户ID',
  `last_message_id` int(11) NULL DEFAULT NULL COMMENT '最后消息ID',
  `last_message_content` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '最后消息内容',
  `last_message_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后消息时间',
  `unread_count` int(11) NULL DEFAULT 0 COMMENT '未读消息数',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_target`(`user_id`, `target_user_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_target_user_id`(`target_user_id`) USING BTREE,
  INDEX `idx_last_message_time`(`last_message_time`) USING BTREE,
  CONSTRAINT `conversations_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `conversations_ibfk_2` FOREIGN KEY (`target_user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 13 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '会话表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of conversations
-- ----------------------------
INSERT INTO `conversations` VALUES (7, 5, 4, 12, '你好', '2026-03-17 10:42:42', 0, '2026-03-17 10:39:14', '2026-03-20 19:26:24');
INSERT INTO `conversations` VALUES (8, 4, 5, 12, '你好', '2026-03-17 12:21:52', 0, '2026-03-17 10:39:14', '2026-03-17 12:21:52');
INSERT INTO `conversations` VALUES (9, 5, 6, 14, '你好朋友', '2026-03-17 12:26:20', 0, '2026-03-17 12:25:40', '2026-03-17 12:26:31');
INSERT INTO `conversations` VALUES (10, 6, 5, 14, '你好朋友', '2026-03-17 12:26:20', 0, '2026-03-17 12:25:40', '2026-03-17 12:26:19');
INSERT INTO `conversations` VALUES (11, 6, 4, 15, '你好', '2026-03-20 11:48:21', 0, '2026-03-20 11:48:21', '2026-03-20 11:48:21');
INSERT INTO `conversations` VALUES (12, 4, 6, 15, '你好', '2026-03-20 11:48:21', 1, '2026-03-20 11:48:21', '2026-03-20 11:48:21');

-- ----------------------------
-- Table structure for creator_settings
-- ----------------------------
DROP TABLE IF EXISTS `creator_settings`;
CREATE TABLE `creator_settings`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `default_category_id` int(11) NULL DEFAULT NULL COMMENT '默认投稿分类ID',
  `auto_publish` tinyint(4) NULL DEFAULT 0 COMMENT '自动发布：1-开启，0-关闭',
  `comment_notify` tinyint(4) NULL DEFAULT 1 COMMENT '评论通知：1-开启，0-关闭',
  `like_notify` tinyint(4) NULL DEFAULT 1 COMMENT '点赞通知：1-开启，0-关闭',
  `follow_notify` tinyint(4) NULL DEFAULT 1 COMMENT '关注通知：1-开启，0-关闭',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_default_category_id`(`default_category_id`) USING BTREE,
  CONSTRAINT `creator_settings_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `creator_settings_ibfk_2` FOREIGN KEY (`default_category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '创作者设置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of creator_settings
-- ----------------------------
INSERT INTO `creator_settings` VALUES (1, 5, NULL, 0, 1, 1, 1, '2026-03-17 14:10:56', '2026-03-17 14:10:56');
INSERT INTO `creator_settings` VALUES (2, 4, NULL, 0, 1, 1, 1, '2026-03-17 14:12:03', '2026-03-17 14:12:03');

-- ----------------------------
-- Table structure for dynamic_comments
-- ----------------------------
DROP TABLE IF EXISTS `dynamic_comments`;
CREATE TABLE `dynamic_comments`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `dynamic_id` int(11) NOT NULL COMMENT '动态ID',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '评论内容',
  `parent_id` int(11) NULL DEFAULT NULL COMMENT '父评论ID（回复时使用）',
  `reply_user_id` int(11) NULL DEFAULT NULL COMMENT '回复目标用户ID',
  `like_count` int(11) NULL DEFAULT 0 COMMENT '点赞数',
  `status` int(11) NULL DEFAULT 0 COMMENT '状态：0-正常，1-删除',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_dynamic_id`(`dynamic_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_parent_id`(`parent_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '动态评论表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of dynamic_comments
-- ----------------------------

-- ----------------------------
-- Table structure for favorite_folders
-- ----------------------------
DROP TABLE IF EXISTS `favorite_folders`;
CREATE TABLE `favorite_folders`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `video_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `user_id`(`user_id`, `name`) USING BTREE,
  CONSTRAINT `favorite_folders_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of favorite_folders
-- ----------------------------
INSERT INTO `favorite_folders` VALUES (1, 5, '数学', 3, '2026-03-12 17:45:36', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (2, 5, 'AI', 3, '2026-03-12 17:45:58', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (3, 4, '默认收藏夹', 0, '2026-03-13 17:34:21', '2026-03-16 21:43:42');
INSERT INTO `favorite_folders` VALUES (4, 4, 'AI', 1, '2026-03-15 14:49:02', '2026-03-20 11:55:27');
INSERT INTO `favorite_folders` VALUES (5, 5, '默认收藏夹', 3, '2026-03-11 20:47:41', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (6, 6, '默认收藏夹', 0, '2026-03-17 09:00:09', '2026-03-17 09:00:09');

-- ----------------------------
-- Table structure for favorite_manuscripts
-- ----------------------------
DROP TABLE IF EXISTS `favorite_manuscripts`;
CREATE TABLE `favorite_manuscripts`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `folder_id` int(11) NOT NULL,
  `manuscript_id` int(11) NOT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_folder_manuscript`(`folder_id`, `manuscript_id`) USING BTREE,
  INDEX `favorite_videos_ibfk_2`(`manuscript_id`) USING BTREE,
  CONSTRAINT `favorite_manuscripts_ibfk_1` FOREIGN KEY (`folder_id`) REFERENCES `favorite_folders` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `favorite_manuscripts_ibfk_2` FOREIGN KEY (`manuscript_id`) REFERENCES `manuscripts` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 28 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of favorite_manuscripts
-- ----------------------------
INSERT INTO `favorite_manuscripts` VALUES (18, 4, 10, '2026-03-16 17:04:45');
INSERT INTO `favorite_manuscripts` VALUES (19, 1, 10, '2026-03-20 12:00:48');
INSERT INTO `favorite_manuscripts` VALUES (20, 2, 10, '2026-03-20 12:00:48');
INSERT INTO `favorite_manuscripts` VALUES (21, 2, 12, '2026-03-20 12:00:57');
INSERT INTO `favorite_manuscripts` VALUES (22, 5, 12, '2026-03-20 12:00:57');
INSERT INTO `favorite_manuscripts` VALUES (23, 1, 13, '2026-03-20 12:01:05');
INSERT INTO `favorite_manuscripts` VALUES (24, 5, 14, '2026-03-20 14:03:20');
INSERT INTO `favorite_manuscripts` VALUES (25, 1, 11, '2026-03-20 20:30:38');
INSERT INTO `favorite_manuscripts` VALUES (26, 2, 11, '2026-03-20 20:30:38');
INSERT INTO `favorite_manuscripts` VALUES (27, 5, 11, '2026-03-20 20:30:38');

-- ----------------------------
-- Table structure for interaction_counts
-- ----------------------------
DROP TABLE IF EXISTS `interaction_counts`;
CREATE TABLE `interaction_counts`  (
  `target_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标类型',
  `target_id` int(11) NOT NULL COMMENT '目标ID',
  `like_count` int(11) NULL DEFAULT 0 COMMENT '点赞数',
  `collect_count` int(11) NULL DEFAULT 0 COMMENT '收藏数',
  `share_count` int(11) NULL DEFAULT 0 COMMENT '分享数',
  `coin_count` int(11) NULL DEFAULT 0 COMMENT '投币数',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`target_type`, `target_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '交互计数表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of interaction_counts
-- ----------------------------
INSERT INTO `interaction_counts` VALUES ('DYNAMIC', 1, 1, 0, 0, 0, '2026-03-16 14:28:43');
INSERT INTO `interaction_counts` VALUES ('VIDEO', 10, 1, 1, 0, 0, '2026-03-16 14:28:48');

-- ----------------------------
-- Table structure for manuscript_collection_relations
-- ----------------------------
DROP TABLE IF EXISTS `manuscript_collection_relations`;
CREATE TABLE `manuscript_collection_relations`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `manuscript_id` int(11) NOT NULL COMMENT '稿件ID',
  `collection_id` int(11) NOT NULL COMMENT '合集ID',
  `collection_order` int(11) NULL DEFAULT 0 COMMENT '在合集中的顺序',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_manuscript_collection`(`manuscript_id`, `collection_id`) USING BTREE,
  INDEX `idx_collection_id`(`collection_id`) USING BTREE,
  INDEX `idx_collection_order`(`collection_id`, `collection_order`) USING BTREE,
  CONSTRAINT `manuscript_collection_relations_ibfk_1` FOREIGN KEY (`manuscript_id`) REFERENCES `manuscripts` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `manuscript_collection_relations_ibfk_2` FOREIGN KEY (`collection_id`) REFERENCES `manuscript_collections` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '稿件与合集关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscript_collection_relations
-- ----------------------------
INSERT INTO `manuscript_collection_relations` VALUES (1, 12, 3, 0, '2026-03-16 23:20:36');
INSERT INTO `manuscript_collection_relations` VALUES (2, 13, 3, 0, '2026-03-16 23:23:05');
INSERT INTO `manuscript_collection_relations` VALUES (3, 14, 3, 0, '2026-03-16 23:28:29');

-- ----------------------------
-- Table structure for manuscript_collections
-- ----------------------------
DROP TABLE IF EXISTS `manuscript_collections`;
CREATE TABLE `manuscript_collections`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '合集标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '合集描述',
  `cover_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '封面图片URL',
  `user_id` int(11) NOT NULL COMMENT '创建用户ID',
  `manuscript_count` int(11) NULL DEFAULT 0 COMMENT '稿件数量',
  `view_count` int(11) NULL DEFAULT 0 COMMENT '浏览次数',
  `status` tinyint(4) NULL DEFAULT 1 COMMENT '状态：0-私密，1-公开',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  CONSTRAINT `manuscript_collections_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '稿件合集表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscript_collections
-- ----------------------------
INSERT INTO `manuscript_collections` VALUES (1, 'test1', '1', '/uploads/images/20260317094017_997d7071-28fe-47e0-8ac5-ac653e4f2e51.jpg', 5, 0, 0, 1, '2026-03-17 09:40:18', '2026-03-17 09:40:18');
INSERT INTO `manuscript_collections` VALUES (2, 'test1', '1', '/uploads/images/20260317094237_a9a41bbc-0933-415e-bb0a-34a92a2e506f.jpg', 5, 0, 0, 1, '2026-03-17 09:42:38', '2026-03-17 09:42:38');
INSERT INTO `manuscript_collections` VALUES (3, 'AI合集', 'AI合集', NULL, 5, 3, 0, 1, '2026-03-17 09:54:07', '2026-03-20 17:02:56');

-- ----------------------------
-- Table structure for manuscripts
-- ----------------------------
DROP TABLE IF EXISTS `manuscripts`;
CREATE TABLE `manuscripts`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '稿件标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '稿件描述',
  `cover_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '稿件封面',
  `user_id` int(11) NOT NULL COMMENT '上传用户ID',
  `category_id` int(11) NOT NULL COMMENT '分类ID',
  `view_count` int(11) NULL DEFAULT 0 COMMENT '总播放量',
  `like_count` int(11) NULL DEFAULT 0 COMMENT '总点赞数',
  `coin_count` int(11) NULL DEFAULT 0 COMMENT '总投币数',
  `collect_count` int(11) NULL DEFAULT 0 COMMENT '总收藏数',
  `share_count` int(11) NULL DEFAULT 0 COMMENT '总分享数',
  `comment_count` int(11) NULL DEFAULT 0 COMMENT '总评论数',
  `danmaku_count` int(11) NULL DEFAULT 0 COMMENT '总弹幕数',
  `status` int(11) NULL DEFAULT 0 COMMENT '0-待审核 1-处理中 2-待上架 3-已上架 4-拒绝 -1-下架',
  `review_status` int(11) NULL DEFAULT 0 COMMENT '0-待审核 1-通过 2-拒绝',
  `review_reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '审核原因',
  `review_time` datetime NULL DEFAULT NULL COMMENT '审核时间',
  `reviewer_id` int(11) NULL DEFAULT NULL COMMENT '审核人ID',
  `process_progress` int(11) NULL DEFAULT 0 COMMENT '处理进度百分比',
  `process_stage` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '处理阶段',
  `collection_id` int(11) NULL DEFAULT NULL COMMENT '所属合集ID（可选）',
  `collection_order` int(11) NULL DEFAULT 0 COMMENT '在合集中的排序',
  `upload_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `duration_seconds` int(11) NULL DEFAULT 0 COMMENT 'Total duration in seconds (sum of all videos)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_category_id`(`category_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_collection_id`(`collection_id`) USING BTREE,
  INDEX `idx_upload_time`(`upload_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 22 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '稿件表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscripts
-- ----------------------------
INSERT INTO `manuscripts` VALUES (10, '无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】', '本视频为无限token版本的openclaw安装教程\r\n任意Windows电脑均可安装 无需付费 无需高配电脑\r\n如果觉得视频对你有帮助 请一键三连加关注 随时关注主播最新教程\r\n如有问题可评论区留言或私信 随缘解答 无套路。小龙虾', '/uploads/manuscripts/10/cover.jpg', 4, 1, 211, 3, 0, 3, 0, 0, 0, 3, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-03-15 11:01:33', '2026-03-22 20:56:25', 384);
INSERT INTO `manuscripts` VALUES (11, 'AI Agent正在重走操作系统的老路', '一个MCP工具号称能省98%的Context Token，社区瞬间炸锅。但深挖后我发现，这不是一个“省钱”的问题——AI Agent开发正在重演操作系统的进化史。Context就是新时代的内存，而我们还停留在手动管理阶段。一次工具调用吃掉50K token，3-4次就烧掉半个窗口。本期拆解“沙盒+索引+按需检索”架构，以及协议层的结构性缺陷。', '/uploads/manuscripts/11/cover.jpg', 4, 1, 82, 2, 0, 1, 0, 0, 0, 3, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-03-16 22:51:00', '2026-03-22 20:05:13', 657);
INSERT INTO `manuscripts` VALUES (12, '把 OpenClaw 翻译成“小龙虾”的，出来挨打！', '', '/uploads/manuscripts/12/cover.jpg', 5, 4, 24, 1, 0, 1, 0, 0, 0, 3, 0, NULL, NULL, NULL, NULL, NULL, 3, 0, '2026-03-16 23:20:36', '2026-03-22 20:29:40', 197);
INSERT INTO `manuscripts` VALUES (13, '数学怎么提分', '数学怎么提分#家长收藏孩子受益 #家长必读 #学习方法 #学霸秘籍', '/uploads/manuscripts/13/cover.jpg', 5, 3, 24, 1, 0, 1, 0, 0, 0, 3, 0, NULL, NULL, NULL, NULL, NULL, 3, 0, '2026-03-16 23:23:05', '2026-03-20 20:09:58', 72);
INSERT INTO `manuscripts` VALUES (14, '每天学习一个电子知识，今天学习如何测电压', '', '/uploads/manuscripts/14/cover.jpg', 5, 2, 133, 2, 0, 1, 0, 0, 0, 3, 0, NULL, NULL, NULL, NULL, NULL, 3, 0, '2026-03-16 23:28:29', '2026-03-22 20:29:43', 27);
INSERT INTO `manuscripts` VALUES (21, '投篮发力差的共性和修改思路', '投篮发力差的共性和修改思路', '/uploads/manuscripts/21/cover.jpg', 5, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, NULL, NULL, NULL, NULL, NULL, NULL, 0, '2026-03-20 20:12:18', '2026-03-20 20:12:18', 0);

-- ----------------------------
-- Table structure for message_settings
-- ----------------------------
DROP TABLE IF EXISTS `message_settings`;
CREATE TABLE `message_settings`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `private_message_notification` tinyint(4) NULL DEFAULT 1 COMMENT '私信通知：1-开启，0-关闭',
  `reply_notification` tinyint(4) NULL DEFAULT 1 COMMENT '回复通知：1-开启，0-关闭',
  `at_notification` tinyint(4) NULL DEFAULT 1 COMMENT '@通知：1-开启，0-关闭',
  `like_notification` tinyint(4) NULL DEFAULT 1 COMMENT '点赞通知：1-开启，0-关闭',
  `system_notification` tinyint(4) NULL DEFAULT 1 COMMENT '系统通知：1-开启，0-关闭',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  CONSTRAINT `message_settings_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '消息设置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of message_settings
-- ----------------------------

-- ----------------------------
-- Table structure for messages
-- ----------------------------
DROP TABLE IF EXISTS `messages`;
CREATE TABLE `messages`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `sender_id` int(11) NOT NULL COMMENT '发送者ID',
  `receiver_id` int(11) NOT NULL COMMENT '接收者ID',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '消息内容',
  `is_read` tinyint(4) NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `message_type` tinyint(4) NULL DEFAULT 1 COMMENT '消息类型：1-文本，2-图片，3-表情',
  `target_id` int(11) NULL DEFAULT NULL,
  `media_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '图片/媒体URL',
  `conversation_id` int(11) NULL DEFAULT NULL COMMENT '所属会话ID',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `sender_id`(`sender_id`) USING BTREE,
  INDEX `receiver_id`(`receiver_id`) USING BTREE,
  CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `messages_ibfk_2` FOREIGN KEY (`receiver_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 45 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '消息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of messages
-- ----------------------------
INSERT INTO `messages` VALUES (11, 5, 4, '你好', 1, '2026-03-17 10:39:14', '2026-03-17 12:18:29', 1, NULL, NULL, 7);
INSERT INTO `messages` VALUES (12, 5, 4, '你好', 1, '2026-03-17 10:42:41', '2026-03-17 12:18:29', 1, NULL, NULL, 7);
INSERT INTO `messages` VALUES (13, 5, 6, '你好兄弟', 1, '2026-03-17 12:25:40', '2026-03-17 12:26:04', 1, NULL, NULL, 9);
INSERT INTO `messages` VALUES (14, 6, 5, '你好朋友', 1, '2026-03-17 12:26:19', '2026-03-17 12:26:31', 1, NULL, NULL, 10);
INSERT INTO `messages` VALUES (15, 6, 4, '你好', 0, '2026-03-20 11:48:21', '2026-03-20 11:48:21', 1, NULL, NULL, 11);
INSERT INTO `messages` VALUES (42, 5, 4, '赞了你的视频《无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】》', 0, '2026-03-20 11:53:38', '2026-03-20 18:25:41', 4, 25, NULL, NULL);
INSERT INTO `messages` VALUES (43, 5, 4, '赞了你的视频《AI Agent正在重走操作系统的老路》', 0, '2026-03-16 23:19:34', '2026-03-20 18:25:45', 4, 26, NULL, NULL);
INSERT INTO `messages` VALUES (44, 4, 5, '赞了你的视频《每天学习一个电子知识，今天学习如何测电压》', 1, '2026-03-20 11:15:41', '2026-03-20 19:26:24', 4, 31, NULL, NULL);

-- ----------------------------
-- Table structure for permissions
-- ----------------------------
DROP TABLE IF EXISTS `permissions`;
CREATE TABLE `permissions`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `method` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `parent_id` int(11) NULL DEFAULT NULL,
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`name`) USING BTREE,
  UNIQUE INDEX `code`(`code`) USING BTREE,
  INDEX `parent_id`(`parent_id`) USING BTREE,
  CONSTRAINT `permissions_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `permissions` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of permissions
-- ----------------------------
INSERT INTO `permissions` VALUES (1, '用户管理', 'user:manage', '/users', 'GET', NULL, '用户管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (2, '视频管理', 'video:manage', '/videos', 'GET', NULL, '视频管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (3, '评论管理', 'comment:manage', '/comments', 'GET', NULL, '评论管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (4, '分类管理', 'category:manage', '/categories', 'GET', NULL, '分类管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (5, '标签管理', 'tag:manage', '/tags', 'GET', NULL, '标签管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (6, '内容审核', 'review:manage', '/review', 'GET', NULL, '内容审核权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (7, '统计分析', 'statistics:manage', '/statistics', 'GET', NULL, '统计分析权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');
INSERT INTO `permissions` VALUES (8, '角色管理', 'role:manage', '/roles', 'GET', NULL, '角色管理权限', '2026-03-06 20:47:32', '2026-03-06 20:47:32');

-- ----------------------------
-- Table structure for prohibited_word
-- ----------------------------
DROP TABLE IF EXISTS `prohibited_word`;
CREATE TABLE `prohibited_word`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `word` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '违禁词',
  `match_type` enum('EXACT','CONTAINS') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'CONTAINS' COMMENT '匹配类型：EXACT-精确匹配 CONTAINS-包含匹配',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '分类：POLITICS-政治 PORN-色情 AD-广告等',
  `is_enabled` tinyint(4) NULL DEFAULT 1 COMMENT '是否启用：0-禁用 1-启用',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_word`(`word`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '违禁词词典表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of prohibited_word
-- ----------------------------

-- ----------------------------
-- Table structure for replies
-- ----------------------------
DROP TABLE IF EXISTS `replies`;
CREATE TABLE `replies`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `comment_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `like_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `reply_to_user_id` int(11) NULL DEFAULT NULL,
  `status` enum('NORMAL','REMOVED') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'NORMAL' COMMENT '状态：NORMAL-正常 REMOVED-已下架待审核',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_replies_comment_id`(`comment_id`) USING BTREE,
  INDEX `reply_to_user_id`(`reply_to_user_id`) USING BTREE,
  CONSTRAINT `replies_ibfk_1` FOREIGN KEY (`comment_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `replies_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `replies_ibfk_3` FOREIGN KEY (`reply_to_user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of replies
-- ----------------------------
INSERT INTO `replies` VALUES (1, 3, 4, '你好', 0, '2026-03-15 14:23:33', '2026-03-15 14:23:33', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (2, 3, 4, '@string：你好', 0, '2026-03-15 14:23:40', '2026-03-15 14:23:40', 4, 'NORMAL');
INSERT INTO `replies` VALUES (3, 6, 4, '测试', 0, '2026-03-16 21:56:39', '2026-03-16 21:56:39', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (4, 6, 4, '@string：@string 测试', 0, '2026-03-16 22:13:13', '2026-03-16 22:13:13', 4, 'NORMAL');
INSERT INTO `replies` VALUES (5, 6, 4, '@string：测试', 0, '2026-03-16 22:19:09', '2026-03-16 22:19:09', 4, 'NORMAL');
INSERT INTO `replies` VALUES (6, 3, 5, '你好', 0, '2026-03-22 20:30:08', '2026-03-22 20:30:08', NULL, 'NORMAL');

-- ----------------------------
-- Table structure for role_permissions
-- ----------------------------
DROP TABLE IF EXISTS `role_permissions`;
CREATE TABLE `role_permissions`  (
  `role_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`role_id`, `permission_id`) USING BTREE,
  INDEX `permission_id`(`permission_id`) USING BTREE,
  CONSTRAINT `role_permissions_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `role_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of role_permissions
-- ----------------------------
INSERT INTO `role_permissions` VALUES (1, 1);
INSERT INTO `role_permissions` VALUES (2, 1);
INSERT INTO `role_permissions` VALUES (1, 2);
INSERT INTO `role_permissions` VALUES (2, 2);
INSERT INTO `role_permissions` VALUES (1, 3);
INSERT INTO `role_permissions` VALUES (2, 3);
INSERT INTO `role_permissions` VALUES (1, 4);
INSERT INTO `role_permissions` VALUES (1, 5);
INSERT INTO `role_permissions` VALUES (1, 6);
INSERT INTO `role_permissions` VALUES (2, 6);
INSERT INTO `role_permissions` VALUES (1, 7);
INSERT INTO `role_permissions` VALUES (1, 8);

-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of roles
-- ----------------------------
INSERT INTO `roles` VALUES (1, '超级管理员', '拥有所有权限', '2026-03-06 20:46:56', '2026-03-06 20:46:56');
INSERT INTO `roles` VALUES (2, '普通管理员', '拥有基本管理权限', '2026-03-06 20:46:56', '2026-03-06 20:46:56');

-- ----------------------------
-- Table structure for tags
-- ----------------------------
DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 30 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tags
-- ----------------------------
INSERT INTO `tags` VALUES (1, '原神', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (2, '我的世界', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (3, '英雄联盟', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (4, '鬼畜视频', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (5, '美食制作', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (6, '动漫推荐', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (7, '游戏攻略', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (8, '音乐MV', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (9, '舞蹈教学', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (10, '科技测评', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (11, '生活日常', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (12, '旅行', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (13, '健身', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (14, '学习', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (15, '职场', '2026-03-02 18:28:07', '2026-03-02 18:28:07');
INSERT INTO `tags` VALUES (16, 'string', '2026-03-04 00:02:23', '2026-03-04 00:02:23');
INSERT INTO `tags` VALUES (17, '11', '2026-03-09 14:43:26', '2026-03-09 14:43:26');
INSERT INTO `tags` VALUES (18, '123', '2026-03-09 15:52:29', '2026-03-09 15:52:29');
INSERT INTO `tags` VALUES (19, 'uioi', '2026-03-09 15:53:37', '2026-03-09 15:53:37');
INSERT INTO `tags` VALUES (20, '111', '2026-03-09 17:30:02', '2026-03-09 17:30:02');
INSERT INTO `tags` VALUES (21, '9999', '2026-03-09 17:31:01', '2026-03-09 17:31:01');
INSERT INTO `tags` VALUES (22, '999', '2026-03-09 17:31:01', '2026-03-09 17:31:01');
INSERT INTO `tags` VALUES (23, '5555', '2026-03-11 12:46:05', '2026-03-11 12:46:05');
INSERT INTO `tags` VALUES (24, '555', '2026-03-11 15:35:16', '2026-03-11 15:35:16');
INSERT INTO `tags` VALUES (25, '1', '2026-03-14 22:36:01', '2026-03-14 22:36:01');
INSERT INTO `tags` VALUES (26, '篮球', '2026-03-20 20:12:18', '2026-03-20 20:12:18');
INSERT INTO `tags` VALUES (27, '投篮', '2026-03-20 20:12:18', '2026-03-20 20:12:18');
INSERT INTO `tags` VALUES (28, '运动', '2026-03-20 20:12:18', '2026-03-20 20:12:18');
INSERT INTO `tags` VALUES (29, '体育', '2026-03-20 20:12:18', '2026-03-20 20:12:18');

-- ----------------------------
-- Table structure for user_dynamics
-- ----------------------------
DROP TABLE IF EXISTS `user_dynamics`;
CREATE TABLE `user_dynamics`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `dynamic_type` tinyint(4) NULL DEFAULT 0 COMMENT '动态类型：0-纯文字，1-图片，2-引用视频',
  `image_url` varchar(2000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '图片URL，多个用逗号分隔',
  `ref_video_id` int(11) NULL DEFAULT NULL COMMENT '引用的视频ID',
  `ref_manuscript_id` int(11) NULL DEFAULT NULL COMMENT '引用的稿件ID',
  `like_count` int(11) NULL DEFAULT 0,
  `comment_count` int(11) NULL DEFAULT 0,
  `share_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE,
  CONSTRAINT `user_dynamics_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_dynamics
-- ----------------------------
INSERT INTO `user_dynamics` VALUES (1, 4, '6666', 0, NULL, NULL, NULL, 1, 0, 0, '2026-03-07 12:15:24', 0);
INSERT INTO `user_dynamics` VALUES (2, 6, '大家好', 0, NULL, NULL, NULL, 0, 0, 0, '2026-03-20 11:52:09', 0);
INSERT INTO `user_dynamics` VALUES (3, 5, '😀😀😀AI agent', 2, '/uploads/images/20260320152632_cc7c57ed-539c-4e58-a15c-211b194cacc8.jpg', NULL, 12, 0, 0, 0, '2026-03-20 15:26:33', 0);

-- ----------------------------
-- Table structure for user_interactions
-- ----------------------------
DROP TABLE IF EXISTS `user_interactions`;
CREATE TABLE `user_interactions`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `target_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标类型：VIDEO/DYNAMIC/COMMENT/REPLY/USER',
  `target_id` int(11) NOT NULL COMMENT '目标ID',
  `interaction_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交互类型：LIKE/COLLECT/FOLLOW/COIN',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_interaction`(`user_id`, `target_type`, `target_id`, `interaction_type`) USING BTREE,
  INDEX `idx_target`(`target_type`, `target_id`, `interaction_type`) USING BTREE,
  INDEX `idx_user`(`user_id`, `interaction_type`, `created_at`) USING BTREE,
  INDEX `idx_user_target`(`user_id`, `target_type`, `target_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 49 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户交互记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_interactions
-- ----------------------------
INSERT INTO `user_interactions` VALUES (1, 4, 'VIDEO', 10, 'LIKE', '2026-03-15 14:33:36');
INSERT INTO `user_interactions` VALUES (3, 4, 'VIDEO', 10, 'COLLECT', '2026-03-15 15:20:25');
INSERT INTO `user_interactions` VALUES (5, 4, 'USER', 5, 'FOLLOW', '2026-03-13 17:51:14');
INSERT INTO `user_interactions` VALUES (13, 4, 'VIDEO', 25, 'COLLECT', '2026-03-16 17:04:45');
INSERT INTO `user_interactions` VALUES (15, 4, 'VIDEO', 25, 'LIKE', '2026-03-16 17:15:11');
INSERT INTO `user_interactions` VALUES (29, 4, 'DYNAMIC', 1, 'LIKE', '2026-03-16 21:27:11');
INSERT INTO `user_interactions` VALUES (30, 4, 'VIDEO', 26, 'LIKE', '2026-03-16 23:18:38');
INSERT INTO `user_interactions` VALUES (31, 5, 'VIDEO', 26, 'LIKE', '2026-03-16 23:19:34');
INSERT INTO `user_interactions` VALUES (32, 5, 'VIDEO', 29, 'LIKE', '2026-03-17 11:57:40');
INSERT INTO `user_interactions` VALUES (35, 4, 'COMMENT', 6, 'LIKE', '2026-03-19 20:32:07');
INSERT INTO `user_interactions` VALUES (36, 5, 'VIDEO', 31, 'LIKE', '2026-03-19 21:48:30');
INSERT INTO `user_interactions` VALUES (37, 4, 'VIDEO', 31, 'LIKE', '2026-03-20 11:15:41');
INSERT INTO `user_interactions` VALUES (39, 6, 'USER', 5, 'FOLLOW', '2026-03-20 11:47:54');
INSERT INTO `user_interactions` VALUES (40, 6, 'USER', 4, 'FOLLOW', '2026-03-20 11:47:59');
INSERT INTO `user_interactions` VALUES (41, 5, 'VIDEO', 25, 'LIKE', '2026-03-20 11:53:38');
INSERT INTO `user_interactions` VALUES (42, 5, 'VIDEO', 25, 'COLLECT', '2026-03-20 12:00:48');
INSERT INTO `user_interactions` VALUES (43, 5, 'VIDEO', 29, 'COLLECT', '2026-03-20 12:00:57');
INSERT INTO `user_interactions` VALUES (44, 5, 'VIDEO', 30, 'COLLECT', '2026-03-20 12:01:05');
INSERT INTO `user_interactions` VALUES (45, 5, 'VIDEO', 30, 'LIKE', '2026-03-20 12:01:06');
INSERT INTO `user_interactions` VALUES (46, 5, 'VIDEO', 31, 'COLLECT', '2026-03-20 14:03:20');
INSERT INTO `user_interactions` VALUES (47, 5, 'USER', 4, 'FOLLOW', '2026-03-20 14:08:14');
INSERT INTO `user_interactions` VALUES (48, 5, 'VIDEO', 26, 'COLLECT', '2026-03-20 20:30:38');

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `nickname` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `gender` tinyint(4) NULL DEFAULT 0 COMMENT '0:未知,1:男,2:女',
  `birthdate` date NULL DEFAULT NULL,
  `signature` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `level` int(11) NULL DEFAULT 1,
  `following_count` int(11) NULL DEFAULT 0,
  `follower_count` int(11) NULL DEFAULT 0,
  `video_count` int(11) NULL DEFAULT 0,
  `liked_count` int(11) NULL DEFAULT 0,
  `coin_count` int(11) NULL DEFAULT 0,
  `point_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  `experience` int(11) NULL DEFAULT 0,
  `bio` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `announcement` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE,
  UNIQUE INDEX `email`(`email`) USING BTREE,
  UNIQUE INDEX `phone`(`phone`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 7 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO `users` VALUES (4, 'string', '$2a$10$4CcJDm2RUaRhXMppYr.Rj.Q7yMK.NQFSRi7QdeH8FUa/ELfLGpwXG', 'string', '/uploads/avatars/4/avatar.jpg', '123@qq.com', '13011011111', 1, '2026-03-17', 'string12', 1, 0, 0, 0, 6, 0, 0, '2026-03-02 22:31:41', '2026-03-20 14:08:14', 1, 0, '', '999');
INSERT INTO `users` VALUES (5, 'test', '$2a$10$EiwLhC5FdeaALJmFEdVdg.fQvbuKaTNM2HWFweipvK6ieoQbSMzV6', 'test666', '/uploads/avatars/5/avatar.jpg', '321@qq.com', '17733311111', 1, '2026-03-09', '6666666', 1, 0, 0, 0, 11, 0, 0, '2026-03-04 20:48:47', '2026-03-20 15:04:20', 1, 0, '666', '666');
INSERT INTO `users` VALUES (6, 'admin', '$2a$10$FpNqkWXQJoaOwAYCeWzK6u5EMMgSypBWwuvU5H7lUbolDB3AXtf.W', 'admin', 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png', '12111@qq.com', '11111111111', NULL, '2026-03-19', NULL, 1, 0, 0, 0, 0, 0, 0, '2026-03-17 09:00:05', '2026-03-20 11:47:59', 0, 0, '', NULL);

-- ----------------------------
-- Table structure for video_tags
-- ----------------------------
DROP TABLE IF EXISTS `video_tags`;
CREATE TABLE `video_tags`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `video_id` int(11) NOT NULL,
  `tag_id` int(11) NOT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_video_tag`(`video_id`, `tag_id`) USING BTREE,
  INDEX `tag_id`(`tag_id`) USING BTREE,
  CONSTRAINT `video_tags_ibfk_1` FOREIGN KEY (`video_id`) REFERENCES `videos` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `video_tags_ibfk_2` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of video_tags
-- ----------------------------
INSERT INTO `video_tags` VALUES (1, 29, 20, '2026-03-16 23:20:36');
INSERT INTO `video_tags` VALUES (2, 30, 20, '2026-03-16 23:23:05');
INSERT INTO `video_tags` VALUES (3, 31, 20, '2026-03-16 23:28:28');
INSERT INTO `video_tags` VALUES (4, 32, 26, '2026-03-20 20:12:18');
INSERT INTO `video_tags` VALUES (5, 32, 27, '2026-03-20 20:12:18');
INSERT INTO `video_tags` VALUES (6, 32, 28, '2026-03-20 20:12:18');
INSERT INTO `video_tags` VALUES (7, 32, 29, '2026-03-20 20:12:18');

-- ----------------------------
-- Table structure for videos
-- ----------------------------
DROP TABLE IF EXISTS `videos`;
CREATE TABLE `videos`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `manuscript_id` int(11) NULL DEFAULT NULL COMMENT '所属稿件ID',
  `video_order` int(11) NULL DEFAULT 0 COMMENT '在稿件中的排序（分P顺序）',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `play_url_hd` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '高清视频URL(1080p)',
  `play_url_sd` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '标清视频URL(720p)',
  `play_url_ld` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '流畅视频URL(480p)',
  `user_id` int(11) NULL DEFAULT NULL COMMENT '用户ID（通过稿件关联，已废弃）',
  `category_id` int(11) NULL DEFAULT NULL COMMENT '分类ID（通过稿件关联，已废弃）',
  `upload_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  `process_progress` int(11) NULL DEFAULT 0 COMMENT '处理进度：0-100',
  `process_stage` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '当前处理阶段',
  `has_subtitle` tinyint(4) NULL DEFAULT 0 COMMENT '是否有字幕',
  `has_summary` tinyint(4) NULL DEFAULT 0 COMMENT '是否有摘要',
  `process_status` int(11) NULL DEFAULT 0 COMMENT 'Processing status: 0-pending 1-transcoding 2-audio extracting 3-subtitle generating 4-AI summarizing 5-completed 6-transcode failed 7-audio failed 8-subtitle failed 9-AI failed',
  `process_error` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'Processing failure reason',
  `source_video_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'Source video URL (for admin preview)',
  `duration_seconds` int(11) NULL DEFAULT 0 COMMENT 'Video duration in seconds',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '视频描述',
  `cover_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '封面URL',
  `play_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '播放URL',
  `review_status` int(11) NULL DEFAULT 0 COMMENT '审核状态',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_videos_user_id`(`user_id`) USING BTREE,
  INDEX `idx_videos_category_id`(`category_id`) USING BTREE,
  INDEX `idx_videos_upload_time`(`upload_time`) USING BTREE,
  INDEX `idx_manuscript_id`(`manuscript_id`) USING BTREE,
  CONSTRAINT `videos_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `videos_ibfk_2` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 33 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of videos
-- ----------------------------
INSERT INTO `videos` VALUES (25, 10, 0, '无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】', '/uploads/manuscripts/10/videos/25/transcoded/1080p.mp4', '/uploads/manuscripts/10/videos/25/transcoded/720p.mp4', '/uploads/manuscripts/10/videos/25/transcoded/480p.mp4', 4, 1, '2026-03-15 11:01:32', '2026-03-20 19:29:48', 0, 0, NULL, 1, 1, 5, '', '/uploads/manuscripts/10/videos/25/source/video.mp4', 384, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (26, 11, 0, 'AI Agent正在重走操作系统的老路', '/uploads/manuscripts/11/videos/26/transcoded/1080p.mp4', '/uploads/manuscripts/11/videos/26/transcoded/720p.mp4', '/uploads/manuscripts/11/videos/26/transcoded/480p.mp4', 4, 1, '2026-03-16 22:50:59', '2026-03-20 19:41:32', 0, 0, NULL, 0, 0, 21, NULL, '/uploads/manuscripts/10/videos/26/source/video.mp4', 271, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (27, 11, 1, '谁在给AI投毒？GEO灰产是如何割韭菜的？', '/uploads/manuscripts/11/videos/27/transcoded/1080p.mp4', '/uploads/manuscripts/11/videos/27/transcoded/720p.mp4', '/uploads/manuscripts/11/videos/27/transcoded/480p.mp4', 4, 1, '2026-03-16 22:50:59', '2026-03-20 19:41:15', 0, 0, NULL, 0, 0, 21, NULL, '/uploads/manuscripts/10/videos/27/source/video.mp4', 230, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (28, 11, 2, '周鸿祎建议大家别只把AI当搜索引擎用', '/uploads/manuscripts/11/videos/28/transcoded/1080p.mp4', '/uploads/manuscripts/11/videos/28/transcoded/720p.mp4', '/uploads/manuscripts/11/videos/28/transcoded/480p.mp4', 4, 1, '2026-03-16 22:50:59', '2026-03-20 18:35:55', 0, 0, NULL, 0, 0, 21, NULL, '/uploads/manuscripts/10/videos/28/source/video.mp4', 155, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (29, 12, 0, '把 OpenClaw 翻译成“小龙虾”的，出来挨打！', '/uploads/manuscripts/12/videos/29/transcoded/1080p.mp4', '/uploads/manuscripts/12/videos/29/transcoded/720p.mp4', '/uploads/manuscripts/12/videos/29/transcoded/480p.mp4', 5, 1, '2026-03-16 23:20:36', '2026-03-20 19:41:49', -1, 0, NULL, 0, 0, 21, NULL, '/uploads/manuscripts/10/videos/29/source/video.mp4', 196, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (30, 13, 0, '数学怎么提分', '/uploads/manuscripts/13/videos/30/transcoded/1080p.mp4', '/uploads/manuscripts/13/videos/30/transcoded/720p.mp4', '/uploads/manuscripts/13/videos/30/transcoded/480p.mp4', 5, 2, '2026-03-16 23:23:05', '2026-03-20 19:41:56', -1, 0, NULL, 0, 0, 21, NULL, '/uploads/manuscripts/10/videos/30/source/video.mp4', 71, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (31, 14, 0, '每天学习一个电子知识，今天学习如何测电压', '/uploads/manuscripts/14/videos/31/transcoded/1080p.mp4', '/uploads/manuscripts/14/videos/31/transcoded/720p.mp4', '/uploads/manuscripts/14/videos/31/transcoded/480p.mp4', 5, 4, '2026-03-16 23:28:28', '2026-03-20 19:39:34', -1, 0, NULL, 1, 1, 5, NULL, '/uploads/manuscripts/10/videos/31/source/video.mp4', 26, NULL, NULL, NULL, 0);
INSERT INTO `videos` VALUES (32, 21, 0, '投篮发力差的共性和修改思路', NULL, NULL, NULL, NULL, NULL, '2026-03-20 20:12:18', '2026-03-20 20:12:18', 0, 0, NULL, 0, 0, 0, NULL, '/uploads/manuscripts/21/videos/32/source/video.mp4', 69, NULL, NULL, NULL, 0);

SET FOREIGN_KEY_CHECKS = 1;
