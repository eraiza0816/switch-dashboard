package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPIGetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"font_size":   "md",
		"log_level":   "INFO",
	})
}

func (s *Server) handleAPISaveSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleAPIGetConfigSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"column_widths": map[string]int{},
		"column_order":  []string{},
		"map_positions": map[string]any{},
	})
}

func (s *Server) handleAPISaveConfigSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleAPIConfigReload(w http.ResponseWriter, r *http.Request) {
	if s.ConfigReload == nil {
		http.Error(w, "reload not available", http.StatusServiceUnavailable)
		return
	}
	if err := s.ConfigReload(); err != nil {
		s.Logger.Error("config reload failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
