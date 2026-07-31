package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPIGetVendors(w http.ResponseWriter, r *http.Request) {
	custom := s.OUI.Custom()
	s.writeJSON(w, http.StatusOK, custom)
}

func (s *Server) handleAPISaveVendors(w http.ResponseWriter, r *http.Request) {
	var vendors map[string]string
	if err := json.NewDecoder(r.Body).Decode(&vendors); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.OUI.SetCustom(vendors)
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIUpdateOUI(w http.ResponseWriter, r *http.Request) {
	if err := s.OUI.Update(); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"count":  len(s.OUI.Entries()),
	})
}
