package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// EmbeddingRetriever 内存向量 TopK（启动时 embed 全部 chunk，查询时再 embed query）。
type EmbeddingRetriever struct {
	chunks  []KnowledgeChunk
	vectors [][]float32
	topK    int
	embedder Embedder
}

// EmbeddingOptions 向量检索配置。
type EmbeddingOptions struct {
	TopK     int
	Embedder Embedder
}

// NewEmbeddingRetriever 构建内存向量索引。
func NewEmbeddingRetriever(opts EmbeddingOptions) (*EmbeddingRetriever, error) {
	if opts.Embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = 3
	}
	chunks, err := loadKnowledgeChunks()
	if err != nil {
		return nil, err
	}
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Title + "\n" + c.Content
	}
	vecs, err := opts.Embedder.Embed(context.Background(), texts)
	if err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}
	return &EmbeddingRetriever{
		chunks:   chunks,
		vectors:  vecs,
		topK:     topK,
		embedder: opts.Embedder,
	}, nil
}

// Retrieve 按余弦相似度返回 TopK。
func (r *EmbeddingRetriever) Retrieve(ctx context.Context, query string) ([]KnowledgeChunk, error) {
	if r == nil || len(r.chunks) == 0 {
		return nil, fmt.Errorf("embedding retriever not initialized")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}
	qvecs, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	qv := qvecs[0]

	type ranked struct {
		score float64
		idx   int
	}
	rankedList := make([]ranked, 0, len(r.chunks))
	for i, dv := range r.vectors {
		s := cosineSimilarity(qv, dv)
		if s > 0 {
			rankedList = append(rankedList, ranked{score: s, idx: i})
		}
	}
	sort.Slice(rankedList, func(i, j int) bool {
		return rankedList[i].score > rankedList[j].score
	})
	k := r.topK
	if k > len(rankedList) {
		k = len(rankedList)
	}
	out := make([]KnowledgeChunk, 0, k)
	for i := 0; i < k; i++ {
		ch := r.chunks[rankedList[i].idx]
		ch.Score = rankedList[i].score
		out = append(out, ch)
	}
	return out, nil
}
