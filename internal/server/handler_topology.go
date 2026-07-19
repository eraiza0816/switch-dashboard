package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleAPITopology(w http.ResponseWriter, r *http.Request) {
	switches := s.Cache.GetSwitches()
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

		for _, entry := range sw.MACTable {
			mac := entry.MAC
			nodes = append(nodes, TopologyNode{
				ID:     mac,
				Name:   "Client " + mac[len(mac)-8:],
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
