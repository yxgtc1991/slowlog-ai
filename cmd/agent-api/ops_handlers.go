package main

import (
	"net/http"
)

func (s *server) handleListInstances(w http.ResponseWriter, _ *http.Request) {
	list := s.instances.List()
	writeJSON(w, http.StatusOK, map[string]any{"instances": list})
}
