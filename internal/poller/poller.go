package poller

import (
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/eraiza0816/switch-dashboard/internal/history"
	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
	"github.com/eraiza0816/switch-dashboard/internal/server"
)

type Poller struct {
	cache        *server.Cache
	client       *rtlplayground.Client
	ip           string
	name         string
	model        string
	interval     int
	mu           sync.Mutex
	stopCh       chan struct{}
	counters     map[string]*CounterState
	history      map[string]*History
	notes        map[string]string
	logger       *slog.Logger
	historyStore *history.Store
}

func (p *Poller) SetHistoryStore(store *history.Store) {
	p.historyStore = store
}

func New(cache *server.Cache, ip, name, model string, interval int) *Poller {
	return &Poller{
		cache:    cache,
		ip:       ip,
		name:     name,
		model:    model,
		interval: interval,
		stopCh:   make(chan struct{}),
		counters: make(map[string]*CounterState),
		history:  make(map[string]*History),
		notes:    make(map[string]string),
		logger:   slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

func NewWithNotes(cache *server.Cache, ip, name, model string, interval int, notes map[string]string) *Poller {
	p := New(cache, ip, name, model, interval)
	if notes != nil {
		p.notes = notes
	}
	return p
}

func NewWithClientAndNotes(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, notes map[string]string) *Poller {
	p := NewWithClient(cache, client, ip, name, model, interval)
	if notes != nil {
		p.notes = notes
	}
	return p
}

func NewWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int) *Poller {
	p := New(cache, ip, name, model, interval)
	p.client = client
	return p
}

func (p *Poller) Start() {
	// Seed initial data so UI has something to show
	p.seedMockData()
	go p.run()
}

func (p *Poller) seedMockData() {
	now := float64(time.Now().UnixNano()) / 1e9
	ports := []server.PortState{
		{Port: "1", Status: "up", Link: "Link Up", Speed: "10G", Duplex: "Full", TXBytes: 1000000, RXBytes: 2000000, CumTX: 1000000000, CumRX: 2000000000, SpeedTX: 800000000, SpeedRX: 1600000000},
		{Port: "2", Status: "down", Link: "Link Down", Speed: "", Duplex: ""},
		{Port: "3", Status: "up", Link: "Link Up", Speed: "2.5G", Duplex: "Full", TXBytes: 500000, RXBytes: 1000000, CumTX: 500000000, CumRX: 1000000000, SpeedTX: 400000000, SpeedRX: 800000000},
		{Port: "4", Status: "disable", Link: "Disabled"},
		{Port: "5", Status: "up", Link: "Link Up", Speed: "1G", Duplex: "Full", TXBytes: 100000, RXBytes: 200000, CumTX: 100000000, CumRX: 200000000, SpeedTX: 80000000, SpeedRX: 160000000},
		{Port: "9", Status: "up", Link: "Link Up", Speed: "10G", Duplex: "Full", TXBytes: 2000000, RXBytes: 4000000, CumTX: 2000000000, CumRX: 4000000000, SpeedTX: 1600000000, SpeedRX: 3200000000},
	}

	if p.historyStore != nil {
		ts := time.Now()
		for _, port := range ports {
			if port.SpeedTX > 0 {
				for i := 120; i >= 0; i-- {
					t := ts.Add(-time.Duration(i) * time.Second)
					variation := 1.0 + float64(i%10)*0.02
					if err := p.historyStore.WriteSample(p.ip, port.Port, t,
						port.CumTX-int64(float64(port.SpeedTX)*float64(i)*variation),
						port.CumRX-int64(float64(port.SpeedRX)*float64(i)*variation),
						int64(float64(port.SpeedTX)*variation),
						int64(float64(port.SpeedRX)*variation),
					); err != nil {
						p.logger.Error("seed history write failed", "ip", p.ip, "port", port.Port, "error", err)
					}
				}
			}
		}
		p.logger.Info("seeded mock history data", "ip", p.ip, "ports", len(ports))
	}

	swData := &server.SwitchData{
		Name:     p.name,
		IP:       p.ip,
		Model:    "RTLPlayground Simulator",
		MAC:      "1c:2a:a3:23:00:02",
		Firmware: "v0.2.19",
		Hostname: "rtlplayground",
		Ports:    ports,
		Status:   "online",
		Timestamp: now,
		MACTable: []server.MACEntry{
			{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001", Vendor: "Intel Corporate"},
			{MAC: "AA:BB:CC:DD:EE:02", Type: "l", Port: "3", VLAN: "001", Vendor: "Raspberry Pi"},
			{MAC: "AA:BB:CC:DD:EE:03", Type: "s", Port: "5", VLAN: "010", Vendor: "Apple Inc"},
		},
	}
	p.cache.UpdateSwitch(p.ip, swData)
}

func (p *Poller) Stop() {
	close(p.stopCh)
}

func (p *Poller) run() {
	if p.interval <= 0 {
		p.interval = 30
	}
	ticker := time.NewTicker(time.Duration(p.interval) * time.Second)
	defer ticker.Stop()

	p.poll()
	for {
		select {
		case <-ticker.C:
			p.poll()
		case <-p.stopCh:
			return
		}
	}
}

func (p *Poller) poll() {
	status, info, _, macTable := p.fetchData()
	if status == nil {
		return
	}

	now := float64(time.Now().UnixNano()) / 1e9
	ports := make([]server.PortState, 0, len(status))

	for _, entry := range status {
		port := entry.PortNum
		key := p.ip + ":" + itoa(int64(port))

		txPackets := parseHex(entry.TxG)
		rxPackets := parseHex(entry.RxG)
		txBytes := txPackets * 800
		rxBytes := rxPackets * 800

		speedStr := linkSpeedToString(entry.Link)
		speedBPS := ParseLinkSpeed(speedStr)

		prev := p.counters[key]
		delta := ComputeDelta(prev, txBytes, rxBytes)

		dt := float64(p.interval)
		if prev != nil {
			dt = now - prev.TS
		}

		maxBytes := SanityCheckMaxBytes(speedBPS, dt)
		deltaTX := ClampDelta(delta.DeltaTX, maxBytes)
		deltaRX := ClampDelta(delta.DeltaRX, maxBytes)

		cumTX := delta.CumTX
		cumRX := delta.CumRX
		if prev == nil {
			cumTX = txBytes
			cumRX = rxBytes
		} else {
			cumTX = prev.CumTX + deltaTX
			cumRX = prev.CumRX + deltaRX
		}

		p.counters[key] = &CounterState{
			Tx: txBytes, Rx: rxBytes,
			CumTX: cumTX, CumRX: cumRX,
			TS: now,
		}

		speedTX, speedRX := ComputeSpeed(deltaTX, deltaRX, dt)
		if math.IsInf(float64(speedTX), 0) || speedTX < 0 {
			speedTX = 0
		}
		if math.IsInf(float64(speedRX), 0) || speedRX < 0 {
			speedRX = 0
		}

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

		noteKey := p.ip + ":" + itoa(int64(port))
		ps := server.PortState{
			Port:      itoa(int64(port)),
			Status:    statusStr,
			Link:      linkStr,
			Speed:     speedStr,
			Duplex:    duplex,
			TXBytes:   txBytes,
			RXBytes:   rxBytes,
			TXPackets: txPackets,
			RXPackets: rxPackets,
			CumTX:     cumTX,
			CumRX:     cumRX,
			SpeedTX:   speedTX,
			SpeedRX:   speedRX,
			Note:      p.notes[noteKey],
			IsSFP:     entry.IsSFP != 0,
			SFPVendor: entry.SFPVendor,
			SFPModel:  entry.SFPModel,
		}
		ports = append(ports, ps)

		histKey := (p.ip + ":" + itoa(int64(port)))
		if p.history[histKey] == nil {
			p.history[histKey] = &History{}
		}
		p.history[histKey].Record(now, cumTX, cumRX, speedTX, speedRX, p.interval)

		if p.historyStore != nil {
			ts := time.Unix(0, int64(now*1e9))
			p.historyStore.WriteSample(p.ip, itoa(int64(port)), ts, cumTX, cumRX, speedTX, speedRX)
		}
	}

	swData := &server.SwitchData{
		Name:     p.name,
		IP:       p.ip,
		Model:    p.model,
		Status:   "online",
		Ports:    ports,
		Timestamp: now,
		DHCP:     server.SnoopingStatus{Enabled: false, Ports: make(map[string]string)},
		IGMP:     server.IGMPStatus{Enabled: false},
		Jumbo:    server.JumboFrameStatus{Enabled: false, Size: "Disabled"},
	}

	if info != nil {
		swData.MAC = info.MACAddress
		swData.Firmware = info.SwVer
		swData.Hostname = info.Hostname
	}

	macEntries := make([]server.MACEntry, 0, len(macTable))
	for _, entry := range macTable {
		macEntries = append(macEntries, server.MACEntry{
			MAC:  entry.MAC,
			Type: entry.Type,
			Port: entry.PortStr(),
			VLAN: entry.VLAN,
		})
	}
	swData.MACTable = macEntries

	// Fetch extra data
	eeeData, vlanList, lagData, mirrorData, bwData, mtuData := p.fetchExtraData()
	if len(eeeData) > 0 {
		_ = eeeData
	}
	if len(vlanList) > 0 {
		_ = vlanList
	}
	if len(lagData) > 0 {
		_ = lagData
	}
	if mirrorData != nil {
		_ = mirrorData
	}
	if len(bwData) > 0 {
		_ = bwData
	}
	if len(mtuData) > 0 {
		for _, m := range mtuData {
			if m.PortNum > 0 {
				swData.Jumbo.Enabled = true
				swData.Jumbo.Size = m.MTU
				break
			}
		}
	}

	p.cache.UpdateSwitch(p.ip, swData)
}

func (p *Poller) fetchExtraData() ([]rtlplayground.EEEEntry, []rtlplayground.VLANListItem, []rtlplayground.LAGEntry, *rtlplayground.MirrorConfig, []rtlplayground.BandwidthEntry, []rtlplayground.MTUEntry) {
	if p.client == nil {
		return nil, nil, nil, nil, nil, nil
	}
	eee, _ := p.client.ScrapeEEE()
	vlan, _ := p.client.ScrapeVLANList()
	lag, _ := p.client.ScrapeLAG()
	mirror, _ := p.client.ScrapeMirror()
	bw, _ := p.client.ScrapeBandwidth()
	mtu, _ := p.client.ScrapeMTU()
	return eee, vlan, lag, mirror, bw, mtu
}

func (p *Poller) fetchData() ([]rtlplayground.StatusEntry, *rtlplayground.Information, []rtlplayground.SFPDiagEntry, []rtlplayground.L2Entry) {
	if p.client == nil {
		return nil, nil, nil, nil
	}

	status, err := p.client.ScrapeStatus()
	if err != nil {
		p.logger.Error("scrape status failed", "ip", p.ip, "error", err)
		return nil, nil, nil, nil
	}

	info, _ := p.client.ScrapeInformation()
	sfpDiag, _ := p.client.ScrapeSFPDiag()
	macTable, _ := p.client.ScrapeAllMACTable()

	return status, info, sfpDiag, macTable
}

func linkSpeedToString(link int) string {
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

func parseHex(s string) int64 {
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
