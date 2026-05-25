package rag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NewDefaultRetriever 默认 TF-IDF；环境变量 SLOWLOG_RAG=mock 时用固定 Mock（如 eval）。
func NewDefaultRetriever() (Retriever, error) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SLOWLOG_RAG")), "mock") {
		return NewMockRetriever(), nil
	}
	topK := 3
	if v := strings.TrimSpace(os.Getenv("SLOWLOG_RAG_TOPK")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topK = n
		}
	}
	return NewTFIDFRetriever(TFIDFOptions{TopK: topK})
}

// MustDefaultRetriever NewDefaultRetriever 失败时 panic（main 用）。
func MustDefaultRetriever() Retriever {
	r, err := NewDefaultRetriever()
	if err != nil {
		panic(fmt.Sprintf("rag retriever: %v", err))
	}
	return r
}
