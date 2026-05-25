// rag-index 构建/查看磁盘 RAG 索引（G13）。
package main

import (
	"ai_slow_log/internal/rag"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	buildEmb := flag.Bool("embedding", false, "同时构建 embedding 索引（需 embedder）")
	statusOnly := flag.Bool("status", false, "仅打印索引状态")
	dir := flag.String("dir", "", "索引目录（默认 data/rag-index 或 SLOWLOG_RAG_INDEX_DIR）")
	flag.Parse()

	indexDir := rag.IndexDirOr(*dir)

	if *statusOnly {
		st, err := rag.GetIndexStatus(indexDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "status: %v\n", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(st)
		return
	}

	if err := rag.BuildAllIndexes(indexDir, *buildEmb); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		os.Exit(1)
	}
	st, _ := rag.GetIndexStatus(indexDir)
	fmt.Println("RAG index built:", indexDir)
	fmt.Printf("  chunks=%d up_to_date=%v tfidf_on_disk=%v embedding_on_disk=%v\n",
		st.Manifest.ChunkCount, st.UpToDate, st.TFIDFOnDisk, st.EmbeddingOnDisk)
}
