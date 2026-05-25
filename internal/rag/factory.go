package rag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// NewDefaultRetriever 默认 TF-IDF；SLOWLOG_RAG=mock|embedding|tfidf。
// G16：SLOWLOG_RAG_MULTI=1 时多 query + RRF；SLOWLOG_RAG_DUAL=1 时 TF-IDF+embedding 双路 RRF。
func NewDefaultRetriever() (Retriever, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SLOWLOG_RAG")))
	topK := ragTopK()
	var base Retriever
	var err error
	switch mode {
	case "mock":
		base = NewMockRetriever()
	case "embedding", "embed", "vector":
		emb, e := NewEmbedderFromEnv()
		if e != nil {
			return nil, e
		}
		base, err = newEmbeddingWithPersist(topK, emb)
	case "tfidf", "":
		base, err = newTFIDFWithPersist(topK)
	default:
		return nil, fmt.Errorf("unknown SLOWLOG_RAG=%q (use tfidf, embedding, or mock)", mode)
	}
	if err != nil {
		return nil, err
	}
	return wrapG16(base, mode, topK)
}

func wrapG16(base Retriever, mode string, topK int) (Retriever, error) {
	if mode == "mock" {
		return base, nil
	}
	if envBool("SLOWLOG_RAG_DUAL") {
		tf, err := newTFIDFWithPersist(topK)
		if err != nil {
			return nil, err
		}
		emb, err := NewEmbedderFromEnv()
		if err != nil {
			return nil, err
		}
		vec, err := newEmbeddingWithPersist(topK, emb)
		if err != nil {
			return nil, err
		}
		return NewDualPathRetriever(DualPathOptions{TFIDF: tf, Embedding: vec, TopK: topK})
	}
	if envBool("SLOWLOG_RAG_MULTI") {
		useRRF := true
		if v := strings.TrimSpace(os.Getenv("SLOWLOG_RAG_RRF")); v == "0" || strings.EqualFold(v, "false") {
			useRRF = false
		}
		return NewMultiPathRetriever(MultiPathOptions{
			Backend: base,
			TopK:    topK,
			UseRRF:  useRRF,
		})
	}
	return base, nil
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
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
