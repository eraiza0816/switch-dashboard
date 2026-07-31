package server

import (
	"net/http"
)

func (s *Server) handleAPISpeeds(w http.ResponseWriter, r *http.Request) {
	data := s.Cache.GetSpeeds()
	s.writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	port := r.URL.Query().Get("port")
	rng := r.URL.Query().Get("range")
	if rng == "" {
		rng = "live"
	}

	if s.HistoryStore == nil {
		s.writeJSON(w, http.StatusOK, []int{})
		return
	}

	points, err := s.HistoryStore.QueryHistory(ip, port, rng)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "query failed")
		return
	}

	result := struct {
		TX         []int64   `json:"tx"`
		RX         []int64   `json:"rx"`
		Timestamps []float64 `json:"timestamps"`
	}{
		TX:         make([]int64, len(points)),
		RX:         make([]int64, len(points)),
		Timestamps: make([]float64, len(points)),
	}
	for i, p := range points {
		result.TX[i] = p.TX
		result.RX[i] = p.RX
		result.Timestamps[i] = p.TS
	}

	s.writeJSON(w, http.StatusOK, result)
}
