package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPIUpdateHost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MAC  string `json:"mac"`
		Host string `json:"host"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	if req.MAC == "" {
		http.Error(w, `{"error":"mac required"}`, http.StatusBadRequest)
		return
	}
	s.clientHostsMu.Lock()
	if s.ClientHosts == nil {
		s.ClientHosts = make(map[string]string)
	}
	s.ClientHosts[req.MAC] = req.Host
	s.clientHostsMu.Unlock()

	s.saveClientHosts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
