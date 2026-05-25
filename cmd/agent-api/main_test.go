package main

import (
	"ai_slow_log/internal/ops"
	"ai_slow_log/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func testServer(t *testing.T) *server {
	t.Helper()
	reg, err := ops.LoadRegistry("")
	if err != nil {
		t.Fatal(err)
	}
	return &server{reportDir: t.TempDir(), instances: reg}
}

func TestHandleHealth(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	s.handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandleAnalyze_emptyBody(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.timeout = 0
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", nil)
	rec := httptest.NewRecorder()
	s.handleAnalyze(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleGetReport_notFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := testServer(t)
	s.reportDir = dir
	req := httptest.NewRequest(http.MethodGet, "/v1/reports/agent-run-missing", nil)
	req.SetPathValue("id", "agent-run-missing")
	rec := httptest.NewRecorder()
	s.handleGetReport(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandleGetReport_ok(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	id := "agent-run-test-001"
	if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s := testServer(t)
	s.reportDir = dir
	req := httptest.NewRequest(http.MethodGet, "/v1/reports/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	s.handleGetReport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestReadSlowLogBody_json(t *testing.T) {
	t.Parallel()
	body := bytes.NewBufferString(`{"slow_log":"# Time: 1\nSELECT 1"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", body)
	b, err := readSlowLogBody(req)
	if err != nil || !bytes.Contains(b, []byte("SELECT 1")) {
		t.Fatalf("b=%s err=%v", b, err)
	}
}

func TestHandleIngest_unauthorized(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.webhookSecret = "secret"
	body := []byte(`{"slow_log":"# Query_time: 2.0\nSELECT 1"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleIngest(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandleIngest_skipThreshold(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.ingestMinQueryTime = 1.0
	s.jobs = service.NewJobStore()
	body := []byte(`{"slow_log":"# Query_time: 0.01\nSELECT 1","async":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.handleIngest(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rec.Code, rec.Body.String())
	}
	var resp ingestResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Status != "skipped" {
		t.Fatalf("%+v", resp)
	}
}

func TestHandleGetJob_ok(t *testing.T) {
	t.Parallel()
	store := service.NewJobStore()
	store.Create("job-x", "fb")
	s := testServer(t)
	s.jobs = store
	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/job-x", nil)
	req.SetPathValue("id", "job-x")
	rec := httptest.NewRecorder()
	s.handleGetJob(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestAPIKey_blocksWithoutKey(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.apiKey = "test-api-key"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/instances", s.handleListInstances)
	req := httptest.NewRequest(http.MethodGet, "/v1/instances", nil)
	rec := httptest.NewRecorder()
	s.wrap(mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKey_acceptsBearer(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.apiKey = "test-api-key"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/instances", s.handleListInstances)
	req := httptest.NewRequest(http.MethodGet, "/v1/instances", nil)
	req.Header.Set("Authorization", "Bearer test-api-key")
	rec := httptest.NewRecorder()
	s.wrap(mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKey_healthPublic(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.apiKey = "test-api-key"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.handleHealth)
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	s.wrap(mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestMetrics_prometheus(t *testing.T) {
	t.Parallel()
	s := testServer(t)
	s.apiKey = "k"
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("X-API-Key", "k")
	rec := httptest.NewRecorder()
	s.wrap(mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("slowlog_http_requests_total")) {
		t.Fatalf("body=%s", rec.Body.Bytes())
	}
}

func TestAnalyzeResponse_shape(t *testing.T) {
	t.Parallel()
	var resp analyzeResponse
	raw := `{"report_id":"agent-run-x","iterations":3,"final_result":"ok"}`
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ReportID != "agent-run-x" {
		t.Fatalf("%+v", resp)
	}
}
