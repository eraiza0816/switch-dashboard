package poller

import (
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/byte4geek/switch-dashboard/internal/rtlplayground"
	"github.com/byte4geek/switch-dashboard/internal/server"
)

type Poller struct {
	cache    *server.Cache
	client   *rtlplayground.Client
	ip       string
	name     string
	model    string
	interval int
	mu       sync.Mutex
	stopCh   chan struct{}
	counters map[string]*CounterState
	history  map[string]*History
	logger   *slog.Logger
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
		logger:   slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

func NewWithClient(cache *server.Cache, client *rtlplayground.Client, ip, name, model string, interval int) *Poller {
	p := New(cache, ip, name, model, interval)
	p.client = client
	return p
}

func (p *Poller) Start() {
	go p.run()
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

		txBytes := parseHex(entry.TxG)
		rxBytes := parseHex(entry.RxG)
		txPackets := parseHex(entry.TxG)
		rxPackets := parseHex(entry.RxG)

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
		}
		ports = append(ports, ps)

		histKey := (p.ip + ":" + itoa(int64(port)))
		if p.history[histKey] == nil {
			p.history[histKey] = &History{}
		}
		p.history[histKey].Record(now, cumTX, cumRX, speedTX, speedRX, p.interval)
	}

	swData := &server.SwitchData{
		Name:     p.name,
		IP:       p.ip,
		Model:    p.model,
		Status:   "online",
		Ports:    ports,
		Timestamp: now,
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

	p.cache.UpdateSwitch(p.ip, swData)
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
