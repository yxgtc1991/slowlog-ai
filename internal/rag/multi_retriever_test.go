package rag

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultiPathRetriever_withSlowLog(t *testing.T) {
	t.Parallel()
	tf, err := NewTFIDFRetriever(TFIDFOptions{TopK: 5})
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMultiPathRetriever(MultiPathOptions{Backend: tf, TopK: 3, UseRRF: true})
	if err != nil {
		t.Fatal(err)
	}
	slow, err := os.ReadFile(filepath.Join("..", "..", "testdata", "slowlog-products.txt"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := ContextWithSlowLog(context.Background(), string(slow))
	// 泛化 rag_query + 慢日志改写后，应能命中 products / 索引相关 chunk
	chunks, err := m.Retrieve(ctx, "全表扫描 如何优化")
	if err != nil {
		t.Fatal(err)
	}
	wantSubs := []string{"products", "最左前缀", "price", "Rows_examined", "索引"}
	found := false
	for _, c := range chunks {
		for _, sub := range wantSubs {
			if strings.Contains(c.Title, sub) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("top=%v", chunkTitles(chunks))
	}
}
