package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
)

type saveConfigData struct {
	Title           string            `json:"title"`
	RefreshInterval int               `json:"refresh_interval"`
	Switches        []saveSwitchEntry `json:"switches"`
}

type saveSwitchEntry struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Password  string `json:"password"`
	PSK       string `json:"psk,omitempty"`
	Model     string `json:"model"`
	PortCount int    `json:"port_count"`
	Enabled   bool   `json:"enabled"`
}

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
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
	psks := r.Form["psk[]"]
	models := r.Form["model[]"]

	for i := 0; i < len(ips); i++ {
		if i >= len(names) {
			break
		}
		entry := saveSwitchEntry{
			Name:      names[i],
			IP:        ips[i],
			Password:  "",
			PSK:       "",
			Model:     "RTLPlayground",
			PortCount: 8,
			Enabled:   true,
		}
		if i < len(passwords) {
			entry.Password = passwords[i]
		}
		if i < len(psks) {
			entry.PSK = psks[i]
		}
		if i < len(models) && models[i] != "" {
			entry.Model = models[i]
		}
		cfg.Switches = append(cfg.Switches, entry)
	}

	configPath := s.ConfigPath
	if configPath == "" {
		configPath = "config.json"
	}

	// Preserve keys that are not edited on the web form (infrastructure
	// devices, unmanaged switches, notes, settings, client overrides, ...).
	existing := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &existing)
	}
	merged := make(map[string]json.RawMessage)
	for k, v := range existing {
		merged[k] = v
	}
	if raw, err := json.Marshal(cfg.Title); err == nil {
		merged["title"] = raw
	}
	if raw, err := json.Marshal(cfg.RefreshInterval); err == nil {
		merged["refresh_interval"] = raw
	}
	if len(cfg.Switches) > 0 {
		if raw, err := json.Marshal(cfg.Switches); err == nil {
			merged["switches"] = raw
		}
	}
	data, err := json.MarshalIndent(merged, "", "  ")
	if err == nil {
		os.WriteFile(configPath, data, 0644)
		s.Logger.Info("config saved", "switches", len(cfg.Switches))
	}

	// Update cache with new switch names/ips so dashboard reflects changes immediately
	for _, sw := range cfg.Switches {
		existing := s.Cache.GetSwitch(sw.IP)
		if existing != nil {
			existing.Name = sw.Name
			existing.Model = sw.Model
			s.Cache.UpdateSwitch(sw.IP, existing)
		}
		// New switches are intentionally NOT seeded: the dashboard shows
		// data only after the first successful scrape, so fabricated
		// "online" data can never appear in production.
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

	// Try to connect to each real switch in the background
	for _, sw := range cfg.Switches {
		sw := sw
		s.Logger.Info("attempting background connection", "ip", sw.IP)
		go func() {
			client, err := newRTLClient(sw.IP, sw.Password, sw.PSK)
			if err != nil {
				s.Logger.Info("switch not reachable, live data will appear after recovery", "ip", sw.IP, "error", err.Error())
				return
			}
			info, err := client.ScrapeInformation()
			if err != nil {
				s.Logger.Warn("initial scrape failed", "ip", sw.IP, "error", err.Error())
				return
			}
			s.Logger.Info("connected to switch, updating live data", "ip", sw.IP, "model", info.HWVer)
			doPoll(client, s.Cache, sw.IP, sw.Name, sw.Model, s.Logger)
		}()
	}

	http.Redirect(w, r, fmt.Sprintf("/config?saved=1&lang=%s", r.FormValue("lang")), http.StatusFound)
}

func newRTLClient(ip, password, pskHex string) (*rtlplayground.Client, error) {
	client, err := rtlplayground.New(ip, password)
	if err != nil {
		return nil, err
	}
	if pskHex != "" {
		if err := client.SetPSK(pskHex); err != nil {
			return nil, fmt.Errorf("invalid psk: %w", err)
		}
	}
	return client, nil
}

func doPoll(client *rtlplayground.Client, cache *Cache, ip, name, model string, logger *slog.Logger) {
	status, err := client.ScrapeStatus()
	if err != nil {
		logger.Error("poll failed", "ip", ip, "error", err)
		return
	}
	info, _ := client.ScrapeInformation()
	sfpDiag, _ := client.ScrapeSFPDiag()
	macTable, _ := client.ScrapeAllMACTable()
	eee, _ := client.ScrapeEEE()
	vlanList, _ := client.ScrapeVLANList()
	lag, _ := client.ScrapeLAG()
	mirror, _ := client.ScrapeMirror()
	bw, _ := client.ScrapeBandwidth()
	mtu, _ := client.ScrapeMTU()

	ports := make([]PortState, 0, len(status))
	for _, entry := range status {
		txPackets := rtlplayground.ParseHex(entry.TxG)
		rxPackets := rtlplayground.ParseHex(entry.RxG)
		txBytes := txPackets * 800
		rxBytes := rxPackets * 800

		statusStr, linkStr, duplex := PortStatus(entry.Link, entry.Enabled)

		speedStr := rtlplayground.LinkSpeedString(entry.Link)
		ports = append(ports, PortState{
			Port:      fmt.Sprintf("%d", entry.PortNum),
			Status:    statusStr,
			Link:      linkStr,
			Speed:     speedStr,
			Duplex:    duplex,
			TXBytes:   txBytes,
			RXBytes:   rxBytes,
			TXPackets: txPackets,
			RXPackets: rxPackets,
			Estimated: true,
			IsSFP:     entry.IsSFP != 0,
			SFPVendor: entry.SFPVendor,
			SFPModel:  entry.SFPModel,
		})
	}

	macEntries := make([]MACEntry, 0, len(macTable))
	for _, entry := range macTable {
		macEntries = append(macEntries, MACEntry{
			MAC:  entry.MAC,
			Type: entry.Type,
			Port: entry.PortStr(),
			VLAN: entry.VLAN,
		})
	}

	jumbo := JumboFrameStatus{Enabled: false, Size: "Disabled"}
	if len(mtu) > 0 {
		maxMTU := int64(0)
		for _, m := range mtu {
			if v := rtlplayground.ParseMTUHex(m.MTU); v > maxMTU {
				maxMTU = v
			}
		}
		if maxMTU > 1518 {
			jumbo = JumboFrameStatus{Enabled: true, Size: fmt.Sprintf("%d", maxMTU)}
		}
	}

	sfpDiagStatus := make([]SFPDiagStatus, 0, len(sfpDiag))
	for _, d := range sfpDiag {
		v := rtlplayground.FormatSFPDiag(d)
		sfpDiagStatus = append(sfpDiagStatus, SFPDiagStatus{
			Port:    v.Port,
			Options: v.Options,
			Temp:    v.Temp,
			VCC:     v.VCC,
			Bias:    v.Bias,
			TXPower: v.TXPower,
			RXPower: v.RXPower,
			State:   v.State,
			HasDDMI: v.HasDDMI,
		})
	}

	swData := &SwitchData{
		Name:     name,
		IP:       ip,
		Model:    model,
		Status:   "online",
		Ports:    ports,
		MACTable: macEntries,
		DHCP:     SnoopingStatus{Enabled: false, Ports: make(map[string]string)},
		IGMP:     IGMPStatus{Enabled: false},
		Jumbo:    jumbo,
		SFPDiag:  sfpDiagStatus,
	}
	if info != nil {
		swData.MAC = info.MACAddress
		swData.Firmware = info.SwVer
		swData.Hostname = info.Hostname
	}
	_ = eee
	_ = vlanList
	_ = lag
	_ = mirror
	_ = bw

	cache.UpdateSwitch(ip, swData)
}
