package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// DeepSeekToolCallingAdapter 将 DeepSeekClient 适配为 ToolCallingClient
// 注意：如果 DeepSeek SDK 不支持原生 Tool Calling，这个适配器会使用 prompt 方式模拟
type DeepSeekToolCallingAdapter struct {
	client *DeepSeekClient
}

// NewDeepSeekToolCallingAdapter 创建适配器
func NewDeepSeekToolCallingAdapter(client *DeepSeekClient) ToolCallingClient {
	return &DeepSeekToolCallingAdapter{client: client}
}

// ChatWithTools 实现 ToolCallingClient 接口
// 如果 DeepSeek SDK 支持原生 Tool Calling，直接使用
// 否则，使用 prompt 方式让 LLM 输出工具调用意图，然后解析
func (a *DeepSeekToolCallingAdapter) ChatWithTools(ctx context.Context, req ChatWithToolsRequest) (*ChatWithToolsResponse, error) {
	// 检查 DeepSeekClient 是否实现了 ChatWithTools 方法
	// 如果实现了，直接使用原生 Tool Calling
	if a.client != nil {
		// 尝试调用原生方法（如果存在）
		// 注意：需要 DeepSeekClient 实现 ChatWithTools 方法
		// 目前先使用 prompt 方式
	}

	// 使用 prompt 方式模拟 Tool Calling
	return a.chatWithToolsViaPrompt(ctx, req)
}

// chatWithToolsViaPrompt 通过 prompt 方式模拟 Tool Calling
func (a *DeepSeekToolCallingAdapter) chatWithToolsViaPrompt(ctx context.Context, req ChatWithToolsRequest) (*ChatWithToolsResponse, error) {
	// 构建包含工具定义的 prompt
	prompt := a.buildToolCallingPrompt(req)

	// 调用普通 Chat 方法
	content, err := a.client.Chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 尝试从响应中解析工具调用
	toolCalls := a.parseToolCallsFromResponse(content, req.Tools)

	return &ChatWithToolsResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Role:      "assistant",
	}, nil
}

// buildToolCallingPrompt 构建包含工具定义的 prompt
func (a *DeepSeekToolCallingAdapter) buildToolCallingPrompt(req ChatWithToolsRequest) string {
	var sb strings.Builder

	// 添加工具定义
	if len(req.Tools) > 0 {
		sb.WriteString("【可用工具 - 必须使用】\n")
		sb.WriteString("**⚠️ 重要：要完成分析任务，你必须调用工具，不要直接输出分析结果！**\n")
		sb.WriteString("**⚠️ 第一轮响应必须包含工具调用，不能只有文本！**\n\n")
		sb.WriteString("如果需要调用工具，请严格按照以下 JSON 格式输出（必须包含在 ```json 代码块中）：\n\n")
		sb.WriteString("```json\n")
		sb.WriteString("{\n")
		sb.WriteString("  \"tool_calls\": [\n")
		sb.WriteString("    {\n")
		sb.WriteString("      \"id\": \"call_xxx\",\n")
		sb.WriteString("      \"type\": \"function\",\n")
		sb.WriteString("      \"function\": {\n")
		sb.WriteString("        \"name\": \"工具名称\",\n")
		sb.WriteString("        \"arguments\": {\"参数名\": \"参数值\"}\n")
		sb.WriteString("      }\n")
		sb.WriteString("    }\n")
		sb.WriteString("  ]\n")
		sb.WriteString("}\n")
		sb.WriteString("```\n\n")

		for _, tool := range req.Tools {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Description))
			if params, ok := tool.Parameters["properties"].(map[string]interface{}); ok {
				sb.WriteString("  参数：\n")
				for paramName, paramInfo := range params {
					if paramMap, ok := paramInfo.(map[string]interface{}); ok {
						paramType := paramMap["type"]
						paramDesc := paramMap["description"]
						sb.WriteString(fmt.Sprintf("    - %s (%s): %s\n", paramName, paramType, paramDesc))
					}
				}
			}
			sb.WriteString("\n")
		}
	}

	// 添加消息历史
	for _, msg := range req.Messages {
		if msg.Role == "tool" {
			// Tool 消息：工具执行结果
			sb.WriteString(fmt.Sprintf("【工具执行结果】\n"))
			if contentStr, ok := msg.Content.(string); ok {
				sb.WriteString(contentStr)
			} else {
				contentBytes, _ := json.Marshal(msg.Content)
				sb.WriteString(string(contentBytes))
			}
		} else if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			// Assistant 消息：包含工具调用
			sb.WriteString(fmt.Sprintf("【助手 - 已调用工具】\n"))
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("- 调用工具: %s\n", tc.Name))
			}
			if contentStr, ok := msg.Content.(string); ok && contentStr != "" {
				sb.WriteString(fmt.Sprintf("说明: %s\n", contentStr))
			}
		} else {
			// User 或其他消息
			sb.WriteString(fmt.Sprintf("【%s】\n", msg.Role))
			if contentStr, ok := msg.Content.(string); ok {
				sb.WriteString(contentStr)
			} else {
				contentBytes, _ := json.Marshal(msg.Content)
				sb.WriteString(string(contentBytes))
			}
		}
		sb.WriteString("\n\n")
	}

	// 如果是第一轮且没有工具调用历史，强制要求调用工具
	if len(req.Messages) == 1 && req.Messages[0].Role == "user" {
		sb.WriteString("【重要提示】\n")
		sb.WriteString("你必须调用工具来完成分析任务。请输出工具调用的 JSON 格式，不要直接输出分析结果。\n")
	}

	return sb.String()
}

