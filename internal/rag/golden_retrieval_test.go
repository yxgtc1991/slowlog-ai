package rag

import (
	"context"
	"strings"
	"testing"
)

// 与 testdata/slowlog-products.txt 及知识库对齐的检索 golden（无 LLM）。
func TestGoldenRetrieval_productsScenario(t *testing.T) {
	t.Parallel()
	r, err := NewTFIDFRetriever(TFIDFOptions{TopK: 5})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		query       string
		wantInTitle string
	}{
		{
			query:       "price 复合索引 最左前缀 左列 code",
			wantInTitle: "最左前缀",
		},
		{
			query:       "ORDER BY created_at LIMIT filesort 排序",
			wantInTitle: "ORDER BY",
		},
		{
			query:       "EXPLAIN type ALL key 为空 全表扫描",
			wantInTitle: "ALL",
		},
		{
			query:       "products price 索引 created_at dry_run",
			wantInTitle: "products",
		},
		{
			query:       "Query_time Lock_time 锁等待",
			wantInTitle: "Query_time",
		},
		{
			query:       "Rows_sent Rows_examined 返回行数",
			wantInTitle: "Rows_sent",
		},
		{
			query:       "Using filesort 排序 Extra",
			wantInTitle: "filesort",
		},
		{
			query:       "Lock_time 锁竞争 长事务",
			wantInTitle: "锁竞争",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.wantInTitle, func(t *testing.T) {
			t.Parallel()
			chunks, err := r.Retrieve(context.Background(), tc.query)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, c := range chunks {
				if strings.Contains(c.Title, tc.wantInTitle) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("query=%q top=%v want title containing %q", tc.query, chunkTitles(chunks), tc.wantInTitle)
			}
		})
	}
}

func chunkTitles(chunks []KnowledgeChunk) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.Title
	}
	return out
}
