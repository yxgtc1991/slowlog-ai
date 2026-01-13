package config

import (
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	// DeepSeek API 配置
	DeepSeekAPIKey string
	DeepSeekModel  string

	// 可选：其他配置项
	LogLevel string
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
	}

	model := os.Getenv("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-chat" // 默认模型
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		DeepSeekAPIKey: apiKey,
		DeepSeekModel:  model,
		LogLevel:       logLevel,
	}, nil
}
