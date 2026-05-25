package rag

import (
	"slices"
	"testing"
)

func TestTokenize_domainPhrases(t *testing.T) {
	t.Parallel()
	toks := tokenize("price 最左前缀 rows_examined 高")
	if !slices.Contains(toks, "最左前缀") {
		t.Fatalf("tokens=%v want 最左前缀 phrase", toks)
	}
	if !slices.Contains(toks, "rows_examined") {
		t.Fatalf("tokens=%v want rows_examined", toks)
	}
}

func TestTokenize_englishWords(t *testing.T) {
	t.Parallel()
	toks := tokenize("ORDER BY LIMIT filesort")
	for _, want := range []string{"order", "limit", "filesort"} {
		if !slices.Contains(toks, want) {
			t.Fatalf("tokens=%v missing %q", toks, want)
		}
	}
}

func TestTermFreq_normalized(t *testing.T) {
	t.Parallel()
	tf := termFreq([]string{"a", "a", "b"})
	if tf["a"] <= tf["b"] {
		t.Fatalf("tf=%v", tf)
	}
	sum := 0.0
	for _, v := range tf {
		sum += v
	}
	if sum < 0.99 || sum > 1.01 {
		t.Fatalf("sum=%v want ~1", sum)
	}
}
