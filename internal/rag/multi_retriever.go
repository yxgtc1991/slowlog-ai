package rag

import (
	"context"
	"fmt"
)

// MultiPathOptions 多路召回 + 可选 RRF。
type MultiPathOptions struct {
	Backend Retriever
	TopK    int
	UseRRF  bool
}

// MultiPathRetriever 对改写后的多条 query 分别检索并融合（G16）。
type MultiPathRetriever struct {
	backend Retriever
	topK    int
	useRRF  bool
}

func NewMultiPathRetriever(opts MultiPathOptions) (*MultiPathRetriever, error) {
	if opts.Backend == nil {
		return nil, fmt.Errorf("multi-path retriever: backend is required")
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = 3
	}
	return &MultiPathRetriever{
		backend: opts.Backend,
		topK:    topK,
		useRRF:  opts.UseRRF,
	}, nil
}

func (m *MultiPathRetriever) Retrieve(ctx context.Context, query string) ([]KnowledgeChunk, error) {
	if m == nil || m.backend == nil {
		return nil, fmt.Errorf("multi-path retriever not initialized")
	}
	queries := RewriteQueries(query, SlowLogFromContext(ctx))
	if len(queries) == 0 {
		return nil, fmt.Errorf("empty query")
	}

	lists := make([][]KnowledgeChunk, 0, len(queries))
	for _, q := range queries {
		chunks, err := m.backend.Retrieve(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("retrieve %q: %w", q, err)
		}
		if len(chunks) > 0 {
			lists = append(lists, chunks)
		}
	}
	if len(lists) == 0 {
		return nil, nil
	}
	if m.useRRF && len(lists) > 1 {
		return MergeRRF(lists, m.topK, defaultRRFK), nil
	}
	// 单路或未开 RRF：取第一条非空结果并截断 topK
	best := lists[0]
	for _, l := range lists[1:] {
		if len(l) > len(best) {
			best = l
		}
	}
	if len(best) > m.topK {
		best = best[:m.topK]
	}
	return best, nil
}

// QueriesForDebug 返回本次将使用的检索 query（供 rag-check 打印）。
func (m *MultiPathRetriever) QueriesForDebug(ctx context.Context, query string) []string {
	return RewriteQueries(query, SlowLogFromContext(ctx))
}
