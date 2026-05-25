package analyzer

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

const briefReasoningMax = 160
const briefOutcomeMax = 200
const briefConclusionMax = 600

// SavedReportPaths 一次运行写入的全部报告路径。
type SavedReportPaths struct {
	JSON      string
	FullMD    string
	FullHTML  string
	BriefMD   string
	BriefHTML string
}

// FormatV6ReportBriefMarkdown 逐轮精简版（给客户 / 快速复盘）。
func FormatV6ReportBriefMarkdown(r *V6AgentRunReport) string {
	var b strings.Builder
	b.WriteString("# Agent 分析过程一览\n\n")
	b.WriteString(fmt.Sprintf("- 生成时间：%s\n", r.GeneratedAt))
	b.WriteString(fmt.Sprintf("- 总轮数：**%d**\n", r.Iterations))
	if r.Trace != nil && r.Trace.TotalDurationMs > 0 {
		b.WriteString(fmt.Sprintf("- 总耗时：**%d ms**\n", r.Trace.TotalDurationMs))
	}
	b.WriteString("\n")

	b.WriteString("## 结论摘要\n\n")
	b.WriteString(mdCodeFence("text", truncate(flattenMarkdownHeadings(escapeMarkdownFences(r.FinalResult)), briefConclusionMax)))
	b.WriteString("\n## 逐轮一览\n\n")
	b.WriteString("| 轮次 | 耗时 | 做了什么 | 为什么 | 结果 |\n")
	b.WriteString("|:--:|------|----------|--------|------|\n")
	for _, round := range r.Rounds {
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			round.Round,
			mdTableCell(briefRoundTiming(&round)),
			mdTableCell(briefWhatDid(&round)),
			mdTableCell(briefWhy(&round)),
			mdTableCell(briefOutcomeLine(&round)),
		))
	}
	b.WriteString("\n## 逐轮详情\n\n")
	for _, round := range r.Rounds {
		b.WriteString(formatBriefRoundSection(&round))
	}
	return limitLineLengthOutsideFences(b.String(), maxMarkdownLineLen)
}

func formatBriefRoundSection(round *AgentRoundRecord) string {
	var b strings.Builder
	na := round.Action
	b.WriteString(fmt.Sprintf("### 第 %d 轮 · %s\n\n", round.Round, briefActionTitle(&na)))
	if s := strings.TrimSpace(round.CurrentState); s != "" {
		b.WriteString(fmt.Sprintf("- **当时状态**：%s\n", mdSafeText(truncate(s, 120))))
	}
	b.WriteString(fmt.Sprintf("- **做了什么**：%s\n", mdSafeText(briefWhatDid(round))))
	b.WriteString(fmt.Sprintf("- **为什么**：%s\n", mdSafeText(briefWhy(round))))
	if timing := FormatRoundTiming(round.Trace); timing != "—" {
		b.WriteString(fmt.Sprintf("- **耗时**：%s\n", mdSafeText(timing)))
	}
	if round.ActionError != "" {
		b.WriteString(fmt.Sprintf("- **错误**：%s\n", mdSafeText(round.ActionError)))
	} else {
		b.WriteString(fmt.Sprintf("- **结果**：%s\n", mdSafeText(briefOutcomeLine(round))))
	}
	return b.String() + "\n"
}

func briefActionTitle(na *promptv6.NextAction) string {
	switch na.Type {
	case promptv6.ActionCallTool:
		return "调用工具 · " + na.ToolName
	case promptv6.ActionRetrieveRAG:
		return "检索知识库"
	case promptv6.ActionAnalyze:
		return "归纳分析"
	case promptv6.ActionAskQuestion:
		return "向用户提问"
	case promptv6.ActionFinish:
		return "完成 · 输出结论"
	default:
		return string(na.Type)
	}
}

func briefRoundTiming(round *AgentRoundRecord) string {
	if d := RoundDurationMs(round.Trace); d > 0 {
		return fmt.Sprintf("%dms", d)
	}
	return "—"
}

func briefWhatDid(round *AgentRoundRecord) string {
	na := round.Action
	switch na.Type {
	case promptv6.ActionCallTool:
		return fmt.Sprintf("调用 `%s`", na.ToolName)
	case promptv6.ActionRetrieveRAG:
		return fmt.Sprintf("RAG 检索「%s」", truncate(na.RAGQuery, 60))
	case promptv6.ActionAnalyze:
		return "基于已有信息归纳根因"
	case promptv6.ActionAskQuestion:
		return fmt.Sprintf("提问：%s", truncate(na.Question, 60))
	case promptv6.ActionFinish:
		return "输出最终诊断报告"
	default:
		return string(na.Type)
	}
}

