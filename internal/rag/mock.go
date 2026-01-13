package rag

import "context"

type MockRetriever struct{}

func NewMockRetriever() *MockRetriever {
	return &MockRetriever{}
}

func (m *MockRetriever) Retrieve(ctx context.Context, query string) ([]KnowledgeChunk, error) {
	return []KnowledgeChunk{
		{
			Title:   "Rows_examined >> Rows_sent",
			Content: "扫描行数远大于返回行数，通常意味着索引未被有效使用，或存在全表扫描。",
			Source:  "pattern",
		},
		{
			Title:   "ORDER BY + LIMIT",
			Content: "即使有 LIMIT，若 ORDER BY 字段未命中索引，仍可能触发 filesort。",
			Source:  "anti-pattern",
		},
	}, nil
}
