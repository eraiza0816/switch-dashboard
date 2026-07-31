package server

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

// ClientHost returns the persisted hostname override for a MAC address
// (legacy helper used by topology rendering).
func (s *Server) ClientHost(mac string) string {
	e, ok := s.GetClient(mac)
	if !ok {
		return ""
	}
	return e.Host
}

// GetClient returns the persisted client entry for a MAC address.
func (s *Server) GetClient(mac string) (config.ClientEntry, bool) {
	key := NormalizeMAC(mac)
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	e, ok := s.Clients[key]
	return e, ok
}

// AllClients returns a snapshot of the persisted client database.
func (s *Server) AllClients() []config.ClientEntry {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	out := make([]config.ClientEntry, 0, len(s.Clients))
	for _, e := range s.Clients {
		out = append(out, e)
	}
	return out
}

func (s *Server) upsertClient(e config.ClientEntry) {
	key := NormalizeMAC(e.MAC)
	if key == "" {
		return
	}
	if e.MAC == "" {
		e.MAC = FormatMAC(key)
	}
	s.clientsMu.Lock()
	s.Clients[key] = e
	s.clientsMu.Unlock()
}

func (s *Server) deleteClient(mac string) bool {
	key := NormalizeMAC(mac)
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	if _, ok := s.Clients[key]; !ok {
		return false
	}
	delete(s.Clients, key)
	return true
}

// updateClientStatuses reconciles the persisted client database against the
// set of active clients (currently present in a switch forwarding table).
// Returns true if the database changed and should be persisted.
func (s *Server) updateClientStatuses(active map[string]config.ClientEntry, liveSwitchIPs map[string]bool) bool {
	now := time.Now().Unix()
	changed := false

	s.clientsMu.Lock()
	for mac, entry := range active {
		key := NormalizeMAC(mac)
		existing, ok := s.Clients[key]
		if ok {
			// Preserve user-provided host override and device type.
			if existing.Host == "" {
				existing.Host = entry.Host
			}
			if existing.DeviceType == "" {
				existing.DeviceType = entry.DeviceType
			}
			existing.MAC = FormatMAC(key)
			existing.IP = entry.IP
			existing.Port = entry.Port
			existing.VLAN = entry.VLAN
			existing.Status = "online"
			existing.LastSeen = now
			s.Clients[key] = existing
		} else {
			s.Clients[key] = config.ClientEntry{
				MAC:      FormatMAC(key),
				Host:     entry.Host,
				IP:       entry.IP,
				Port:     entry.Port,
				VLAN:     entry.VLAN,
				Status:   "online",
				LastSeen: now,
			}
		}
		changed = true
	}

	// Mark previously-online clients offline when their parent switch is still
	// being polled but the MAC is no longer present.
	for mac, e := range s.Clients {
		lastIP := e.IP
		if lastIP == "" {
			lastIP = e.ScannerIP
		}
		if e.Status == "online" && lastIP != "" && liveSwitchIPs[lastIP] {
			if _, isActive := active[mac]; !isActive {
				e.Status = "offline"
				s.Clients[mac] = e
				changed = true
			}
		}
	}
	s.clientsMu.Unlock()

	return changed
}

func (s *Server) loadClients() {
	if s.ClientsPath == "" {
		return
	}
	data, err := os.ReadFile(s.ClientsPath)
	if err != nil {
		return
	}

	// Preferred format: map[mac]ClientEntry
	var entries map[string]config.ClientEntry
	if err := json.Unmarshal(data, &entries); err == nil {
		s.clientsMu.Lock()
		if entries == nil {
			entries = make(map[string]config.ClientEntry)
		}
		for k, e := range entries {
			s.Clients[NormalizeMAC(k)] = e
		}
		s.clientsMu.Unlock()
		return
	}

	// Legacy format: map[mac]hostname string -> migrate
	var legacy map[string]string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return
	}
	s.clientsMu.Lock()
	for k, host := range legacy {
		key := NormalizeMAC(k)
		s.Clients[key] = config.ClientEntry{
			MAC:  FormatMAC(key),
			Host: host,
		}
	}
	s.clientsMu.Unlock()
}

func (s *Server) saveClients() {
	if s.ClientsPath == "" {
		return
	}
	s.clientsMu.RLock()
	out := make(map[string]config.ClientEntry, len(s.Clients))
	for k, e := range s.Clients {
		out[k] = e
	}
	s.clientsMu.RUnlock()
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(s.ClientsPath, data, 0644); err != nil {
		s.Logger.Error("failed to save clients", "error", err)
	}
}

// isIgnoredMAC reports whether a MAC matches one of the configured ignore
// patterns (e.g. "AA:BB:CC:*" or a full MAC).
func (s *Server) isIgnoredMAC(mac string) bool {
	clean := NormalizeMAC(mac)
	if clean == "" {
		return false
	}
	for _, pat := range s.Config.IgnoredMACs() {
		p := strings.TrimSpace(strings.ToUpper(pat))
		if p == "" {
			continue
		}
		if strings.HasSuffix(p, "*") {
			prefix := NormalizeMAC(strings.TrimSuffix(p, "*"))
			if prefix != "" && strings.HasPrefix(clean, prefix) {
				return true
			}
		} else {
			full := NormalizeMAC(p)
			if full != "" && clean == full {
				return true
			}
		}
	}
	return false
}

// NormalizeMAC strips separators and upper-cases a MAC address.
func NormalizeMAC(mac string) string {
	var b strings.Builder
	for _, c := range mac {
		if c >= 'a' && c <= 'z' {
			b.WriteRune(c - ('a' - 'A'))
		} else if c != ':' && c != '-' && c != ' ' && c != '.' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// FormatMAC renders a normalized MAC (12 hex chars) as AA:BB:CC:DD:EE:FF.
func FormatMAC(clean string) string {
	if len(clean) != 12 {
		return clean
	}
	var b strings.Builder
	for i := 0; i < len(clean); i += 2 {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteString(clean[i : i+2])
	}
	return b.String()
}
