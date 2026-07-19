package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPISwitches(w http.ResponseWriter, r *http.Request) {
	data := s.Cache.GetSwitches()
	if data == nil {
		data = []*SwitchData{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleAPIRefreshMAC(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		http.Error(w, "missing ip", http.StatusBadRequest)
		return
	}
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		http.Error(w, "switch not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok", "count": len(sw.MACTable), "mac_table": sw.MACTable,
	})
}

func (s *Server) handleAPISwitchSFP(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		http.Error(w, "missing ip", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "index.html", nil)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "config.html", nil)
}

func (s *Server) handleBackups(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "backups.html", nil)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "logs.html", nil)
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "api_docs.html", nil)
}

