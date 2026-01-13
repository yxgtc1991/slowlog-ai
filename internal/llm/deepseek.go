package llm

import (
	"context"
	"fmt"

	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
)

type DeepSeekClient struct {
	client deepseek.Client
	model  string
}

// NewDeepSeekClient 创建 DeepSeek 客户端
// apiKey: DeepSeek API 密钥
// model: 模型名称，如果为空则使用默认模型
func NewDeepSeekClient(apiKey, model string) (*DeepSeekClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}

	if model == "" {
		model = deepseek.DEEPSEEK_CHAT_MODEL
	}

	c, err := deepseek.NewClient(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create deepseek client: %w", err)
	}
	return &DeepSeekClient{
		client: c,
		model:  model,
	}, nil
}

func (d *DeepSeekClient) Chat(ctx context.Context, prompt string) (string, error) {
	req := &request.ChatCompletionsRequest{
		Messages: []*request.Message{
			{
				Role:    request.RoleUser,
				Content: prompt,
			},
		},
		Model:  d.model,
		Stream: false,
	}

	resp, err := d.client.CallChatCompletionsChat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
