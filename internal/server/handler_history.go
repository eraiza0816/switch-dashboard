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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}
