package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
)

// TFIDFRetriever 基于知识库 Markdown 的 TF-IDF 检索（无向量库依赖）。
type TFIDFRetriever struct {
	chunks []scoredChunk
	topK   int
}

type scoredChunk struct {
	chunk KnowledgeChunk
	tf    map[string]float64
}

// TFIDFOptions 检索器配置。
type TFIDFOptions struct {
	TopK int
}

// NewTFIDFRetriever 从嵌入的 slowlog/docs 构建索引（按 ## 切 chunk）。
func NewTFIDFRetriever(opts TFIDFOptions) (*TFIDFRetriever, error) {
	topK := opts.TopK
	if topK <= 0 {
		topK = 3
	}
	raw, err := loadKnowledgeChunks()
	if err != nil {
		return nil, err
	}
	df := make(map[string]int)
	chunks := make([]scoredChunk, 0, len(raw))
	for _, doc := range raw {
		tokens := tokenize(doc.Title + " " + doc.Content)
		tf := termFreq(tokens)
		for term := range tf {
			df[term]++
		}
		chunks = append(chunks, scoredChunk{
			chunk: doc,
			tf:    tf,
		})
	}
	nDocs := float64(len(chunks))
	for i := range chunks {
		for term, tf := range chunks[i].tf {
			idf := 1.0 + math.Log2((nDocs+1.0)/(float64(df[term])+1.0))
			chunks[i].tf[term] = tf * idf
		}
	}
	return &TFIDFRetriever{chunks: chunks, topK: topK}, nil
}

// Retrieve 按查询返回 TopK 知识块（按相关度排序）。
func (r *TFIDFRetriever) Retrieve(_ context.Context, query string) ([]KnowledgeChunk, error) {
	if r == nil || len(r.chunks) == 0 {
		return nil, fmt.Errorf("tfidf retriever not initialized")
	}
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil, fmt.Errorf("empty query")
	}
	qtf := termFreq(qTokens)

	type ranked struct {
		score float64
		idx   int
	}
	rankedList := make([]ranked, 0, len(r.chunks))
	for i, c := range r.chunks {
		s := dotProduct(qtf, c.tf)
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
		ch := r.chunks[rankedList[i].idx].chunk
		ch.Score = rankedList[i].score
		out = append(out, ch)
	}
	return out, nil
}

func dotProduct(a, b map[string]float64) float64 {
	var s float64
	for term, w := range a {
		if bw, ok := b[term]; ok {
			s += w * bw
		}
	}
	return s
}
