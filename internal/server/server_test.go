package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testLogger struct{}

func (t testLogger) Info(msg string, args ...any)  {}
func (t testLogger) Warn(msg string, args ...any)  {}
func (t testLogger) Error(msg string, args ...any) {}
func (t testLogger) Debug(msg string, args ...any) {}

type testConfig struct{}

func (t testConfig) Title() string                { return "Test Dashboard" }
func (t testConfig) RefreshInterval() int         { return 30 }
func (t testConfig) EnabledColumns() []string     { return []string{"port", "status", "speed"} }
func (t testConfig) GridColumns() string          { return "auto" }
func (t testConfig) PortsWrapThreshold() int      { return 0 }
func (t testConfig) ColumnWidths() map[string]int { return nil }
func (t testConfig) ColumnOrder() []string        { return nil }
func (t testConfig) Version() string              { return "test" }

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
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data []map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
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
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/speeds", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	if len(data) != 1 {
		t.Fatalf("expected 1 switch in speeds, got %d", len(data))
	}
}

func TestAPITopology(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var top Topology
	json.Unmarshal(w.Body.Bytes(), &top)
	if len(top.Nodes) < 2 {
		t.Fatalf("expected >=2 nodes, got %d", len(top.Nodes))
	}
}

func TestAPISettings(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/settings", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIOpenAPI(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/openapi.json", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var spec map[string]any
	json.Unmarshal(w.Body.Bytes(), &spec)
	paths := spec["paths"].(map[string]any)
	if len(paths) < 15 {
		t.Fatalf("expected >=15 API paths, got %d", len(paths))
	}
	info := spec["info"].(map[string]any)
	if info["version"] != "0.1.0" {
		t.Fatalf("version: got %q, want 0.1.0", info["version"])
	}
}

func TestAPIRefreshMAC(t *testing.T) {
	s := newTestServer()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/switches/192.168.1.1/refresh_mac", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	if data["status"] != "ok" {
		t.Fatalf("expected ok, got %v", data["status"])
	}
}

func TestAPIHistory(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/history?ip=192.168.1.1&port=1&range=live", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPINotes(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader(`{"key":"192.168.1.1:1","note":"uplink"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/notes", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIReset(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/reset", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIBackup(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/switches/192.168.1.1/backup", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIReboot(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/switches/192.168.1.1/reboot", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPILogs(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/logs", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPILogsLevel(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader(`{"level":"DEBUG"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/logs/level", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPILogsClear(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/logs/clear", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIListBackups(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/backups", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIVendors(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/vendors", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIUpdateOUI(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/vendors/update_oui", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPISettingsSave(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader(`{"font_size":"lg"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/settings", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAPIConfigSettings(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/config/settings", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSwitchNotFound(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/switches/9.9.9.9/refresh_mac", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown switch, got %d", w.Code)
	}
}

func Test404(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/nonexistent", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestConfigSave(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader("title=Test&refresh_interval=30")
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/config", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusFound {
		t.Fatalf("config POST: expected 302, got %d", w.Code)
	}
}

func TestSFPHandler(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches/192.168.1.1/transceiver", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	if data["vendor_name"] == nil {
		t.Fatal("expected vendor_name in transceiver response")
	}
}

func TestCmdEndpoint(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader(`{"cmd":"show"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/switches/192.168.1.1/cmd", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", data["status"])
	}
}

func TestHTMLPages(t *testing.T) {
	s := newTestServer()
	for _, path := range []string{"/", "/logs", "/backups", "/config", "/api-docs"} {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", path, nil)
		s.Router.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("page %s returned %d", path, w.Code)
		}
		if len(w.Body.Bytes()) < 40 {
			t.Errorf("page %s body too short: %d bytes", path, len(w.Body.Bytes()))
		}
	}
}
