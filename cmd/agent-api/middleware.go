package main

import (
	"ai_slow_log/internal/ops"
	"ai_slow_log/internal/service"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

func (s *server) wrap(mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if rid == "" {
			rid = newRequestID()
		}
		ctx := ops.ContextWithRequestID(r.Context(), rid)
		w.Header().Set("X-Request-ID", rid)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "req-" + hex.EncodeToString(b[:])
}

func (s *server) resolveInstance(r *http.Request) (ops.Instance, error) {
	id := strings.TrimSpace(r.Header.Get("X-Instance-ID"))
	if id == "" {
		id = strings.TrimSpace(r.URL.Query().Get("instance_id"))
	}
	if s.requireInstance && id == "" {
		return ops.Instance{}, errMissingInstance
	}
	return s.instances.Resolve(id)
}

func (s *server) runMeta(r *http.Request, instID string) service.RunMeta {
	return service.RunMeta{
		InstanceID: instID,
		RequestID:  ops.RequestIDFromContext(r.Context()),
		Actor:      strings.TrimSpace(r.Header.Get("X-Actor")),
		ClientIP:   clientIP(r),
	}
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return ""
}

func (s *server) audit(action, status, instID, reportID, detail string, r *http.Request) {
	if s.auditor == nil {
		return
	}
	_ = s.auditor.Log(ops.AuditEvent{
		Action:     action,
		Status:     status,
		InstanceID: instID,
		RequestID:  ops.RequestIDFromContext(r.Context()),
		ReportID:   reportID,
		ClientIP:   clientIP(r),
		Actor:      strings.TrimSpace(r.Header.Get("X-Actor")),
		Detail:     detail,
	})
}

func (s *server) requireAdmin(r *http.Request) bool {
	if s.adminToken == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	return strings.TrimPrefix(auth, "Bearer ") == s.adminToken
}
