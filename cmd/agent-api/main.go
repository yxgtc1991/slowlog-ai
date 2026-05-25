// agent-api 最小 HTTP 服务：POST /v1/analyze 提交慢日志，返回 report_id。
package main

import (
	"ai_slow_log/internal/config"
	"ai_slow_log/internal/llm"
	"ai_slow_log/internal/ops"
	"ai_slow_log/internal/rag"
	"ai_slow_log/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxSlowLogBytes = 512 << 10 // 512KB

func main() {
	_ = config.LoadDotEnv(".env")

	addr := getenv("SLOWLOG_API_ADDR", ":8080")
	reportDir := getenv("SLOWLOG_REPORT_DIR", "reports")
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is required")
	}

	llmClient, err := llm.NewDeepSeekClient(apiKey, "")
	if err != nil {
		log.Fatal(err)
	}

	timeout := analyzeTimeout()
	instReg, err := ops.LoadRegistry(strings.TrimSpace(os.Getenv("SLOWLOG_INSTANCES_FILE")))
	if err != nil {
		log.Fatal(err)
	}
	srv := &server{
		llm:                llmClient,
		reportDir:          reportDir,
		timeout:            timeout,
		jobs:               service.NewJobStore(),
		webhookSecret:      strings.TrimSpace(os.Getenv("SLOWLOG_WEBHOOK_SECRET")),
		ingestMinQueryTime: ingestMinQueryTimeFromEnv(),
		publicBase:         strings.TrimSpace(os.Getenv("SLOWLOG_API_PUBLIC_URL")),
		ragIndexDir:        rag.IndexDirOr(""),
		instances:          instReg,
		auditor:            ops.NewAuditor(strings.TrimSpace(os.Getenv("SLOWLOG_AUDIT_PATH"))),
		requireInstance:    strings.TrimSpace(os.Getenv("SLOWLOG_REQUIRE_INSTANCE_ID")) == "1",
		adminToken:         strings.TrimSpace(os.Getenv("SLOWLOG_ADMIN_TOKEN")),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", srv.handleHealth)
	mux.HandleFunc("GET /v1/instances", srv.handleListInstances)
	mux.HandleFunc("POST /v1/analyze", srv.handleAnalyze)
	mux.HandleFunc("POST /v1/ingest", srv.handleIngest)
	mux.HandleFunc("GET /v1/jobs/{id}", srv.handleGetJob)
	mux.HandleFunc("GET /v1/reports/{id}", srv.handleGetReport)
	mux.HandleFunc("GET /v1/reports/{id}/brief.html", srv.handleGetBriefHTML)
	mux.HandleFunc("GET /v1/rag/status", srv.handleRAGStatus)
	mux.HandleFunc("POST /v1/rag/rebuild", srv.handleRAGRebuild)

	log.Printf("agent-api listening on %s (report_dir=%s)", addr, reportDir)
	if err := http.ListenAndServe(addr, srv.wrap(mux)); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	llm                *llm.DeepSeekClient
	reportDir          string
	timeout            time.Duration
	jobs               *service.JobStore
	webhookSecret      string
	ingestMinQueryTime float64
	publicBase         string
	ragIndexDir        string
	instances          *ops.Registry
	auditor            *ops.Auditor
	requireInstance    bool
	adminToken         string
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type analyzeRequest struct {
	SlowLog string `json:"slow_log"`
	Guided  *bool  `json:"guided,omitempty"`
}

type analyzeResponse struct {
	ReportID    string `json:"report_id"`
	Iterations  int    `json:"iterations"`
	FinalResult string `json:"final_result"`
	JSONPath    string `json:"json_path"`
	BriefHTML   string `json:"brief_html_path"`
}

func (s *server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	inst, err := s.resolveInstance(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	body, err := readSlowLogBody(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	guided := true
	if strings.EqualFold(r.URL.Query().Get("guided"), "false") {
		guided = false
	}

	var req analyzeRequest
	if json.Unmarshal(body, &req) == nil && strings.TrimSpace(req.SlowLog) != "" {
		body = []byte(req.SlowLog)
		if req.Guided != nil {
			guided = *req.Guided
		}
	}

	s.audit("analyze", "started", inst.ID, "", "", r)

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	result, err := service.RunV6(ctx, s.llm, service.RunV6Config{
		SlowLog:        string(body),
		ReportDir:      s.reportDir,
		Guided:         guided,
		HITL:           false,
		AnalyzeTimeout: s.timeout,
		Meta:           s.runMeta(r, inst.ID),
	})
	if err != nil {
		s.audit("analyze", "failed", inst.ID, "", err.Error(), r)
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	s.audit("analyze", "ok", inst.ID, result.ReportID, "", r)
	writeJSON(w, http.StatusOK, analyzeResponse{
		ReportID:    result.ReportID,
		Iterations:  result.Iterations,
		FinalResult: result.FinalResult,
		JSONPath:    result.Paths.JSON,
		BriefHTML:   result.Paths.BriefHTML,
	})
}

func (s *server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || strings.Contains(id, "..") {
		writeErr(w, http.StatusBadRequest, errors.New("invalid report id"))
		return
	}
	path := filepath.Join(s.reportDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeErr(w, http.StatusNotFound, err)
			return
		}
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *server) handleGetBriefHTML(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || strings.Contains(id, "..") {
		writeErr(w, http.StatusBadRequest, errors.New("invalid report id"))
		return
	}
	path := filepath.Join(s.reportDir, id+".brief.html")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeErr(w, http.StatusNotFound, err)
			return
		}
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func readSlowLogBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	limited := io.LimitReader(r.Body, maxSlowLogBytes+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("empty body")
	}
	if len(b) > maxSlowLogBytes {
		return nil, fmt.Errorf("body exceeds %d bytes", maxSlowLogBytes)
	}
	return b, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func analyzeTimeout() time.Duration {
	if v := os.Getenv("SLOWLOG_ANALYZE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return 15 * time.Minute
}
