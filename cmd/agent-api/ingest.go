package main

import (
	"ai_slow_log/internal/service"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ingestResponse struct {
	JobID       string `json:"job_id,omitempty"`
	ReportID    string `json:"report_id,omitempty"`
	Status      string `json:"status"`
	Iterations  int    `json:"iterations,omitempty"`
	FinalResult string `json:"final_result,omitempty"`
	SkipReason  string `json:"skip_reason,omitempty"`
	ReportURL   string `json:"report_url,omitempty"`
}

func (s *server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if !s.checkWebhookSecret(r) {
		writeErr(w, http.StatusUnauthorized, errUnauthorized)
		return
	}

	body, err := readBodyLimited(r, maxSlowLogBytes)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	req, err := service.ParseIngestRequest(body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	inst, err := s.resolveInstance(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}

	if !service.PassesIngestThreshold(req.SlowLog, s.ingestMinQueryTime) {
		s.audit("ingest", "skipped", inst.ID, "", "below query_time threshold", r)
		writeJSON(w, http.StatusOK, ingestResponse{
			Status:     string(service.JobSkipped),
			SkipReason: "query_time below SLOWLOG_INGEST_MIN_QUERY_TIME",
		})
		return
	}

	meta := s.runMeta(r, inst.ID)
	cfg := service.RunV6Config{
		SlowLog:        req.SlowLog,
		ReportDir:      s.reportDir,
		Guided:         req.GuidedEnabled(),
		HITL:           false,
		AnalyzeTimeout: s.timeout,
		Meta:           meta,
	}

	if req.IngestAsyncDefault(true) {
		jobID := service.NewJobID()
		s.jobs.Create(jobID, req.Source)
		s.audit("ingest", "async_started", inst.ID, "", jobID, r)
		service.RunIngestAsync(r.Context(), s.jobs, jobID, s.llm, cfg, strings.TrimSpace(req.CallbackURL))
		writeJSON(w, http.StatusAccepted, ingestResponse{
			JobID:  jobID,
			Status: string(service.JobPending),
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()
	result, err := service.RunV6(ctx, s.llm, cfg)
	if err != nil {
		s.audit("ingest", "failed", inst.ID, "", err.Error(), r)
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	s.audit("ingest", "ok", inst.ID, result.ReportID, "", r)
	writeJSON(w, http.StatusOK, ingestResponse{
		ReportID:    result.ReportID,
		Status:      string(service.JobCompleted),
		Iterations:  result.Iterations,
		FinalResult: result.FinalResult,
		ReportURL:   s.reportURL(result.ReportID),
	})
}

func (s *server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || strings.Contains(id, "..") {
		writeErr(w, http.StatusBadRequest, errInvalidID)
		return
	}
	j, ok := s.jobs.Get(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errNotFound)
		return
	}
	resp := ingestResponse{
		JobID:       j.ID,
		Status:      string(j.Status),
		ReportID:    j.ReportID,
		FinalResult: j.FinalResult,
		SkipReason:  j.SkipReason,
	}
	if j.ReportID != "" {
		resp.ReportURL = s.reportURL(j.ReportID)
	}
	if j.Error != "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"job_id": j.ID, "status": j.Status, "error": j.Error,
		})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) checkWebhookSecret(r *http.Request) bool {
	if s.webhookSecret == "" {
		return true
	}
	return r.Header.Get("X-Webhook-Secret") == s.webhookSecret
}

func (s *server) reportURL(reportID string) string {
	if s.publicBase == "" {
		return ""
	}
	return strings.TrimRight(s.publicBase, "/") + "/v1/reports/" + reportID
}

func readBodyLimited(r *http.Request, max int) ([]byte, error) {
	defer r.Body.Close()
	b, err := io.ReadAll(io.LimitReader(r.Body, int64(max+1)))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errEmptyBody
	}
	if len(b) > max {
		return nil, errBodyTooLarge
	}
	return b, nil
}

func ingestMinQueryTimeFromEnv() float64 {
	v := strings.TrimSpace(os.Getenv("SLOWLOG_INGEST_MIN_QUERY_TIME"))
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}
