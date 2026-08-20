-- 清理废弃的 search_documents 表（旧 SearchEngine 抽象层，未被实际查询）
DROP INDEX IF EXISTS idx_search_documents_gin;
DROP TABLE IF EXISTS search_documents;