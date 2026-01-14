package analyzer

import (
	"ai_slow_log/internal/rag"
	"context"
)

// RAGRetrieverAdapter 将 rag.Retriever 适配到 analyzer.Retriever
type RAGRetrieverAdapter struct {
	retriever rag.Retriever
}

// NewRAGRetrieverAdapter 创建 RAG 检索器适配器
func NewRAGRetrieverAdapter(r rag.Retriever) Retriever {
	return &RAGRetrieverAdapter{retriever: r}
}

// Retrieve 实现 analyzer.Retriever 接口
func (a *RAGRetrieverAdapter) Retrieve(ctx context.Context, query string) ([]rag.KnowledgeChunk, error) {
	return a.retriever.Retrieve(ctx, query)
}
