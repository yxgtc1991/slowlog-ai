package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ActionType V6 版本的行动类型
// LLM 可以自主决定下一步要做什么
type ActionType string

const (
	ActionCallTool    ActionType = "call_tool"    // 调用工具
	ActionRetrieveRAG ActionType = "retrieve_rag" // 检索 RAG 知识库
	ActionAnalyze     ActionType = "analyze"      // 继续分析（基于已有信息）
	ActionAskQuestion ActionType = "ask_question" // 提出问题（需要用户输入）
	ActionFinish      ActionType = "finish"       // 完成并输出最终结果
)

// NextAction V6 版本的下一步行动决策
// LLM 输出这个结构，系统根据行动类型执行对应操作
type NextAction struct {
	Type      ActionType             `json:"type"`                // 行动类型
	Reasoning string                 `json:"reasoning"`           // 为什么选择这个行动
	ToolName  string                 `json:"tool_name,omitempty"` // 如果 type=call_tool，指定工具名
	ToolArgs  map[string]interface{} `json:"tool_args,omitempty"` // 如果 type=call_tool，工具参数
	RAGQuery  string                 `json:"rag_query,omitempty"` // 如果 type=retrieve_rag，检索查询
	Analysis  string                 `json:"analysis,omitempty"`  // 如果 type=analyze，分析内容
	Question  string                 `json:"question,omitempty"`  // 如果 type=ask_question，问题内容
	Result    string                 `json:"result,omitempty"`    // 如果 type=finish，最终结果
}

// AgentDecision V6 版本的 Agent 决策
// 包含当前状态、下一步行动和上下文信息
type AgentDecision struct {
	CurrentState string      `json:"current_state"`     // 当前状态描述
	NextAction   NextAction  `json:"next_action"`       // 下一步行动
	Context      interface{} `json:"context,omitempty"` // 上下文信息（可选）
}

// BuildAgentPromptV6 构建 V6 版本的 Agent prompt
// V6 版本让 LLM 自主决定下一步行动，而不仅仅是调用工具
func BuildAgentPromptV6(
	slowLog string,
	availableTools []CapabilityV4,
	conversationHistory []string, // 对话历史（可选）
	currentContext map[string]interface{}, // 当前上下文（已执行的工具结果、RAG 检索结果等）
) string {
	var sb strings.Builder

	// 1. Agent 角色设定
	sb.WriteString(`你是一个【MySQL 慢日志分析 Agent】，你的任务是自主决策如何分析慢日志。

【核心原则】
1. 你需要自主规划分析步骤，决定下一步要做什么
2. 你可以调用工具、检索知识库、继续分析、提问或完成分析
3. 每一步都要有明确的理由（reasoning）
4. 当你有足够信息给出最终分析结果时，选择 finish

`)

	// 2. 可用工具列表
	if len(availableTools) > 0 {
		sb.WriteString("【可用工具】\n")
		for i, tool := range availableTools {
			meta := DescribeCapabilityV4(tool)
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, meta.Name))
			sb.WriteString(fmt.Sprintf("   说明：%s\n", meta.Description))
			sb.WriteString("   参数：\n")
			for k, v := range meta.InputSchema {
				sb.WriteString(fmt.Sprintf("     - %s: %s\n", k, v))
			}
			sb.WriteString("\n")
		}
	}

	// 3. 行动类型说明
	sb.WriteString(`【可执行的行动类型】

1. call_tool - 调用工具
   - 当你需要调用系统工具时使用
   - 必须提供 tool_name 和 tool_args
   - 例如：调用 analyze_slow_log 工具分析慢日志

2. retrieve_rag - 检索知识库
   - 当你需要查询相关知识（模式、反模式、指标说明等）时使用
   - 必须提供 rag_query（检索查询字符串）
   - 例如：查询"rows_examined 高"相关的优化建议

3. analyze - 继续分析
   - 当你基于已有信息（工具结果、RAG 检索结果）继续分析时使用
   - 必须提供 analysis（分析内容）
   - 例如：基于工具分析结果，进一步推断问题原因

4. ask_question - 提出问题
   - 当你需要用户提供额外信息时使用
   - 必须提供 question（问题内容）
   - 例如：询问表结构、索引信息等

5. finish - 完成分析
   - 当你已经收集足够信息，可以给出最终分析结果时使用
   - 必须提供 result（最终分析结果，包含问题诊断和优化建议）
   - 这是分析的终点

`)

	// 4. 当前上下文（已执行的操作和结果）
	if len(currentContext) > 0 {
		sb.WriteString("【当前上下文】\n")
		sb.WriteString("以下是已经执行的操作和结果：\n\n")
		for key, value := range currentContext {
			sb.WriteString(fmt.Sprintf("- %s:\n", key))
			valueJSON, _ := json.MarshalIndent(value, "  ", "  ")
			sb.WriteString(fmt.Sprintf("  %s\n\n", string(valueJSON)))
		}
	}

	// 5. 对话历史（如果有）
	if len(conversationHistory) > 0 {
		sb.WriteString("【对话历史】\n")
		for i, msg := range conversationHistory {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, msg))
		}
		sb.WriteString("\n")
	}

	// 6. 慢日志内容
	sb.WriteString("【慢日志内容】\n")
	sb.WriteString(slowLog)
	sb.WriteString("\n\n")

	// 7. 输出格式要求
	sb.WriteString(`【输出格式】
你必须输出严格的 JSON 格式，结构如下：
{
  "current_state": "当前分析状态描述（例如：已调用 analyze_slow_log 工具，获得了初步分析结果）",
  "next_action": {
    "type": "call_tool | retrieve_rag | analyze | ask_question | finish",
    "reasoning": "为什么选择这个行动（必须详细说明）",
    "tool_name": "如果 type=call_tool，指定工具名",
    "tool_args": { "如果 type=call_tool，工具参数" },
    "rag_query": "如果 type=retrieve_rag，检索查询",
    "analysis": "如果 type=analyze，分析内容",
    "question": "如果 type=ask_question，问题内容",
    "result": "如果 type=finish，最终分析结果（包含问题诊断和优化建议）"
  }
}

【重要提示】
1. 每一步都要有明确的 reasoning（为什么选择这个行动）
2. 如果选择 call_tool，必须提供正确的 tool_name 和 tool_args
3. 如果选择 retrieve_rag，rag_query 应该是有针对性的查询（例如："rows_examined 高如何优化"）
4. 如果选择 finish，result 应该是完整的分析报告（问题诊断 + 优化建议）
5. 输出必须是有效的 JSON，可以直接被程序解析

现在请分析慢日志，并决定下一步要做什么：

`)

	return sb.String()
}

