package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

func (s *Server) handleAPITopology(w http.ResponseWriter, r *http.Request) {
	switches := s.Cache.GetSwitches()

	switchesByIP := make(map[string]*SwitchData, len(switches))
	switchMACToIP := make(map[string]string)
	for _, sw := range switches {
		switchesByIP[sw.IP] = sw
		if m := NormalizeMAC(sw.MAC); m != "" {
			switchMACToIP[m] = sw.IP
		}
	}

	liveSwitchIPs := make(map[string]bool, len(switchesByIP))
	for ip := range switchesByIP {
		liveSwitchIPs[ip] = true
	}

	// Infrastructure devices from config
	infraByMAC := make(map[string]config.InfraDevice)
	for _, dev := range s.Config.InfrastructureDevices() {
		m := NormalizeMAC(dev.MAC)
		if m == "" {
			continue
		}
		// Skip infra devices that are actually monitored switches
		if _, ok := switchMACToIP[m]; ok {
			continue
		}
		infraByMAC[m] = dev
	}

	// Static switch uplinks from config: these ports carry a link to another
	// switch and must never be treated as client ports.
	staticParent := make(map[string]struct {
		parentIP   string
		parentPort string
		upPort     string
	})
	uplinkPorts := make(map[string]bool)
	for _, sw := range s.Config.Switches() {
		if !sw.Enabled || sw.ParentIP == "" || sw.ParentPort == "" {
			continue
		}
		staticParent[sw.IP] = struct {
			parentIP   string
			parentPort string
			upPort     string
		}{parentIP: sw.ParentIP, parentPort: sw.ParentPort, upPort: sw.UplinkPort}
		uplinkPorts[sw.ParentIP+":"+sw.ParentPort] = true
		if sw.UplinkPort != "" {
			uplinkPorts[sw.IP+":"+sw.UplinkPort] = true
		}
	}

	// Group forwarding table entries per switch/port, skipping CPU port and
	// ignored MACs.
	type macRef struct {
		mac  string
		vlan string
	}
	swPortMACs := make(map[string]map[string][]macRef)
	for ip, sw := range switchesByIP {
		pm := make(map[string][]macRef)
		for _, e := range sw.MACTable {
			port := e.Port
			if port == "9" {
				continue // CPU port
			}
			mac := NormalizeMAC(e.MAC)
			if mac == "" || s.isIgnoredMAC(e.MAC) {
				continue
			}
			pm[port] = append(pm[port], macRef{mac: mac, vlan: e.VLAN})
		}
		swPortMACs[ip] = pm
	}

	// Active clients: MACs learned on ports that do not carry another switch.
	activeClients := make(map[string]config.ClientEntry)
	for ip, pm := range swPortMACs {
		for port, refs := range pm {
			if uplinkPorts[ip+":"+port] {
				continue
			}
			hasSwitch := false
			for _, ref := range refs {
				if _, ok := switchMACToIP[ref.mac]; ok {
					hasSwitch = true
					break
				}
			}
			if hasSwitch {
				continue
			}
			for _, ref := range refs {
				if _, ok := switchMACToIP[ref.mac]; ok {
					continue
				}
				if _, ok := infraByMAC[ref.mac]; ok {
					continue
				}
				e := activeClients[ref.mac]
				e.MAC = FormatMAC(ref.mac)
				e.IP = ip
				e.Port = port
				e.VLAN = ref.vlan
				activeClients[ref.mac] = e
			}
		}
	}

	// Persist online/offline status into the client database.
	if s.updateClientStatuses(activeClients, liveSwitchIPs) {
		s.saveClients()
	}

	nodes := make([]TopologyNode, 0, len(switches)+len(infraByMAC)+len(activeClients))
	links := make([]TopologyLink, 0)

	// ---------- Switch nodes ----------
	for ip, sw := range switchesByIP {
		nodes = append(nodes, TopologyNode{
			ID:     ip,
			Name:   sw.Name,
			Type:   "switch",
			IP:     ip,
			MAC:    sw.MAC,
			Model:  sw.Model,
			Status: sw.Status,
		})
	}

	// ---------- Infra devices ----------
	infraConn := make(map[string]string) // mac -> node id it is attached to
	for mac, dev := range infraByMAC {
		status := "offline"
		for ip, pm := range swPortMACs {
			for _, refs := range pm {
				hasSwitch := false
				for _, ref := range refs {
					if _, ok := switchMACToIP[ref.mac]; ok {
						hasSwitch = true
						break
					}
				}
				if hasSwitch {
					continue
				}
				for _, ref := range refs {
					if ref.mac == mac {
						status = "online"
						infraConn[mac] = ip
						break
					}
				}
			}
		}
		vendor := ""
		if s.OUI != nil {
			vendor = s.OUI.Lookup(dev.MAC)
		}
		nodes = append(nodes, TopologyNode{
			ID:     mac,
			Name:   dev.Name,
			Type:   dev.Type,
			MAC:    FormatMAC(mac),
			Vendor: vendor,
			Status: status,
		})
	}

	// ---------- Unmanaged switches ----------
	type unmanagedInfo struct {
		id         string
		name       string
		parentIP   string
		parentPort string
	}
	unmanagedByPort := make(map[string]unmanagedInfo)
	for _, us := range s.Config.UnmanagedSwitches() {
		if us.Name == "" || us.ParentIP == "" || us.ParentPort == "" {
			continue
		}
		key := us.ParentIP + ":" + us.ParentPort
		info := unmanagedInfo{
			id:         "unmanaged_" + key,
			name:       us.Name,
			parentIP:   us.ParentIP,
			parentPort: us.ParentPort,
		}
		unmanagedByPort[key] = info
		parentStatus := "offline"
		if _, ok := switchesByIP[us.ParentIP]; ok {
			parentStatus = "online"
		}
		nodes = append(nodes, TopologyNode{
			ID:         info.id,
			Name:       us.Name,
			Type:       "unmanaged_switch",
			Status:     parentStatus,
			ParentIP:   us.ParentIP,
			ParentPort: us.ParentPort,
		})
		links = append(links, TopologyLink{
			Source:     us.ParentIP,
			Target:     info.id,
			SourcePort: "Port " + us.ParentPort,
			Type:       "infra",
			Speed:      s.portSpeed(us.ParentIP, us.ParentPort),
		})
	}

	// ---------- Static switch uplinks ----------
	for swIP, st := range staticParent {
		speed, tx, rx := s.portInfo(st.parentIP, st.parentPort)
		if speed == "" && st.upPort != "" {
			speed, tx, rx = s.portInfo(swIP, st.upPort)
		}
		links = append(links, TopologyLink{
			Source:     st.parentIP,
			Target:     swIP,
			SourcePort: "Port " + st.parentPort,
			TargetPort: st.upPort,
			Speed:      speed,
			TXBPS:      tx,
			RXBPS:      rx,
			Type:       "uplink",
		})
	}

	// ---------- Auto-discovered switch links ----------
	processed := make(map[string]bool)
	for _, l := range links {
		if l.Source != "" && l.Target != "" {
			processed[linkKey(l.Source, l.Target)] = true
		}
	}
	for srcIP, pm := range swPortMACs {
		for port, refs := range pm {
			var neighborIPs []string
			seen := make(map[string]bool)
			for _, ref := range refs {
				if dstIP, ok := switchMACToIP[ref.mac]; ok && !seen[dstIP] {
					seen[dstIP] = true
					neighborIPs = append(neighborIPs, dstIP)
				}
			}
			if len(neighborIPs) == 0 {
				continue
			}
			for _, dstIP := range neighborIPs {
				if dstIP == srcIP {
					continue
				}
				// Respect configured static parent relationships.
				if st, ok := staticParent[dstIP]; ok && st.parentIP != srcIP {
					continue
				}
				key := linkKey(srcIP, dstIP)
				if processed[key] {
					continue
				}
				processed[key] = true
				speed, tx, rx := s.portInfo(srcIP, port)
				links = append(links, TopologyLink{
					Source:     srcIP,
					Target:     dstIP,
					SourcePort: "Port " + port,
					Speed:      speed,
					TXBPS:      tx,
					RXBPS:      rx,
					Type:       "uplink",
				})
			}
		}
	}

	// ---------- Infra links ----------
	for mac, ip := range infraConn {
		port := s.clientPortOn(ip, mac)
		speed := s.portSpeed(ip, port)
		source := ip
		sourcePort := "Port " + port
		if info, ok := unmanagedByPort[ip+":"+port]; ok {
			source = info.id
			sourcePort = ""
		}
		links = append(links, TopologyLink{
			Source:     source,
			Target:     mac,
			SourcePort: sourcePort,
			Speed:      speed,
			Type:       "infra",
		})
	}

	// ---------- Client nodes & links ----------
	clientSet := make(map[string]bool)
	for mac := range activeClients {
		clientSet[mac] = true
	}

	knownClients := s.AllClients()
	sort.Slice(knownClients, func(i, j int) bool {
		return knownClients[i].MAC < knownClients[j].MAC
	})

	for _, c := range knownClients {
		key := NormalizeMAC(c.MAC)
		if c.MAC == "" || key == "" {
			continue
		}
		if s.isIgnoredMAC(c.MAC) {
			continue
		}
		if _, ok := switchMACToIP[key]; ok {
			continue
		}
		if _, ok := infraByMAC[key]; ok {
			continue
		}

		isActive := clientSet[key]
		ip := c.IP
		port := c.Port
		status := c.Status
		if isActive {
			ip = activeClients[key].IP
			port = activeClients[key].Port
			status = "online"
		}
		if status == "" && isActive {
			status = "online"
		}
		if status == "" && !isActive {
			status = "offline"
		}
		if ip == "" {
			continue
		}
		if _, ok := switchesByIP[ip]; !ok {
			continue // parent switch no longer present
		}

		vendor := c.Vendor
		if vendor == "" && s.OUI != nil {
			vendor = s.OUI.Lookup(c.MAC)
		}
		name := clientName(c.MAC, c.Host, vendor)

		nodes = append(nodes, TopologyNode{
			ID:           key,
			Name:         name,
			Type:         "client",
			MAC:          FormatMAC(key),
			Vendor:       vendor,
			Host:         c.Host,
			DeviceType:   c.DeviceType,
			Status:       status,
			LastSeenIP:   ip,
			LastSeenPort: port,
			LastSeenTime: float64(c.LastSeen),
		})

		// Link the client to the correct parent node.
		source := ip
		sourcePort := "Port " + port
		if info, ok := unmanagedByPort[ip+":"+port]; ok {
			source = info.id
			sourcePort = ""
		}
		speed := ""
		if isActive {
			speed = s.portSpeed(ip, port)
		}
		links = append(links, TopologyLink{
			Source:     source,
			Target:     key,
			SourcePort: sourcePort,
			Speed:      speed,
			Type:       "client",
		})
	}

	// Make output deterministic, with switches rendered first.
	sort.Slice(nodes, func(i, j int) bool {
		ri, rj := nodeTypeRank(nodes[i].Type), nodeTypeRank(nodes[j].Type)
		if ri != rj {
			return ri < rj
		}
		return nodes[i].ID < nodes[j].ID
	})
	sort.Slice(links, func(i, j int) bool {
		a, b := links[i], links[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Target < b.Target
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Topology{Nodes: nodes, Links: links})
}

// clientPortOn returns the port on switch ip where a given MAC is learned.
func (s *Server) clientPortOn(ip, mac string) string {
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		return ""
	}
	clean := NormalizeMAC(mac)
	for _, e := range sw.MACTable {
		if NormalizeMAC(e.MAC) == clean {
			return e.Port
		}
	}
	return ""
}

// portSpeed returns the display speed for a switch port.
func (s *Server) portSpeed(ip, port string) string {
	speed, _, _ := s.portInfo(ip, port)
	return speed
}

// portInfo returns speed, TX and RX bps for a switch port.
func (s *Server) portInfo(ip, port string) (string, int64, int64) {
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		return "", 0, 0
	}
	p := strings.ToLower(port)
	for _, ps := range sw.Ports {
		if strings.ToLower(ps.Port) == p || strings.ToLower(ps.Port) == "port "+p {
			return ps.Speed, ps.SpeedTX, ps.SpeedRX
		}
	}
	return "", 0, 0
}

func linkKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func nodeTypeRank(t string) int {
	switch t {
	case "switch":
		return 0
	case "router", "internet":
		return 1
	case "unmanaged_switch":
		return 2
	case "repeater":
		return 3
	default:
		return 4
	}
}

func clientName(mac, host, vendor string) string {
	if host != "" {
		return host
	}
	if vendor != "" {
		return vendor
	}
	if len(mac) >= 8 {
		return "Client " + mac[len(mac)-8:]
	}
	return "Client " + mac
}
