package rag

import (
	"context"
	"fmt"
)

// DualPathOptions TF-IDF + Embedding 双路召回（均对改写后的 queries 检索后 RRF）。
type DualPathOptions struct {
	TFIDF     Retriever
	Embedding Retriever
	TopK      int
}

// DualPathRetriever G16 可选：稀疏 + 向量两路融合。
type DualPathRetriever struct {
	tfidf     Retriever
	embedding Retriever
	topK      int
}

func NewDualPathRetriever(opts DualPathOptions) (*DualPathRetriever, error) {
	if opts.TFIDF == nil || opts.Embedding == nil {
		return nil, fmt.Errorf("dual-path retriever: tfidf and embedding required")
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = 3
	}
	return &DualPathRetriever{tfidf: opts.TFIDF, embedding: opts.Embedding, topK: topK}, nil
}

func (d *DualPathRetriever) Retrieve(ctx context.Context, query string) ([]KnowledgeChunk, error) {
	if d == nil {
		return nil, fmt.Errorf("dual-path retriever not initialized")
	}
	queries := RewriteQueries(query, SlowLogFromContext(ctx))
	var lists [][]KnowledgeChunk
	for _, q := range queries {
		tf, err := d.tfidf.Retrieve(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("tfidf %q: %w", q, err)
		}
		if len(tf) > 0 {
			lists = append(lists, tf)
		}
		em, err := d.embedding.Retrieve(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("embedding %q: %w", q, err)
		}
		if len(em) > 0 {
			lists = append(lists, em)
		}
	}
	if len(lists) == 0 {
		return nil, nil
	}
	return MergeRRF(lists, d.topK, defaultRRFK), nil
}
