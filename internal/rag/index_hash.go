package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// ComputeEmbeddedDocsHash 对嵌入知识库全部 .md 内容做 SHA256（排序路径保证稳定）。
func ComputeEmbeddedDocsHash() (string, error) {
	h := sha256.New()
	var paths []string
	err := fs.WalkDir(slowlogDocsFS, "slowlog/docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !fs.ValidPath(path) {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, path := range paths {
		b, err := fs.ReadFile(slowlogDocsFS, path)
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func docsHashOrError() (string, error) {
	hash, err := ComputeEmbeddedDocsHash()
	if err != nil {
		return "", fmt.Errorf("docs hash: %w", err)
	}
	return hash, nil
}
