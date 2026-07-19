package server

import (
	"sync"
	"time"
)

type PortState struct {
	Port      string `json:"port"`
	Status    string `json:"status"`
	Link      string `json:"link"`
	Speed     string `json:"speed"`
	Duplex    string `json:"duplex"`
	TXBytes   int64  `json:"tx_bytes"`
	RXBytes   int64  `json:"rx_bytes"`
	TXPackets int64  `json:"tx_packets"`
	RXPackets int64  `json:"rx_packets"`
	CumTX     int64  `json:"cum_tx"`
	CumRX     int64  `json:"cum_rx"`
	SpeedTX   int64  `json:"speed_tx_bps"`
	SpeedRX   int64  `json:"speed_rx_bps"`
	Note      string `json:"note,omitempty"`
}

type SwitchData struct {
	Name       string            `json:"name"`
	IP         string            `json:"ip"`
	Model      string            `json:"model"`
	MAC        string            `json:"mac"`
	Uptime     string            `json:"uptime"`
	Firmware   string            `json:"firmware"`
	Hostname   string            `json:"hostname"`
	Ports      []PortState       `json:"ports"`
	MACTable   []MACEntry        `json:"mac_table"`
	MACScraped int64             `json:"mac_timestamp"`
	Status     string            `json:"status"`
	Error      string            `json:"error,omitempty"`
	Timestamp  float64           `json:"timestamp"`
	DHCP       interface{}       `json:"dhcp_snooping"`
	IGMP       interface{}       `json:"igmp"`
	Jumbo      interface{}       `json:"jumbo_frame"`
}

type MACEntry struct {
	MAC    string `json:"mac"`
	Type   string `json:"type"`
	Port   string `json:"port"`
	VLAN   string `json:"vlan"`
	Vendor string `json:"vendor,omitempty"`
	Host   string `json:"host,omitempty"`
}

type TopologyNode struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	IP     string `json:"ip,omitempty"`
	MAC    string `json:"mac,omitempty"`
	Model  string `json:"model,omitempty"`
	Status string `json:"status"`
}

type TopologyLink struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	SourcePort  string `json:"source_port"`
	TargetPort  string `json:"target_port"`
	Speed       string `json:"speed"`
	Type        string `json:"type"`
	TXBPS       int64  `json:"tx_bps,omitempty"`
	RXBPS       int64  `json:"rx_bps,omitempty"`
}

type Topology struct {
	Nodes []TopologyNode `json:"nodes"`
	Links []TopologyLink `json:"links"`
}

type HistoryPoint struct {
	TS float64 `json:"ts"`
	TX int64   `json:"tx"`
	RX int64   `json:"rx"`
}

type Cache struct {
	mu        sync.RWMutex
	switches  map[string]*SwitchData
	speeds    map[string]map[string]PortSpeeds
	lastPoll  time.Time
}

type PortSpeeds struct {
	SpeedTX int64 `json:"speed_tx"`
	SpeedRX int64 `json:"speed_rx"`
}

func NewCache() *Cache {
	return &Cache{
		switches: make(map[string]*SwitchData),
		speeds:   make(map[string]map[string]PortSpeeds),
	}
}

func (c *Cache) UpdateSwitch(ip string, data *SwitchData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.switches[ip] = data
	c.lastPoll = time.Now()

	swSpeeds := make(map[string]PortSpeeds)
	for _, p := range data.Ports {
		swSpeeds[p.Port] = PortSpeeds{SpeedTX: p.SpeedTX, SpeedRX: p.SpeedRX}
	}
	c.speeds[ip] = swSpeeds
}

func (c *Cache) RemoveSwitch(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.switches, ip)
	delete(c.speeds, ip)
}

func (c *Cache) GetSwitches() []*SwitchData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*SwitchData, 0, len(c.switches))
	for _, sw := range c.switches {
		result = append(result, sw)
	}
	return result
}

func (c *Cache) GetSwitch(ip string) *SwitchData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.switches[ip]
}

func (c *Cache) GetSpeeds() map[string]map[string]PortSpeeds {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]map[string]PortSpeeds)
	for ip, speeds := range c.speeds {
		sw := make(map[string]PortSpeeds)
		for port, s := range speeds {
			sw[port] = s
		}
		c.speeds[ip] = sw
	}
	_ = result
	return c.speeds
}

func (c *Cache) GetAllIPs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ips := make([]string, 0, len(c.switches))
	for ip := range c.switches {
		ips = append(ips, ip)
	}
	return ips
}
