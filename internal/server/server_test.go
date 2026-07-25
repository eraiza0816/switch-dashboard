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

func TestAPIHistoryNilStore(t *testing.T) {
	s := newTestServer()
	s.HistoryStore = nil
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/history?ip=192.168.1.1&port=1&range=live", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with nil store, got %d", w.Code)
	}
}

func TestAPIHistoryAllRanges(t *testing.T) {
	s := newTestServer()
	for _, rng := range []string{"live", "1h", "24h", ""} {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/history?ip=10.0.0.1&port=5&range="+rng, nil)
		s.Router.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("range=%q: expected 200, got %d", rng, w.Code)
		}
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

func TestAPIUpdateHost(t *testing.T) {
	s := newTestServer()

	// Save a host override for a MAC
	body := `{"mac":"AA:BB:CC:DD:EE:01","host":"MyDevice"}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/clients/update_host", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify topology shows the override
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w2, r2)
	var topo Topology
	if err := json.Unmarshal(w2.Body.Bytes(), &topo); err != nil {
		t.Fatalf("unmarshal topology: %v", err)
	}
	found := false
	for _, n := range topo.Nodes {
		if n.MAC == "AA:BB:CC:DD:EE:01" {
			found = true
			if n.Name != "MyDevice" {
				t.Fatalf("expected name 'MyDevice', got %q", n.Name)
			}
		}
	}
	if !found {
		t.Fatal("client node AA:BB:CC:DD:EE:01 not found in topology")
	}
}

func TestClientHostPersistence(t *testing.T) {
	path := t.TempDir() + "/clients.json"

	cache1 := NewCache()
	cache1.UpdateSwitch("192.168.1.1", &SwitchData{
		Name: "Test Switch", IP: "192.168.1.1", Model: "RTLPlayground",
		MAC: "AA:BB:CC:DD:EE:FF", Hostname: "test-switch", Status: "online",
		MACTable: []MACEntry{{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001"}},
	})
	s1 := NewServer(cache1, testConfig{}, testLogger{}, nil, nil)
	s1.ClientHostsPath = path
	body := `{"mac":"AA:BB:CC:DD:EE:01","host":"MyDevice"}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/clients/update_host", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s1.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("save: expected 200, got %d", w.Code)
	}

	// Create a new server and load from file
	cache2 := NewCache()
	cache2.UpdateSwitch("192.168.1.1", &SwitchData{
		Name: "Test Switch", IP: "192.168.1.1", Model: "RTLPlayground",
		MAC: "AA:BB:CC:DD:EE:FF", Hostname: "test-switch", Status: "online",
		MACTable: []MACEntry{{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001"}},
	})
	s2 := NewServer(cache2, testConfig{}, testLogger{}, nil, nil)
	s2.ClientHostsPath = path
	s2.loadClientHosts()

	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s2.Router.ServeHTTP(w2, r2)
	var topo Topology
	json.Unmarshal(w2.Body.Bytes(), &topo)
	var node *TopologyNode
	for i := range topo.Nodes {
		if topo.Nodes[i].MAC == "AA:BB:CC:DD:EE:01" {
			node = &topo.Nodes[i]
			break
		}
	}
	if node == nil {
		t.Fatal("client node not found after reload")
	}
	if node.Name != "MyDevice" {
		t.Fatalf("expected MyDevice after reload, got %s", node.Name)
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

func TestAPIGetLayoutPositions(t *testing.T) {
	s := newTestServer()
	s.LayoutPositionsPath = t.TempDir() + "/layout.json"

	// Empty positions should return empty object
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/layout_positions", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]map[string]float64
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestAPISaveLayoutPositions(t *testing.T) {
	s := newTestServer()
	s.LayoutPositionsPath = t.TempDir() + "/layout.json"

	// Save some positions
	body := `{"node1":{"x":100,"y":200},"node2":{"x":300,"y":400}}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/layout_positions", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify positions were saved
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/layout_positions", nil)
	s.Router.ServeHTTP(w2, r2)
	var result map[string]map[string]float64
	if err := json.NewDecoder(w2.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["node1"]["x"] != 100 || result["node1"]["y"] != 200 {
		t.Fatalf("unexpected node1 position: %+v", result["node1"])
	}
	if result["node2"]["x"] != 300 || result["node2"]["y"] != 400 {
		t.Fatalf("unexpected node2 position: %+v", result["node2"])
	}
}

func TestAPISaveLayoutPositionsInvalidBody(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/layout_positions", strings.NewReader(`not json`))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}
