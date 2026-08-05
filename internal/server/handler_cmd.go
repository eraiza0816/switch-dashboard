package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPISwitchCmd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	if req.Command == "" {
		s.writeError(w, http.StatusBadRequest, "missing cmd")
		return
	}
	ip := chi.URLParam(r, "ip")
	client, err := s.clientFor(ip)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := client.ExecuteCommand(req.Command); err != nil {
		s.writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
