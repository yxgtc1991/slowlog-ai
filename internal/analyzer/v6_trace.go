package analyzer

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// V6AgentOption 配置 V6 Agent（如开启轨迹输出）。
type V6AgentOption func(*V6AgentAnalyzer)

// WithAgentVerbose 每轮打印决策、工具/RAG 执行与上下文变化（用于理解 Agent 流程）。
func WithAgentVerbose(verbose bool) V6AgentOption {
	return func(a *V6AgentAnalyzer) {
		a.verbose = verbose
	}
}

// WithAgentRecordRounds 记录每轮完整输出，供 SaveV6RunReport 持久化。
func WithAgentRecordRounds(record bool) V6AgentOption {
	return func(a *V6AgentAnalyzer) {
		a.recordRounds = record
	}
}

// WithAgentGuide 在 Prompt 中注入推荐分析流程（见 prompt.GuidedSlowLogPreamble）。
func WithAgentGuide(guide string) V6AgentOption {
	return func(a *V6AgentAnalyzer) {
		a.extraGuide = guide
	}
}

func (a *V6AgentAnalyzer) tracef(format string, args ...interface{}) {
	if !a.verbose {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func (a *V6AgentAnalyzer) traceRoundStart(iter int) {
	a.tracef("\n%s\n", strings.Repeat("=", 62))
	a.tracef("🔄 第 %d / %d 轮  阶段=%s\n", iter, a.maxIterations, a.state.Phase)
	a.tracef("%s\n", strings.Repeat("-", 62))
}

func (a *V6AgentAnalyzer) traceDecision(decision *promptv6.AgentDecision, rawLLM string) {
	a.tracef("【LLM 决策】\n")
	if decision.CurrentState != "" {
		a.tracef("  当前状态: %s\n", decision.CurrentState)
	}
	na := decision.NextAction
	a.tracef("  行动类型: %s\n", na.Type)
	a.tracef("  理由: %s\n", na.Reasoning)
	switch na.Type {
	case promptv6.ActionCallTool:
		a.tracef("  工具: %s\n", na.ToolName)
		if len(na.ToolArgs) > 0 {
			b, _ := json.MarshalIndent(na.ToolArgs, "  ", "  ")
			a.tracef("  参数:\n  %s\n", string(b))
		}
	case promptv6.ActionRetrieveRAG:
		a.tracef("  RAG 查询: %s\n", na.RAGQuery)
	case promptv6.ActionAnalyze:
		a.traceDumpText("  分析片段", na.Analysis, 400)
	case promptv6.ActionAskQuestion:
		a.tracef("  提问: %s\n", na.Question)
	case promptv6.ActionFinish:
		a.traceDumpText("  最终结果预览", na.Result.String(), 500)
	}
	if a.verbose && rawLLM != "" {
		a.tracef("\n【LLM 原始 JSON 响应】\n")
		if len(rawLLM) > 2000 {
			a.tracef("%s\n...(截断)\n", rawLLM[:2000])
		} else {
			a.tracef("%s\n", rawLLM)
		}
	}
}

func (a *V6AgentAnalyzer) traceAfterAction(action promptv6.NextAction, toolResults map[string]interface{}, lastRAG *RAGResult) {
	a.tracef("\n【本步执行结果】\n")
	switch action.Type {
	case promptv6.ActionCallTool:
		if ent, ok := a.state.Tools[action.ToolName]; ok && ent.Error != "" {
			retry := "勿重试"
			if ent.Retryable {
				retry = "可重试"
			}
			a.tracef("  ❌ 工具 %s 失败 [%s] %s: %s\n", action.ToolName, ent.Code, retry, ent.Error)
			return
		}
		if res, ok := toolResults[action.ToolName]; ok {
			a.traceDumpJSON(fmt.Sprintf("  ✅ 工具 %s 返回", action.ToolName), res, 1200)
		}
	case promptv6.ActionRetrieveRAG:
		for i := len(a.state.RAG) - 1; i >= 0; i-- {
			r := a.state.RAG[i]
			if r.Query != action.RAGQuery {
				continue
			}
			if r.Error != "" {
				a.tracef("  ❌ RAG 失败: %s\n", r.Error)
				return
			}
			a.tracef("  ✅ RAG 命中 %d 条: %s\n", r.ChunkCount, strings.Join(r.Titles, "; "))
			return
		}
		if lastRAG != nil {
			a.tracef("  ✅ RAG 命中 %d 条\n", len(lastRAG.Chunks))
		}
	case promptv6.ActionAnalyze:
		a.traceDumpText("  写入状态 analysis", action.Analysis, 300)
	case promptv6.ActionAskQuestion:
		a.tracef("  已记录问题（演示模式不等待用户输入）\n")
	case promptv6.ActionFinish:
		a.tracef("  进入 finish，结束循环\n")
	}
	a.traceStateSummary()
}

func (a *V6AgentAnalyzer) traceStateSummary() {
	if a.state == nil {
		return
	}
	a.tracef("\n【状态机】%s\n", a.state.Phase)
	sum := a.state.PromptSummary(300)
	if sum != "" {
		a.tracef("【Prompt 摘要】\n%s\n", sum)
	}
}

func (a *V6AgentAnalyzer) traceDumpJSON(prefix string, v interface{}, max int) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		a.tracef("%s: %v\n", prefix, v)
		return
	}
	s := string(b)
	if len(s) > max {
		s = s[:max] + "\n...(截断)"
	}
	a.tracef("%s:\n%s\n", prefix, s)
}

