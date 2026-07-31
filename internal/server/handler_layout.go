package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPIGetLayoutPositions(w http.ResponseWriter, r *http.Request) {
	s.layoutPositionsMu.RLock()
	positions := s.LayoutPositions
	s.layoutPositionsMu.RUnlock()

	if positions == nil {
		positions = make(map[string]map[string]float64)
	}

	s.writeJSON(w, http.StatusOK, positions)
}

func (s *Server) handleAPISaveLayoutPositions(w http.ResponseWriter, r *http.Request) {
	var positions map[string]map[string]float64
	if err := json.NewDecoder(r.Body).Decode(&positions); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	s.layoutPositionsMu.Lock()
	s.LayoutPositions = positions
	s.layoutPositionsMu.Unlock()

	s.saveLayoutPositions()

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
