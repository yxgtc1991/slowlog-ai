package rag

import "testing"

func TestMergeRRF_prefersConsensus(t *testing.T) {
	t.Parallel()
	a := KnowledgeChunk{Title: "A", Source: "pattern", Score: 1}
	b := KnowledgeChunk{Title: "B", Source: "pattern", Score: 1}
	c := KnowledgeChunk{Title: "C", Source: "metric", Score: 1}
	out := MergeRRF([][]KnowledgeChunk{
		{a, b, c},
		{a, c, b},
	}, 2, 60)
	if len(out) != 2 || out[0].Title != "A" {
		t.Fatalf("got %+v", out)
	}
}
