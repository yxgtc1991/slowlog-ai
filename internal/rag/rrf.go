package rag

import (
	"sort"
	"strings"
)

const defaultRRFK = 60.0

// MergeRRF 将多路 TopK 列表按 Reciprocal Rank Fusion 合并。
func MergeRRF(lists [][]KnowledgeChunk, topK int, k float64) []KnowledgeChunk {
	if k <= 0 {
		k = defaultRRFK
	}
	if topK <= 0 {
		topK = 3
	}
	if len(lists) == 0 {
		return nil
	}
	if len(lists) == 1 {
		out := lists[0]
		if len(out) > topK {
			out = out[:topK]
		}
		return out
	}

	type acc struct {
		chunk KnowledgeChunk
		score float64
	}
	byKey := map[string]*acc{}

	for _, list := range lists {
		for rank, ch := range list {
			key := chunkKey(ch)
			a, ok := byKey[key]
			if !ok {
				cp := ch
				cp.Score = 0
				byKey[key] = &acc{chunk: cp}
				a = byKey[key]
			}
			a.score += 1.0 / (k + float64(rank+1))
		}
	}

	merged := make([]acc, 0, len(byKey))
	for _, a := range byKey {
		a.chunk.Score = a.score
		merged = append(merged, *a)
	}
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].score > merged[j].score
	})
	if len(merged) > topK {
		merged = merged[:topK]
	}
	out := make([]KnowledgeChunk, len(merged))
	for i := range merged {
		out[i] = merged[i].chunk
	}
	return out
}

func chunkKey(ch KnowledgeChunk) string {
	return strings.ToLower(strings.TrimSpace(ch.Title)) + "|" + strings.TrimSpace(ch.Source)
}
