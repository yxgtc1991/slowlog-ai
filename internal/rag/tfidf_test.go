package rag

import (
	"context"
	"strings"
	"testing"
)

func TestTFIDFRetriever_rowsExaminedQuery(t *testing.T) {
	t.Parallel()
	r, err := NewTFIDFRetriever(TFIDFOptions{TopK: 3})
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
	if !strings.Contains(chunks[0].Title, "Rows_examined") {
		t.Fatalf("top=%q want Rows_examined related", chunks[0].Title)
	}
	if chunks[0].Score <= 0 {
		t.Fatalf("score=%v", chunks[0].Score)
	}
}

func TestTFIDFRetriever_limitQuery(t *testing.T) {
	t.Parallel()
	r, err := NewTFIDFRetriever(TFIDFOptions{TopK: 2})
	if err != nil {
		t.Fatal(err)
	}
	chunks, err := r.Retrieve(context.Background(), "LIMIT ORDER BY filesort")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range chunks {
		if strings.Contains(c.Title, "LIMIT") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("chunks=%v", chunks)
	}
}

func TestNewDefaultRetriever_mockEnv(t *testing.T) {
	t.Setenv("SLOWLOG_RAG", "mock")
	r, err := NewDefaultRetriever()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*MockRetriever); !ok {
		t.Fatalf("type %T", r)
	}
}
