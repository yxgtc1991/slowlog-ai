package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// Embedder 将文本映射为向量（用于内存 TopK，无需独立向量库）。
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// LocalEmbedder 确定性本地向量（测试 / 离线对比，不调 API）。
type LocalEmbedder struct {
	Dim int
}

func NewLocalEmbedder(dim int) *LocalEmbedder {
	if dim <= 0 {
		dim = 128
	}
	return &LocalEmbedder{Dim: dim}
}

func (e *LocalEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, text := range texts {
		vec := make([]float32, e.Dim)
		for term, w := range termFreq(tokenize(text)) {
			h := fnv.New32a()
			_, _ = h.Write([]byte(term))
			idx := int(h.Sum32() % uint32(e.Dim))
			vec[idx] += float32(w)
		}
		normalizeVec(vec)
		out[i] = vec
	}
	return out, nil
}

// HTTPEmbedder 调用 OpenAI 兼容的 /v1/embeddings。
type HTTPEmbedder struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewHTTPEmbedder(apiKey, baseURL, model string) *HTTPEmbedder {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &HTTPEmbedder{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (e *HTTPEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(map[string]any{
		"model": e.Model,
		"input": texts,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embeddings HTTP %d: %s", resp.StatusCode, truncateContent(string(raw), 200))
	}

	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	out := make([][]float32, len(texts))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(texts) {
			continue
		}
		vec := make([]float32, len(item.Embedding))
		for i, v := range item.Embedding {
			vec[i] = float32(v)
		}
		normalizeVec(vec)
		out[item.Index] = vec
	}
	for i, vec := range out {
		if vec == nil {
			return nil, fmt.Errorf("missing embedding for input[%d]", i)
		}
	}
	return out, nil
}

// NewEmbedderFromEnv 按 SLOWLOG_EMBEDDING_PROVIDER 选择实现（默认 local）。
func NewEmbedderFromEnv() (Embedder, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_PROVIDER")))
	if provider == "" {
		provider = "local"
	}
	switch provider {
	case "local":
		return NewLocalEmbedder(128), nil
	case "http", "openai", "api":
		key := strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_API_KEY"))
		if key == "" {
			key = strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
		}
		if key == "" {
			return nil, fmt.Errorf("SLOWLOG_EMBEDDING_PROVIDER=http requires API key")
		}
		base := strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_BASE_URL"))
		model := strings.TrimSpace(os.Getenv("SLOWLOG_EMBEDDING_MODEL"))
		return NewHTTPEmbedder(key, base, model), nil
	default:
		return nil, fmt.Errorf("unknown SLOWLOG_EMBEDDING_PROVIDER=%q", provider)
	}
}

func normalizeVec(v []float32) {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return
	}
	inv := float32(1.0 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot
}
