package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPISwitches(w http.ResponseWriter, r *http.Request) {
	data := s.Cache.GetSwitches()
	if data == nil {
		data = []*SwitchData{}
	}
	for _, sw := range data {
		filtered := sw.MACTable[:0]
		for _, entry := range sw.MACTable {
			if s.isIgnoredMAC(entry.MAC) {
				continue
			}
			if entry.Vendor == "" {
				entry.Vendor = s.OUI.Lookup(entry.MAC)
			}
			filtered = append(filtered, entry)
		}
		sw.MACTable = filtered
	}
	s.writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleAPIRefreshMAC(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		s.writeError(w, http.StatusBadRequest, "missing ip")
		return
	}
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		s.writeError(w, http.StatusNotFound, "switch not found")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok", "count": len(sw.MACTable), "mac_table": sw.MACTable,
	})
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
	_ = switches // fallback: read from config.json below

	// Read switches from saved config.json (not from the live cache)
	type cfgSwitch struct {
		Name      string `json:"name"`
		IP        string `json:"ip"`
		Password  string `json:"password"`
		PSK       string `json:"psk"`
		Model     string `json:"model"`
		PortCount int    `json:"port_count"`
		Enabled   bool   `json:"enabled"`
	}
	type cfgRoot struct {
		Title           string      `json:"title"`
		RefreshInterval int         `json:"refresh_interval"`
		Switches        []cfgSwitch `json:"switches"`
	}

	configPath := s.ConfigPath
	if configPath == "" {
		configPath = "config.json"
	}
	diskCfg := cfgRoot{RefreshInterval: 30}
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &diskCfg)
	}

	for i, sw := range diskCfg.Switches {
		switches = append(switches, SwitchFormData{
			Index:     i,
			Name:      sw.Name,
			IP:        sw.IP,
			Password:  sw.Password,
			PSK:       sw.PSK,
			Model:     sw.Model,
			PortCount: sw.PortCount,
			Enabled:   sw.Enabled,
		})
	}

	title := diskCfg.Title
	if title == "" {
		title = s.Config.Title()
	}

	data := PageData{
		Title:    title,
		Saved:    r.URL.Query().Get("saved") == "1",
		Version:  Version,
		Refresh:  diskCfg.RefreshInterval,
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

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   s.Config.Title(),
		Version: Version,
	}
	s.renderTemplate(w, "api_docs.html", data)
}

func (s *Server) handleMap(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   s.Config.Title(),
		Version: Version,
	}
	s.renderTemplate(w, "map.html", data)
}
