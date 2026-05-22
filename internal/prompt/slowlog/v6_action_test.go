package prompt

import (
	"strings"
	"testing"
)

func TestNormalizeNextAction_toolNameAsType(t *testing.T) {
	t.Parallel()
	na := NextAction{Type: "analyze_slow_log", ToolArgs: map[string]interface{}{"slow_log": "x"}}
	normalizeNextAction(&na)
	if na.Type != ActionCallTool {
		t.Fatalf("type=%s want call_tool", na.Type)
	}
	if na.ToolName != "analyze_slow_log" {
		t.Fatalf("tool_name=%s", na.ToolName)
	}
}

func TestParseAgentDecision_finishResultObject(t *testing.T) {
	t.Parallel()
	raw := `{
  "current_state": "done",
  "next_action": {
    "type": "finish",
    "reasoning": "all info collected",
    "result": {
      "diagnosis": "全表扫描",
      "index": "(price, created_at)"
    }
  }
}`
	d, err := ParseAgentDecision(raw)
	if err != nil {
		t.Fatal(err)
	}
	if d.NextAction.Type != ActionFinish {
		t.Fatalf("type=%s", d.NextAction.Type)
	}
	if d.NextAction.Result.String() == "" {
		t.Fatal("expected non-empty result")
	}
	if !strings.Contains(d.NextAction.Result.String(), "全表扫描") {
		t.Fatalf("result=%s", d.NextAction.Result)
	}
}
