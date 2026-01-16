package analyzer

import (
	"ai_slow_log/internal/llm"
	promptv5 "ai_slow_log/internal/prompt/slowlog"
	"context"
	"encoding/json"
	"fmt"
)

// V5ToolCallingAnalyzer v5 版本的分析器（使用真正的 Tool Calling）
// 让 LLM 通过 Tool Calling 协议直接调用工具，而不是在文本中输出意图
type V5ToolCallingAnalyzer struct {
	llm                ToolCallingLLMClient
	retriever          Retriever
	capabilityExecutor CapabilityExecutor
	availableCaps      []promptv5.CapabilityV4
	maxIterations      int
	currentIteration   int
}

// ToolCallingLLMClient 支持工具调用的 LLM 客户端接口
// 直接使用 llm 包中定义的接口
type ToolCallingLLMClient = llm.ToolCallingClient

// ToolCallingRequest 工具调用请求
type ToolCallingRequest struct {
	Messages   []ToolCallingMessage
	Tools      []ToolDefinition
	ToolChoice string // "auto", "none", 或具体工具名
}

// ToolCallingMessage 工具调用消息
type ToolCallingMessage struct {
	Role      string
	Content   string
	ToolCalls []ToolCall
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Type        string
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// ToolCallingResponse 工具调用响应
type ToolCallingResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// V5ToolCallingResult v5 工具调用版本的分析结果
type V5ToolCallingResult struct {
	Analysis    string                 // 最终分析结果
	ToolCalls   []ToolCall             // 工具调用列表
	ToolResults map[string]interface{} // 工具执行结果
	Iterations  int                    // 实际迭代次数
	RawOutput   string                 // 原始 LLM 输出
}

// NewV5ToolCallingAnalyzer 创建 v5 工具调用版本的分析器
func NewV5ToolCallingAnalyzer(
	llm ToolCallingLLMClient,
	retriever Retriever,
	executor CapabilityExecutor,
	caps []promptv5.CapabilityV4,
) *V5ToolCallingAnalyzer {
	return &V5ToolCallingAnalyzer{
		llm:                llm,
		retriever:          retriever,
		capabilityExecutor: executor,
		availableCaps:      caps,
		maxIterations:      3,
	}
}

// Analyze 执行 v5 工具调用版本的分析
// 流程：LLM 分析 → 输出工具调用 → 系统执行 → 返回结果给 LLM → LLM 继续分析
func (a *V5ToolCallingAnalyzer) Analyze(ctx context.Context, slowLog string) (*V5ToolCallingResult, error) {
	if slowLog == "" {
		return nil, fmt.Errorf("slow log is empty")
	}

	a.currentIteration = 0
	toolResults := make(map[string]interface{})
	allToolCalls := make([]ToolCall, 0) // 保存所有轮次的工具调用
	var finalAnalysis string

	// 1. 构建工具定义
	tools := a.buildToolDefinitions()

	// 2. 构建初始消息
	messages := []ToolCallingMessage{
		{
			Role:    "user",
			Content: a.buildInitialPrompt(slowLog),
		},
	}

	// 3. 多轮交互
	for a.currentIteration < a.maxIterations {
		a.currentIteration++

		// 调用 LLM（带工具定义）
		// 第一轮必须调用工具，后续轮次可以是 auto
		toolChoice := "auto"
		if a.currentIteration == 1 {
			// 第一轮强制要求调用工具（如果 SDK 支持）
			// 如果不支持，通过 prompt 强制
			toolChoice = "required" // 或使用第一个工具名
		}

		// 转换为 llm 包的类型
		llmMessages := a.convertMessagesToLLM(messages)
		llmTools := a.convertToolsToLLM(tools)

		llmReq := llm.ChatWithToolsRequest{
			Messages:   llmMessages,
			Tools:      llmTools,
			ToolChoice: toolChoice,
		}

		llmResp, err := a.llm.ChatWithTools(ctx, llmReq)
		if err != nil {
			return nil, fmt.Errorf("failed to call LLM: %w", err)
		}

		// 转换响应
		resp := a.convertLLMResponse(llmResp)

		// 4. 检查是否有工具调用
		if len(resp.ToolCalls) == 0 {
			// 第一轮如果没有工具调用，说明还没进入 MCP/Agent 模式
			if a.currentIteration == 1 {
				// 强制要求调用工具，重新发送请求
				messages = append(messages, ToolCallingMessage{
					Role:    "user",
					Content: "你必须调用工具来分析慢日志，不要直接输出分析结果。请调用 analyze_slow_log 工具。",
				})
				continue // 重新循环，强制调用工具
			}
			// 后续轮次如果没有工具调用，说明已经完成
			finalAnalysis = resp.Content
			break
		}

		// 5. 保存工具调用（用于最终结果）
		allToolCalls = append(allToolCalls, resp.ToolCalls...)

		// 6. 执行工具调用
		for _, toolCall := range resp.ToolCalls {
			result, err := a.capabilityExecutor.ExecuteCapability(ctx, toolCall.Name, toolCall.Arguments)
			if err != nil {
				toolResults[toolCall.Name] = map[string]interface{}{
					"error": err.Error(),
				}
			} else {
				toolResults[toolCall.Name] = result
			}
		}

		// 7. 将工具执行结果添加到消息历史
		// 添加 assistant 消息（包含工具调用）
		messages = append(messages, ToolCallingMessage{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// 添加 tool 消息（工具执行结果）
		for _, toolCall := range resp.ToolCalls {
			result := toolResults[toolCall.Name]
			resultJSON, _ := json.Marshal(result)
			messages = append(messages, ToolCallingMessage{
				Role:    "tool",
				Content: string(resultJSON),
			})
		}

		// 继续下一轮，让 LLM 基于工具结果继续分析
	}

	// 如果达到最大迭代次数，使用最后一轮的内容
	if finalAnalysis == "" && len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		finalAnalysis = lastMsg.Content
	}

	return &V5ToolCallingResult{
		Analysis:    finalAnalysis,
		ToolCalls:   allToolCalls, // 保存所有轮次的工具调用
		ToolResults: toolResults,
		Iterations:  a.currentIteration,
		RawOutput:   finalAnalysis,
	}, nil
}

// buildToolDefinitions 构建工具定义
func (a *V5ToolCallingAnalyzer) buildToolDefinitions() []ToolDefinition {
	tools := make([]ToolDefinition, 0, len(a.availableCaps))

	for _, cap := range a.availableCaps {
		meta := promptv5.DescribeCapabilityV4(cap)

		// 构建参数 schema
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for paramName, paramDesc := range meta.InputSchema {
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
			Name:        meta.Name,
			Description: meta.Description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		})
	}

	return tools
}

// buildInitialPrompt 构建初始 prompt
// 关键：明确告诉 LLM 必须调用工具，而不是先输出文本分析
func (a *V5ToolCallingAnalyzer) buildInitialPrompt(slowLog string) string {
	return fmt.Sprintf(`你是一个【MySQL 慢查询分析助手】。

你的任务是分析慢日志。**重要：你必须使用工具来完成分析，不要直接输出分析结果。**

【慢日志内容】
%s

【重要规则】
1. 你必须调用 analyze_slow_log 工具来分析这段慢日志
2. 不要直接输出分析结果，必须通过工具调用
3. 工具会返回分析结果，然后你可以基于结果给出总结和建议

现在请调用 analyze_slow_log 工具，参数 slow_log 就是上面的慢日志内容。`, slowLog)
}

// ParamDesc 参数描述
type ParamDesc struct {
	Type        string
	Description string
}

// parseParamDesc 解析参数描述（格式：type // description）
func parseParamDesc(desc string) ParamDesc {
	// 查找 "//" 分隔符
	for i := 0; i < len(desc)-1; i++ {
		if desc[i] == '/' && desc[i+1] == '/' {
			typePart := trim(desc[:i])
			descPart := trim(desc[i+2:])
			return ParamDesc{
				Type:        typePart,
				Description: descPart,
			}
		}
	}
	return ParamDesc{Type: "string", Description: desc}
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

// convertMessagesToLLM 将内部消息格式转换为 llm 包的消息格式
func (a *V5ToolCallingAnalyzer) convertMessagesToLLM(msgs []ToolCallingMessage) []llm.Message {
	result := make([]llm.Message, 0, len(msgs))
	for _, msg := range msgs {
		llmMsg := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if len(msg.ToolCalls) > 0 {
			llmToolCalls := make([]llm.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				llmToolCalls = append(llmToolCalls, llm.ToolCall{
					ID:        tc.ID,
					Type:      "function",
					Name:      tc.Name,
					Arguments: tc.Arguments,
				})
			}
			llmMsg.ToolCalls = llmToolCalls
		}
		result = append(result, llmMsg)
	}
	return result
}

// convertToolsToLLM 将内部工具格式转换为 llm 包的工具格式
func (a *V5ToolCallingAnalyzer) convertToolsToLLM(tools []ToolDefinition) []llm.ToolDefinition {
	result := make([]llm.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		result = append(result, llm.ToolDefinition{
			Type:        tool.Type,
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}
	return result
}

// convertLLMResponse 将 llm 包的响应转换为内部响应格式
func (a *V5ToolCallingAnalyzer) convertLLMResponse(llmResp *llm.ChatWithToolsResponse) *ToolCallingResponse {
	toolCalls := make([]ToolCall, 0, len(llmResp.ToolCalls))
	for _, tc := range llmResp.ToolCalls {
		args, _ := llm.ParseToolCallArguments(tc.Arguments)
		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Name,
			Arguments: args,
		})
	}
	return &ToolCallingResponse{
		Content:   llmResp.Content,
		ToolCalls: toolCalls,
	}
}
