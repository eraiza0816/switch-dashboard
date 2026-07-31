package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPISpeeds(w http.ResponseWriter, r *http.Request) {
	data := s.Cache.GetSpeeds()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	port := r.URL.Query().Get("port")
	rng := r.URL.Query().Get("range")
	if rng == "" {
		rng = "live"
	}

	w.Header().Set("Content-Type", "application/json")

	if s.HistoryStore == nil {
		w.Write([]byte("[]"))
		return
	}

	points, err := s.HistoryStore.QueryHistory(ip, port, rng)
	if err != nil {
		http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
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

	json.NewEncoder(w).Encode(result)
}
