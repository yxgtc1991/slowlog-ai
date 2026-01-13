package slowlog

import (
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"context"
	"fmt"
	"sync"
)

var (
	clientInstance *llm.DeepSeekClient
	clientOnce     sync.Once
	clientErr      error
)

// getClient 获取单例的 LLM 客户端
func getClient() (*llm.DeepSeekClient, error) {
	clientOnce.Do(func() {
		cfg, err := config.Load()
		if err != nil {
			clientErr = fmt.Errorf("failed to load config: %w", err)
			return
		}
		clientInstance, clientErr = llm.NewDeepSeekClient(cfg.DeepSeekAPIKey, cfg.DeepSeekModel)
		if clientErr != nil {
			clientErr = fmt.Errorf("failed to create LLM client: %w", clientErr)
		}
	})
	return clientInstance, clientErr
}

// AnalyzeSlowLog 分析慢日志并返回优化建议
// prompt: 构造好的 prompt 文本
func AnalyzeSlowLog(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	client, err := getClient()
	if err != nil {
		return "", fmt.Errorf("failed to get LLM client: %w", err)
	}

	result, err := client.Chat(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to chat with LLM: %w", err)
	}

	return result, nil
}
