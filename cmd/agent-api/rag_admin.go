package main

import (
	"ai_slow_log/internal/rag"
	"net/http"
)

func (s *server) handleRAGStatus(w http.ResponseWriter, r *http.Request) {
	if !s.checkAdminOrWebhook(r) {
		writeErr(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	st, err := rag.GetIndexStatus(s.ragIndexDir)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *server) handleRAGRebuild(w http.ResponseWriter, r *http.Request) {
	if !s.checkAdminOrWebhook(r) {
		writeErr(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	s.audit("rag_rebuild", "started", "", "", r.URL.RawQuery, r)
	withEmb := r.URL.Query().Get("embedding") == "true"
	if err := rag.BuildAllIndexes(s.ragIndexDir, withEmb); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	st, _ := rag.GetIndexStatus(s.ragIndexDir)
	s.audit("rag_rebuild", "ok", "", "", "", r)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"status": st,
	})
}

func (s *server) checkAdminOrWebhook(r *http.Request) bool {
	if s.adminToken != "" {
		return s.requireAdmin(r)
	}
	return s.checkWebhookSecret(r)
}