// ParseAgentDecision 解析 V6 版本的 Agent 决策
func ParseAgentDecision(llmOutput string) (*AgentDecision, error) {
	// 尝试提取 JSON（可能包含 markdown 代码块）
	jsonStr := extractJSON(llmOutput)
	if jsonStr == "" {
		// 如果提取失败，输出前 500 字符用于调试
		debugOutput := llmOutput
		if len(debugOutput) > 500 {
			debugOutput = debugOutput[:500] + "..."
		}
		return nil, fmt.Errorf("no JSON found in LLM output. First 500 chars: %s", debugOutput)
	}

	var decision AgentDecision
	if err := json.Unmarshal([]byte(jsonStr), &decision); err != nil {
		// 输出提取到的 JSON 字符串用于调试
		debugJSON := jsonStr
		if len(debugJSON) > 500 {
			debugJSON = debugJSON[:500] + "..."
		}
		return nil, fmt.Errorf("failed to parse JSON: %w. Extracted JSON (first 500 chars): %s", err, debugJSON)
	}

	// 验证行动类型
	if decision.NextAction.Type != ActionCallTool &&
		decision.NextAction.Type != ActionRetrieveRAG &&
		decision.NextAction.Type != ActionAnalyze &&
		decision.NextAction.Type != ActionAskQuestion &&
		decision.NextAction.Type != ActionFinish {
		return nil, fmt.Errorf("invalid action type: %s", decision.NextAction.Type)
	}

	return &decision, nil
}

// extractJSON 从文本中提取 JSON（可能包含在 markdown 代码块中）
func extractJSON(text string) string {
	// 1. 尝试查找 ```json ... ``` 代码块
	if idx := strings.Index(text, "```json"); idx != -1 {
		// 找到 ```json 后的换行
		start := idx + 7 // len("```json")
		for start < len(text) && (text[start] == ' ' || text[start] == '\n' || text[start] == '\r') {
			start++
		}
		// 找到结束的 ```
		if end := strings.Index(text[start:], "```"); end != -1 {
			jsonStr := strings.TrimSpace(text[start : start+end])
			// 验证是否是有效的 JSON
			var test interface{}
			if json.Unmarshal([]byte(jsonStr), &test) == nil {
				return jsonStr
			}
		}
	}

	// 2. 尝试查找 ``` ... ``` 代码块（无语言标识）
	if idx := strings.Index(text, "```"); idx != -1 {
		// 跳过开头的 ```
		start := idx + 3
		// 跳过可能的语言标识和换行
		for start < len(text) && text[start] != '\n' && text[start] != '\r' {
			start++
		}
		// 跳过换行
		for start < len(text) && (text[start] == '\n' || text[start] == '\r') {
			start++
		}
		// 找到结束的 ```
		if end := strings.Index(text[start:], "```"); end != -1 {
			jsonStr := strings.TrimSpace(text[start : start+end])
			// 验证是否是有效的 JSON
			var test interface{}
			if json.Unmarshal([]byte(jsonStr), &test) == nil {
				return jsonStr
			}
		}
	}

	// 3. 尝试直接解析整个文本
	var test interface{}
	if json.Unmarshal([]byte(text), &test) == nil {
		return text
	}

	// 4. 尝试查找第一个 { 到最后一个 }（更智能的匹配）
	startIdx := strings.Index(text, "{")
	if startIdx == -1 {
		return ""
	}

	// 从后往前找匹配的 }
	braceCount := 0
	endIdx := -1
	for i := startIdx; i < len(text); i++ {
		if text[i] == '{' {
			braceCount++
		} else if text[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i
				break
			}
		}
	}

	if endIdx != -1 && endIdx > startIdx {
		jsonStr := strings.TrimSpace(text[startIdx : endIdx+1])
		// 验证是否是有效的 JSON
		var test interface{}
		if json.Unmarshal([]byte(jsonStr), &test) == nil {
			return jsonStr
		}
	}

	return ""
}
