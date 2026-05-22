package analyzer

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"context"
	"fmt"
)

// V6AgentAnalyzer V6 版本的 Agent 分析器
// LLM 自主决定下一步行动，而不仅仅是调用工具
type V6AgentAnalyzer struct {
	llm                 LLMClient
	retriever           Retriever
	capabilityExecutor  CapabilityExecutor
	availableTools      []promptv6.CapabilityV4
	verbose             bool
	recordRounds        bool
	extraGuide          string
	maxIterations       int
	currentIteration    int
	conversationHistory []string
	state               *AgentState
}

// V6AgentResult V6 Agent 版本的分析结果
type V6AgentResult struct {
	FinalResult         string
	Actions             []promptv6.NextAction
	Rounds              []AgentRoundRecord
	ToolResults         map[string]interface{}
	RAGResults          []RAGResult
	Iterations          int
	ConversationHistory []string
	FinalPhase          AgentPhase
	State               *AgentState
}

// RAGResult RAG 检索结果
type RAGResult struct {
	Query  string
	Chunks []interface{}
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
		maxIterations:       10,
		state:               NewAgentState(),
		conversationHistory: make([]string, 0),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Analyze 执行 V6 Agent 版本的分析
func (a *V6AgentAnalyzer) Analyze(ctx context.Context, slowLog string) (*V6AgentResult, error) {
	if slowLog == "" {
		return nil, fmt.Errorf("slow log is empty")
	}

	a.currentIteration = 0
	a.state = NewAgentState()
	a.conversationHistory = make([]string, 0)
	allActions := make([]promptv6.NextAction, 0)
	rounds := make([]AgentRoundRecord, 0)
	toolResults := make(map[string]interface{})
	ragResults := make([]RAGResult, 0)
	var finalResult string

	for a.currentIteration < a.maxIterations {
		a.currentIteration++
		a.traceRoundStart(a.currentIteration)

		prompt := promptv6.BuildAgentPromptV6(
			slowLog,
			a.availableTools,
			a.conversationHistory,
			a.state.PromptSummary(400),
			a.extraGuide,
		)

		llmOutput, err := a.llm.Chat(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to call LLM: %w", err)
		}

		decision, err := promptv6.ParseAgentDecision(llmOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent decision: %w", err)
		}

		a.traceDecision(decision, llmOutput)

		roundRec := AgentRoundRecord{
			Round:        a.currentIteration,
			LLMRaw:       llmOutput,
			CurrentState: decision.CurrentState,
			AgentPhase:   string(a.state.Phase),
			Action:       decision.NextAction,
		}

		allActions = append(allActions, decision.NextAction)
		a.conversationHistory = append(a.conversationHistory,
			fmt.Sprintf("决策 %d: %s - %s", a.currentIteration, decision.NextAction.Type, decision.NextAction.Reasoning))

		var lastRAG *RAGResult
		switch decision.NextAction.Type {
		case promptv6.ActionCallTool:
			toolArgs := decision.NextAction.ToolArgs
			if decision.NextAction.ToolName == "explain_mysql_query" {
				toolArgs = NormalizeExplainArgs(toolArgs, slowLog)
			}
			result, err := a.capabilityExecutor.ExecuteCapability(
				ctx,
				decision.NextAction.ToolName,
				toolArgs,
			)
			if err != nil {
				roundRec.ActionError = err.Error()
				a.state.RecordTool(decision.NextAction.ToolName, nil, err)
			} else {
				roundRec.ActionOutcome = result
				toolResults[decision.NextAction.ToolName] = result
				a.state.RecordTool(decision.NextAction.ToolName, result, nil)
			}
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("执行工具: %s", decision.NextAction.ToolName))

		case promptv6.ActionRetrieveRAG:
			chunks, err := a.retriever.Retrieve(ctx, decision.NextAction.RAGQuery)
			if err != nil {
				roundRec.ActionError = err.Error()
				a.state.RecordRAG(decision.NextAction.RAGQuery, nil, err)
				a.conversationHistory = append(a.conversationHistory,
					fmt.Sprintf("检索 RAG 失败: %s - %v", decision.NextAction.RAGQuery, err))
			} else {
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
				roundRec.ActionOutcome = lastRAG
				a.state.RecordRAG(decision.NextAction.RAGQuery, chunks, nil)
				a.conversationHistory = append(a.conversationHistory,
					fmt.Sprintf("检索 RAG: %s (找到 %d 个知识块)", decision.NextAction.RAGQuery, len(chunks)))
			}

		case promptv6.ActionAnalyze:
			roundRec.ActionOutcome = decision.NextAction.Analysis
			a.state.RecordAnalyze(decision.NextAction.Analysis)
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("分析: %s", decision.NextAction.Analysis))

		case promptv6.ActionAskQuestion:
			a.state.RecordQuestion(decision.NextAction.Question)
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("问题: %s", decision.NextAction.Question))

		case promptv6.ActionFinish:
			finalResult = decision.NextAction.Result.String()
			a.state.MarkFinished()
			a.conversationHistory = append(a.conversationHistory,
				fmt.Sprintf("完成: %s", finalResult))
		}

		a.traceAfterAction(decision.NextAction, toolResults, lastRAG)

		if a.recordRounds {
			roundRec.AgentPhase = string(a.state.Phase)
			roundRec.ContextKeys = a.state.contextKeys()
			rounds = append(rounds, roundRec)
		}

		if decision.NextAction.Type == promptv6.ActionFinish {
			break
		}
	}

	if finalResult == "" && len(allActions) > 0 {
		lastAction := allActions[len(allActions)-1]
		if lastAction.Analysis != "" {
			finalResult = lastAction.Analysis
		} else if lastAction.Result.String() != "" {
			finalResult = lastAction.Result.String()
		} else {
			finalResult = "分析未完成（达到最大迭代次数）"
		}
	}

	return &V6AgentResult{
		FinalResult:         finalResult,
		Actions:             allActions,
		Rounds:              rounds,
		ToolResults:         toolResults,
		RAGResults:          ragResults,
		Iterations:          a.currentIteration,
		ConversationHistory: a.conversationHistory,
		FinalPhase:          a.state.Phase,
		State:               a.state,
	}, nil
}

func (s *AgentState) contextKeys() []string {
	var keys []string
	for _, r := range s.RAG {
		keys = append(keys, "rag:"+r.Query)
	}
	for name, t := range s.Tools {
		if t.Error != "" {
			keys = append(keys, "tool:"+name+":error")
		} else {
			keys = append(keys, "tool:"+name+":ok")
		}
	}
	if s.Analysis != "" {
		keys = append(keys, "analysis")
	}
	if s.Question != "" {
		keys = append(keys, "question")
	}
	keys = append(keys, "phase:"+string(s.Phase))
	return keys
}
