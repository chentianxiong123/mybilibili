-- AI 多渠道 API 配置
CREATE TABLE IF NOT EXISTS `ai_api_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '渠道名称，如 DeepSeek / OpenAI / 七牛云',
  `base_url` varchar(255) NOT NULL COMMENT 'API 根地址，如 https://api.deepseek.com',
  `api_key` varchar(255) NOT NULL COMMENT 'API 密钥',
  `model` varchar(100) NOT NULL DEFAULT 'deepseek-chat' COMMENT '模型名称',
  `max_tokens` int NOT NULL DEFAULT 2000,
  `temperature` double NOT NULL DEFAULT 0.7,
  `enabled` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI API 多渠道配置';

-- AI 功能绑定渠道
CREATE TABLE IF NOT EXISTS `ai_bindings` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feature` varchar(30) NOT NULL COMMENT 'CHAT / REVIEW / SUMMARY',
  `api_config_id` bigint NOT NULL COMMENT '绑定的渠道 ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_feature` (`feature`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 功能与渠道绑定关系';
