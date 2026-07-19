package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type saveConfigData struct {
	Title           string              `json:"title"`
	RefreshInterval int                 `json:"refresh_interval"`
	Switches        []saveSwitchEntry   `json:"switches"`
}

type saveSwitchEntry struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Password string `json:"password"`
	Model    string `json:"model"`
	PortCount int   `json:"port_count"`
	Enabled  bool   `json:"enabled"`
}

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	cfg := saveConfigData{
		RefreshInterval: 30,
	}

	if t := r.FormValue("title"); t != "" {
		cfg.Title = t
	}
	if v := r.FormValue("refresh_interval"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RefreshInterval = n
		}
	}

	names := r.Form["name[]"]
	ips := r.Form["ip[]"]
	passwords := r.Form["password[]"]
	models := r.Form["model[]"]

	for i := 0; i < len(ips); i++ {
		if i >= len(names) {
			break
		}
		entry := saveSwitchEntry{
			Name:      names[i],
			IP:        ips[i],
			Password:  "",
			Model:     "RTLPlayground",
			PortCount: 8,
			Enabled:   true,
		}
		if i < len(passwords) {
			entry.Password = passwords[i]
		}
		if i < len(models) && models[i] != "" {
			entry.Model = models[i]
		}
		cfg.Switches = append(cfg.Switches, entry)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		os.WriteFile("config.json", data, 0644)
		s.Logger.Info("config saved", "switches", len(cfg.Switches))
	}

	// Update cache with new switch names/ips so dashboard reflects changes immediately
	for _, sw := range cfg.Switches {
		existing := s.Cache.GetSwitch(sw.IP)
		if existing != nil {
			existing.Name = sw.Name
			existing.Model = sw.Model
			s.Cache.UpdateSwitch(sw.IP, existing)
		} else {
			// New switch: seed with basic data
			s.Cache.UpdateSwitch(sw.IP, &SwitchData{
				Name:   sw.Name,
				IP:     sw.IP,
				Model:  sw.Model,
				Status: "online",
				Ports:  []PortState{},
			})
		}
	}

	// Remove cached switches that are no longer in config
	for _, ip := range s.Cache.GetAllIPs() {
		found := false
		for _, sw := range cfg.Switches {
			if sw.IP == ip {
				found = true
				break
			}
		}
		if !found {
			s.Cache.RemoveSwitch(ip)
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/config?saved=1&lang=%s", r.FormValue("lang")), http.StatusFound)
}
