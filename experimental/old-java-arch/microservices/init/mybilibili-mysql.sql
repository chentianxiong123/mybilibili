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

 Date: 24/05/2026 12:18:14
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
-- Table structure for ai_chat_messages
-- ----------------------------
DROP TABLE IF EXISTS `ai_chat_messages`;
CREATE TABLE `ai_chat_messages`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `conversation_id` bigint(20) NOT NULL COMMENT '所属会话ID',
  `role` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'user 或 assistant',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '消息内容',
  `token_count` int(11) NOT NULL DEFAULT 0 COMMENT '消耗token数',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_conversation_id`(`conversation_id`) USING BTREE,
  CONSTRAINT `fk_ai_msg_conversation` FOREIGN KEY (`conversation_id`) REFERENCES `ai_conversations` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'AI客服消息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of ai_chat_messages
-- ----------------------------

-- ----------------------------
-- Table structure for ai_conversations
-- ----------------------------
DROP TABLE IF EXISTS `ai_conversations`;
CREATE TABLE `ai_conversations`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'AI客服对话' COMMENT '会话标题',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '1=活跃, 0=归档',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'AI客服会话表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of ai_conversations
-- ----------------------------

-- ----------------------------
-- Table structure for banner_images
-- ----------------------------
DROP TABLE IF EXISTS `banner_images`;
CREATE TABLE `banner_images`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `image_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `link_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `sort_order` int(11) NULL DEFAULT 0,
  `status` int(11) NULL DEFAULT 1,
  `type` int(11) NOT NULL COMMENT '1:首页轮播 2:分类轮播 3:顶部背景 4:用户主页背景',
  `category_id` int(11) NULL DEFAULT NULL,
  `start_time` datetime NULL DEFAULT NULL,
  `end_time` datetime NULL DEFAULT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_type`(`type`) USING BTREE,
  INDEX `idx_category_id`(`category_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of banner_images
-- ----------------------------

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
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE,
  INDEX `idx_comments_video_id`(`manuscript_id`) USING BTREE,
  CONSTRAINT `comments_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_comments_manuscript` FOREIGN KEY (`manuscript_id`) REFERENCES `manuscripts` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 18 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of comments
-- ----------------------------
INSERT INTO `comments` VALUES (3, 10, 4, '你好', 2, 3, '2026-03-15 14:23:23', '2026-04-07 12:51:18', 0);
INSERT INTO `comments` VALUES (6, NULL, 4, '测试', 1, 3, '2026-03-16 21:56:24', '2026-03-20 18:23:08', 0);
INSERT INTO `comments` VALUES (7, 14, 4, '交流电压跟直流电压有什么区别呀？😁', 1, 2, '2026-04-06 22:50:22', '2026-05-09 19:42:10', 0);
INSERT INTO `comments` VALUES (8, 10, 4, '有人吗？', 0, 0, '2026-04-07 12:51:34', '2026-04-07 12:51:34', 0);
INSERT INTO `comments` VALUES (9, 13, 4, '数学好难呀', 1, 3, '2026-04-10 17:48:37', '2026-04-10 18:16:28', 0);
INSERT INTO `comments` VALUES (10, 11, 4, '大家都在用什么工具写代码呀？', 1, 0, '2026-04-10 18:05:37', '2026-05-06 23:52:33', 0);
INSERT INTO `comments` VALUES (11, 29, 4, '你好', 0, 0, '2026-04-12 22:32:57', '2026-04-12 22:32:57', 0);
INSERT INTO `comments` VALUES (12, 29, 4, '我会打乒乓球', 2, 0, '2026-04-12 22:39:51', '2026-05-09 17:41:09', 0);
INSERT INTO `comments` VALUES (13, 29, 4, '乒乓球怎么打？', 1, 1, '2026-04-12 23:31:55', '2026-05-09 17:41:07', 0);
INSERT INTO `comments` VALUES (14, 27, 4, 'hello', 0, 0, '2026-04-12 23:39:32', '2026-04-12 23:39:32', 0);
INSERT INTO `comments` VALUES (15, 27, 4, '你好', 1, 1, '2026-04-12 23:55:50', '2026-05-09 19:07:41', 0);
INSERT INTO `comments` VALUES (16, 12, 4, 'crawfish', 2, 0, '2026-04-29 15:55:17', '2026-05-09 15:48:54', 0);
INSERT INTO `comments` VALUES (17, 12, 5, '小龙虾 飞书机器人', 0, 0, '2026-05-09 22:51:11', '2026-05-09 22:51:11', 0);

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
INSERT INTO `conversations` VALUES (7, 5, 4, 90, '在吗', '2026-05-09 19:41:24', 0, '2026-03-17 10:39:14', '2026-05-09 19:41:23');
INSERT INTO `conversations` VALUES (8, 4, 5, 90, '在吗', '2026-05-09 19:41:24', 0, '2026-03-17 10:39:14', '2026-05-16 20:35:54');
INSERT INTO `conversations` VALUES (9, 5, 6, 91, '在吗在吗', '2026-05-09 19:41:37', 0, '2026-03-17 12:25:40', '2026-05-09 19:41:37');
INSERT INTO `conversations` VALUES (10, 6, 5, 91, '在吗在吗', '2026-05-09 19:41:38', 1, '2026-03-17 12:25:40', '2026-05-09 19:41:37');
INSERT INTO `conversations` VALUES (11, 6, 4, 83, '你好，请问找我有什么事吗', '2026-04-10 21:00:06', 1, '2026-03-20 11:48:21', '2026-04-10 21:00:05');
INSERT INTO `conversations` VALUES (12, 4, 6, 83, '你好，请问找我有什么事吗', '2026-04-10 21:00:05', 0, '2026-03-20 11:48:21', '2026-04-10 21:00:05');

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
) ENGINE = InnoDB AUTO_INCREMENT = 18 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '动态评论表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of dynamic_comments
-- ----------------------------
INSERT INTO `dynamic_comments` VALUES (1, 3, 5, '你好', NULL, NULL, 0, 1, '2026-03-30 22:07:36', '2026-03-30 22:07:36');
INSERT INTO `dynamic_comments` VALUES (6, 1, 4, '你们好😄', NULL, NULL, 0, 0, '2026-04-08 12:11:31', '2026-04-08 12:11:31');
INSERT INTO `dynamic_comments` VALUES (7, 1, 4, '你们好', NULL, NULL, 0, 0, '2026-04-08 12:11:48', '2026-04-08 12:11:48');
INSERT INTO `dynamic_comments` VALUES (8, 1, 4, '你们好啊', NULL, NULL, 0, 0, '2026-04-22 14:04:26', '2026-04-22 14:04:26');
INSERT INTO `dynamic_comments` VALUES (9, 1, 4, '你们好啊', NULL, NULL, 1, 0, '2026-04-22 14:05:02', '2026-05-02 22:56:39');
INSERT INTO `dynamic_comments` VALUES (10, 3, 4, '这是什么图片？', NULL, NULL, 1, 0, '2026-04-30 10:32:57', '2026-04-30 13:46:45');
INSERT INTO `dynamic_comments` VALUES (11, 2, 4, '你好', NULL, NULL, 1, 0, '2026-04-30 13:47:01', '2026-04-30 13:47:03');
INSERT INTO `dynamic_comments` VALUES (12, 3, 4, '哈喽哈喽', 10, NULL, 0, 0, '2026-04-30 15:33:59', '2026-04-30 15:33:59');
INSERT INTO `dynamic_comments` VALUES (13, 3, 4, '你好', 10, NULL, 0, 0, '2026-04-30 15:54:04', '2026-04-30 15:54:04');
INSERT INTO `dynamic_comments` VALUES (14, 3, 4, '怎么说', 10, 4, 0, 0, '2026-04-30 15:54:14', '2026-04-30 15:54:14');
INSERT INTO `dynamic_comments` VALUES (15, 3, 4, '你好', 10, NULL, 0, 0, '2026-04-30 15:59:54', '2026-04-30 15:59:54');
INSERT INTO `dynamic_comments` VALUES (16, 1, 4, '6666', 9, NULL, 1, 0, '2026-05-02 22:57:02', '2026-05-02 22:57:04');
INSERT INTO `dynamic_comments` VALUES (17, 1, 4, '66666', 9, 4, 0, 0, '2026-05-09 14:32:52', '2026-05-09 14:32:52');

-- ----------------------------
-- Table structure for favorite_folders
-- ----------------------------
DROP TABLE IF EXISTS `favorite_folders`;
CREATE TABLE `favorite_folders`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `collect_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `user_id`(`user_id`, `name`) USING BTREE,
  CONSTRAINT `favorite_folders_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 15 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of favorite_folders
-- ----------------------------
INSERT INTO `favorite_folders` VALUES (1, 5, '数学', 3, '2026-03-12 17:45:36', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (2, 5, 'AI', 3, '2026-03-12 17:45:58', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (3, 4, '默认收藏夹', 3, '2026-03-13 17:34:21', '2026-04-30 22:40:05');
INSERT INTO `favorite_folders` VALUES (4, 4, 'AI', 1, '2026-03-15 14:49:02', '2026-05-08 14:41:35');
INSERT INTO `favorite_folders` VALUES (5, 5, '默认收藏夹', 3, '2026-03-11 20:47:41', '2026-03-20 20:30:38');
INSERT INTO `favorite_folders` VALUES (6, 6, '默认收藏夹', 0, '2026-03-17 09:00:09', '2026-03-17 09:00:09');
INSERT INTO `favorite_folders` VALUES (7, 5, '计算机', 1, '2026-05-09 16:52:06', '2026-05-09 17:41:57');
INSERT INTO `favorite_folders` VALUES (8, 5, '篮球', 1, '2026-05-09 16:52:14', '2026-05-09 16:52:16');
INSERT INTO `favorite_folders` VALUES (14, 7, '默认收藏夹', 0, '2026-05-20 23:34:58', '2026-05-20 23:34:58');

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
  INDEX `fk_favorite_manuscripts_manuscript`(`manuscript_id`) USING BTREE,
  CONSTRAINT `fk_favorite_manuscripts_folder` FOREIGN KEY (`folder_id`) REFERENCES `favorite_folders` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `fk_favorite_manuscripts_manuscript` FOREIGN KEY (`manuscript_id`) REFERENCES `manuscripts` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 36 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of favorite_manuscripts
-- ----------------------------
INSERT INTO `favorite_manuscripts` VALUES (19, 1, 10, '2026-03-20 12:00:48');
INSERT INTO `favorite_manuscripts` VALUES (20, 2, 10, '2026-03-20 12:00:48');
INSERT INTO `favorite_manuscripts` VALUES (21, 2, 12, '2026-03-20 12:00:57');
INSERT INTO `favorite_manuscripts` VALUES (22, 5, 12, '2026-03-20 12:00:57');
INSERT INTO `favorite_manuscripts` VALUES (23, 1, 13, '2026-03-20 12:01:05');
INSERT INTO `favorite_manuscripts` VALUES (24, 5, 14, '2026-03-20 14:03:20');
INSERT INTO `favorite_manuscripts` VALUES (25, 1, 11, '2026-03-20 20:30:38');
INSERT INTO `favorite_manuscripts` VALUES (26, 2, 11, '2026-03-20 20:30:38');
INSERT INTO `favorite_manuscripts` VALUES (27, 5, 11, '2026-03-20 20:30:38');
INSERT INTO `favorite_manuscripts` VALUES (29, 3, 14, '2026-03-28 14:26:41');
INSERT INTO `favorite_manuscripts` VALUES (31, 3, 11, '2026-04-11 12:21:57');
INSERT INTO `favorite_manuscripts` VALUES (32, 4, 11, '2026-04-11 12:21:57');
INSERT INTO `favorite_manuscripts` VALUES (33, 3, 29, '2026-04-30 22:40:05');
INSERT INTO `favorite_manuscripts` VALUES (34, 8, 27, '2026-05-09 16:52:16');
INSERT INTO `favorite_manuscripts` VALUES (35, 7, 14, '2026-05-09 17:41:57');

-- ----------------------------
-- Table structure for live_linkmic
-- ----------------------------
DROP TABLE IF EXISTS `live_linkmic`;
CREATE TABLE `live_linkmic`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `room_id` bigint(20) NOT NULL COMMENT '直播间ID',
  `streamer_id` bigint(20) NOT NULL COMMENT '主播ID',
  `viewer_id` bigint(20) NOT NULL COMMENT '观众ID（连麦者）',
  `viewer_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '观众用户名',
  `status` tinyint(4) NULL DEFAULT 0 COMMENT '状态: 0=申请中 1=已连接 2=已断开 3=被拒绝',
  `audio_enabled` tinyint(4) NULL DEFAULT 0 COMMENT '音频是否开启: 0=关闭 1=开启',
  `video_enabled` tinyint(4) NULL DEFAULT 0 COMMENT '视频是否开启: 0=关闭 1=开启',
  `max_linkmics` int(11) NULL DEFAULT 3 COMMENT '最大连麦人数',
  `apply_time` datetime NULL DEFAULT NULL COMMENT '申请时间',
  `connect_time` datetime NULL DEFAULT NULL COMMENT '连接时间',
  `end_time` datetime NULL DEFAULT NULL COMMENT '结束时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_room_id`(`room_id`) USING BTREE,
  INDEX `idx_viewer_id`(`viewer_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '直播连麦表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of live_linkmic
-- ----------------------------

-- ----------------------------
-- Table structure for live_rooms
-- ----------------------------
DROP TABLE IF EXISTS `live_rooms`;
CREATE TABLE `live_rooms`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL COMMENT '主播用户ID',
  `room_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '直播间名称',
  `stream_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '推流域名/流密钥 (唯一)',
  `status` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'offline' COMMENT 'offline/live',
  `cover_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '封面图',
  `viewer_count` int(11) NOT NULL DEFAULT 0 COMMENT '在线观众数',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '直播分类',
  `scheduled_at` datetime NULL DEFAULT NULL COMMENT '定时开播时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_stream_key`(`stream_key`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '直播间表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of live_rooms
-- ----------------------------
INSERT INTO `live_rooms` VALUES (1, 4, '我的直播间', 'e4a53048e3fd45c6a91964b44df6262b', 'live', NULL, 0, NULL, NULL, '2026-05-19 11:31:55', '2026-05-19 12:22:28');
INSERT INTO `live_rooms` VALUES (2, 5, '我的直播间', '7d42fcf7f4d84cbf9f24b763b46a34a9', 'offline', NULL, 0, NULL, NULL, '2026-05-20 22:57:40', '2026-05-20 22:57:40');

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
) ENGINE = InnoDB AUTO_INCREMENT = 1637826567 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '稿件与合集关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscript_collection_relations
-- ----------------------------
INSERT INTO `manuscript_collection_relations` VALUES (1, 12, 3, 0, '2026-03-16 23:20:36');
INSERT INTO `manuscript_collection_relations` VALUES (2, 13, 3, 0, '2026-03-16 23:23:05');
INSERT INTO `manuscript_collection_relations` VALUES (1637826563, 11, 1444888585, 0, '2026-04-07 21:03:00');
INSERT INTO `manuscript_collection_relations` VALUES (1637826564, 27, 3, 3, '2026-05-08 12:57:20');
INSERT INTO `manuscript_collection_relations` VALUES (1637826565, 10, 1444888586, 0, '2026-05-09 13:19:44');
INSERT INTO `manuscript_collection_relations` VALUES (1637826566, 11, 1444888586, 1, '2026-05-09 14:32:42');

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
) ENGINE = InnoDB AUTO_INCREMENT = 1444888587 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '稿件合集表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscript_collections
-- ----------------------------
INSERT INTO `manuscript_collections` VALUES (3, 'AI合集', 'AI合集', '/uploads/manuscripts/27/cover.jpg', 5, 3, 0, 1, '2026-03-17 09:54:07', '2026-05-08 13:34:01');
INSERT INTO `manuscript_collections` VALUES (1444888585, '人工智能', '', NULL, 4, 1, NULL, 1, '2026-04-07 21:03:00', '2026-04-29 16:51:07');
INSERT INTO `manuscript_collections` VALUES (1444888586, '电子', '烧录工具', '/uploads/manuscripts/10/cover.jpg', 4, 2, NULL, 1, '2026-05-09 13:19:44', '2026-05-09 14:32:42');

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
  `upload_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `duration` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '总时长显示字符串',
  `duration_seconds` int(11) NULL DEFAULT 0 COMMENT '总时长秒数',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE,
  INDEX `idx_category_id`(`category_id`) USING BTREE,
  INDEX `idx_status`(`status`) USING BTREE,
  INDEX `idx_upload_time`(`upload_time`) USING BTREE,
  CONSTRAINT `fk_manuscripts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 33 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '稿件表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of manuscripts
