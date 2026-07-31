package server

import (
	"net/http"
	"sort"
	"strings"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

// topologyState holds all intermediate data needed to build the topology
// graph from the switch cache and configuration.
type topologyState struct {
	s               *Server
	switchesByIP    map[string]*SwitchData
	switchMACToIP   map[string]string
	liveSwitchIPs   map[string]bool
	infraByMAC      map[string]config.InfraDevice
	staticParent    map[string]staticUplink
	uplinkPorts     map[string]bool
	unmanagedByPort map[string]unmanagedInfo
	swPortMACs      map[string]map[string][]macRef
	switchPorts     map[string]bool // "ip:port" -> port carries another switch
}

type macRef struct {
	mac  string
	vlan string
}

type staticUplink struct {
	parentIP   string
	parentPort string
	upPort     string
}

type unmanagedInfo struct {
	id         string
	name       string
	parentIP   string
	parentPort string
}

// infraConn records where an infrastructure device is attached.
type infraConn struct {
	ip   string
	port string
}

func (s *Server) newTopologyState(switches []*SwitchData) *topologyState {
	st := &topologyState{
		s:               s,
		switchesByIP:    make(map[string]*SwitchData, len(switches)),
		switchMACToIP:   make(map[string]string),
		liveSwitchIPs:   make(map[string]bool, len(switches)),
		infraByMAC:      make(map[string]config.InfraDevice),
		staticParent:    make(map[string]staticUplink),
		uplinkPorts:     make(map[string]bool),
		unmanagedByPort: make(map[string]unmanagedInfo),
		switchPorts:     make(map[string]bool),
	}
	st.collectSwitches(switches)
	st.collectInfra()
	st.collectStaticParents()
	st.collectUnmanaged()
	st.buildSwPortMACs()
	return st
}

func (st *topologyState) collectSwitches(switches []*SwitchData) {
	for _, sw := range switches {
		st.switchesByIP[sw.IP] = sw
		st.liveSwitchIPs[sw.IP] = true
		if m := NormalizeMAC(sw.MAC); m != "" {
			st.switchMACToIP[m] = sw.IP
		}
	}
}

// collectInfra loads infrastructure devices from config, skipping devices
// that are actually monitored switches.
func (st *topologyState) collectInfra() {
	for _, dev := range st.s.Config.InfrastructureDevices() {
		m := NormalizeMAC(dev.MAC)
		if m == "" {
			continue
		}
		if _, ok := st.switchMACToIP[m]; ok {
			continue
		}
		st.infraByMAC[m] = dev
	}
}

// collectStaticParents loads static switch uplinks from config. These ports
// carry a link to another switch and must never be treated as client ports.
func (st *topologyState) collectStaticParents() {
	for _, sw := range st.s.Config.Switches() {
		if !sw.Enabled || sw.ParentIP == "" || sw.ParentPort == "" {
			continue
		}
		st.staticParent[sw.IP] = staticUplink{
			parentIP:   sw.ParentIP,
			parentPort: sw.ParentPort,
			upPort:     sw.UplinkPort,
		}
		st.uplinkPorts[sw.ParentIP+":"+sw.ParentPort] = true
		if sw.UplinkPort != "" {
			st.uplinkPorts[sw.IP+":"+sw.UplinkPort] = true
		}
	}
}

func (st *topologyState) collectUnmanaged() {
	for _, us := range st.s.Config.UnmanagedSwitches() {
		if us.Name == "" || us.ParentIP == "" || us.ParentPort == "" {
			continue
		}
		key := us.ParentIP + ":" + us.ParentPort
		st.unmanagedByPort[key] = unmanagedInfo{
			id:         "unmanaged_" + key,
			name:       us.Name,
			parentIP:   us.ParentIP,
			parentPort: us.ParentPort,
		}
	}
}

// buildSwPortMACs groups forwarding table entries per switch/port, skipping
// the CPU port and ignored MACs, and precomputes which ports carry a switch.
func (st *topologyState) buildSwPortMACs() {
	st.swPortMACs = make(map[string]map[string][]macRef, len(st.switchesByIP))
	for ip, sw := range st.switchesByIP {
		pm := make(map[string][]macRef)
		for _, e := range sw.MACTable {
			port := e.Port
			if port == "9" {
				continue // CPU port
			}
			mac := NormalizeMAC(e.MAC)
			if mac == "" || st.s.isIgnoredMAC(e.MAC) {
				continue
			}
			if _, ok := st.switchMACToIP[mac]; ok {
				st.switchPorts[ip+":"+port] = true
			}
			pm[port] = append(pm[port], macRef{mac: mac, vlan: e.VLAN})
		}
		st.swPortMACs[ip] = pm
	}
}

// activeClients returns MACs learned on ports that do not carry a switch or
// an infrastructure device.
func (st *topologyState) activeClients() map[string]config.ClientEntry {
	active := make(map[string]config.ClientEntry)
	for ip, pm := range st.swPortMACs {
		for port, refs := range pm {
			if st.uplinkPorts[ip+":"+port] || st.switchPorts[ip+":"+port] {
				continue
			}
			for _, ref := range refs {
				if _, ok := st.switchMACToIP[ref.mac]; ok {
					continue
				}
				if _, ok := st.infraByMAC[ref.mac]; ok {
					continue
				}
				e := active[ref.mac]
				e.MAC = FormatMAC(ref.mac)
				e.IP = ip
				e.Port = port
				e.VLAN = ref.vlan
				active[ref.mac] = e
			}
		}
	}
	return active
}

func (st *topologyState) buildSwitchNodes() []TopologyNode {
	nodes := make([]TopologyNode, 0, len(st.switchesByIP))
	for ip, sw := range st.switchesByIP {
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
	return nodes
}

// buildInfraNodes emits infrastructure device nodes and where each is attached.
func (st *topologyState) buildInfraNodes() ([]TopologyNode, map[string]infraConn) {
	nodes := make([]TopologyNode, 0, len(st.infraByMAC))
	conns := make(map[string]infraConn)
	for mac, dev := range st.infraByMAC {
		status := "offline"
		if ip, port, ok := st.findInfraConn(mac); ok {
			status = "online"
			conns[mac] = infraConn{ip: ip, port: port}
		}
		vendor := ""
		if st.s.OUI != nil {
			vendor = st.s.OUI.Lookup(dev.MAC)
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
	return nodes, conns
}

func (st *topologyState) findInfraConn(mac string) (string, string, bool) {
	for ip, pm := range st.swPortMACs {
		for port, refs := range pm {
			if st.switchPorts[ip+":"+port] {
				continue
			}
			for _, ref := range refs {
				if ref.mac == mac {
					return ip, port, true
				}
			}
		}
	}
	return "", "", false
}

func (st *topologyState) buildUnmanagedNodes() ([]TopologyNode, []TopologyLink) {
	nodes := make([]TopologyNode, 0, len(st.unmanagedByPort))
	links := make([]TopologyLink, 0, len(st.unmanagedByPort))
	for _, info := range st.unmanagedByPort {
		parentStatus := "offline"
		if _, ok := st.switchesByIP[info.parentIP]; ok {
			parentStatus = "online"
		}
		nodes = append(nodes, TopologyNode{
			ID:         info.id,
			Name:       info.name,
			Type:       "unmanaged_switch",
			Status:     parentStatus,
			ParentIP:   info.parentIP,
			ParentPort: info.parentPort,
		})
		links = append(links, TopologyLink{
			Source:     info.parentIP,
			Target:     info.id,
			SourcePort: "Port " + info.parentPort,
			Type:       "infra",
			Speed:      st.s.portSpeed(info.parentIP, info.parentPort),
		})
	}
	return nodes, links
}

func (st *topologyState) buildStaticUplinkLinks() []TopologyLink {
	links := make([]TopologyLink, 0, len(st.staticParent))
	for swIP, up := range st.staticParent {
		speed, tx, rx := st.s.portInfo(up.parentIP, up.parentPort)
		if speed == "" && up.upPort != "" {
			speed, tx, rx = st.s.portInfo(swIP, up.upPort)
		}
		links = append(links, TopologyLink{
			Source:     up.parentIP,
			Target:     swIP,
			SourcePort: "Port " + up.parentPort,
			TargetPort: up.upPort,
			Speed:      speed,
			TXBPS:      tx,
			RXBPS:      rx,
			Type:       "uplink",
		})
	}
	return links
}

// buildAutoLinks discovers switch-to-switch links from the forwarding table,
// skipping links already emitted by static uplinks.
func (st *topologyState) buildAutoLinks(processed map[string]bool) []TopologyLink {
	links := make([]TopologyLink, 0)
	for srcIP, pm := range st.swPortMACs {
		for port, refs := range pm {
			var neighborIPs []string
			seen := make(map[string]bool)
			for _, ref := range refs {
				if dstIP, ok := st.switchMACToIP[ref.mac]; ok && !seen[dstIP] {
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
				if up, ok := st.staticParent[dstIP]; ok && up.parentIP != srcIP {
					continue
				}
				key := linkKey(srcIP, dstIP)
				if processed[key] {
					continue
				}
				processed[key] = true
				speed, tx, rx := st.s.portInfo(srcIP, port)
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
	return links
}

func (st *topologyState) buildInfraLinks(conns map[string]infraConn) []TopologyLink {
	links := make([]TopologyLink, 0, len(conns))
	for mac, conn := range conns {
		speed := st.s.portSpeed(conn.ip, conn.port)
		source, sourcePort := st.resolveSource(conn.ip, conn.port)
		links = append(links, TopologyLink{
			Source:     source,
			Target:     mac,
			SourcePort: sourcePort,
			Speed:      speed,
			Type:       "infra",
		})
	}
	return links
}

func (st *topologyState) buildClientNodes(activeClients map[string]config.ClientEntry) ([]TopologyNode, []TopologyLink) {
	clientSet := make(map[string]bool, len(activeClients))
	for mac := range activeClients {
		clientSet[mac] = true
	}

	knownClients := st.s.AllClients()
	sort.Slice(knownClients, func(i, j int) bool {
		return knownClients[i].MAC < knownClients[j].MAC
	})

	nodes := make([]TopologyNode, 0, len(knownClients))
	links := make([]TopologyLink, 0, len(knownClients))
	for _, c := range knownClients {
		node, link, ok := st.buildClientNode(c, clientSet, activeClients)
		if !ok {
			continue
		}
		nodes = append(nodes, node)
		links = append(links, link)
	}
	return nodes, links
}

func (st *topologyState) buildClientNode(c config.ClientEntry, clientSet map[string]bool, activeClients map[string]config.ClientEntry) (TopologyNode, TopologyLink, bool) {
	key := NormalizeMAC(c.MAC)
	if st.skipClient(c.MAC, key) {
		return TopologyNode{}, TopologyLink{}, false
	}

	isActive := clientSet[key]
	ip := c.IP
	port := c.Port
	status := c.Status
	if isActive {
		ac := activeClients[key]
		ip = ac.IP
		port = ac.Port
		status = "online"
	}
	if status == "" {
		if isActive {
			status = "online"
		} else {
			status = "offline"
		}
	}
	if ip == "" {
		return TopologyNode{}, TopologyLink{}, false
	}
	if _, ok := st.switchesByIP[ip]; !ok {
		return TopologyNode{}, TopologyLink{}, false // parent switch no longer present
	}

	vendor := c.Vendor
	if vendor == "" && st.s.OUI != nil {
		vendor = st.s.OUI.Lookup(c.MAC)
	}
	name := clientName(c.MAC, c.Host, vendor)

	node := TopologyNode{
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
	}

	source, sourcePort := st.resolveSource(ip, port)
	speed := ""
	if isActive {
		speed = st.s.portSpeed(ip, port)
	}
	link := TopologyLink{
		Source:     source,
		Target:     key,
		SourcePort: sourcePort,
		Speed:      speed,
		Type:       "client",
	}
	return node, link, true
}

// skipClient reports whether a known client must be excluded from the graph.
func (st *topologyState) skipClient(rawMAC, key string) bool {
	if rawMAC == "" || key == "" {
		return true
	}
	if st.s.isIgnoredMAC(rawMAC) {
		return true
	}
	if _, ok := st.switchMACToIP[key]; ok {
		return true
	}
	if _, ok := st.infraByMAC[key]; ok {
		return true
	}
	return false
}

// resolveSource rewrites the parent node for links whose port belongs to an
// unmanaged switch.
func (st *topologyState) resolveSource(ip, port string) (string, string) {
	if info, ok := st.unmanagedByPort[ip+":"+port]; ok {
		return info.id, ""
	}
	return ip, "Port " + port
}

func (s *Server) handleAPITopology(w http.ResponseWriter, r *http.Request) {
	st := s.newTopologyState(s.Cache.GetSwitches())
	activeClients := st.activeClients()

	// Persist online/offline status into the client database.
	if s.updateClientStatuses(activeClients, st.liveSwitchIPs) {
		s.saveClients()
	}

	nodes := st.buildSwitchNodes()
	infraNodes, infraConn := st.buildInfraNodes()
	unmanagedNodes, unmanagedLinks := st.buildUnmanagedNodes()
	clientNodes, clientLinks := st.buildClientNodes(activeClients)
	nodes = append(nodes, infraNodes...)
	nodes = append(nodes, unmanagedNodes...)
	nodes = append(nodes, clientNodes...)

	links := unmanagedLinks
	links = append(links, st.buildStaticUplinkLinks()...)

	processed := make(map[string]bool)
	for _, l := range links {
		if l.Source != "" && l.Target != "" {
			processed[linkKey(l.Source, l.Target)] = true
		}
	}
	links = append(links, st.buildAutoLinks(processed)...)
	links = append(links, st.buildInfraLinks(infraConn)...)
	links = append(links, clientLinks...)

	sortNodes(nodes)
	sortLinks(links)

	s.writeJSON(w, http.StatusOK, Topology{Nodes: nodes, Links: links})
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

func sortNodes(nodes []TopologyNode) {
	sort.Slice(nodes, func(i, j int) bool {
		ri, rj := nodeTypeRank(nodes[i].Type), nodeTypeRank(nodes[j].Type)
		if ri != rj {
			return ri < rj
		}
		return nodes[i].ID < nodes[j].ID
	})
}

func sortLinks(links []TopologyLink) {
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
