package ops

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIMetrics_prometheus(t *testing.T) {
	t.Parallel()
	m := NewAPIMetrics()
	m.RecordHTTP("POST", "/v1/analyze", 200)
	m.ObserveAnalyze(2e9, true)
	m.IncRateLimit()
	rec := httptest.NewRecorder()
	m.WritePrometheus(rec)
	body := rec.Body.String()
	for _, sub := range []string{
		"slowlog_http_requests_total",
		"slowlog_analyze_total",
		"slowlog_analyze_duration_seconds_sum",
		"slowlog_rate_limit_rejected_total 1",
	} {
		if !strings.Contains(body, sub) {
			t.Fatalf("missing %q in:\n%s", sub, body)
		}
	}
}
