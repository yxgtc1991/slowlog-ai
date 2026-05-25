package analyzer

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunTrace_BeginRecordsDuration(t *testing.T) {
	t.Parallel()
	tr := NewRunTrace()
	end := tr.Begin("llm.chat", 1, nil)
	time.Sleep(2 * time.Millisecond)
	end(nil)
	if len(tr.Spans) != 1 {
		t.Fatalf("spans=%d", len(tr.Spans))
	}
	if tr.Spans[0].DurationMs < 1 || tr.Spans[0].Status != "ok" {
		t.Fatalf("%+v", tr.Spans[0])
	}
}

func TestRunTrace_errorStatus(t *testing.T) {
	t.Parallel()
	tr := NewRunTrace()
	end := tr.Begin("tool.explain_mysql_query", 2, map[string]string{"tool": "explain_mysql_query"})
	end(errors.New("fail"))
	if tr.Spans[0].Status != "error" {
		t.Fatalf("status=%s", tr.Spans[0].Status)
	}
}

func TestFormatRoundTiming(t *testing.T) {
	t.Parallel()
	s := FormatRoundTiming([]TraceSpan{
		{Name: "llm.chat", DurationMs: 100},
		{Name: "tool.connect_mysql_instance", DurationMs: 20},
	})
	if s == "—" || !strings.Contains(s, "100") || !strings.Contains(s, "20") {
		t.Fatalf("got %q", s)
	}
}
