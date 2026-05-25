package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const indexFormatVersion = 1

// IndexManifest 持久化索引元数据。
type IndexManifest struct {
	Version    int       `json:"version"`
	Mode       string    `json:"mode"`
	DocsHash   string    `json:"docs_hash"`
	ChunkCount int       `json:"chunk_count"`
	BuiltAt    time.Time `json:"built_at"`
	EmbedModel string    `json:"embed_model,omitempty"`
}

type serializedChunkTF struct {
	Chunk KnowledgeChunk   `json:"chunk"`
	TF    map[string]float64 `json:"tf"`
}

type tfidfIndexFile struct {
	Manifest IndexManifest       `json:"manifest"`
	Chunks   []serializedChunkTF `json:"chunks"`
}

type embeddingIndexFile struct {
	Manifest IndexManifest    `json:"manifest"`
	Chunks   []KnowledgeChunk `json:"chunks"`
	Vectors  [][]float32      `json:"vectors"`
}

func tfidfIndexPath(dir string) string {
	return filepath.Join(dir, "tfidf-index.json")
}

func embeddingIndexPath(dir string) string {
	return filepath.Join(dir, "embedding-index.json")
}

// SaveTFIDFIndex 将 TF-IDF 索引写入目录。
func SaveTFIDFIndex(dir string, data *tfidfIndexData) error {
	if data == nil {
		return fmt.Errorf("nil index data")
	}
	hash, err := docsHashOrError()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := tfidfIndexFile{
		Manifest: IndexManifest{
			Version:    indexFormatVersion,
			Mode:       "tfidf",
			DocsHash:   hash,
			ChunkCount: len(data.chunks),
			BuiltAt:    time.Now(),
		},
		Chunks: make([]serializedChunkTF, len(data.chunks)),
	}
	for i, c := range data.chunks {
		out.Chunks[i] = serializedChunkTF{Chunk: c.chunk, TF: c.tf}
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tfidfIndexPath(dir), b, 0o644)
}

// LoadTFIDFIndex 从目录加载；第二返回值 false 表示需重建。
func LoadTFIDFIndex(dir string, topK int) (*TFIDFRetriever, bool, error) {
	b, err := os.ReadFile(tfidfIndexPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var file tfidfIndexFile
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, false, err
	}
	currentHash, err := docsHashOrError()
	if err != nil {
		return nil, false, err
	}
	if file.Manifest.DocsHash != currentHash {
		return nil, false, nil
	}
	chunks := make([]scoredChunk, len(file.Chunks))
	for i, c := range file.Chunks {
		chunks[i] = scoredChunk{chunk: c.Chunk, tf: c.TF}
	}
	return (&tfidfIndexData{chunks: chunks}).retriever(topK), true, nil
}

// SaveEmbeddingIndex 持久化向量索引。
func SaveEmbeddingIndex(dir string, chunks []KnowledgeChunk, vectors [][]float32, model string) error {
	hash, err := docsHashOrError()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := embeddingIndexFile{
		Manifest: IndexManifest{
			Version:    indexFormatVersion,
			Mode:       "embedding",
			DocsHash:   hash,
			ChunkCount: len(chunks),
			BuiltAt:    time.Now(),
			EmbedModel: model,
		},
		Chunks:  chunks,
		Vectors: vectors,
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(embeddingIndexPath(dir), b, 0o644)
}

// LoadEmbeddingIndex 加载向量索引；model 须与构建时一致。
func LoadEmbeddingIndex(dir string, topK int, embedder Embedder, expectModel string) (*EmbeddingRetriever, bool, error) {
	b, err := os.ReadFile(embeddingIndexPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var file embeddingIndexFile
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, false, err
	}
	currentHash, err := docsHashOrError()
	if err != nil {
		return nil, false, err
	}
	if file.Manifest.DocsHash != currentHash {
		return nil, false, nil
	}
	if expectModel != "" && file.Manifest.EmbedModel != "" && file.Manifest.EmbedModel != expectModel {
		return nil, false, nil
	}
	if len(file.Chunks) != len(file.Vectors) {
		return nil, false, fmt.Errorf("embedding index corrupt: chunk/vector count mismatch")
	}
	if topK <= 0 {
		topK = 3
	}
	return &EmbeddingRetriever{
		chunks:   file.Chunks,
		vectors:  file.Vectors,
		topK:     topK,
		embedder: embedder,
	}, true, nil
}

// ReadIndexManifest 读取已存在的 manifest（tfidf 优先）。
func ReadIndexManifest(dir string) (*IndexManifest, error) {
	for _, p := range []string{tfidfIndexPath(dir), embeddingIndexPath(dir)} {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var m struct {
			Manifest IndexManifest `json:"manifest"`
		}
		if json.Unmarshal(b, &m) == nil {
			return &m.Manifest, nil
		}
	}
	return nil, os.ErrNotExist
}

// IndexStatus 当前索引与嵌入知识库是否一致。
type IndexStatus struct {
	IndexDir       string `json:"index_dir"`
	CurrentDocsHash string `json:"current_docs_hash"`
	StoredDocsHash  string `json:"stored_docs_hash,omitempty"`
	UpToDate       bool   `json:"up_to_date"`
	TFIDFOnDisk    bool   `json:"tfidf_on_disk"`
	EmbeddingOnDisk bool  `json:"embedding_on_disk"`
	Manifest       *IndexManifest `json:"manifest,omitempty"`
}

func GetIndexStatus(dir string) (IndexStatus, error) {
	dir = IndexDirOr(dir)
	current, err := docsHashOrError()
	if err != nil {
		return IndexStatus{}, err
	}
	st := IndexStatus{
		IndexDir:        dir,
		CurrentDocsHash: current,
		TFIDFOnDisk:     fileExists(tfidfIndexPath(dir)),
		EmbeddingOnDisk: fileExists(embeddingIndexPath(dir)),
	}
	if m, err := ReadIndexManifest(dir); err == nil {
		st.Manifest = m
		st.StoredDocsHash = m.DocsHash
		st.UpToDate = m.DocsHash == current
	}
	return st, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
