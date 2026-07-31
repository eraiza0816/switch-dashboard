package poller

import (
	"log/slog"
	"math"
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

func New(cache *server.Cache, ip, name, model string, interval int, logger *slog.Logger) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
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
		logger:   logger,
	}
}

func NewWithNotes(cache *server.Cache, ip, name, model string, interval int, notes map[string]string, logger *slog.Logger) *Poller {
	p := New(cache, ip, name, model, interval, logger)
	if notes != nil {
		p.notes = notes
	}
	return p
}

func NewWithClientAndNotes(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, notes map[string]string, logger *slog.Logger) *Poller {
	p := NewWithClient(cache, client, ip, name, model, interval, logger)
	if notes != nil {
		p.notes = notes
	}
	return p
}

func NewWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int, logger *slog.Logger) *Poller {
	p := New(cache, ip, name, model, interval, logger)
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
		Name:      p.name,
		IP:        p.ip,
		Model:     "RTLPlayground Simulator",
		MAC:       "1c:2a:a3:23:00:02",
		Firmware:  "v0.2.19",
		Hostname:  "rtlplayground",
		Ports:     ports,
		Status:    "online",
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

		txPackets := rtlplayground.ParseHex(entry.TxG)
		rxPackets := rtlplayground.ParseHex(entry.RxG)
		txBytes := txPackets * 800
		rxBytes := rxPackets * 800

		speedStr := rtlplayground.LinkSpeedString(entry.Link)
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

		statusStr, linkStr, duplex := server.PortStatus(entry.Link, entry.Enabled)

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
		Name:      p.name,
		IP:        p.ip,
		Model:     p.model,
		Status:    "online",
		Ports:     ports,
		Timestamp: now,
		DHCP:      server.SnoopingStatus{Enabled: false, Ports: make(map[string]string)},
		IGMP:      server.IGMPStatus{Enabled: false},
		Jumbo:     server.JumboFrameStatus{Enabled: false, Size: "Disabled"},
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

	eeeStatus := make([]server.EEEStatus, 0, len(eeeData))
	for _, e := range eeeData {
		eeeStatus = append(eeeStatus, server.EEEStatus{
			Port:   itoa(int64(e.PortNum)),
			Active: e.Active != 0,
			Status: e.EEE,
			LP:     e.EEELP,
		})
	}
	swData.EEE = eeeStatus

	vlanItems := make([]server.VLANItem, 0, len(vlanList))
	for _, v := range vlanList {
		vlanItems = append(vlanItems, server.VLANItem{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	swData.VLANList = vlanItems

	lagStatus := make([]server.LAGStatus, 0, len(lagData))
	for _, l := range lagData {
		lagStatus = append(lagStatus, server.LAGStatus{
			Number:  l.LAGNum,
			Members: l.Members,
			Hash:    l.Hash,
		})
	}
	swData.LAG = lagStatus

	if mirrorData != nil {
		swData.Mirror = &server.MirrorStatus{
			Enabled:  mirrorData.Enabled != 0,
			Port:     itoa(int64(mirrorData.MPort)),
			MirrorRX: mirrorData.MirrorRX,
			MirrorTX: mirrorData.MirrorTX,
		}
	}

	bwStatus := make([]server.BWStatus, 0, len(bwData))
	for _, b := range bwData {
		bwStatus = append(bwStatus, server.BWStatus{
			Port:     itoa(int64(b.PortNum)),
			InLimit:  b.ILimited != 0,
			InBW:     b.IBW,
			OutLimit: b.ELimited != 0,
			OutBW:    b.EBW,
		})
	}
	swData.Bandwidth = bwStatus

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
