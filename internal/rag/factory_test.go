package rag

import (
	"testing"
)

func TestNewDefaultRetriever_tfidfDefault(t *testing.T) {
	t.Setenv("SLOWLOG_RAG", "")
	t.Setenv("SLOWLOG_EMBEDDING_PROVIDER", "")
	r, err := NewDefaultRetriever()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*TFIDFRetriever); !ok {
		t.Fatalf("type %T", r)
	}
}

func TestNewDefaultRetriever_unknownMode(t *testing.T) {
	t.Setenv("SLOWLOG_RAG", "unknown-mode")
	_, err := NewDefaultRetriever()
	if err == nil {
		t.Fatal("expected error")
	}
}
