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
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "response": "command sent (mock)"})
}