func (a *V6AgentAnalyzer) traceDumpText(prefix, text string, max int) {
	if text == "" {
		return
	}
	if len(text) > max {
		text = text[:max] + "..."
	}
	a.tracef("%s: %s\n", prefix, text)
}

// PrintV6AgentSummary 在分析结束后输出完整轨迹（stdout）。
func PrintV6AgentSummary(result *V6AgentResult, verbose bool) {
	if result == nil {
		return
	}
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("🤖 Agent 执行轨迹：%d 轮迭代，%d 个行动，终态=%s\n",
		result.Iterations, len(result.Actions), result.FinalPhase)
	for i, action := range result.Actions {
		fmt.Printf("\n  %d. [%s] %s\n", i+1, action.Type, action.Reasoning)
		switch action.Type {
		case promptv6.ActionCallTool:
			fmt.Printf("     工具：%s\n", action.ToolName)
			if verbose && len(action.ToolArgs) > 0 {
				b, _ := json.Marshal(action.ToolArgs)
				fmt.Printf("     参数：%s\n", string(b))
			}
			if res, ok := result.ToolResults[action.ToolName]; ok && verbose {
				b, _ := json.Marshal(res)
				if len(b) > 800 {
					b = append(b[:800], []byte("...")...)
				}
				fmt.Printf("     返回：%s\n", string(b))
			}
		case promptv6.ActionRetrieveRAG:
			fmt.Printf("     查询：%s\n", action.RAGQuery)
			for _, rr := range result.RAGResults {
				if rr.Query == action.RAGQuery {
					for j, ch := range rr.Chunks {
						if m, ok := ch.(map[string]interface{}); ok {
							fmt.Printf("     知识[%d]: %v\n", j+1, m["title"])
						}
					}
				}
			}
		case promptv6.ActionAnalyze:
			text := action.Analysis
			if len(text) > 120 {
				text = text[:120] + "..."
			}
			fmt.Printf("     分析：%s\n", text)
		case promptv6.ActionAskQuestion:
			fmt.Printf("     问题：%s\n", action.Question)
		case promptv6.ActionFinish:
			fmt.Printf("     完成\n")
		}
	}
	if verbose && len(result.ConversationHistory) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("📜 对话历史（写入 Prompt 的摘要）")
		for _, line := range result.ConversationHistory {
			fmt.Printf("  · %s\n", line)
		}
	}
	if verbose && result.State != nil {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("📋 终态上下文摘要（与 Prompt 一致）")
		fmt.Println(result.State.PromptSummary(500))
	}
}
