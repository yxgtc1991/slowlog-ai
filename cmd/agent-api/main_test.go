package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	t.Parallel()
	s := &server{reportDir: t.TempDir()}
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	s.handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandleAnalyze_emptyBody(t *testing.T) {
	t.Parallel()
	s := &server{reportDir: t.TempDir(), timeout: 0}
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
	s := &server{reportDir: dir}
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
	s := &server{reportDir: dir}
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
