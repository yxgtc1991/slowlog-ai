package analyzer

import (
	"fmt"
	"html"
	"strings"

	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

// FormatV6ReportHTML 生成可在浏览器 / GoLand 中直接预览的 HTML 报告。
func FormatV6ReportHTML(r *V6AgentRunReport, jsonFile string) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>V6 Agent 运行报告</title>\n")
	b.WriteString(`<style>
:root { color-scheme: light dark; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; max-width: 920px; margin: 2rem auto; padding: 0 1.25rem; line-height: 1.6; }
@media (prefers-color-scheme: dark) {
  body { background: #1e1e1e; color: #e0e0e0; }
  pre, code { background: #2b2b2b; color: #d4d4d4; }
  a { color: #6cb6ff; }
}
@media (prefers-color-scheme: light) {
  body { background: #fff; color: #222; }
  pre, code { background: #f6f8fa; color: #24292f; }
  a { color: #0969da; }
}
h1 { font-size: 1.5rem; border-bottom: 1px solid #8884; padding-bottom: 0.3em; }
h2 { font-size: 1.2rem; margin-top: 1.6em; }
h3 { font-size: 1.05rem; }
pre { padding: 1em; overflow-x: auto; border-radius: 6px; white-space: pre-wrap; word-break: break-word; }
.meta { color: #888; font-size: 0.9rem; }
.round { border-left: 3px solid #4a9eff88; padding-left: 1em; margin: 1.2em 0; }
.note { background: #4a9eff18; padding: 0.75em 1em; border-radius: 6px; margin: 1em 0; }
.conclusion { white-space: pre-wrap; }
</style>
</head>
<body>
`)
	b.WriteString("<h1>V6 Agent 运行报告</h1>\n")
	b.WriteString("<p class=\"note\">完整 LLM 原文见 <code>")
	b.WriteString(html.EscapeString(jsonFile))
	b.WriteString("</code> 中 <code>rounds[].llm_raw</code>。本页为可视化报告，不依赖 GoLand Markdown 预览。</p>\n")
	b.WriteString("<ul class=\"meta\">\n")
	b.WriteString(fmt.Sprintf("<li>生成时间：%s</li>\n", html.EscapeString(r.GeneratedAt)))
	b.WriteString(fmt.Sprintf("<li>迭代轮数：%d</li>\n", r.Iterations))
	if len(r.AvailableTools) > 0 {
		b.WriteString(fmt.Sprintf("<li>可用 MCP：%s</li>\n", html.EscapeString(strings.Join(r.AvailableTools, ", "))))
	}
	b.WriteString("</ul>\n")

	b.WriteString("<h2>慢日志输入</h2>\n<pre>")
	b.WriteString(html.EscapeString(strings.TrimSpace(r.SlowLog)))
	b.WriteString("</pre>\n")

	b.WriteString("<h2>逐轮轨迹</h2>\n")
	for _, round := range r.Rounds {
		b.WriteString(fmt.Sprintf("<section class=\"round\"><h3>第 %d 轮</h3>\n", round.Round))
		b.WriteString("<p><strong>当前状态</strong>：")
		b.WriteString(html.EscapeString(emptyFallback(round.CurrentState, "—")))
		b.WriteString("</p>\n")
		na := round.Action
		b.WriteString(fmt.Sprintf("<p><strong>行动</strong>：<code>%s</code></p>\n", html.EscapeString(string(na.Type))))
		b.WriteString("<p><strong>理由</strong>：")
		b.WriteString(html.EscapeString(na.Reasoning))
		b.WriteString("</p>\n")
		switch na.Type {
		case promptv6.ActionCallTool:
			b.WriteString(fmt.Sprintf("<p>工具：<code>%s</code></p>\n", html.EscapeString(na.ToolName)))
		case promptv6.ActionRetrieveRAG:
			b.WriteString("<p>RAG 查询：")
			b.WriteString(html.EscapeString(na.RAGQuery))
			b.WriteString("</p>\n")
		case promptv6.ActionAnalyze:
			b.WriteString("<p><strong>分析</strong></p><pre>")
			b.WriteString(html.EscapeString(na.Analysis))
			b.WriteString("</pre>\n")
		case promptv6.ActionFinish:
			b.WriteString("<p><strong>结果预览</strong></p><pre class=\"conclusion\">")
			b.WriteString(html.EscapeString(truncate(na.Result.String(), 800)))
			b.WriteString("</pre>\n")
		}
		if round.ActionError != "" {
			b.WriteString("<p><strong>错误</strong>：")
			b.WriteString(html.EscapeString(round.ActionError))
			b.WriteString("</p>\n")
		}
		if round.ActionOutcome != nil {
			b.WriteString("<p><strong>执行结果</strong></p><pre>")
			b.WriteString(html.EscapeString(formatOutcomePlain(round.ActionOutcome)))
			b.WriteString("</pre>\n")
		}
		b.WriteString(fmt.Sprintf("<p class=\"meta\">LLM JSON → rounds[%d].llm_raw</p></section>\n", round.Round-1))
	}

	b.WriteString("<h2>最终结论</h2>\n<pre class=\"conclusion\">")
	b.WriteString(html.EscapeString(r.FinalResult))
	b.WriteString("</pre>\n")

	b.WriteString("<h2>RAG 检索汇总</h2>\n")
	if len(r.RAGResults) == 0 {
		b.WriteString("<p>（无）</p>\n")
	} else {
		for _, rag := range r.RAGResults {
			b.WriteString("<h3>")
			b.WriteString(html.EscapeString(rag.Query))
			b.WriteString("</h3><ul>\n")
			for _, ch := range rag.Chunks {
				if m, ok := ch.(map[string]interface{}); ok {
					b.WriteString("<li><strong>")
					b.WriteString(html.EscapeString(fmt.Sprint(m["title"])))
					b.WriteString("</strong> — ")
					b.WriteString(html.EscapeString(fmt.Sprint(m["content"])))
					b.WriteString("</li>\n")
				}
			}
			b.WriteString("</ul>\n")
		}
	}

	b.WriteString("<h2>MCP 工具结果汇总</h2>\n")
	for name, res := range r.ToolResults {
		body := formatOutcomePlain(res)
		b.WriteString("<h3>")
		b.WriteString(html.EscapeString(name))
		b.WriteString("</h3><pre>")
		b.WriteString(html.EscapeString(body))
		b.WriteString("</pre>\n")
	}

	b.WriteString("</body></html>\n")
	return b.String()
}
