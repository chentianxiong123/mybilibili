CREATE TABLE IF NOT EXISTS `manuscript_edit_versions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `manuscript_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `before_snapshot` longtext NOT NULL,
  `after_snapshot` longtext NOT NULL,
  `changed_fields` text NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'PENDING',
  `reviewer_id` int(11) DEFAULT NULL,
  `review_reason` varchar(500) DEFAULT NULL,
  `reviewed_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_manuscript_edit_versions_manuscript` (`manuscript_id`),
  KEY `idx_manuscript_edit_versions_user` (`user_id`),
  KEY `idx_manuscript_edit_versions_status` (`status`),
  KEY `idx_manuscript_edit_versions_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
