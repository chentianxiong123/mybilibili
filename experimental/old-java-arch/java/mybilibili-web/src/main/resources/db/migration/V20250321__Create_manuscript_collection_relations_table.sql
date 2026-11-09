-- 创建稿件与合集关联表（多对多关系）
CREATE TABLE IF NOT EXISTS manuscript_collection_relations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    manuscript_id INT NOT NULL COMMENT '稿件ID',
    collection_id INT NOT NULL COMMENT '合集ID',
    collection_order INT DEFAULT 0 COMMENT '在合集中的顺序',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
    FOREIGN KEY (manuscript_id) REFERENCES manuscripts(id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES manuscript_collections(id) ON DELETE CASCADE,
    UNIQUE KEY uk_manuscript_collection (manuscript_id, collection_id),
    INDEX idx_collection_id (collection_id),
    INDEX idx_collection_order (collection_id, collection_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='稿件与合集关联表';

-- 迁移现有数据（如果有）
-- 将 manuscripts 表中的 collection_id 和 collection_order 迁移到中间表
INSERT INTO manuscript_collection_relations (manuscript_id, collection_id, collection_order, created_at)
SELECT id, collection_id, collection_order, upload_time
FROM manuscripts
WHERE collection_id IS NOT NULL
ON DUPLICATE KEY UPDATE collection_order = VALUES(collection_order);
