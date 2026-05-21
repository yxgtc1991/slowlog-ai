package analyzer

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"context"
	"fmt"
)

// V6AgentAnalyzer V6 版本的 Agent 分析器
// LLM 自主决定下一步行动，而不仅仅是调用工具
type V6AgentAnalyzer struct {
	llm                 LLMClient               // 使用普通 LLM 客户端（不需要 Tool Calling）
	retriever           Retriever               // RAG 检索器
	capabilityExecutor  CapabilityExecutor      // 能力执行器
	availableTools      []promptv6.CapabilityV4 // 可用工具列表
	verbose             bool                    // 每轮轨迹输出到 stderr
	maxIterations       int                     // 最大迭代次数
	currentIteration    int                     // 当前迭代次数
	conversationHistory []string                // 对话历史
	context             map[string]interface{}  // 当前上下文（工具结果、RAG 结果等）
}

// V6AgentResult V6 Agent 版本的分析结果
type V6AgentResult struct {
	FinalResult         string                 // 最终分析结果
	Actions             []promptv6.NextAction  // 执行的所有行动
	ToolResults         map[string]interface{} // 工具执行结果
	RAGResults          []RAGResult            // RAG 检索结果
	Iterations          int                    // 实际迭代次数
	ConversationHistory []string               // 完整对话历史
}

// RAGResult RAG 检索结果
type RAGResult struct {
	Query  string
	Chunks []interface{} // []rag.KnowledgeChunk
}

