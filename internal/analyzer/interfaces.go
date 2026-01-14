package analyzer

import (
	"ai_slow_log/internal/rag"
	"context"
)

// LLMClient 抽象 LLM 能力
type LLMClient interface {
	Chat(ctx context.Context, prompt string) (string, error)
}

// Retriever 抽象 RAG 检索能力
type Retriever interface {
	Retrieve(ctx context.Context, query string) ([]rag.KnowledgeChunk, error)
}

// PromptBuilder 抽象 Prompt 构造
type PromptBuilder interface {
	Build(slowLog string, chunks []rag.KnowledgeChunk) string
}
