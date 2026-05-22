package analyzer

import (
	"strings"
	"testing"

	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

func TestFormatV6ReportBriefMarkdown_table(t *testing.T) {
	t.Parallel()
	md := FormatV6ReportBriefMarkdown(&V6AgentRunReport{
		GeneratedAt: "2026-01-01T00:00:00Z",
		Iterations:  2,
		FinalResult: "## 结论\n根因是索引未命中",
		Rounds: []AgentRoundRecord{
			{
				Round: 1,
				Action: promptv6.NextAction{
					Type:      promptv6.ActionRetrieveRAG,
					Reasoning: "需要确认左前缀原则",
					RAGQuery:  "全表扫描",
				},
				ActionOutcome: map[string]interface{}{
					"Chunks": []interface{}{
						map[string]interface{}{"title": "Rows_examined >> Rows_sent"},
					},
				},
			},
			{
				Round: 2,
				Action: promptv6.NextAction{
					Type:      promptv6.ActionFinish,
					Reasoning: "信息已够",
					Result:    "建议加索引 (price, created_at)",
				},
			},
		},
	})
	if !strings.Contains(md, "| 轮次 | 做了什么 | 为什么 | 结果 |") {
		t.Fatal("want summary table")
	}
	if !strings.Contains(md, "检索知识库") {
		t.Fatal("want rag action")
	}
	if !strings.Contains(md, "结论摘要") {
		t.Fatal("want conclusion section")
	}
}

func TestBriefToolOutcome_explain(t *testing.T) {
	t.Parallel()
	got := briefToolOutcome(&AgentRoundRecord{
		Action: promptv6.NextAction{ToolName: "explain_mysql_query"},
		ActionOutcome: map[string]interface{}{
			"rows": []interface{}{
				map[string]interface{}{"type": "ALL", "key": nil, "Extra": "Using filesort"},
			},
		},
	})
	if !strings.Contains(got, "ALL") || !strings.Contains(got, "filesort") {
		t.Fatalf("got %q", got)
	}
}
