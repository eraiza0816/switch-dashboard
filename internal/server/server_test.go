package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testLogger struct{}

func (t testLogger) Info(msg string, args ...any)  {}
func (t testLogger) Warn(msg string, args ...any)  {}
func (t testLogger) Error(msg string, args ...any) {}
func (t testLogger) Debug(msg string, args ...any) {}

type testConfig struct{}

func (t testConfig) Title() string                     { return "Test Dashboard" }
func (t testConfig) RefreshInterval() int              { return 30 }
func (t testConfig) EnabledColumns() []string          { return []string{"port", "status", "speed"} }
func (t testConfig) GridColumns() string               { return "auto" }
func (t testConfig) PortsWrapThreshold() int           { return 0 }
func (t testConfig) ColumnWidths() map[string]int      { return nil }
func (t testConfig) ColumnOrder() []string             { return nil }
func (t testConfig) Version() string                   { return "test" }

func newTestServer() *Server {
	cache := NewCache()
	cache.UpdateSwitch("192.168.1.1", &SwitchData{
		Name:     "Test Switch",
		IP:       "192.168.1.1",
		Model:    "RTLPlayground",
		MAC:      "AA:BB:CC:DD:EE:FF",
		Hostname: "test-switch",
		Status:   "online",
		Ports: []PortState{
			{Port: "1", Status: "up", Speed: "1G", TXBytes: 1000, RXBytes: 2000, SpeedTX: 800, SpeedRX: 1600},
			{Port: "2", Status: "down", Speed: "", TXBytes: 0, RXBytes: 0},
		},
		MACTable: []MACEntry{
			{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001"},
		},
	})

	return NewServer(cache, testConfig{}, testLogger{}, nil, nil)
}

func TestAPISwitches(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/switches", nil)
	s.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var data []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 switch, got %d", len(data))
	}
	if data[0]["name"] != "Test Switch" {
		t.Fatalf("name: got %q", data[0]["name"])
	}
	ports := data[0]["ports"].([]any)
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
}

func TestAPISpeeds(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/speeds", nil)
	s.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPITopology(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var top Topology
	if err := json.Unmarshal(w.Body.Bytes(), &top); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if len(top.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes (switch + client), got %d", len(top.Nodes))
	}
}

func TestAPISettings(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/settings", nil)
	s.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func Test404(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/nonexistent", nil)
	s.Router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
