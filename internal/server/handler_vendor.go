package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPIGetVendors(w http.ResponseWriter, r *http.Request) {
	custom := s.OUI.Custom()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(custom)
}

func (s *Server) handleAPISaveVendors(w http.ResponseWriter, r *http.Request) {
	var vendors map[string]string
	if err := json.NewDecoder(r.Body).Decode(&vendors); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.OUI.SetCustom(vendors)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleAPIUpdateOUI(w http.ResponseWriter, r *http.Request) {
	if err := s.OUI.Update(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"count":  len(s.OUI.Entries()),
	})
}
