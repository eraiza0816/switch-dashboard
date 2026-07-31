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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func (s *Server) handleAPISaveLayoutPositions(w http.ResponseWriter, r *http.Request) {
	var positions map[string]map[string]float64
	if err := json.NewDecoder(r.Body).Decode(&positions); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}

	s.layoutPositionsMu.Lock()
	s.LayoutPositions = positions
	s.layoutPositionsMu.Unlock()

	s.saveLayoutPositions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
