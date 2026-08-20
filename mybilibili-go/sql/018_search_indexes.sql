-- 中文全文搜索：zhparser 分词
CREATE EXTENSION IF NOT EXISTS zhparser;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_ts_config WHERE cfgname = 'zh_cn') THEN
        CREATE TEXT SEARCH CONFIGURATION zh_cn (PARSER = zhparser);
        ALTER TEXT SEARCH CONFIGURATION zh_cn ADD MAPPING FOR n,v,a,i,e,l,t,b WITH simple;
    END IF;
END
$$;

-- 全文搜索索引：稿件标题+描述
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_manuscripts_search ON manuscripts USING GIN(search_vector);

CREATE OR REPLACE FUNCTION update_manuscript_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('zh_cn', COALESCE(NEW.title, '') || ' ' || COALESCE(NEW.description, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_manuscript_search_vector ON manuscripts;
CREATE TRIGGER trg_manuscript_search_vector
    BEFORE INSERT OR UPDATE OF title, description ON manuscripts
    FOR EACH ROW EXECUTE FUNCTION update_manuscript_search_vector();

-- 对已有数据做一次初始化
UPDATE manuscripts SET search_vector = to_tsvector('zh_cn', COALESCE(title, '') || ' ' || COALESCE(description, ''))
WHERE search_vector IS NULL;

-- 模糊搜索：稿件标题 + 用户昵称
CREATE INDEX IF NOT EXISTS idx_manuscripts_title_trgm ON manuscripts USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_nickname_trgm ON users USING GIN (nickname gin_trgm_ops);

-- 评论全文搜索
CREATE INDEX IF NOT EXISTS idx_comments_content_trgm ON comments USING GIN (content gin_trgm_ops);

-- manuscripts 向量列（用于 AI 推荐）
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS embedding vector(384);
CREATE INDEX IF NOT EXISTS idx_manuscripts_embedding ON manuscripts USING hnsw (embedding vector_cosine_ops);