package server

import (
	"net/http"
)

func (s *Server) handleAPIGetSettings(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"font_size": "md",
		"log_level": "INFO",
	})
}

func (s *Server) handleAPISaveSettings(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIGetConfigSettings(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"column_widths": map[string]int{},
		"column_order":  []string{},
		"map_positions": map[string]any{},
	})
}

func (s *Server) handleAPISaveConfigSettings(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIConfigReload(w http.ResponseWriter, r *http.Request) {
	if s.ConfigReload == nil {
		s.writeError(w, http.StatusServiceUnavailable, "reload not available")
		return
	}
	if err := s.ConfigReload(); err != nil {
		s.Logger.Error("config reload failed", "error", err)
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