// NewV6AgentAnalyzer 创建 V6 Agent 版本的分析器
func NewV6AgentAnalyzer(
	llm LLMClient,
	retriever Retriever,
	executor CapabilityExecutor,
	tools []promptv6.CapabilityV4,
	opts ...V6AgentOption,
) *V6AgentAnalyzer {
	a := &V6AgentAnalyzer{
		llm:                 llm,
		retriever:           retriever,
		capabilityExecutor:  executor,
		availableTools:      tools,
		maxIterations:       10, // Agent 可能需要更多轮次
		context:             make(map[string]interface{}),
		conversationHistory: make([]string, 0),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Analyze 执行 V6 Agent 版本的分析
// 流程：LLM 决策 → 执行行动 → 更新上下文 → LLM 继续决策 → ... → 完成
func (a *V6AgentAnalyzer) Analyze(ctx context.Context, slowLog string) (*V6AgentResult, error) {
	if slowLog == "" {
		return nil, fmt.Errorf("slow log is empty")
	}

	// 重置状态
	a.currentIteration = 0
	a.context = make(map[string]interface{})
	a.conversationHistory = make([]string, 0)
	allActions := make([]promptv6.NextAction, 0)
	toolResults := make(map[string]interface{})
	ragResults := make([]RAGResult, 0)
	var finalResult string

	// Agent 循环：LLM 决策 → 执行 → 更新上下文 → 继续
	for a.currentIteration < a.maxIterations {
		a.currentIteration++
		a.traceRoundStart(a.currentIteration)

		// 1. 构建 Agent prompt，让 LLM 决定下一步行动
		prompt := promptv6.BuildAgentPromptV6(
			slowLog,
			a.availableTools,
			a.conversationHistory,
			a.context,
		)

		// 2. 调用 LLM 获取决策
		llmOutput, err := a.llm.Chat(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to call LLM: %w", err)
		}

		// 3. 解析 LLM 的决策
		decision, err := promptv6.ParseAgentDecision(llmOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent decision: %w", err)
		}

		a.traceDecision(decision, llmOutput)

		// 4. 记录决策
		allActions = append(allActions, decision.NextAction)
		a.conversationHistory = append(a.conversationHistory,
			fmt.Sprintf("决策 %d: %s - %s", a.currentIteration, decision.NextAction.Type, decision.NextAction.Reasoning))

		// 5. 执行行动
		var lastRAG *RAGResult
		switch decision.NextAction.Type {
		case promptv6.ActionCallTool:
			// 调用工具
			result, err := a.capabilityExecutor.ExecuteCapability(
				ctx,
				decision.NextAction.ToolName,
				decision.NextAction.ToolArgs,
			)
			if err != nil {
				a.context[fmt.Sprintf("tool_%s_error", decision.NextAction.ToolName)] = err.Error()
			} else {
				toolResults[decision.NextAction.ToolName] = result
				a.context[fmt.Sprintf("tool_%s_result", decision.NextAction.ToolName)] = result
			}
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("执行工具: %s", decision.NextAction.ToolName))

		case promptv6.ActionRetrieveRAG:
			// 检索 RAG
			chunks, err := a.retriever.Retrieve(ctx, decision.NextAction.RAGQuery)
			if err != nil {
				a.context[fmt.Sprintf("rag_%s_error", decision.NextAction.RAGQuery)] = err.Error()
				a.conversationHistory = append(a.conversationHistory,
					fmt.Sprintf("检索 RAG 失败: %s - %v", decision.NextAction.RAGQuery, err))
			} else {
				// 转换为 interface{} 以便存储
				chunkInterfaces := make([]interface{}, len(chunks))
				for i, chunk := range chunks {
					chunkInterfaces[i] = map[string]interface{}{
						"title":   chunk.Title,
						"content": chunk.Content,
						"source":  chunk.Source,
						"score":   chunk.Score,
					}
				}
				rr := RAGResult{
					Query:  decision.NextAction.RAGQuery,
					Chunks: chunkInterfaces,
				}
				ragResults = append(ragResults, rr)
				lastRAG = &ragResults[len(ragResults)-1]
				a.context[fmt.Sprintf("rag_%s_result", decision.NextAction.RAGQuery)] = chunkInterfaces
				a.conversationHistory = append(a.conversationHistory,
					fmt.Sprintf("检索 RAG: %s (找到 %d 个知识块)", decision.NextAction.RAGQuery, len(chunks)))
			}

		case promptv6.ActionAnalyze:
			// 继续分析（基于已有信息）
			a.context["analysis"] = decision.NextAction.Analysis
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("分析: %s", decision.NextAction.Analysis))

		case promptv6.ActionAskQuestion:
			// 提出问题（需要用户输入）
			// 在实际应用中，这里可以暂停等待用户输入
			// 当前实现中，我们记录问题并继续
			a.context["question"] = decision.NextAction.Question
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("问题: %s", decision.NextAction.Question))
			// 注意：在真实场景中，这里应该等待用户回答

		case promptv6.ActionFinish:
			// 完成分析
			finalResult = decision.NextAction.Result
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("完成: %s", finalResult))
		}

		a.traceAfterAction(decision.NextAction, toolResults, lastRAG)

		// 如果已经完成，跳出循环
		if decision.NextAction.Type == promptv6.ActionFinish {
			break
		}
	}

	// 如果达到最大迭代次数但未完成，使用最后一次的分析或结果
	if finalResult == "" && len(allActions) > 0 {
		lastAction := allActions[len(allActions)-1]
		if lastAction.Analysis != "" {
			finalResult = lastAction.Analysis
		} else if lastAction.Result != "" {
			finalResult = lastAction.Result
		} else {
			finalResult = "分析未完成（达到最大迭代次数）"
		}
	}

	return &V6AgentResult{
		FinalResult:         finalResult,
		Actions:             allActions,
		ToolResults:         toolResults,
		RAGResults:          ragResults,
		Iterations:          a.currentIteration,
		ConversationHistory: a.conversationHistory,
	}, nil
}

// formatContextForPrompt 格式化上下文用于 prompt（内部方法，如果需要）
func (a *V6AgentAnalyzer) formatContextForPrompt() map[string]interface{} {
	// 简化上下文，只保留关键信息
	simplified := make(map[string]interface{})
	for k, v := range a.context {
		// 如果值太大，截断
		valueStr := fmt.Sprintf("%v", v)
		if len(valueStr) > 500 {
			valueStr = valueStr[:500] + "..."
		}
		simplified[k] = valueStr
	}
	return simplified
}
