package analyzer

import (
	"fmt"
	"strings"
	"time"

	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

// TraceSpan 单段可观测 span（写入报告 JSON）。
type TraceSpan struct {
	Name       string            `json:"name"`
	Round      int               `json:"round"`
	Status     string            `json:"status"` // ok | error
	DurationMs int64             `json:"duration_ms"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// RunTrace 一次 Agent 运行的结构化轨迹。
type RunTrace struct {
	TotalDurationMs int64       `json:"total_duration_ms"`
	Spans           []TraceSpan `json:"spans"`
}

// NewRunTrace 创建轨迹收集器。
func NewRunTrace() *RunTrace {
	return &RunTrace{Spans: make([]TraceSpan, 0)}
}

// Finish 在 Analyze 结束时写入总耗时。
func (t *RunTrace) Finish(started time.Time) {
	if t == nil {
		return
	}
	t.TotalDurationMs = time.Since(started).Milliseconds()
}

// Begin 记录 span；结束时调用 returned 函数并传入 error（nil 表示 ok）。
func (t *RunTrace) Begin(name string, round int, attrs map[string]string) func(error) {
	if t == nil {
		return func(error) {}
	}
	start := time.Now()
	return func(err error) {
		status := "ok"
		if err != nil {
			status = "error"
		}
		t.Spans = append(t.Spans, TraceSpan{
			Name:       name,
			Round:      round,
			Status:     status,
			DurationMs: time.Since(start).Milliseconds(),
			Attributes: copyAttrs(attrs),
		})
	}
}

// SpansForRound 返回某轮的全部 span。
func (t *RunTrace) SpansForRound(round int) []TraceSpan {
	if t == nil {
		return nil
	}
	out := make([]TraceSpan, 0)
	for _, s := range t.Spans {
		if s.Round == round {
			out = append(out, s)
		}
	}
	return out
}

// RoundDurationMs 某轮 span 耗时合计。
func RoundDurationMs(spans []TraceSpan) int64 {
	var n int64
	for _, s := range spans {
		n += s.DurationMs
	}
	return n
}

// FormatRoundTiming 人类可读耗时（如 `LLM 820ms · 工具 45ms`）。
func FormatRoundTiming(spans []TraceSpan) string {
	if len(spans) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(spans))
	for _, s := range spans {
		label := spanShortLabel(s.Name)
		parts = append(parts, fmt.Sprintf("%s %dms", label, s.DurationMs))
	}
	return strings.Join(parts, " · ")
}

func spanShortLabel(name string) string {
	switch {
	case strings.HasPrefix(name, "llm."):
		return "LLM"
	case strings.HasPrefix(name, "tool."):
		return "工具"
	case strings.HasPrefix(name, "rag."):
		return "RAG"
	case strings.HasPrefix(name, "action."):
		return "行动"
	default:
		return name
	}
}

func copyAttrs(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func actionTraceAttrs(na promptv6.NextAction) map[string]string {
	attrs := map[string]string{"type": string(na.Type)}
	switch na.Type {
	case promptv6.ActionCallTool:
		attrs["tool"] = na.ToolName
	case promptv6.ActionRetrieveRAG:
		attrs["query"] = truncateText(na.RAGQuery, 80)
	}
	return attrs
}

// actionSpanName 生成 action 执行 span 名。
func actionSpanName(na promptv6.NextAction) string {
	switch na.Type {
	case promptv6.ActionCallTool:
		return "tool." + na.ToolName
	case promptv6.ActionRetrieveRAG:
		return "rag.retrieve"
	case promptv6.ActionAnalyze:
		return "action.analyze"
	case promptv6.ActionAskQuestion:
		return "action.ask_question"
	case promptv6.ActionFinish:
		return "action.finish"
	default:
		return "action." + string(na.Type)
	}
}