// parseToolCallsFromResponse 从响应中解析工具调用
func (a *DeepSeekToolCallingAdapter) parseToolCallsFromResponse(content string, tools []ToolDefinition) []ToolCall {
	// 方法1：尝试从 JSON 代码块中解析
	start := strings.Index(content, "```json")
	if start != -1 {
		start += 7 // 跳过 ```json
		// 跳过空白
		for start < len(content) && (content[start] == ' ' || content[start] == '\n' || content[start] == '\r') {
			start++
		}
		// 查找结束标记
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr := strings.TrimSpace(content[start : start+end])
			if toolCalls := a.parseToolCallsJSON(jsonStr); len(toolCalls) > 0 {
				return toolCalls
			}
		}
	}

	// 方法2：尝试从普通代码块中解析
	start = strings.Index(content, "```")
	if start != -1 {
		start += 3 // 跳过 ```
		// 跳过空白和语言标识
		for start < len(content) && (content[start] == ' ' || content[start] == '\n' || content[start] == '\r' ||
			(content[start] >= 'a' && content[start] <= 'z') || (content[start] >= 'A' && content[start] <= 'Z')) {
			start++
		}
		// 查找结束标记
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr := strings.TrimSpace(content[start : start+end])
			if toolCalls := a.parseToolCallsJSON(jsonStr); len(toolCalls) > 0 {
				return toolCalls
			}
		}
	}

	// 方法3：尝试直接解析整个内容（可能是纯 JSON）
	if strings.TrimSpace(content)[0] == '{' {
		if toolCalls := a.parseToolCallsJSON(content); len(toolCalls) > 0 {
			return toolCalls
		}
	}

	// 方法4：尝试查找 JSON 对象（可能被包裹在文本中）
	// 查找第一个 {
	start = strings.Index(content, "{")
	if start != -1 {
		// 查找匹配的 }
		braceCount := 0
		for i := start; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
			} else if content[i] == '}' {
				braceCount--
				if braceCount == 0 {
					jsonStr := strings.TrimSpace(content[start : i+1])
					if toolCalls := a.parseToolCallsJSON(jsonStr); len(toolCalls) > 0 {
						return toolCalls
					}
					break
				}
			}
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseToolCallsJSON 解析工具调用的 JSON
func (a *DeepSeekToolCallingAdapter) parseToolCallsJSON(jsonStr string) []ToolCall {
	// 尝试多种 JSON 格式
	// 格式1: { "tool_calls": [...] }
	var result1 struct {
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result1); err == nil && len(result1.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, 0, len(result1.ToolCalls))
		for _, tc := range result1.ToolCalls {
			toolCalls = append(toolCalls, ToolCall{
				ID:        tc.ID,
				Type:      tc.Type,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
		return toolCalls
	}

	// 格式2: 直接是数组 [{ "id": ..., "function": {...} }]
	var result2 []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"function"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result2); err == nil && len(result2) > 0 {
		toolCalls := make([]ToolCall, 0, len(result2))
		for _, tc := range result2 {
			toolCalls = append(toolCalls, ToolCall{
				ID:        tc.ID,
				Type:      tc.Type,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
		return toolCalls
	}

	return nil
}
