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
			// New switch: seed with mock data so UI shows something
			s.Cache.UpdateSwitch(sw.IP, &SwitchData{
				Name:   sw.Name,
				IP:     sw.IP,
				Model:  sw.Model,
				Status: "online",
				Ports: []PortState{
					{Port: "1", Status: "up", Link: "Link Up", Speed: "10G", Duplex: "Full", CumTX: 100000, CumRX: 200000, SpeedTX: 800, SpeedRX: 1600},
					{Port: "2", Status: "down", Link: "Link Down"},
					{Port: "3", Status: "up", Link: "Link Up", Speed: "2.5G", Duplex: "Full", CumTX: 50000, CumRX: 100000, SpeedTX: 400, SpeedRX: 800},
					{Port: "4", Status: "disable", Link: "Disabled"},
					{Port: "5", Status: "up", Link: "Link Up", Speed: "10G", Duplex: "Full", CumTX: 200000, CumRX: 400000, SpeedTX: 1600, SpeedRX: 3200, IsSFP: true},
					{Port: "6", Status: "up", Link: "Link Up", Speed: "10G", Duplex: "Full", CumTX: 200000, CumRX: 400000, SpeedTX: 1600, SpeedRX: 3200, IsSFP: true},
				},
				MACTable: []MACEntry{
					{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001"},
					{MAC: "AA:BB:CC:DD:EE:02", Type: "l", Port: "3", VLAN: "001"},
					{MAC: "AA:BB:CC:DD:EE:03", Type: "s", Port: "5", VLAN: "010"},
				},
				MACScraped: 0,
				DHCP: SnoopingStatus{Enabled: false, Ports: make(map[string]string)},
				IGMP: IGMPStatus{Enabled: false},
				Jumbo: JumboFrameStatus{Enabled: false, Size: "Disabled"},
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