func briefWhy(round *AgentRoundRecord) string {
	return truncate(strings.TrimSpace(round.Action.Reasoning), briefReasoningMax)
}

func briefOutcomeLine(round *AgentRoundRecord) string {
	if round.ActionError != "" {
		return "失败：" + truncate(round.ActionError, briefOutcomeMax)
	}
	na := round.Action
	switch na.Type {
	case promptv6.ActionCallTool:
		return truncate(briefToolOutcome(round), briefOutcomeMax)
	case promptv6.ActionRetrieveRAG:
		return truncate(briefRAGOutcome(round.ActionOutcome), briefOutcomeMax)
	case promptv6.ActionAnalyze:
		first := strings.TrimSpace(na.Analysis)
		if idx := strings.IndexAny(first, "\n"); idx > 0 {
			first = first[:idx]
		}
		return truncate(first, briefOutcomeMax)
	case promptv6.ActionFinish:
		return truncate(oneLinePlain(na.Result.String()), briefOutcomeMax)
	case promptv6.ActionAskQuestion:
		return "等待用户补充信息"
	default:
		return "—"
	}
}

func briefToolOutcome(round *AgentRoundRecord) string {
	na := round.Action
	switch na.ToolName {
	case "connect_mysql_instance":
		if m, ok := round.ActionOutcome.(map[string]interface{}); ok {
			if v, ok := m["current_database"].(string); ok {
				return fmt.Sprintf("已连接，当前库 %s", v)
			}
		}
		return "MySQL 连接成功"
	case "explain_mysql_query":
		if m, ok := round.ActionOutcome.(map[string]interface{}); ok {
			if rows, ok := m["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					return fmt.Sprintf("type=%s key=%s Extra=%s",
						formatGoValueForMD(row["type"]),
						formatGoValueForMD(row["key"]),
						formatGoValueForMD(row["Extra"]))
				}
			}
		}
		return "已获取 EXPLAIN 执行计划"
	case "add_mysql_index":
		if m, ok := round.ActionOutcome.(map[string]interface{}); ok {
			if ddl, ok := m["ddl"].(string); ok {
				return "dry_run DDL: " + ddl
			}
		}
		return "已生成索引 DDL（dry_run）"
	case "analyze_slow_log":
		if m, ok := round.ActionOutcome.(map[string]interface{}); ok {
			if raw, ok := m["RawOutput"].(string); ok {
				var nested map[string]interface{}
				if json.Unmarshal([]byte(raw), &nested) == nil {
					if s, ok := nested["summary"].(string); ok {
						return s
					}
				}
			}
		}
		return "慢日志结构化分析完成"
	default:
		return "工具执行成功"
	}
}

