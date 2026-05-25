package main

import (
	"net/http"
	"strings"
)

// isPublicAPIPath 无需 SLOWLOG_API_KEY 的路径。
func isPublicAPIPath(path string) bool {
	return path == "/v1/health"
}

func (s *server) checkAPIKey(r *http.Request) bool {
	if s.apiKey == "" {
		return true
	}
	if k := strings.TrimSpace(r.Header.Get("X-API-Key")); k != "" && k == s.apiKey {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ") == s.apiKey
	}
	return false
}

func (s *server) apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey == "" || isPublicAPIPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if !s.checkAPIKey(r) {
			writeErr(w, http.StatusUnauthorized, errUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
