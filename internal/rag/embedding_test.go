package rag

import (
	"context"
	"strings"
	"testing"
)

func TestEmbeddingRetriever_rowsExaminedQuery(t *testing.T) {
	t.Parallel()
	r, err := NewEmbeddingRetriever(EmbeddingOptions{
		TopK:     3,
		Embedder: NewLocalEmbedder(128),
	})
	if err != nil {
		t.Fatal(err)
	}
	chunks, err := r.Retrieve(context.Background(), "rows_examined 高 全表扫描 索引")
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Fatal("no chunks")
	}
	found := false
	for _, c := range chunks {
		if strings.Contains(c.Title, "Rows_examined") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("chunks=%v", chunks)
	}
}

func TestNewDefaultRetriever_embeddingEnv(t *testing.T) {
	t.Setenv("SLOWLOG_RAG", "embedding")
	t.Setenv("SLOWLOG_EMBEDDING_PROVIDER", "local")
	r, err := NewDefaultRetriever()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*EmbeddingRetriever); !ok {
		t.Fatalf("type %T", r)
	}
}
