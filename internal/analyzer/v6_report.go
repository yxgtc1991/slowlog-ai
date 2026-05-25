package analyzer

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// maxMarkdownLineLen GoLand 等对超长单行 Markdown 易渲染为空白，写入前折行。
const maxMarkdownLineLen = 240

// AgentRoundRecord 单轮 Agent 完整记录（便于复盘，无需重跑）。
type AgentRoundRecord struct {
	Round         int                 `json:"round"`
	LLMRaw        string              `json:"llm_raw"`
	CurrentState  string              `json:"current_state,omitempty"`
	AgentPhase    string              `json:"agent_phase,omitempty"`
	Action        promptv6.NextAction `json:"action"`
	ActionError   string              `json:"action_error,omitempty"`
	ActionOutcome interface{}         `json:"action_outcome,omitempty"`
	ContextKeys   []string            `json:"context_keys_after,omitempty"`
	Trace         []TraceSpan         `json:"trace,omitempty"`
}

// V6AgentRunReport 一次完整运行的可持久化报告。
type V6AgentRunReport struct {
	GeneratedAt    string                 `json:"generated_at"`
	SlowLog        string                 `json:"slow_log"`
	Iterations     int                    `json:"iterations"`
	FinalResult    string                 `json:"final_result"`
	Actions        []promptv6.NextAction  `json:"actions"`
	Rounds         []AgentRoundRecord     `json:"rounds"`
	ToolResults    map[string]interface{} `json:"tool_results"`
	RAGResults     []RAGResult            `json:"rag_results"`
	History        []string               `json:"conversation_history"`
	AvailableTools []string               `json:"available_tools,omitempty"`
	FinalPhase     string                 `json:"final_phase,omitempty"`
	AgentState     *AgentState            `json:"agent_state,omitempty"`
	Trace          *RunTrace              `json:"trace,omitempty"`
	InstanceID     string                 `json:"instance_id,omitempty"`
	RequestID      string                 `json:"request_id,omitempty"`
	Actor          string                 `json:"actor,omitempty"`
	ClientIP       string                 `json:"client_ip,omitempty"`
}

// BuildV6RunReport 从 Analyze 结果组装报告（需 Analyze 时开启 round 记录）。
func BuildV6RunReport(slowLog string, result *V6AgentResult, toolNames []string) *V6AgentRunReport {
	if result == nil {
		return nil
	}
	return &V6AgentRunReport{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		SlowLog:        slowLog,
		Iterations:     result.Iterations,
		FinalResult:    result.FinalResult,
		Actions:        result.Actions,
		Rounds:         result.Rounds,
		ToolResults:    result.ToolResults,
		RAGResults:     result.RAGResults,
		History:        result.ConversationHistory,
		AvailableTools: toolNames,
		FinalPhase:     string(result.FinalPhase),
		AgentState:     result.State,
		Trace:          result.Trace,
	}
}

