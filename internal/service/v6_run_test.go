package service

import "testing"

func TestReportIDFromJSONPath(t *testing.T) {
	t.Parallel()
	id := ReportIDFromJSONPath("reports/agent-run-20260525-181138.json")
	if id != "agent-run-20260525-181138" {
		t.Fatalf("got %q", id)
	}
}

func TestRunV6_emptySlowLog(t *testing.T) {
	t.Parallel()
	_, err := RunV6(t.Context(), nil, RunV6Config{SlowLog: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
}
