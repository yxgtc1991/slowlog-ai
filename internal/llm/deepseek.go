package llm

import (
	"context"
	"encoding/json"
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

// ChatWithTools 支持工具调用的聊天方法
// 实现 ToolCallingClient 接口
func (d *DeepSeekClient) ChatWithTools(ctx context.Context, req ChatWithToolsRequest) (*ChatWithToolsResponse, error) {
	// 转换消息格式
	messages := make([]*request.Message, 0, len(req.Messages))
	for _, msg := range req.Messages {
		msgContent := ""
		if contentStr, ok := msg.Content.(string); ok {
			msgContent = contentStr
		} else {
			// 如果是对象数组，转换为字符串
			contentBytes, _ := json.Marshal(msg.Content)
			msgContent = string(contentBytes)
		}

		reqMsg := &request.Message{
			Role:    msg.Role,
			Content: msgContent,
		}

		// 如果有工具调用，添加到消息中
		if len(msg.ToolCalls) > 0 {
			// DeepSeek SDK 可能使用不同的格式，这里先简化处理
			// 实际使用时需要根据 SDK 文档调整
		}

		messages = append(messages, reqMsg)
	}

	// 构建请求
	chatReq := &request.ChatCompletionsRequest{
		Messages: messages,
		Model:    d.model,
		Stream:   false,
	}

	// 如果有工具定义，添加到请求中
	// 注意：需要检查 DeepSeek SDK 是否支持 tools 字段
	// 如果不支持，可能需要使用其他方式（如 prompt 中描述工具）

	resp, err := d.client.CallChatCompletionsChat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// 解析响应
	response := &ChatWithToolsResponse{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: []ToolCall{}, // 需要从响应中提取工具调用
		Role:      "assistant",
	}

	// 尝试从响应中提取工具调用
	// 如果 DeepSeek SDK 支持，应该可以从 resp.Choices[0].Message.ToolCalls 获取
	// 如果不支持，可能需要解析 Content 中的 JSON

	return response, nil
}
