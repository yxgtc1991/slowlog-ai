// rag-check 本地验证 RAG 检索（不跑 LLM）。
package main

import (
	"ai_slow_log/internal/rag"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	compare := false
	if len(args) > 0 && (args[0] == "-compare" || args[0] == "compare") {
		compare = true
		args = args[1:]
	}
	query := "rows_examined 高 全表扫描 复合索引"
	if len(args) > 0 {
		query = strings.Join(args, " ")
	}

	if compare {
		_ = os.Setenv("SLOWLOG_EMBEDDING_PROVIDER", "local")
		for _, mode := range []string{"tfidf", "embedding"} {
			_ = os.Setenv("SLOWLOG_RAG", mode)
			fmt.Printf("=== SLOWLOG_RAG=%s ===\n", mode)
			runOnce(query)
			fmt.Println()
		}
		return
	}

	runOnce(query)
}

func runOnce(query string) {
	r, err := rag.NewDefaultRetriever()
	if err != nil {
		log.Fatal(err)
	}
	chunks, err := r.Retrieve(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	mode := strings.TrimSpace(os.Getenv("SLOWLOG_RAG"))
	if mode == "" {
		mode = "tfidf"
	}
	fmt.Printf("mode: %s\nquery: %s\n\n", mode, query)
	for i, c := range chunks {
		fmt.Printf("%d. [%.4f] %s (%s)\n   %s\n\n", i+1, c.Score, c.Title, c.Source, truncate(c.Content, 120))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
