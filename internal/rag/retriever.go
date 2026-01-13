package rag

import "context"

// KnowledgeChunk 知识块，包含从 RAG 检索到的相关信息
type KnowledgeChunk struct {
	Title   string  // 知识块标题
	Content string  // 知识块内容
	Source  string  // 来源类型：pattern / anti-pattern / metric / action / boundary
	Score   float64 // 相关性分数（可选，用于排序）
}

// Retriever RAG 检索器接口
// 根据查询字符串检索相关的知识块
type Retriever interface {
	Retrieve(ctx context.Context, query string) ([]KnowledgeChunk, error)
}
