package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// SkillIndex 用 langchaingo(pgvector vectorstore + embeddings) 做客服技能语义检索。
// 未配置 AI_PGVECTOR_URL 或模型不支持嵌入时，NewSkillIndex 返回 error，调用方回退关键词匹配。
type SkillIndex struct {
	store pgvector.Store
}

// NewSkillIndex 构造技能索引。llm 需实现 embeddings.EmbedderClient（如 langchaingo openai.LLM）。
// connURL 为 pgx 连接串（如 postgres://postgres:postgres@127.0.0.1:5432/mybilibili）。
func NewSkillIndex(llm llms.Model, connURL string) (*SkillIndex, error) {
	client, ok := llm.(embeddings.EmbedderClient)
	if !ok || llm == nil {
		return nil, fmt.Errorf("skill index: llm 不支持嵌入")
	}
	if connURL == "" {
		return nil, fmt.Errorf("skill index: AI_PGVECTOR_URL 未设置")
	}
	embedder, err := embeddings.NewEmbedder(client)
	if err != nil {
		return nil, fmt.Errorf("skill index: embedder: %w", err)
	}
	store, err := pgvector.New(context.Background(),
		pgvector.WithConnectionURL(connURL),
		pgvector.WithEmbedder(embedder),
		pgvector.WithCollectionName("ai_skills_embeddings"),
		pgvector.WithVectorDimensions(1536),
	)
	if err != nil {
		return nil, fmt.Errorf("skill index: pgvector store: %w", err)
	}
	return &SkillIndex{store: store}, nil
}

// AddSkill 将技能文本嵌入并加入向量库。
func (s *SkillIndex) AddSkill(ctx context.Context, id int64, text string) error {
	_, err := s.store.AddDocuments(ctx, []schema.Document{{
		PageContent: text,
		Metadata:    map[string]any{"skill_id": id},
	}})
	return err
}

// Search 语义检索技能，返回相似度降序文档（PageContent + Metadata["skill_id"]）。
func (s *SkillIndex) Search(ctx context.Context, query string, limit int) ([]schema.Document, error) {
	if limit <= 0 {
		limit = 5
	}
	return s.store.SimilaritySearch(ctx, query, limit)
}

// vectorStoreEnabled 判断是否配置了 pgvector 向量库（供 main.go 懒初始化）。
func vectorStoreEnabled() bool { return os.Getenv("AI_PGVECTOR_URL") != "" }

var _ vectorstores.VectorStore = pgvector.Store{}
