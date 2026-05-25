package main

import (
	"ai_slow_log/internal/rag"
	"net/http"
)

func (s *server) handleRAGStatus(w http.ResponseWriter, r *http.Request) {
	if !s.checkWebhookSecret(r) {
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
	if !s.checkWebhookSecret(r) {
		writeErr(w, http.StatusUnauthorized, errUnauthorized)
		return
	}
	withEmb := r.URL.Query().Get("embedding") == "true"
	if err := rag.BuildAllIndexes(s.ragIndexDir, withEmb); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	st, _ := rag.GetIndexStatus(s.ragIndexDir)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"status": st,
	})
}
