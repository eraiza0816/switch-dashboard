package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPISwitchCmd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "response": "command sent (mock)"})
}