-- ----------------------------
INSERT INTO `manuscripts` VALUES (10, '无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】', '本视频为无限token版本的openclaw安装教程\r\n任意Windows电脑均可安装 无需付费 无需高配电脑\r\n如果觉得视频对你有帮助 请一键三连加关注 随时关注主播最新教程\r\n如有问题可评论区留言或私信 随缘解答 无套路。小龙虾', '/uploads/manuscripts/10/cover.jpg', 4, 1, 321, 3, 1, 3, 0, 2, 0, 3, 0, NULL, NULL, NULL, '2026-03-15 11:01:33', '2026-05-16 23:49:46', '6:24', 384);
INSERT INTO `manuscripts` VALUES (11, 'AI Agent正在重走操作系统的老路', '一个MCP工具号称能省98%的Context Token，社区瞬间炸锅。但深挖后我发现，这不是一个“省钱”的问题——AI Agent开发正在重演操作系统的进化史。Context就是新时代的内存，而我们还停留在手动管理阶段。一次工具调用吃掉50K token，3-4次就烧掉半个窗口。本期拆解“沙盒+索引+按需检索”架构，以及协议层的结构性缺陷。', '/uploads/manuscripts/11/cover.jpg', 4, 1, 122, 2, 1, 2, 1, 1, 0, 3, 0, NULL, NULL, NULL, '2026-03-16 22:51:00', '2026-05-16 20:33:57', '4:31', 271);
INSERT INTO `manuscripts` VALUES (12, '把 OpenClaw 翻译成“小龙虾”的，出来挨打！', '', '/uploads/manuscripts/12/cover.jpg', 5, 4, 42, 2, 0, 1, 1, 2, 0, 3, 0, NULL, NULL, NULL, '2026-03-16 23:20:36', '2026-05-10 09:27:49', '3:16', 196);
INSERT INTO `manuscripts` VALUES (13, '数学怎么提分', '数学怎么提分#家长收藏孩子受益 #家长必读 #学习方法 #学霸秘籍', '/uploads/manuscripts/13/cover.jpg', 5, 3, 29, 2, 0, 1, 0, 1, 0, 3, 0, NULL, NULL, NULL, '2026-03-16 23:23:05', '2026-05-09 16:49:14', '1:11', 71);
INSERT INTO `manuscripts` VALUES (14, '每天学习一个电子知识，今天学习如何测电压', '', '/uploads/manuscripts/14/cover.jpg', 5, 2, 195, 3, 0, 2, 0, 1, 0, 3, 0, NULL, NULL, NULL, '2026-03-16 23:28:29', '2026-05-10 00:02:32', '0:26', 26);
INSERT INTO `manuscripts` VALUES (27, '投篮力线错误？这样改！', '投篮力线错误？这样改！', '/uploads/manuscripts/27/cover.jpg', 5, 5, 33, 2, 0, 1, 0, 2, 0, 3, 1, NULL, NULL, NULL, '2026-04-03 18:02:58', '2026-05-10 07:55:40', '01:09', 69);
INSERT INTO `manuscripts` VALUES (29, '少儿乒乓球训练', '少儿乒乓球训练', '/uploads/manuscripts/29/cover.jpg', 5, 5, 78, 2, 1, 1, 2, 3, 0, 3, 0, NULL, NULL, NULL, '2026-04-03 19:12:28', '2026-05-09 20:13:44', '00:38', 38);
INSERT INTO `manuscripts` VALUES (31, 'github短视频生成工具', 'github短视频生成工具', '/uploads/manuscripts/31/cover.jpg', 4, 7, 3, 1, 0, 0, 0, 0, 0, 3, 1, NULL, NULL, NULL, '2026-05-09 23:37:44', '2026-05-16 20:33:07', '00:47', 47);
INSERT INTO `manuscripts` VALUES (32, 'Skills', '开源Skills', '/uploads/manuscripts/32/cover.jpg', 4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, NULL, NULL, NULL, '2026-05-10 08:04:36', '2026-05-10 08:05:27', '00:27', 27);

