package ops

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// APIMetrics 进程内计数（Prometheus 文本 exposition，PoC）。
type APIMetrics struct {
	mu sync.Mutex

	reqByLabel       map[string]int64
	analyzeByResult  map[string]int64
	analyzeDurSumNs  int64
	analyzeDurCount  int64
	rateLimitReject  int64
}

func NewAPIMetrics() *APIMetrics {
	return &APIMetrics{
		reqByLabel:      map[string]int64{},
		analyzeByResult: map[string]int64{},
	}
}

func (m *APIMetrics) IncRateLimit() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.rateLimitReject++
	m.mu.Unlock()
}

func (m *APIMetrics) RecordHTTP(method, route string, status int) {
	if m == nil {
		return
	}
	key := fmt.Sprintf("%s|%s|%d", strings.ToUpper(method), route, status)
	m.mu.Lock()
	m.reqByLabel[key]++
	m.mu.Unlock()
}

func (m *APIMetrics) ObserveAnalyze(durationNs int64, ok bool) {
	if m == nil {
		return
	}
	result := "error"
	if ok {
		result = "success"
	}
	m.mu.Lock()
	m.analyzeByResult[result]++
	m.analyzeDurSumNs += durationNs
	m.analyzeDurCount++
	m.mu.Unlock()
}

// WritePrometheus 输出 text/plain; version 0.0.4 风格。
func (m *APIMetrics) WritePrometheus(w http.ResponseWriter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = fmt.Fprintln(w, "# HELP slowlog_http_requests_total HTTP requests by method, route pattern, status code.")
	_, _ = fmt.Fprintln(w, "# TYPE slowlog_http_requests_total counter")
	for k, v := range m.reqByLabel {
		method, route, code := splitReqKey(k)
		_, _ = fmt.Fprintf(w, "slowlog_http_requests_total{method=%q,route=%q,code=%q} %d\n", method, route, code, v)
	}
	_, _ = fmt.Fprintln(w, "# HELP slowlog_analyze_total V6 analyze runs by result.")
	_, _ = fmt.Fprintln(w, "# TYPE slowlog_analyze_total counter")
	for result, v := range m.analyzeByResult {
		_, _ = fmt.Fprintf(w, "slowlog_analyze_total{result=%q} %d\n", result, v)
	}
	_, _ = fmt.Fprintln(w, "# HELP slowlog_analyze_duration_seconds_sum Total analyze latency in seconds.")
	_, _ = fmt.Fprintln(w, "# TYPE slowlog_analyze_duration_seconds_sum counter")
	sumSec := float64(m.analyzeDurSumNs) / 1e9
	_, _ = fmt.Fprintf(w, "slowlog_analyze_duration_seconds_sum %.6f\n", sumSec)
	_, _ = fmt.Fprintln(w, "# HELP slowlog_analyze_duration_seconds_count Analyze run count for duration sum.")
	_, _ = fmt.Fprintln(w, "# TYPE slowlog_analyze_duration_seconds_count counter")
	_, _ = fmt.Fprintf(w, "slowlog_analyze_duration_seconds_count %d\n", m.analyzeDurCount)
	_, _ = fmt.Fprintln(w, "# HELP slowlog_rate_limit_rejected_total Requests rejected by rate limiter.")
	_, _ = fmt.Fprintln(w, "# TYPE slowlog_rate_limit_rejected_total counter")
	_, _ = fmt.Fprintf(w, "slowlog_rate_limit_rejected_total %d\n", m.rateLimitReject)
}

func splitReqKey(k string) (method, route, code string) {
	parts := strings.Split(k, "|")
	if len(parts) != 3 {
		return "GET", "unknown", "0"
	}
	return parts[0], parts[1], parts[2]
}

// RouteLabel 优先 Go 1.22+ 路由 Pattern。
func RouteLabel(r *http.Request) string {
	if r != nil && r.Pattern != "" {
		return r.Pattern
	}
	if r != nil {
		return r.URL.Path
	}
	return "unknown"
}
