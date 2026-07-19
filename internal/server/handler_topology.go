package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPITopology(w http.ResponseWriter, r *http.Request) {
	switches := s.Cache.GetSwitches()
	seenMAC := make(map[string]bool)
	nodes := []TopologyNode{}
	links := []TopologyLink{}

	for _, sw := range switches {
		nodes = append(nodes, TopologyNode{
			ID:     sw.IP,
			Name:   sw.Name,
			Type:   "switch",
			IP:     sw.IP,
			MAC:    sw.MAC,
			Model:  sw.Model,
			Status: sw.Status,
		})

		swMAC := normalizeMAC(sw.MAC)

		for _, entry := range sw.MACTable {
			mac := entry.MAC
			normMAC := normalizeMAC(mac)

			// Skip CPU port (port 9 on RTL8372/3) and switch's own MAC
			if entry.Port == "9" || normMAC == swMAC || seenMAC[normMAC] {
				continue
			}
			seenMAC[normMAC] = true

			nodes = append(nodes, TopologyNode{
				ID:     mac,
				Name:   clientName(mac, entry.Host, entry.Vendor),
				Type:   "client",
				MAC:    mac,
				Status: "online",
			})
			links = append(links, TopologyLink{
				Source:     sw.IP,
				Target:     mac,
				SourcePort: "Port " + entry.Port,
				Type:       "client",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Topology{Nodes: nodes, Links: links})
}

func normalizeMAC(mac string) string {
	if len(mac) < 2 {
		return ""
	}
	b := make([]byte, 0, len(mac))
	for _, c := range mac {
		if c != ':' && c != '-' && c != ' ' {
			b = append(b, byte(c))
		}
	}
	return string(b)
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