-- ----------------------------
-- Table structure for meeting_participant
-- ----------------------------
DROP TABLE IF EXISTS `meeting_participant`;
CREATE TABLE `meeting_participant`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `room_id` bigint(20) NOT NULL COMMENT '会议室ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `user_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
  `user_avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '用户头像',
  `role` tinyint(4) NULL DEFAULT 0 COMMENT '角色: 0=参与者 1=主持人',
  `audio_enabled` tinyint(4) NULL DEFAULT 0 COMMENT '音频是否开启: 0=关闭 1=开启',
  `video_enabled` tinyint(4) NULL DEFAULT 0 COMMENT '视频是否开启: 0=关闭 1=开启',
  `screen_share_enabled` tinyint(4) NULL DEFAULT 0 COMMENT '屏幕共享是否开启: 0=关闭 1=开启',
  `join_time` datetime NULL DEFAULT NULL COMMENT '加入时间',
  `leave_time` datetime NULL DEFAULT NULL COMMENT '离开时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_room_id`(`room_id`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会议参与者表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of meeting_participant
-- ----------------------------

-- ----------------------------
-- Table structure for meeting_room
-- ----------------------------
DROP TABLE IF EXISTS `meeting_room`;
CREATE TABLE `meeting_room`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `room_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会议室名称',
  `room_code` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邀请码(6位数字)',
  `creator_id` bigint(20) NOT NULL COMMENT '创建者用户ID',
  `creator_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建者用户名',
  `max_participants` int(11) NULL DEFAULT 5 COMMENT '最大参与人数',
  `status` tinyint(4) NULL DEFAULT 0 COMMENT '状态: 0=未开始/已通过 1=进行中 2=已结束 3=待审批 4=已拒绝',
  `start_time` datetime NULL DEFAULT NULL COMMENT '开始时间',
  `end_time` datetime NULL DEFAULT NULL COMMENT '结束时间',
  `scheduled_start` datetime NULL DEFAULT NULL COMMENT '预约开始时间',
  `scheduled_end` datetime NULL DEFAULT NULL COMMENT '预约结束时间',
  `scheduled_reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '预约事由',
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `room_code`(`room_code`) USING BTREE,
  INDEX `idx_room_code`(`room_code`) USING BTREE,
  INDEX `idx_creator_id`(`creator_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会议室表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of meeting_room
-- ----------------------------

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
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '消息设置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of message_settings
-- ----------------------------
INSERT INTO `message_settings` VALUES (1, 5, 1, 1, 1, 1, 1, '2026-03-27 21:00:02', '2026-03-27 21:00:02');
INSERT INTO `message_settings` VALUES (2, 4, 1, 1, 1, 1, 1, '2026-04-10 09:45:42', '2026-05-09 13:10:00');

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
  `comment_id` int(11) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `sender_id`(`sender_id`) USING BTREE,
  INDEX `receiver_id`(`receiver_id`) USING BTREE,
  CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `messages_ibfk_2` FOREIGN KEY (`receiver_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 96 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '消息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of messages
-- ----------------------------
INSERT INTO `messages` VALUES (11, 5, 4, '你好', 1, '2026-03-17 10:39:14', '2026-03-17 12:18:29', 1, NULL, NULL, 7, NULL);
INSERT INTO `messages` VALUES (12, 5, 4, '你好', 1, '2026-03-17 10:42:41', '2026-03-17 12:18:29', 1, NULL, NULL, 7, NULL);
INSERT INTO `messages` VALUES (13, 5, 6, '你好兄弟', 1, '2026-03-17 12:25:40', '2026-03-17 12:26:04', 1, NULL, NULL, 9, NULL);
INSERT INTO `messages` VALUES (14, 6, 5, '你好朋友', 1, '2026-03-17 12:26:19', '2026-03-17 12:26:31', 1, NULL, NULL, 10, NULL);
INSERT INTO `messages` VALUES (15, 6, 4, '你好', 1, '2026-03-20 11:48:21', '2026-03-28 14:22:31', 1, NULL, NULL, 11, NULL);
INSERT INTO `messages` VALUES (42, 5, 4, '赞了你的视频《无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】》', 1, '2026-03-20 11:53:38', '2026-03-28 14:22:29', 4, 25, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (43, 5, 4, '赞了你的视频《AI Agent正在重走操作系统的老路》', 1, '2026-03-16 23:19:34', '2026-03-28 14:22:29', 4, 26, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (73, 5, 4, '赞了你的视频《无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】', 1, '2026-03-20 11:53:38', '2026-04-10 16:19:03', 4, 25, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (74, 5, 4, '赞了你的视频《AI Agent正在重走操作系统的老路》', 1, '2026-03-16 23:19:34', '2026-04-10 16:19:03', 4, 26, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (76, 5, 4, '赞了你的视频《周鸿祎建议大家别只把AI当搜索引擎用》', 1, '2026-04-01 10:00:00', '2026-04-10 16:19:03', 4, 27, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (77, 6, 4, '回复了你的评论：你们好😄', 1, '2026-04-08 12:15:00', '2026-04-10 16:19:01', 2, 1, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (78, 5, 4, '回复了你的评论：你好', 1, '2026-03-30 22:10:00', '2026-04-10 16:19:01', 2, 1, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (79, 5, 4, '您的稿件《无限零成本token，我给OpenClaw换了个永动机芯！》已通过审核并成功上架啦！', 1, '2026-03-15 10:00:00', '2026-04-10 16:19:16', 5, NULL, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (80, 6, 4, '您的稿件《AI Agent正在重走操作系统的老路》已通过审核并成功上架啦！', 1, '2026-03-16 20:00:00', '2026-04-10 16:19:16', 5, NULL, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (81, 4, 5, '您的稿件《每天学习一个电子知识，今天学习如何测电压》已通过审核并成功上架啦！', 1, '2026-03-18 15:00:00', '2026-05-04 22:26:45', 5, NULL, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (82, 6, 5, '您的稿件《投篮发力差的共性和修改思路》已通过审核并成功上架啦！', 1, '2026-04-05 09:00:00', '2026-05-04 22:26:45', 5, NULL, NULL, NULL, NULL);
INSERT INTO `messages` VALUES (83, 4, 6, '你好，请问找我有什么事吗', 0, '2026-04-10 21:00:05', '2026-04-10 21:00:05', 1, NULL, NULL, 12, NULL);
INSERT INTO `messages` VALUES (84, 4, 5, '你好，找我有什么事吗？', 1, '2026-04-12 14:47:27', '2026-05-04 22:26:47', 1, NULL, NULL, 8, NULL);
INSERT INTO `messages` VALUES (85, 5, 4, '赞了你的评论\"怎么说\"', 0, '2026-05-09 14:41:56', '2026-05-09 14:41:56', 6, NULL, NULL, NULL, 13);
INSERT INTO `messages` VALUES (86, 5, 4, '赞了你的评论\"crawfish\"', 0, '2026-05-09 15:48:54', '2026-05-09 15:48:54', 6, NULL, NULL, NULL, 16);
INSERT INTO `messages` VALUES (87, 5, 4, '赞了你的评论\"乒乓球怎么打？\"', 0, '2026-05-09 17:41:06', '2026-05-09 17:41:06', 6, NULL, NULL, NULL, 13);
INSERT INTO `messages` VALUES (88, 5, 4, '赞了你的评论\"我会打乒乓球\"', 0, '2026-05-09 17:41:09', '2026-05-09 17:41:09', 6, NULL, NULL, NULL, 12);
INSERT INTO `messages` VALUES (89, 5, 4, '赞了你的评论\"你好\"', 0, '2026-05-09 19:07:31', '2026-05-09 19:07:31', 6, NULL, NULL, NULL, 15);
INSERT INTO `messages` VALUES (90, 5, 4, '在吗', 1, '2026-05-09 19:41:23', '2026-05-16 20:35:47', 1, NULL, NULL, 7, NULL);
INSERT INTO `messages` VALUES (91, 5, 6, '在吗在吗', 0, '2026-05-09 19:41:37', '2026-05-09 19:41:37', 1, NULL, NULL, 9, NULL);
INSERT INTO `messages` VALUES (92, 5, 4, '赞了你的评论\"哈哈哈\"', 0, '2026-05-09 19:41:59', '2026-05-09 19:41:59', 6, NULL, NULL, NULL, 7);
INSERT INTO `messages` VALUES (93, 5, 4, '赞了你的评论\"交流电压跟直流电压有什么区别呀？😁\"', 0, '2026-05-09 19:41:59', '2026-05-09 19:41:59', 6, NULL, NULL, NULL, 7);
INSERT INTO `messages` VALUES (94, 5, 4, '回复了你：@测试用户：交流电 直流电🤨', 1, '2026-05-09 19:42:10', '2026-05-16 20:35:54', 2, 14, NULL, NULL, 7);
INSERT INTO `messages` VALUES (95, 5, 4, '赞了你的评论\"crawfish\"', 0, '2026-05-09 22:50:52', '2026-05-09 22:50:52', 6, NULL, NULL, NULL, 16);

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
INSERT INTO `permissions` VALUES (9, '管理员管理', 'admin:manage', '/admin', 'GET', NULL, '管理员账号管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (10, '安全设置', 'security:manage', '/security-settings', 'GET', NULL, '安全设置和登录日志管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (11, '直播管理', 'live:manage', '/live', 'GET', NULL, '直播间管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (12, '会议管理', 'meeting:manage', '/meeting', 'GET', NULL, '会议室管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (13, '存储管理', 'storage:manage', '/storage', 'POST', NULL, '对象存储迁移管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (14, '轮播图管理', 'banner:manage', '/banner-images', 'GET', NULL, '轮播图和背景图管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (15, '搜索索引管理', 'search:manage', '/search/admin/index', 'GET', NULL, '搜索索引管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (16, 'AI管理', 'ai:manage', '/ai', 'GET', NULL, 'AI配置、技能和处理任务管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (17, '消息管理', 'message:manage', '/message', 'POST', NULL, '系统消息广播管理权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (18, '审计日志管理', 'audit:manage', '/audit-logs', 'GET', NULL, '后台操作审计日志查询权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');
INSERT INTO `permissions` VALUES (19, '任务中心管理', 'operation:manage', '/operation-tasks', 'GET', NULL, '统一任务中心查询权限', '2026-05-31 00:00:00', '2026-05-31 00:00:00');

-- ----------------------------
-- Table structure for audit_logs
-- ----------------------------
DROP TABLE IF EXISTS `audit_logs`;
CREATE TABLE `audit_logs`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `operator_id` int(11) NULL DEFAULT NULL,
  `operator_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `operator_role` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `module` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `action` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `target_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `request_method` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `request_uri` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `client_ip` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `user_agent` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `result` tinyint(4) NOT NULL COMMENT '1成功 0失败',
  `message` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `detail` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_audit_operator`(`operator_id`) USING BTREE,
  INDEX `idx_audit_module_action`(`module`, `action`) USING BTREE,
  INDEX `idx_audit_result`(`result`) USING BTREE,
  INDEX `idx_audit_target`(`target_type`, `target_id`) USING BTREE,
  INDEX `idx_audit_created_at`(`created_at`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for operation_tasks
-- ----------------------------
DROP TABLE IF EXISTS `operation_tasks`;
CREATE TABLE `operation_tasks`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `target_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `progress` int(11) NOT NULL DEFAULT 0,
  `stage` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `message` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `operator_id` int(11) NULL DEFAULT NULL,
  `operator_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `started_at` datetime NULL DEFAULT NULL,
  `finished_at` datetime NULL DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_operation_task_key`(`task_key`) USING BTREE,
  INDEX `idx_operation_task_type`(`task_type`) USING BTREE,
  INDEX `idx_operation_task_status`(`status`) USING BTREE,
  INDEX `idx_operation_task_target`(`target_type`, `target_id`) USING BTREE,
  INDEX `idx_operation_task_created_at`(`created_at`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

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
-- Table structure for prohibited_words
-- ----------------------------
DROP TABLE IF EXISTS `prohibited_words`;
CREATE TABLE `prohibited_words`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `word` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `match_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'CONTAINS',
  `category` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `is_enabled` int(11) NULL DEFAULT 1,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_word`(`word`) USING BTREE,
  INDEX `idx_category`(`category`) USING BTREE,
  INDEX `idx_is_enabled`(`is_enabled`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of prohibited_words
-- ----------------------------
INSERT INTO `prohibited_words` VALUES (1, '骗子', 'CONTAINS', 'FRAUD', 1, '2026-04-24 15:08:49', '2026-04-24 15:08:49');
INSERT INTO `prohibited_words` VALUES (2, '傻子', 'EXACT', 'OTHER', 1, '2026-04-24 15:08:49', '2026-04-24 15:08:49');
INSERT INTO `prohibited_words` VALUES (3, '广告', 'CONTAINS', 'AD', 1, '2026-04-24 15:08:49', '2026-04-24 15:08:49');
INSERT INTO `prohibited_words` VALUES (4, '政治敏感', 'CONTAINS', 'POLITICS', 1, '2026-04-24 15:08:49', '2026-04-24 15:08:49');
INSERT INTO `prohibited_words` VALUES (5, '测试违禁词', 'CONTAINS', 'TEST', 0, '2026-04-24 15:08:49', '2026-04-24 15:08:49');

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
) ENGINE = InnoDB AUTO_INCREMENT = 14 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of replies
-- ----------------------------
INSERT INTO `replies` VALUES (1, 3, 4, '你好', 0, '2026-03-15 14:23:33', '2026-03-15 14:23:33', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (2, 3, 4, '@string：你好', 0, '2026-03-15 14:23:40', '2026-03-15 14:23:40', 4, 'NORMAL');
INSERT INTO `replies` VALUES (3, 6, 4, '测试', 0, '2026-03-16 21:56:39', '2026-03-16 21:56:39', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (4, 6, 4, '@string：@string 测试', 0, '2026-03-16 22:13:13', '2026-03-16 22:13:13', 4, 'NORMAL');
INSERT INTO `replies` VALUES (5, 6, 4, '@string：测试', 0, '2026-03-16 22:19:09', '2026-03-16 22:19:09', 4, 'NORMAL');
INSERT INTO `replies` VALUES (6, 3, 5, '你好', 0, '2026-03-22 20:30:08', '2026-03-22 20:30:08', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (7, 7, 4, '哈哈哈', 1, '2026-04-10 17:53:39', '2026-05-09 19:41:59', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (8, 9, 4, '数学如何快速提分？', 1, '2026-04-10 18:06:33', '2026-04-10 18:25:29', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (9, 9, 4, '数学如何快速提分？', 0, '2026-04-10 18:06:52', '2026-04-10 18:06:52', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (10, 9, 4, '大家是怎么学数学的?', 0, '2026-04-10 18:16:28', '2026-04-10 18:16:28', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (11, 13, 4, '怎么说', 1, '2026-05-02 14:18:26', '2026-05-09 14:41:56', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (12, 15, 5, '😗😗😗', 1, '2026-05-09 19:07:41', '2026-05-09 19:07:46', NULL, 'NORMAL');
INSERT INTO `replies` VALUES (13, 7, 5, '@测试用户：交流电 直流电🤨', 0, '2026-05-09 19:42:10', '2026-05-09 19:42:10', 4, 'NORMAL');

-- ----------------------------
-- Table structure for reports
-- ----------------------------
DROP TABLE IF EXISTS `reports`;
CREATE TABLE `reports`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `reporter_id` int(11) NOT NULL COMMENT '举报人ID',
  `target_type` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '举报目标类型: MANUSCRIPT/COMMENT/REPLY/DYNAMIC_COMMENT',
  `target_id` int(11) NOT NULL COMMENT '举报目标ID',
  `manuscript_id` int(11) NULL DEFAULT NULL COMMENT '关联稿件ID',
  `reason` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '举报原因',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '补充说明',
  `status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'PENDING' COMMENT '状态: PENDING/RESOLVED/REJECTED',
  `admin_remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '管理员备注',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `processed_at` datetime NULL DEFAULT NULL COMMENT '处理时间',
  `ai_review_status` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'PENDING' COMMENT 'AI审核状态: PENDING/COMPLETED/FAILED',
  `ai_verdict` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'AI审核结果',
  `ai_risk_level` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'AI风险等级: HIGH/MEDIUM/LOW',
  `ai_reviewed_at` datetime NULL DEFAULT NULL COMMENT 'AI审核时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of reports
-- ----------------------------
INSERT INTO `reports` VALUES (1, 5, 'MANUSCRIPT', 27, 27, '欺诈诈骗', '引战', 'PENDING', NULL, '2026-05-09 19:31:22', NULL, 'PENDING', NULL, NULL, NULL);

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
INSERT INTO `role_permissions` VALUES (2, 7);
INSERT INTO `role_permissions` VALUES (1, 8);
INSERT INTO `role_permissions` VALUES (1, 9);
INSERT INTO `role_permissions` VALUES (1, 10);
INSERT INTO `role_permissions` VALUES (1, 11);
INSERT INTO `role_permissions` VALUES (2, 11);
INSERT INTO `role_permissions` VALUES (1, 12);
INSERT INTO `role_permissions` VALUES (2, 12);
INSERT INTO `role_permissions` VALUES (1, 13);
INSERT INTO `role_permissions` VALUES (1, 14);
INSERT INTO `role_permissions` VALUES (1, 15);
INSERT INTO `role_permissions` VALUES (1, 16);
INSERT INTO `role_permissions` VALUES (2, 16);
INSERT INTO `role_permissions` VALUES (1, 17);
INSERT INTO `role_permissions` VALUES (1, 18);
INSERT INTO `role_permissions` VALUES (2, 18);
INSERT INTO `role_permissions` VALUES (1, 19);
INSERT INTO `role_permissions` VALUES (2, 19);


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
-- Table structure for shares
-- ----------------------------
DROP TABLE IF EXISTS `shares`;
CREATE TABLE `shares`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NULL DEFAULT NULL,
  `manuscript_id` int(11) NOT NULL,
  `channel` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'unknown',
  `ip_address` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_manuscript_id`(`manuscript_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of shares
-- ----------------------------
INSERT INTO `shares` VALUES (1, 4, 29, 'unknown', NULL, '2026-05-01 13:29:26');
INSERT INTO `shares` VALUES (2, 4, 29, 'unknown', NULL, '2026-05-02 14:17:56');
INSERT INTO `shares` VALUES (3, 4, 29, 'unknown', NULL, '2026-05-02 14:18:01');
INSERT INTO `shares` VALUES (4, 4, 29, 'unknown', NULL, '2026-05-02 14:20:44');
INSERT INTO `shares` VALUES (5, 4, 29, 'unknown', NULL, '2026-05-02 14:20:51');
INSERT INTO `shares` VALUES (6, 4, 11, 'unknown', NULL, '2026-05-08 15:05:53');
INSERT INTO `shares` VALUES (7, 5, 29, 'unknown', NULL, '2026-05-09 14:45:17');
INSERT INTO `shares` VALUES (8, 5, 29, 'unknown', NULL, '2026-05-09 14:45:30');
INSERT INTO `shares` VALUES (9, 5, 12, 'unknown', NULL, '2026-05-09 15:48:49');

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
) ENGINE = InnoDB AUTO_INCREMENT = 2120134658 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tags
-- ----------------------------
INSERT INTO `tags` VALUES (-2111918079, 'AI', '2026-05-09 23:38:34', '2026-05-09 23:38:34');
INSERT INTO `tags` VALUES (-2044809214, 'github', '2026-05-09 23:38:34', '2026-05-09 23:38:34');
INSERT INTO `tags` VALUES (-1650552831, '训练', '2026-04-03 19:12:28', '2026-04-03 19:12:28');
INSERT INTO `tags` VALUES (-1587638271, '乒乓球', '2026-04-03 19:12:28', '2026-04-03 19:12:28');
INSERT INTO `tags` VALUES (-161566718, '线性代数', '2026-05-09 22:54:22', '2026-05-09 22:54:22');
INSERT INTO `tags` VALUES (-161566717, '线代', '2026-05-09 22:54:22', '2026-05-09 22:54:22');
INSERT INTO `tags` VALUES (-98652158, '高数', '2026-05-09 22:54:22', '2026-05-09 22:54:22');
INSERT INTO `tags` VALUES (-98652157, '数学', '2026-05-09 22:54:22', '2026-05-09 22:54:22');
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
INSERT INTO `tags` VALUES (400424961, '人工智能', '2026-05-10 08:05:27', '2026-05-10 08:05:27');
INSERT INTO `tags` VALUES (438173698, '开源', '2026-05-10 08:05:27', '2026-05-10 08:05:27');
INSERT INTO `tags` VALUES (2120134657, '软件', '2026-05-09 23:38:34', '2026-05-09 23:38:34');

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
  `ref_manuscript_id` int(11) NULL DEFAULT NULL COMMENT '引用的稿件ID',
  `like_count` int(11) NULL DEFAULT 0,
  `comment_count` int(11) NULL DEFAULT 0,
  `share_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id`) USING BTREE,
  CONSTRAINT `user_dynamics_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 1520365570 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_dynamics
-- ----------------------------
INSERT INTO `user_dynamics` VALUES (-576823294, 4, 'hermes agent', 1, NULL, NULL, 1, 0, 0, '2026-05-09 23:32:37', 0);
INSERT INTO `user_dynamics` VALUES (1, 4, '6666', 0, NULL, NULL, 2, 4, 0, '2026-03-07 12:15:24', 0);
INSERT INTO `user_dynamics` VALUES (2, 6, '大家好', 0, NULL, NULL, 1, 1, 0, '2026-03-20 11:52:09', 0);
INSERT INTO `user_dynamics` VALUES (3, 5, '😀😀😀AI agent', 2, '/uploads/images/20260320152632_cc7c57ed-539c-4e58-a15c-211b194cacc8.jpg', 12, 2, 1, 3, '2026-03-20 15:26:33', 0);
INSERT INTO `user_dynamics` VALUES (1520365569, 5, '海思芯片', 1, '/covers/29ab6969-d9d0-41c1-9690-8eb7905c15a6.png', NULL, 1, 0, 0, '2026-05-05 22:24:51', 0);

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
) ENGINE = InnoDB AUTO_INCREMENT = 2053143553313259523 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户交互记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_interactions
-- ----------------------------
INSERT INTO `user_interactions` VALUES (15, 4, 'MANUSCRIPT', 10, 'LIKE', '2026-03-16 17:15:11');
INSERT INTO `user_interactions` VALUES (29, 4, 'DYNAMIC', 1, 'LIKE', '2026-03-16 21:27:11');
INSERT INTO `user_interactions` VALUES (31, 5, 'MANUSCRIPT', 11, 'LIKE', '2026-03-16 23:19:34');
INSERT INTO `user_interactions` VALUES (32, 5, 'MANUSCRIPT', 12, 'LIKE', '2026-03-17 11:57:40');
INSERT INTO `user_interactions` VALUES (35, 4, 'COMMENT', 6, 'LIKE', '2026-03-19 20:32:07');
INSERT INTO `user_interactions` VALUES (36, 5, 'MANUSCRIPT', 14, 'LIKE', '2026-03-19 21:48:30');
INSERT INTO `user_interactions` VALUES (37, 4, 'MANUSCRIPT', 14, 'LIKE', '2026-03-20 11:15:41');
INSERT INTO `user_interactions` VALUES (39, 6, 'USER', 5, 'FOLLOW', '2026-03-20 11:47:54');
INSERT INTO `user_interactions` VALUES (40, 6, 'USER', 4, 'FOLLOW', '2026-03-20 11:47:59');
INSERT INTO `user_interactions` VALUES (42, 5, 'MANUSCRIPT', 10, 'COLLECT', '2026-03-20 12:00:48');
INSERT INTO `user_interactions` VALUES (43, 5, 'MANUSCRIPT', 12, 'COLLECT', '2026-03-20 12:00:57');
INSERT INTO `user_interactions` VALUES (44, 5, 'MANUSCRIPT', 13, 'COLLECT', '2026-03-20 12:01:05');
INSERT INTO `user_interactions` VALUES (45, 5, 'MANUSCRIPT', 13, 'LIKE', '2026-03-20 12:01:06');
INSERT INTO `user_interactions` VALUES (46, 5, 'MANUSCRIPT', 14, 'COLLECT', '2026-03-20 14:03:20');
INSERT INTO `user_interactions` VALUES (48, 5, 'MANUSCRIPT', 11, 'COLLECT', '2026-03-20 20:30:38');
INSERT INTO `user_interactions` VALUES (2037433097214943234, 1, 'MANUSCRIPT', 10, 'LIKE', '2026-03-27 15:34:38');
INSERT INTO `user_interactions` VALUES (2037435642637352962, 1, 'MANUSCRIPT', 10, 'COLLECT', '2026-03-27 15:44:45');
INSERT INTO `user_interactions` VALUES (2037533243521048577, 5, 'MANUSCRIPT', 10, 'LIKE', '2026-03-27 22:12:35');
INSERT INTO `user_interactions` VALUES (2037777389754122242, 4, 'MANUSCRIPT', 13, 'LIKE', '2026-03-28 14:22:44');
INSERT INTO `user_interactions` VALUES (2037778383862890497, 4, 'MANUSCRIPT', 14, 'COLLECT', '2026-03-28 14:26:41');
INSERT INTO `user_interactions` VALUES (2038926608628125698, 5, 'COMMENT', 3, 'LIKE', '2026-03-31 18:29:19');
INSERT INTO `user_interactions` VALUES (2041375449751134209, 4, 'DYNAMIC_COMMENT', 2, 'LIKE', '2026-04-07 12:40:08');
INSERT INTO `user_interactions` VALUES (2041378259502800898, 4, 'COMMENT', 3, 'LIKE', '2026-04-07 12:51:18');
INSERT INTO `user_interactions` VALUES (2041505816273395714, 4, 'DYNAMIC_COMMENT', 1, 'LIKE', '2026-04-07 21:18:10');
INSERT INTO `user_interactions` VALUES (2042129247121657857, 4, 'MANUSCRIPT', 12, 'LIKE', '2026-04-09 14:35:28');
INSERT INTO `user_interactions` VALUES (2042544719495352322, 4, 'COMMENT', 9, 'LIKE', '2026-04-10 18:06:24');
INSERT INTO `user_interactions` VALUES (2042549522392862721, 4, 'REPLY', 8, 'LIKE', '2026-04-10 18:25:29');
INSERT INTO `user_interactions` VALUES (2042550318928941058, 4, 'MANUSCRIPT', 10, 'COIN', '2026-04-10 18:28:39');
INSERT INTO `user_interactions` VALUES (2042820136252936194, 4, 'MANUSCRIPT', 11, 'COIN', '2026-04-11 12:20:48');
INSERT INTO `user_interactions` VALUES (2042820424225460226, 4, 'MANUSCRIPT', 11, 'COLLECT', '2026-04-11 12:21:57');
INSERT INTO `user_interactions` VALUES (2049405363599355905, 4, 'COMMENT', 16, 'LIKE', '2026-04-29 16:28:09');
INSERT INTO `user_interactions` VALUES (2049423707421470722, 5, 'MANUSCRIPT', 29, 'LIKE', '2026-04-29 17:41:02');
INSERT INTO `user_interactions` VALUES (2049503931651805185, 4, 'MANUSCRIPT', 27, 'LIKE', '2026-04-29 22:59:49');
INSERT INTO `user_interactions` VALUES (2049727135955755010, 4, 'DYNAMIC_COMMENT', 10, 'LIKE', '2026-04-30 13:46:45');
INSERT INTO `user_interactions` VALUES (2049727211285454849, 4, 'DYNAMIC_COMMENT', 11, 'LIKE', '2026-04-30 13:47:03');
INSERT INTO `user_interactions` VALUES (2049861294334361602, 4, 'MANUSCRIPT', 29, 'COIN', '2026-04-30 22:39:51');
INSERT INTO `user_interactions` VALUES (2049861351209123842, 4, 'MANUSCRIPT', 29, 'COLLECT', '2026-04-30 22:40:05');
INSERT INTO `user_interactions` VALUES (2049861654306308098, 4, 'COMMENT', 12, 'LIKE', '2026-04-30 22:41:17');
INSERT INTO `user_interactions` VALUES (2049865328583720962, 4, 'COMMENT', 13, 'LIKE', '2026-04-30 22:55:53');
INSERT INTO `user_interactions` VALUES (2050085165146738689, 4, 'MANUSCRIPT', 29, 'SHARE', '2026-05-01 13:29:26');
INSERT INTO `user_interactions` VALUES (2050460485855141890, 4, 'MANUSCRIPT', 29, 'LIKE', '2026-05-02 14:20:50');
INSERT INTO `user_interactions` VALUES (2050590296812511233, 4, 'DYNAMIC_COMMENT', 9, 'LIKE', '2026-05-02 22:56:39');
INSERT INTO `user_interactions` VALUES (2050590400021749761, 4, 'DYNAMIC_COMMENT', 16, 'LIKE', '2026-05-02 22:57:04');
INSERT INTO `user_interactions` VALUES (2051300263848775681, 4, 'MANUSCRIPT', 11, 'LIKE', '2026-05-04 21:57:48');
INSERT INTO `user_interactions` VALUES (2052037123977973762, 5, 'user', 4, 'follow', '2026-05-06 22:45:50');
INSERT INTO `user_interactions` VALUES (2052053913630101506, 5, 'COMMENT', 10, 'LIKE', '2026-05-06 23:52:33');
INSERT INTO `user_interactions` VALUES (2052646150667677698, 4, 'MANUSCRIPT', 11, 'SHARE', '2026-05-08 15:05:53');
INSERT INTO `user_interactions` VALUES (2052979238669025282, 4, 'user', 5, 'follow', '2026-05-09 13:09:27');
INSERT INTO `user_interactions` VALUES (2053002510399918081, 5, 'REPLY', 11, 'LIKE', '2026-05-09 14:41:56');
INSERT INTO `user_interactions` VALUES (2053003356332314625, 5, 'MANUSCRIPT', 29, 'SHARE', '2026-05-09 14:45:17');
INSERT INTO `user_interactions` VALUES (2053019345874800642, 5, 'MANUSCRIPT', 12, 'SHARE', '2026-05-09 15:48:49');
INSERT INTO `user_interactions` VALUES (2053019363348271105, 5, 'COMMENT', 16, 'LIKE', '2026-05-09 15:48:54');
INSERT INTO `user_interactions` VALUES (2053035215422390274, 5, 'MANUSCRIPT', 27, 'LIKE', '2026-05-09 16:51:53');
INSERT INTO `user_interactions` VALUES (2053035313002872833, 5, 'MANUSCRIPT', 27, 'COLLECT', '2026-05-09 16:52:16');
INSERT INTO `user_interactions` VALUES (2053047614149943297, 5, 'COMMENT', 12, 'LIKE', '2026-05-09 17:41:09');
INSERT INTO `user_interactions` VALUES (2053069347250688001, 5, 'COMMENT', 15, 'LIKE', '2026-05-09 19:07:31');
INSERT INTO `user_interactions` VALUES (2053069412103016449, 5, 'REPLY', 12, 'LIKE', '2026-05-09 19:07:46');
INSERT INTO `user_interactions` VALUES (2053078020538646529, 5, 'REPLY', 7, 'LIKE', '2026-05-09 19:41:59');
INSERT INTO `user_interactions` VALUES (2053078023940227073, 5, 'COMMENT', 7, 'LIKE', '2026-05-09 19:41:59');
INSERT INTO `user_interactions` VALUES (2053116048023904258, -69271551, 'MANUSCRIPT', 14, 'LIKE', '2026-05-09 22:13:05');
INSERT INTO `user_interactions` VALUES (2053116721885020162, -69271551, 'user', 5, 'follow', '2026-05-09 22:15:46');
INSERT INTO `user_interactions` VALUES (2053143553313259522, 4, 'MANUSCRIPT', 31, 'LIKE', '2026-05-10 00:02:23');

-- ----------------------------
-- Table structure for user_privacy_settings
-- ----------------------------
DROP TABLE IF EXISTS `user_privacy_settings`;
CREATE TABLE `user_privacy_settings`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `public_collection` tinyint(1) NULL DEFAULT 1,
  `public_birthday_tags` tinyint(1) NULL DEFAULT 0,
  `public_coin_videos` tinyint(1) NULL DEFAULT 0,
  `public_like_videos` tinyint(1) NULL DEFAULT 0,
  `public_following_list` tinyint(1) NULL DEFAULT 0,
  `public_followers_list` tinyint(1) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_privacy_settings
-- ----------------------------
INSERT INTO `user_privacy_settings` VALUES (-2145452031, -1767964670, 1, 0, 0, 0, 0, 0, '2026-05-20 22:45:34', '2026-05-20 22:45:34');
INSERT INTO `user_privacy_settings` VALUES (-2019622911, -69271551, 1, 0, 0, 0, 0, 0, '2026-05-09 22:12:46', '2026-05-09 22:12:46');
INSERT INTO `user_privacy_settings` VALUES (-1906376703, 4, 1, 0, 0, 0, 0, 0, '2026-05-02 17:55:22', '2026-05-02 17:55:22');
INSERT INTO `user_privacy_settings` VALUES (-1713426431, 5, 1, 0, 0, 0, 0, 0, '2026-05-04 22:26:50', '2026-05-04 22:26:50');

-- ----------------------------
-- Table structure for user_tags
-- ----------------------------
DROP TABLE IF EXISTS `user_tags`;
CREATE TABLE `user_tags`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `tag_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_tag`(`user_id`, `tag_name`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 752762882 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_tags
-- ----------------------------
INSERT INTO `user_tags` VALUES (-1629548543, 5, '全栈工程师', '2026-05-06 22:45:16');
INSERT INTO `user_tags` VALUES (-731967487, 5, '前端开发者', '2026-05-06 22:45:09');
INSERT INTO `user_tags` VALUES (752762881, 4, '程序员', '2026-05-03 23:07:56');

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
  `signature` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT '' COMMENT '个人签名',
  `birthdate` date NULL DEFAULT NULL,
  `level` int(11) NULL DEFAULT 1,
  `following_count` int(11) NULL DEFAULT 0,
  `follower_count` int(11) NULL DEFAULT 0,
  `manuscript_count` int(11) NULL DEFAULT 0,
  `liked_count` int(11) NULL DEFAULT 0,
  `coin_count` int(11) NULL DEFAULT 0,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `status` int(11) NULL DEFAULT 0,
  `pinned_video_id` int(11) NULL DEFAULT NULL COMMENT '置顶视频ID',
  `experience` int(11) NULL DEFAULT 0,
  `bio` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  `announcement` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username`) USING BTREE,
  UNIQUE INDEX `email`(`email`) USING BTREE,
  UNIQUE INDEX `phone`(`phone`) USING BTREE,
  INDEX `idx_users_pinned_video_id`(`pinned_video_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO `users` VALUES (4, 'string', '$2a$10$4CcJDm2RUaRhXMppYr.Rj.Q7yMK.NQFSRi7QdeH8FUa/ELfLGpwXG', '测试用户', '/uploads/avatars/4/avatar.jpg', '123@qq.com', '13011011111', 1, '你好朋友们', '2026-03-17', 1, 1, 1, 0, 11, 207, '2026-03-02 22:31:41', '2026-05-20 00:00:01', 1, 10, 57, '', '999');
INSERT INTO `users` VALUES (5, 'test', '$2a$10$EiwLhC5FdeaALJmFEdVdg.fQvbuKaTNM2HWFweipvK6ieoQbSMzV6', 'test666', '/uploads/avatars/5/avatar.jpg', '321@qq.com', '17733311111', 1, '哈哈哈哈哈', '2026-03-09', 1, 1, 2, 0, 13, 210, '2026-03-04 20:48:47', '2026-05-20 00:00:01', 1, 14, 9, '666', '666');
INSERT INTO `users` VALUES (6, 'admin', '$2a$10$FpNqkWXQJoaOwAYCeWzK6u5EMMgSypBWwuvU5H7lUbolDB3AXtf.W', 'admin', 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png', '12111@qq.com', '11111111111', NULL, '', '2026-03-19', 1, 0, 0, 0, 0, 210, '2026-03-17 09:00:05', '2026-05-20 23:36:33', 1, NULL, 0, '', NULL);
INSERT INTO `users` VALUES (7, 'stringstring', '$2a$10$pnyDK2JNckAbQljEOiKuxeb0wZmJgiOdPIlOt8.5nLQGC7Az3vkUG', 'stringstring', 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png', '123123@123.com', NULL, 0, '', NULL, 1, 0, 0, 0, 0, 0, '2026-05-20 23:34:50', '2026-05-20 23:36:33', 1, NULL, 0, '', NULL);

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
) ENGINE = InnoDB AUTO_INCREMENT = 22 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of video_tags
-- ----------------------------
INSERT INTO `video_tags` VALUES (1, 29, 20, '2026-03-16 23:20:36');
INSERT INTO `video_tags` VALUES (2, 30, 20, '2026-03-16 23:23:05');
INSERT INTO `video_tags` VALUES (3, 31, 20, '2026-03-16 23:28:28');
INSERT INTO `video_tags` VALUES (9, 34, 26, '2026-04-03 18:02:58');
INSERT INTO `video_tags` VALUES (10, 36, -1650552831, '2026-04-03 19:12:28');
INSERT INTO `video_tags` VALUES (11, 36, -1587638271, '2026-04-03 19:12:28');
INSERT INTO `video_tags` VALUES (16, 38, 2120134657, '2026-05-09 23:38:34');
INSERT INTO `video_tags` VALUES (17, 38, -2111918079, '2026-05-09 23:38:34');
INSERT INTO `video_tags` VALUES (18, 38, -2044809214, '2026-05-09 23:38:34');
INSERT INTO `video_tags` VALUES (19, 39, 400424961, '2026-05-10 08:05:27');
INSERT INTO `video_tags` VALUES (20, 39, 2120134657, '2026-05-10 08:05:27');
INSERT INTO `video_tags` VALUES (21, 39, 438173698, '2026-05-10 08:05:27');

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
  `upload_time` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `process_progress` int(11) NULL DEFAULT 0 COMMENT '处理进度：0-100',
  `process_stage` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '当前处理阶段',
  `has_subtitle` tinyint(4) NULL DEFAULT 0 COMMENT '是否有字幕',
  `has_summary` tinyint(4) NULL DEFAULT 0 COMMENT '是否有摘要',
  `process_status` int(11) NULL DEFAULT 0 COMMENT 'Processing status: 0-pending 1-transcoding 2-audio extracting 3-subtitle generating 4-AI summarizing 5-completed 6-transcode failed 7-audio failed 8-subtitle failed 9-AI failed',
  `process_error` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'Processing failure reason',
  `source_video_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'Source video URL (for admin preview)',
  `duration_seconds` int(11) NULL DEFAULT 0 COMMENT 'Video duration in seconds',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_videos_upload_time`(`upload_time`) USING BTREE,
  INDEX `idx_manuscript_id`(`manuscript_id`) USING BTREE,
  CONSTRAINT `fk_videos_manuscript` FOREIGN KEY (`manuscript_id`) REFERENCES `manuscripts` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 40 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of videos
-- ----------------------------
INSERT INTO `videos` VALUES (25, 10, 0, '无限零成本token，我给OpenClaw换了个永动机芯！【小白安装教程】', '/uploads/manuscripts/10/videos/25/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/10/videos/25/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/10/videos/25/transcoded/480p/playlist.m3u8', '2026-03-15 11:01:32', '2026-04-11 10:49:04', 0, NULL, 1, 1, 5, '', '/uploads/manuscripts/10/videos/25/source/video.mp4', 384);
INSERT INTO `videos` VALUES (26, 11, 0, 'AI Agent正在重走操作系统的老路', '/uploads/manuscripts/11/videos/26/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/11/videos/26/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/11/videos/26/transcoded/480p/playlist.m3u8', '2026-03-16 22:50:59', '2026-04-11 10:49:04', 0, NULL, 0, 1, 5, NULL, '/uploads/manuscripts/10/videos/26/source/video.mp4', 271);
INSERT INTO `videos` VALUES (27, 11, 1, '谁在给AI投毒？GEO灰产是如何割韭菜的？', '/uploads/manuscripts/11/videos/27/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/11/videos/27/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/11/videos/27/transcoded/480p/playlist.m3u8', '2026-03-16 22:50:59', '2026-04-11 16:17:49', 0, NULL, 0, 1, 5, NULL, '/uploads/manuscripts/10/videos/27/source/video.mp4', 230);
INSERT INTO `videos` VALUES (28, 11, 2, '周鸿祎建议大家别只把AI当搜索引擎用', '/uploads/manuscripts/11/videos/28/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/11/videos/28/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/11/videos/28/transcoded/480p/playlist.m3u8', '2026-03-16 22:50:59', '2026-04-11 17:13:39', 0, NULL, 1, 1, 5, NULL, '/uploads/manuscripts/10/videos/28/source/video.mp4', 155);
INSERT INTO `videos` VALUES (29, 12, 0, '把 OpenClaw 翻译成“小龙虾”的，出来挨打！', '/uploads/manuscripts/12/videos/29/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/12/videos/29/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/12/videos/29/transcoded/480p/playlist.m3u8', '2026-03-16 23:20:36', '2026-04-11 22:21:28', 0, NULL, 0, 1, 5, NULL, '/uploads/manuscripts/10/videos/29/source/video.mp4', 196);
INSERT INTO `videos` VALUES (30, 13, 0, '数学怎么提分', '/uploads/manuscripts/13/videos/30/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/13/videos/30/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/13/videos/30/transcoded/480p/playlist.m3u8', '2026-03-16 23:23:05', '2026-04-11 22:35:43', 0, NULL, 0, 1, 5, NULL, '/uploads/manuscripts/10/videos/30/source/video.mp4', 71);
INSERT INTO `videos` VALUES (31, 14, 0, '每天学习一个电子知识，今天学习如何测电压', '/uploads/manuscripts/14/videos/31/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/14/videos/31/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/14/videos/31/transcoded/480p/playlist.m3u8', '2026-03-16 23:28:28', '2026-04-11 10:49:04', 0, NULL, 1, 1, 5, NULL, '/uploads/manuscripts/10/videos/31/source/video.mp4', 26);
INSERT INTO `videos` VALUES (34, 27, 0, '投篮力线错误？这样改！', '/uploads/manuscripts/27/videos/34/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/27/videos/34/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/27/videos/34/transcoded/480p/playlist.m3u8', '2026-04-03 18:02:58', '2026-05-09 14:01:34', 100, 'AI_SUCCESS', 1, 1, 5, NULL, '/uploads/manuscripts/27/videos/34/source/video.mp4', 69);
INSERT INTO `videos` VALUES (36, 29, 0, '9', '/uploads/manuscripts/29/videos/36/transcoded/1080p/playlist.m3u8', '/uploads/manuscripts/29/videos/36/transcoded/720p/playlist.m3u8', '/uploads/manuscripts/29/videos/36/transcoded/480p/playlist.m3u8', '2026-04-03 19:12:28', '2026-05-09 23:11:28', 100, 'AI_SUCCESS', 1, 1, 5, NULL, '/uploads/manuscripts/29/videos/36/source/video.mp4', 38);
INSERT INTO `videos` VALUES (38, 31, 0, 'github短视频生成工具', NULL, NULL, NULL, '2026-05-09 23:38:34', '2026-05-09 23:54:47', 0, 'AI_SUCCESS', 1, 1, 5, NULL, '/uploads/manuscripts/31/videos/38/source/video.mp4', 47);
INSERT INTO `videos` VALUES (39, 32, 0, '开源Skills', NULL, NULL, NULL, '2026-05-10 08:05:28', '2026-05-10 08:05:27', 0, NULL, NULL, NULL, 0, NULL, '/uploads/manuscripts/32/videos/39/source/video.mp4', 27);

SET FOREIGN_KEY_CHECKS = 1;

