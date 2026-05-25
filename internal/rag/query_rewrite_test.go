package rag

import (
	"strings"
	"testing"
)

func TestRewriteQueries_slowLogHints(t *testing.T) {
	t.Parallel()
	slow := `# Rows_examined: 48000
SELECT * FROM products WHERE price >= 100 ORDER BY created_at`
	qs := RewriteQueries("如何优化", slow)
	if len(qs) < 2 {
		t.Fatalf("want multiple queries, got %v", qs)
	}
	joined := strings.Join(qs, " ")
	for _, want := range []string{"products", "Rows_examined", "如何优化"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, qs)
		}
	}
}

func TestRewriteQueries_dedup(t *testing.T) {
	t.Parallel()
	qs := RewriteQueries("price 索引", "FROM products WHERE price")
	seen := map[string]struct{}{}
	for _, q := range qs {
		k := strings.ToLower(q)
		if _, ok := seen[k]; ok {
			t.Fatalf("duplicate %q in %v", q, qs)
		}
		seen[k] = struct{}{}
	}
}