func oneLinePlain(s string) string {
	s = flattenMarkdownHeadings(escapeMarkdownFences(s))
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func briefRAGOutcome(out interface{}) string {
	m, ok := out.(map[string]interface{})
	if !ok {
		return "检索完成"
	}
	chunks, ok := m["Chunks"].([]interface{})
	if !ok || len(chunks) == 0 {
		return "未命中知识块"
	}
	titles := make([]string, 0, len(chunks))
	for _, c := range chunks {
		if cm, ok := c.(map[string]interface{}); ok {
			if t, ok := cm["title"].(string); ok && t != "" {
				titles = append(titles, t)
			}
		}
	}
	if len(titles) == 0 {
		return fmt.Sprintf("命中 %d 条知识", len(chunks))
	}
	return fmt.Sprintf("命中 %d 条：%s", len(chunks), strings.Join(titles, "；"))
}

// FormatV6ReportBriefHTML 精简版 HTML（适合浏览器给客户看）。
func FormatV6ReportBriefHTML(r *V6AgentRunReport) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>Agent 分析过程一览</title>\n")
	b.WriteString(`<style>
:root { color-scheme: light dark; }
body { font-family: -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; max-width: 960px; margin: 1.5rem auto; padding: 0 1rem; line-height: 1.55; }
@media (prefers-color-scheme: dark) { body { background: #1a1a1a; color: #e8e8e8; } th { background: #2a2a2a; } td, th { border-color: #444; } .card { background: #242424; } .conclusion { background: #2a2a2a; } }
@media (prefers-color-scheme: light) { body { background: #fafafa; color: #222; } th { background: #eef3f8; } .card { background: #fff; } .conclusion { background: #f0f4f8; } }
h1 { font-size: 1.4rem; margin-bottom: 0.25rem; }
.meta { color: #666; font-size: 0.9rem; }
table { width: 100%; border-collapse: collapse; font-size: 0.92rem; margin: 1rem 0; }
th, td { border: 1px solid #ccc; padding: 0.5rem 0.65rem; vertical-align: top; }
th { text-align: left; }
td:first-child { text-align: center; white-space: nowrap; width: 3rem; }
.card { border-radius: 8px; padding: 1rem 1.1rem; margin: 1rem 0; box-shadow: 0 1px 3px #0002; }
.card h3 { margin: 0 0 0.6rem; font-size: 1.05rem; color: #2b6cb0; }
.card ul { margin: 0; padding-left: 1.2rem; }
.card li { margin: 0.35rem 0; }
.conclusion { padding: 1rem; border-radius: 8px; white-space: pre-wrap; margin: 0.5rem 0 1.5rem; }
.tag { display: inline-block; background: #2b6cb022; color: #2b6cb0; padding: 0.1em 0.5em; border-radius: 4px; font-size: 0.85rem; }
</style></head><body>`)
	b.WriteString("<h1>Agent 分析过程一览</h1>\n")
	meta := fmt.Sprintf("生成时间：%s · 共 %d 轮", html.EscapeString(r.GeneratedAt), r.Iterations)
	if r.Trace != nil && r.Trace.TotalDurationMs > 0 {
		meta += fmt.Sprintf(" · 总耗时 %d ms", r.Trace.TotalDurationMs)
	}
	b.WriteString("<p class=\"meta\">" + meta + "</p>\n")
	b.WriteString("<h2>结论摘要</h2>\n<div class=\"conclusion\">")
	b.WriteString(html.EscapeString(truncate(r.FinalResult, briefConclusionMax)))
	b.WriteString("</div>\n<h2>逐轮一览</h2>\n<table><thead><tr><th>轮次</th><th>耗时</th><th>做了什么</th><th>为什么</th><th>结果</th></tr></thead><tbody>\n")
	for _, round := range r.Rounds {
		b.WriteString("<tr><td>")
		b.WriteString(fmt.Sprintf("%d", round.Round))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(briefRoundTiming(&round)))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(briefWhatDid(&round)))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(briefWhy(&round)))
		b.WriteString("</td><td>")
		b.WriteString(html.EscapeString(briefOutcomeLine(&round)))
		b.WriteString("</td></tr>\n")
	}
	b.WriteString("</tbody></table>\n<h2>逐轮详情</h2>\n")
	for _, round := range r.Rounds {
		na := round.Action
		b.WriteString("<div class=\"card\"><h3>")
		b.WriteString(fmt.Sprintf("第 %d 轮 · <span class=\"tag\">%s</span>", round.Round, html.EscapeString(briefActionTitle(&na))))
		b.WriteString("</h3><ul>\n")
		if s := strings.TrimSpace(round.CurrentState); s != "" {
			b.WriteString("<li><strong>当时状态</strong>：")
			b.WriteString(html.EscapeString(truncate(s, 120)))
			b.WriteString("</li>\n")
		}
		b.WriteString("<li><strong>做了什么</strong>：")
		b.WriteString(html.EscapeString(briefWhatDid(&round)))
		b.WriteString("</li><li><strong>为什么</strong>：")
		b.WriteString(html.EscapeString(briefWhy(&round)))
		if timing := FormatRoundTiming(round.Trace); timing != "—" {
			b.WriteString("</li><li><strong>耗时</strong>：")
			b.WriteString(html.EscapeString(timing))
		}
		b.WriteString("</li><li><strong>结果</strong>：")
		if round.ActionError != "" {
			b.WriteString(html.EscapeString("失败：" + round.ActionError))
		} else {
			b.WriteString(html.EscapeString(briefOutcomeLine(&round)))
		}
		b.WriteString("</li></ul></div>\n")
	}
	b.WriteString("</body></html>\n")
	return b.String()
}
