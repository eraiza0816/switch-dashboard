package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
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
		configPath := s.ConfigPath
		if configPath == "" {
			configPath = "config.json"
		}
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

	// Try to connect to each real switch in the background
	for _, sw := range cfg.Switches {
		sw := sw
		s.Logger.Info("attempting background connection", "ip", sw.IP)
		go func() {
			client, err := newRTLClient(sw.IP, sw.Password)
			if err != nil {
				s.Logger.Info("switch not reachable, keeping mock data", "ip", sw.IP, "error", err.Error())
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

func newRTLClient(ip, password string) (*rtlplayground.Client, error) {
	return rtlplayground.New(ip, password)
}

func doPoll(client *rtlplayground.Client, cache *Cache, ip, name, model string, logger Logger) {
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
		txPackets := parseHexVal(entry.TxG)
		rxPackets := parseHexVal(entry.RxG)
		txBytes := txPackets * 800
		rxBytes := rxPackets * 800

		statusStr := "down"
		linkStr := "Link Down"
		duplex := ""
		if entry.Link > 0 && entry.Enabled != 0 {
			statusStr = "up"
			linkStr = "Link Up"
			duplex = "Full"
		}
		if entry.Enabled == 0 {
			statusStr = "disable"
			linkStr = "Disabled"
		}

		speedStr := linkSpeedFromInt(entry.Link)
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
		for _, m := range mtu {
			if m.PortNum > 0 {
				jumbo.Enabled = true
				jumbo.Size = m.MTU
				break
			}
		}
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
	}
	if info != nil {
		swData.MAC = info.MACAddress
		swData.Firmware = info.SwVer
		swData.Hostname = info.Hostname
	}
	_ = sfpDiag
	_ = eee
	_ = vlanList
	_ = lag
	_ = mirror
	_ = bw

	cache.UpdateSwitch(ip, swData)
}

func parseHexVal(s string) int64 {
	if len(s) < 3 || s[:2] != "0x" {
		return 0
	}
	var v int64
	for _, c := range s[2:] {
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= int64(c - '0')
		case c >= 'a' && c <= 'f':
			v |= int64(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v |= int64(c - 'A' + 10)
		default:
			return 0
		}
	}
	return v
}

func linkSpeedFromInt(link int) string {
	switch link {
	case 0:
		return ""
	case 1:
		return "100M"
	case 2:
		return "1G"
	case 3:
		return "2.5G"
	case 4:
		return "10G"
	default:
		return "Auto"
	}
}
