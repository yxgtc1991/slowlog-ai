package analyzer

import (
	"strings"
	"testing"

	promptv6 "ai_slow_log/internal/prompt/slowlog"
)

func TestMdSafeText_escapesNilAngle(t *testing.T) {
	t.Parallel()
	got := mdSafeText("key=<nil> Extra=Using filesort")
	if strings.Contains(got, "<nil") {
		t.Fatalf("raw angle brackets remain: %q", got)
	}
	if !strings.Contains(got, "&lt;") {
		t.Fatalf("want escaped lt: %q", got)
	}
}

func TestMdTableCell_noPipeOrNewline(t *testing.T) {
	t.Parallel()
	got := mdTableCell("a|b\nc")
	if strings.Contains(got, "|") || strings.Contains(got, "\n") {
		t.Fatalf("got %q", got)
	}
}

func TestLimitLineLengthOutsideFences_preservesTableRow(t *testing.T) {
	t.Parallel()
	in := "| 1 | " + strings.Repeat("x", 300) + " |\n"
	got := limitLineLengthOutsideFences(in, 50)
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("table row split across %d lines: %q", len(lines), got)
	}
}

func TestLimitLineLengthOutsideFences_preservesCodeFence(t *testing.T) {
	t.Parallel()
	in := "short\n```json\n{\"key\":\"" + strings.Repeat("x", 300) + "\"}\n```\n"
	got := limitLineLengthOutsideFences(in, 50)
	if !strings.Contains(got, strings.Repeat("x", 300)) {
		t.Fatal("fence body should stay intact")
	}
}

func TestFormatV6ReportBriefMarkdown_noRawAngleBrackets(t *testing.T) {
	t.Parallel()
	md := FormatV6ReportBriefMarkdown(&V6AgentRunReport{
		GeneratedAt: "t",
		Iterations:  1,
		FinalResult: "ok",
		Rounds: []AgentRoundRecord{{
			Round: 1,
			Action: promptv6.NextAction{
				Type:      promptv6.ActionCallTool,
				ToolName:  "explain_mysql_query",
				Reasoning: "test",
			},
			ActionOutcome: map[string]interface{}{
				"rows": []interface{}{
					map[string]interface{}{"type": "ALL", "key": nil, "Extra": "filesort"},
				},
			},
		}},
	})
	if strings.Contains(md, "<nil") {
		t.Fatalf("brief md must not contain <nil>: %s", md)
	}
}
