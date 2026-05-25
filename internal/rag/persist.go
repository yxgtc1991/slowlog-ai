package rag

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// IndexDirOr 解析索引目录。
func IndexDirOr(dir string) string {
	if strings.TrimSpace(dir) != "" {
		return dir
	}
	if v := strings.TrimSpace(os.Getenv("SLOWLOG_RAG_INDEX_DIR")); v != "" {
		return v
	}
	return "data/rag-index"
}

// PersistEnabled 是否优先加载/写入磁盘索引。
func PersistEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SLOWLOG_RAG_PERSIST"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// BuildAllIndexes 构建并保存 TF-IDF（必选）与 embedding（可选）索引。
func BuildAllIndexes(dir string, withEmbedding bool) error {
	dir = IndexDirOr(dir)
	idx, err := buildTFIDFIndex()
	if err != nil {
		return err
	}
	if err := SaveTFIDFIndex(dir, idx); err != nil {
		return fmt.Errorf("save tfidf: %w", err)
	}
	if !withEmbedding {
		return nil
	}
	emb, err := NewEmbedderFromEnv()
	if err != nil {
		return fmt.Errorf("embedding build: %w", err)
	}
	chunks, err := loadKnowledgeChunks()
	if err != nil {
		return err
	}
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Title + "\n" + c.Content
	}
	vecs, err := emb.Embed(context.Background(), texts)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	model := strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_MODEL"))
	if model == "" {
		model = "local"
	}
	return SaveEmbeddingIndex(dir, chunks, vecs, model)
}

func newTFIDFWithPersist(topK int) (*TFIDFRetriever, error) {
	dir := IndexDirOr("")
	if PersistEnabled() || fileExists(tfidfIndexPath(dir)) {
		if r, ok, err := LoadTFIDFIndex(dir, topK); err != nil {
			return nil, err
		} else if ok {
			return r, nil
		}
	}
	idx, err := buildTFIDFIndex()
	if err != nil {
		return nil, err
	}
	if PersistEnabled() {
		_ = SaveTFIDFIndex(dir, idx)
	}
	return idx.retriever(topK), nil
}

func newEmbeddingWithPersist(topK int, emb Embedder) (*EmbeddingRetriever, error) {
	dir := IndexDirOr("")
	model := strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_MODEL"))
	if model == "" {
		model = "local"
	}
	if PersistEnabled() || fileExists(embeddingIndexPath(dir)) {
		if r, ok, err := LoadEmbeddingIndex(dir, topK, emb, model); err != nil {
			return nil, err
		} else if ok {
			return r, nil
		}
	}
	r, err := NewEmbeddingRetriever(EmbeddingOptions{TopK: topK, Embedder: emb})
	if err != nil {
		return nil, err
	}
	if PersistEnabled() {
		chunks, vectors := r.Snapshot()
		_ = SaveEmbeddingIndex(dir, chunks, vectors, model)
	}
	return r, nil
}
