package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
)

type DeepSeekClient struct {
	client deepseek.Client
}

func NewDeepSeekClient() (*DeepSeekClient, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is not set")
	}

	c, err := deepseek.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	return &DeepSeekClient{client: c}, nil
}

func (d *DeepSeekClient) Chat(ctx context.Context, prompt string) (string, error) {
	req := &request.ChatCompletionsRequest{
		Messages: []*request.Message{
			{
				Role:    request.RoleUser,
				Content: prompt,
			},
		},
		Model:  deepseek.DEEPSEEK_CHAT_MODEL,
		Stream: false,
	}

	resp, err := d.client.CallChatCompletionsChat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
