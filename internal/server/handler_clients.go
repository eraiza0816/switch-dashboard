package server

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

func (s *Server) handleAPIUpdateHost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MAC  string `json:"mac"`
		Host string `json:"host"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	key := NormalizeMAC(req.MAC)
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "mac required")
		return
	}

	s.clientsMu.Lock()
	e, ok := s.Clients[key]
	if !ok {
		e = config.ClientEntry{MAC: FormatMAC(key), Status: "offline"}
	}
	e.Host = strings.TrimSpace(req.Host)
	s.Clients[key] = e
	s.clientsMu.Unlock()

	s.saveClients()
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIUpdateType(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MAC  string `json:"mac"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	key := NormalizeMAC(req.MAC)
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "mac required")
		return
	}

	s.clientsMu.Lock()
	e, ok := s.Clients[key]
	if !ok {
		e = config.ClientEntry{MAC: FormatMAC(key), Status: "offline"}
	}
	e.DeviceType = strings.TrimSpace(req.Type)
	s.Clients[key] = e
	s.clientsMu.Unlock()

	s.saveClients()
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIClientDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MAC string `json:"mac"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	key := NormalizeMAC(req.MAC)
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "mac required")
		return
	}

	s.clientsMu.Lock()
	delete(s.Clients, key)
	s.clientsMu.Unlock()
	s.saveClients()

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleAPIClientImportCSV(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "no file part in the request")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	imported := 0

	s.clientsMu.Lock()
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.clientsMu.Unlock()
			s.writeError(w, http.StatusBadRequest, "invalid csv")
			return
		}
		if len(row) < 2 {
			continue
		}
		host := strings.TrimSpace(row[0])
		macRaw := strings.TrimSpace(row[1])
		c0 := strings.ToLower(host)
		c1 := strings.ToLower(macRaw)
		if (strings.Contains(c0, "host") || strings.Contains(c0, "name")) && strings.Contains(c1, "mac") {
			continue // header row
		}
		key := NormalizeMAC(macRaw)
		if len(key) != 12 {
			continue
		}
		e, ok := s.Clients[key]
		if !ok {
			e = config.ClientEntry{MAC: FormatMAC(key), Status: "offline"}
		}
		e.Host = host
		s.Clients[key] = e
		imported++
	}
	s.clientsMu.Unlock()

	if imported > 0 {
		s.saveClients()
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "imported": imported})
}
