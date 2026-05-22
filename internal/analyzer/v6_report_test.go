package analyzer

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncate_preservesValidUTF8(t *testing.T) {
	t.Parallel()
	s := "综合现有信息分析如下：扫描行数远大于返回行数"
	got := truncate(s, 8)
	if !utf8.ValidString(got) {
		t.Fatalf("invalid utf8: %q", got)
	}
}

func TestEscapeMarkdownFences(t *testing.T) {
	t.Parallel()
	got := escapeMarkdownFences("见 ```sql\nSELECT 1")
	if strings.Contains(got, "```") {
		t.Fatalf("still has fence: %q", got)
	}
}

func TestLimitLineLengthOutsideFences_plain(t *testing.T) {
	t.Parallel()
	in := strings.Repeat("a", 300)
	got := limitLineLengthOutsideFences(in, 100)
	for _, line := range strings.Split(got, "\n") {
		if utf8.RuneCountInString(line) > 100 {
			t.Fatalf("line too long: %d runes", utf8.RuneCountInString(line))
		}
	}
}

func TestFormatOutcomePlain_rawOutput(t *testing.T) {
	t.Parallel()
	out := map[string]interface{}{
		"RawOutput": `{"summary":"ok","metrics":{"rows":1}}`,
	}
	got := formatOutcomePlain(out)
	if !strings.Contains(got, "summary") {
		t.Fatalf("want parsed json, got %s", got)
	}
}

func TestFormatV6ReportMarkdownForFile_usesFencesNotDetails(t *testing.T) {
	t.Parallel()
	md := FormatV6ReportMarkdownForFile(&V6AgentRunReport{
		GeneratedAt: "2026-01-01T00:00:00Z",
		SlowLog:     "SELECT 1",
		FinalResult: "## 结论\n```sql\nSELECT 1;\n```",
	}, "run.json")
	if strings.Contains(md, "<details>") {
		t.Fatal("should not contain details")
	}
	if !strings.Contains(md, "```text") {
		t.Fatal("should use fenced code blocks")
	}
	if strings.Contains(md, "<nil") {
		t.Fatal("should not contain raw angle nil")
	}
}
