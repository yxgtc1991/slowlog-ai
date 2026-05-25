package rag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NewDefaultRetriever 默认 TF-IDF；SLOWLOG_RAG=mock|embedding|tfidf。
func NewDefaultRetriever() (Retriever, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SLOWLOG_RAG")))
	topK := ragTopK()
	switch mode {
	case "mock":
		return NewMockRetriever(), nil
	case "embedding", "embed", "vector":
		emb, err := NewEmbedderFromEnv()
		if err != nil {
			return nil, err
		}
		return newEmbeddingWithPersist(topK, emb)
	case "tfidf", "":
		return newTFIDFWithPersist(topK)
	default:
		return nil, fmt.Errorf("unknown SLOWLOG_RAG=%q (use tfidf, embedding, or mock)", mode)
	}
}

// MustDefaultRetriever NewDefaultRetriever 失败时 panic（main 用）。
func MustDefaultRetriever() Retriever {
	r, err := NewDefaultRetriever()
	if err != nil {
		panic(fmt.Sprintf("rag retriever: %v", err))
	}
	return r
}

func ragTopK() int {
	topK := 3
	if v := strings.TrimSpace(os.Getenv("SLOWLOG_RAG_TOPK")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topK = n
		}
	}
	return topK
}
