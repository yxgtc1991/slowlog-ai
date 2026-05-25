package analyzer

import (
	promptv6 "ai_slow_log/internal/prompt/slowlog"
	"testing"
)

type stubCap struct{ name string }

func (s stubCap) Name() string        { return s.name }
func (s stubCap) Description() string { return "stub" }
func (s stubCap) InputSchema() map[string]string {
	return map[string]string{"x": "y"}
}

func TestToolsForPhase_initAll(t *testing.T) {
	t.Parallel()
	all := []promptv6.CapabilityV4{
		stubCap{"connect_mysql_instance"},
		stubCap{"explain_mysql_query"},
	}
	got := ToolsForPhase(PhaseInit, all)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestToolsForPhase_dbReadySubset(t *testing.T) {
	t.Parallel()
	all := []promptv6.CapabilityV4{
		stubCap{"connect_mysql_instance"},
		stubCap{"explain_mysql_query"},
		stubCap{"add_mysql_index"},
	}
	got := ToolsForPhase(PhaseDBReady, all)
	if len(got) != 2 {
		t.Fatalf("got %d tools", len(got))
	}
}

func TestToolsForPhase_analyzedEmpty(t *testing.T) {
	t.Parallel()
	all := []promptv6.CapabilityV4{stubCap{"connect_mysql_instance"}}
	got := ToolsForPhase(PhaseAnalyzed, all)
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
