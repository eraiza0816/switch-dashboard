package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var _ = json.Marshal // ensure import

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
	data := PageData{
		Title:              s.Config.Title(),
		Version:            Version,
		Refresh:            s.Config.RefreshInterval(),
		GridColumns:        s.Config.GridColumns(),
		EnabledColumns:     s.Config.EnabledColumns(),
		PortsWrapThreshold: s.Config.PortsWrapThreshold(),
		ColumnWidths:       "",
		ColumnOrder:        "",
	}
	s.renderTemplate(w, "index.html", data)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	var switches []SwitchFormData
	for i, sw := range s.Cache.GetSwitches() {
		switches = append(switches, SwitchFormData{
			Index:      i,
			Name:       sw.Name,
			IP:         sw.IP,
			Model:      sw.Model,
			PortCount:  8,
			Enabled:    true,
		})
	}
	data := PageData{
		Title:    s.Config.Title(),
		Version:  Version,
		Refresh:  s.Config.RefreshInterval(),
		Switches: switches,
	}
	s.renderTemplate(w, "config.html", data)
}

func (s *Server) handleBackups(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   s.Config.Title(),
		Version: Version,
	}
	s.renderTemplate(w, "backups.html", data)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:           s.Config.Title(),
		Version:         Version,
		CurrentLogLevel: "INFO",
	}
	s.renderTemplate(w, "logs.html", data)
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   s.Config.Title(),
		Version: Version,
	}
	s.renderTemplate(w, "api_docs.html", data)
}

