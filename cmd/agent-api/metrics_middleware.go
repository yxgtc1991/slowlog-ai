package main

import (
	"ai_slow_log/internal/ops"
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if s.metrics == nil {
		s.metrics = ops.NewAPIMetrics()
	}
	s.metrics.WritePrometheus(w)
}

func (s *server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if s.metrics != nil {
			code := rec.status
			if code == 0 {
				code = http.StatusOK
			}
			s.metrics.RecordHTTP(r.Method, ops.RouteLabel(r), code)
		}
	})
}
