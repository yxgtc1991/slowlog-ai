package rag

import "testing"

func TestLoadKnowledgeChunks_splitBySection(t *testing.T) {
	t.Parallel()
	chunks, err := loadKnowledgeChunks()
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 10 {
		t.Fatalf("chunks=%d want >=10 after ## split", len(chunks))
	}
	found := false
	for _, c := range chunks {
		if c.Title != "" && c.Content != "" && c.Source == "pattern" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected non-empty pattern chunk")
	}
}
