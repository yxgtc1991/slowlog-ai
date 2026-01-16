package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolCall 工具调用结构
type ToolCall struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`      // 通常是 "function"
	Name      string                 `json:"name"`      // 工具名称
	Arguments map[string]interface{} `json:"arguments"` // 工具参数（JSON 字符串或对象）
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Type        string                 `json:"type"`        // 通常是 "function"
	Name        string                 `json:"name"`        // 工具名称
	Description string                 `json:"description"` // 工具描述
	Parameters  map[string]interface{} `json:"parameters"`  // 参数 schema
}

// ChatWithToolsRequest 带工具调用的聊天请求
type ChatWithToolsRequest struct {
	Messages   []Message        `json:"messages"`
	Tools      []ToolDefinition `json:"tools,omitempty"`       // 可用工具列表
	ToolChoice string           `json:"tool_choice,omitempty"` // "auto", "none", 或具体工具名
}

// Message 消息结构
type Message struct {
	Role       string      `json:"role"`                   // "user", "assistant", "tool"
	Content    interface{} `json:"content"`                // 字符串或对象数组
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`   // 工具调用列表
	ToolCallID string      `json:"tool_call_id,omitempty"` // 工具调用 ID（tool role 时使用）
}

// ChatWithToolsResponse 带工具调用的聊天响应
type ChatWithToolsResponse struct {
	Content   string     `json:"content"`    // 文本内容
	ToolCalls []ToolCall `json:"tool_calls"` // 工具调用列表
	Role      string     `json:"role"`       // 通常是 "assistant"
}

// ToolCallingClient 支持工具调用的 LLM 客户端接口
type ToolCallingClient interface {
	// ChatWithTools 支持工具调用的聊天方法
	ChatWithTools(ctx context.Context, req ChatWithToolsRequest) (*ChatWithToolsResponse, error)
}

// ConvertCapabilitiesToTools 将能力列表转换为工具定义
func ConvertCapabilitiesToTools(capabilities []CapabilityInfo) []ToolDefinition {
	tools := make([]ToolDefinition, 0, len(capabilities))

	for _, cap := range capabilities {
		// 构建参数 schema
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for paramName, paramDesc := range cap.InputSchema {
			// 解析参数描述（格式：type // description）
			parts := parseParamDesc(paramDesc)
			properties[paramName] = map[string]interface{}{
				"type":        parts.Type,
				"description": parts.Description,
			}
			required = append(required, paramName)
		}

		tools = append(tools, ToolDefinition{
			Type:        "function",
			Name:        cap.Name,
			Description: cap.Description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		})
	}

	return tools
}

// CapabilityInfo 能力信息（用于转换为工具定义）
type CapabilityInfo struct {
	Name        string
	Description string
	InputSchema map[string]string
}

// ParamDesc 参数描述
type ParamDesc struct {
	Type        string
	Description string
}

// parseParamDesc 解析参数描述（格式：type // description）
func parseParamDesc(desc string) ParamDesc {
	parts := splitParamDesc(desc)
	if len(parts) >= 2 {
		return ParamDesc{
			Type:        trim(parts[0]),
			Description: trim(parts[1]),
		}
	}
	if len(parts) == 1 {
		return ParamDesc{
			Type:        trim(parts[0]),
			Description: "",
		}
	}
	return ParamDesc{Type: "string", Description: desc}
}

// splitParamDesc 分割参数描述
func splitParamDesc(desc string) []string {
	// 查找 "//" 分隔符
	for i := 0; i < len(desc)-1; i++ {
		if desc[i] == '/' && desc[i+1] == '/' {
			return []string{desc[:i], desc[i+2:]}
		}
	}
	return []string{desc}
}

// trim 去除首尾空格
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// ParseToolCallArguments 解析工具调用的参数（可能是 JSON 字符串或对象）
func ParseToolCallArguments(args interface{}) (map[string]interface{}, error) {
	switch v := args.(type) {
	case string:
		// 如果是字符串，尝试解析为 JSON
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("failed to parse arguments as JSON: %w", err)
		}
		return result, nil
	case map[string]interface{}:
		// 如果已经是 map，直接返回
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported arguments type: %T", args)
	}
}
