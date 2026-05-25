// rag-check 本地验证 TF-IDF 检索（不跑 LLM）。
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
	query := "rows_examined 高 全表扫描 复合索引"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}
	r, err := rag.NewDefaultRetriever()
	if err != nil {
		log.Fatal(err)
	}
	chunks, err := r.Retrieve(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("query: %s\n\n", query)
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