// SaveV6RunReport 写入 JSON、完整报告、精简逐轮报告，返回全部路径。
func SaveV6RunReport(dir string, report *V6AgentRunReport) (*SavedReportPaths, error) {
	if report == nil {
		return nil, fmt.Errorf("report is nil")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	stamp := time.Now().Format("20060102-150405")
	base := "agent-run-" + stamp
	paths := &SavedReportPaths{
		JSON:      filepath.Join(dir, base+".json"),
		FullMD:    filepath.Join(dir, base+".md"),
		FullHTML:  filepath.Join(dir, base+".html"),
		BriefMD:   filepath.Join(dir, base+".brief.md"),
		BriefHTML: filepath.Join(dir, base+".brief.html"),
	}

	jb, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths.JSON, jb, 0o644); err != nil {
		return nil, err
	}
	jsonBase := filepath.Base(paths.JSON)
	if err := os.WriteFile(paths.FullMD, []byte(FormatV6ReportMarkdownForFile(report, jsonBase)), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths.FullHTML, []byte(FormatV6ReportHTML(report, jsonBase)), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths.BriefMD, []byte(FormatV6ReportBriefMarkdown(report)), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(paths.BriefHTML, []byte(FormatV6ReportBriefHTML(report)), 0o644); err != nil {
		return nil, err
	}
	return paths, nil
}

// FormatV6ReportMarkdownForFile 生成可在 GoLand 预览的 Markdown。
func FormatV6ReportMarkdownForFile(r *V6AgentRunReport, jsonFile string) string {
	return limitLineLengthOutsideFences(formatV6ReportMarkdown(r, jsonFile), maxMarkdownLineLen)
}

func formatV6ReportMarkdown(r *V6AgentRunReport, jsonFile string) string {
	var b strings.Builder
	b.WriteString("# V6 Agent 运行报告\n\n")
	b.WriteString("完整 LLM 原文见 `" + jsonFile + "`；逐轮速览见同名 `.brief.md`。\n\n")
	b.WriteString(fmt.Sprintf("- 生成时间：`%s`\n", mdSafeText(r.GeneratedAt)))
	b.WriteString(fmt.Sprintf("- 迭代轮数：**%d**\n", r.Iterations))
	if r.Trace != nil && r.Trace.TotalDurationMs > 0 {
		b.WriteString(fmt.Sprintf("- 总耗时：**%d ms**（见 JSON `trace.spans`）\n", r.Trace.TotalDurationMs))
	}
	if len(r.AvailableTools) > 0 {
		b.WriteString(fmt.Sprintf("- 可用 MCP：`%s`\n", strings.Join(r.AvailableTools, "`, `")))
	}
	b.WriteString("\n## 慢日志输入\n\n")
	b.WriteString(mdCodeFence("text", strings.TrimSpace(r.SlowLog)))
	b.WriteString("\n## 逐轮轨迹\n\n")
	for _, round := range r.Rounds {
		b.WriteString(fmt.Sprintf("### 第 %d 轮\n\n", round.Round))
		b.WriteString(fmt.Sprintf("**当前状态**：%s\n\n", mdSafeText(emptyFallback(round.CurrentState, "—"))))
		if round.AgentPhase != "" {
			b.WriteString(fmt.Sprintf("**状态机阶段**：`%s`\n\n", mdSafeText(round.AgentPhase)))
		}
		na := round.Action
		b.WriteString(fmt.Sprintf("**行动**：`%s`\n\n", mdSafeText(string(na.Type))))
		b.WriteString(fmt.Sprintf("**理由**：%s\n\n", mdSafeText(na.Reasoning)))
		switch na.Type {
		case promptv6.ActionCallTool:
			b.WriteString(fmt.Sprintf("- 工具：`%s`\n", na.ToolName))
			if len(na.ToolArgs) > 0 {
				args, _ := json.MarshalIndent(na.ToolArgs, "", "  ")
				b.WriteString("- 参数：\n\n")
				b.WriteString(mdCodeFence("json", string(args)))
			}
		case promptv6.ActionRetrieveRAG:
			b.WriteString(fmt.Sprintf("- RAG 查询：`%s`\n", mdSafeText(na.RAGQuery)))
		case promptv6.ActionAnalyze:
			b.WriteString("- 分析片段：\n\n")
			b.WriteString(mdCodeFence("text", truncate(na.Analysis, 800)))
		case promptv6.ActionAskQuestion:
			b.WriteString(fmt.Sprintf("- 问题：%s\n", mdSafeText(na.Question)))
		case promptv6.ActionFinish:
			b.WriteString("- 最终结果预览：\n\n")
			b.WriteString(mdCodeFence("text", truncate(escapeMarkdownFences(na.Result.String()), 500)))
		}
		if round.ActionError != "" {
			b.WriteString(fmt.Sprintf("\n**执行错误**：%s\n", mdSafeText(round.ActionError)))
		}
		if round.ActionOutcome != nil {
			b.WriteString("\n**执行结果**：\n\n")
			b.WriteString(mdCodeFence("json", formatOutcomePlain(round.ActionOutcome)))
		}
		if len(round.ContextKeys) > 0 {
			b.WriteString(fmt.Sprintf("\n**上下文键**：%s\n", strings.Join(round.ContextKeys, ", ")))
		}
		b.WriteString(fmt.Sprintf("\n*LLM 完整 JSON 见 `%s` → `rounds[%d].llm_raw`*\n\n", jsonFile, round.Round-1))
	}

	b.WriteString("## 最终结论\n\n")
	b.WriteString(mdSafeText(flattenMarkdownHeadings(escapeMarkdownFences(r.FinalResult))))
	b.WriteString("\n")
	b.WriteString("\n\n## RAG 检索汇总\n\n")
	if len(r.RAGResults) == 0 {
		b.WriteString("（无）\n")
	} else {
		for _, rag := range r.RAGResults {
			b.WriteString(fmt.Sprintf("### 查询：`%s`\n\n", mdSafeText(rag.Query)))
			for i, ch := range rag.Chunks {
				if m, ok := ch.(map[string]interface{}); ok {
					title := strings.ReplaceAll(fmt.Sprint(m["title"]), ">>", "»")
					b.WriteString(fmt.Sprintf("%d. **%s** — %s\n", i+1, mdSafeText(title), mdSafeText(truncate(fmt.Sprint(m["content"]), 200))))
				}
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("## MCP 工具结果汇总\n\n")
	if len(r.ToolResults) == 0 {
		b.WriteString("（无）\n")
	} else {
		for name, res := range r.ToolResults {
			body, _ := json.MarshalIndent(res, "", "  ")
			b.WriteString(fmt.Sprintf("### `%s`\n\n", mdSafeText(name)))
			b.WriteString(mdCodeFence("json", string(body)))
		}
	}

	b.WriteString("## 对话历史（写入 Prompt 的摘要）\n\n")
	for i, h := range r.History {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, mdSafeText(truncate(escapeMarkdownFences(h), 300))))
	}
	return b.String()
}

func emptyFallback(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func truncate(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			b.WriteString("...")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// flattenMarkdownHeadings 将结论里的 # 标题降为加粗，避免与报告标题冲突导致预览异常。
func flattenMarkdownHeadings(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#### "):
			lines[i] = "**" + strings.TrimPrefix(trimmed, "#### ") + "**"
		case strings.HasPrefix(trimmed, "### "):
			lines[i] = "**" + strings.TrimPrefix(trimmed, "### ") + "**"
		case strings.HasPrefix(trimmed, "## "):
			lines[i] = "**" + strings.TrimPrefix(trimmed, "## ") + "**"
		case strings.HasPrefix(trimmed, "# "):
			lines[i] = "**" + strings.TrimPrefix(trimmed, "# ") + "**"
		}
	}
	return strings.Join(lines, "\n")
}

// escapeMarkdownFences 避免文内 ``` 破坏 Markdown 解析（GoLand 会整页空白）。
func escapeMarkdownFences(s string) string {
	s = strings.ReplaceAll(s, "```", "'''")
	return sanitizeUTF8(s)
}

func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				b.WriteRune(' ')
				continue
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func formatOutcomePlain(out interface{}) string {
	switch v := out.(type) {
	case string:
		return sanitizeUTF8(v)
	default:
		if m, ok := v.(map[string]interface{}); ok {
			if raw, ok := m["RawOutput"].(string); ok && strings.TrimSpace(raw) != "" {
				var nested interface{}
				if err := json.Unmarshal([]byte(raw), &nested); err == nil {
					pretty, err := json.MarshalIndent(nested, "", "  ")
					if err == nil {
						return sanitizeUTF8(string(pretty))
					}
				}
			}
		}
		body, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return sanitizeUTF8(fmt.Sprint(v))
		}
		return sanitizeUTF8(string(body))
	}
}
