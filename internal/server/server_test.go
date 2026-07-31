package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eraiza0816/switch-dashboard/internal/config"
	"github.com/eraiza0816/switch-dashboard/internal/logbuf"
	"github.com/eraiza0816/switch-dashboard/internal/oui"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

func newLogBuf() *logbuf.LogBuffer {
	return logbuf.New(slog.NewTextHandler(io.Discard, nil), 100, slog.LevelDebug)
}

func newTestServerWithLogger(lb *logbuf.LogBuffer) *Server {
	cache := NewCache()
	cache.UpdateSwitch("192.168.1.1", &SwitchData{
		Name: "Test Switch", IP: "192.168.1.1", Model: "RTLPlayground",
		MAC: "AA:BB:CC:DD:EE:FF", Hostname: "test-switch", Status: "online",
		Ports: []PortState{
			{Port: "1", Status: "up", Speed: "1G", TXBytes: 1000, RXBytes: 2000, SpeedTX: 800, SpeedRX: 1600},
			{Port: "2", Status: "down", Speed: "", TXBytes: 0, RXBytes: 0},
		},
		MACTable: []MACEntry{{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "001"}},
	})
	lbLogger := slog.New(lb)
	srv := NewServer(cache, testConfig{}, lbLogger, lb, nil, nil)
	srv.OUI = oui.New()
	return srv
}

type testConfig struct{}

func (t testConfig) Title() string                { return "Test Dashboard" }
func (t testConfig) RefreshInterval() int         { return 30 }
func (t testConfig) EnabledColumns() []string     { return []string{"port", "status", "speed"} }
func (t testConfig) GridColumns() string          { return "auto" }
func (t testConfig) PortsWrapThreshold() int      { return 0 }
func (t testConfig) ColumnWidths() map[string]int { return nil }
func (t testConfig) ColumnOrder() []string        { return nil }
func (t testConfig) Version() string              { return "test" }
func (t testConfig) Switches() []config.SwitchConfig {
	return nil
}
func (t testConfig) InfrastructureDevices() []config.InfraDevice { return nil }
func (t testConfig) UnmanagedSwitches() []config.UnmanagedSwitch { return nil }
func (t testConfig) IgnoredMACs() []string                       { return nil }

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
	srv := NewServer(cache, testConfig{}, testLogger, newLogBuf(), nil, nil)
	srv.OUI = oui.New()
	return srv
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

func TestAPISwitchesExtraFields(t *testing.T) {
	cache := NewCache()
	cache.UpdateSwitch("192.168.1.1", &SwitchData{
		Name: "Test Switch", IP: "192.168.1.1", Model: "RTLPlayground",
		MAC: "AA:BB:CC:DD:EE:FF", Hostname: "test-switch", Status: "online",
		Ports: []PortState{
			{Port: "1", Status: "up", Speed: "1G", TXBytes: 1000, RXBytes: 2000},
		},
		EEE: []EEEStatus{
			{Port: "1", Active: true, Status: "Enabled", LP: "Supported"},
			{Port: "2", Active: false, Status: "Disabled", LP: "Not Supported"},
		},
		VLANList: []VLANItem{
			{ID: 1, Name: "Default"},
			{ID: 100, Name: "Users"},
		},
		LAG: []LAGStatus{
			{Number: 1, Members: "1,2", Hash: "src-mac"},
		},
		Mirror: &MirrorStatus{
			Enabled: true, Port: "5", MirrorRX: "1,2", MirrorTX: "3,4",
		},
		Bandwidth: []BWStatus{
			{Port: "1", InLimit: true, InBW: "100M", OutLimit: false, OutBW: ""},
		},
	})
	srv := NewServer(cache, testConfig{}, testLogger, newLogBuf(), nil, nil)
	srv.OUI = oui.New()

	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches", nil)
	srv.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var data []map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	if len(data) != 1 {
		t.Fatalf("expected 1 switch, got %d", len(data))
	}

	sw := data[0]

	eee := sw["eee"].([]any)
	if len(eee) != 2 {
		t.Fatalf("expected 2 eee entries, got %d", len(eee))
	}
	e0 := eee[0].(map[string]any)
	if e0["port"] != "1" || e0["active"] != true || e0["status"] != "Enabled" {
		t.Fatalf("unexpected eee[0]: %+v", e0)
	}

	vlan := sw["vlan_list"].([]any)
	if len(vlan) != 2 {
		t.Fatalf("expected 2 vlan entries, got %d", len(vlan))
	}
	v0 := vlan[0].(map[string]any)
	if v0["id"] != float64(1) || v0["name"] != "Default" {
		t.Fatalf("unexpected vlan[0]: %+v", v0)
	}

	lag := sw["lag"].([]any)
	if len(lag) != 1 {
		t.Fatalf("expected 1 lag entry, got %d", len(lag))
	}
	l0 := lag[0].(map[string]any)
	if l0["number"] != float64(1) || l0["members"] != "1,2" {
		t.Fatalf("unexpected lag[0]: %+v", l0)
	}

	mirror := sw["mirror"].(map[string]any)
	if mirror["enabled"] != true || mirror["port"] != "5" {
		t.Fatalf("unexpected mirror: %+v", mirror)
	}

	bw := sw["bandwidth"].([]any)
	if len(bw) != 1 {
		t.Fatalf("expected 1 bandwidth entry, got %d", len(bw))
	}
	b0 := bw[0].(map[string]any)
	if b0["port"] != "1" || b0["in_limited"] != true || b0["in_bw"] != "100M" {
		t.Fatalf("unexpected bandwidth[0]: %+v", b0)
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
	s1 := NewServer(cache1, testConfig{}, testLogger, newLogBuf(), nil, nil)
	s1.OUI = oui.New()
	s1.ClientsPath = path
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
	s2 := NewServer(cache2, testConfig{}, testLogger, newLogBuf(), nil, nil)
	s2.OUI = oui.New()
	s2.ClientsPath = path
	s2.loadClients()

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

func TestHealthz(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/healthz", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("body: got %q", w.Body.String())
	}
}

func TestReadyzReady(t *testing.T) {
	s := newTestServer()
	s.DuckDBReady = true
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/readyz", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var data map[string]string
	json.NewDecoder(w.Body).Decode(&data)
	if data["status"] != "ready" {
		t.Fatalf("status: got %q", data["status"])
	}
}

func TestReadyzNotReady(t *testing.T) {
	s := newTestServer()
	s.DuckDBReady = false
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/readyz", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	var data map[string]string
	json.NewDecoder(w.Body).Decode(&data)
	if data["status"] != "not ready" {
		t.Fatalf("status: got %q", data["status"])
	}
}

func TestAPILogsBody(t *testing.T) {
	lb := newLogBuf()
	s := newTestServerWithLogger(lb)
	s.Logger.Info("test log entry")
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/logs", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var entries []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) < 1 {
		t.Fatal("expected at least 1 log entry")
	}
	found := false
	for _, e := range entries {
		if e["message"] == "test log entry" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("test log entry not found in /api/logs response")
	}
}

func TestAPILogsLevelInvalid(t *testing.T) {
	s := newTestServer()
	body := strings.NewReader(`{"level":"INVALID"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/logs/level", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid level, got %d", w.Code)
	}
}

func TestAPILogsLevelChangesBuffer(t *testing.T) {
	lb := newLogBuf()
	s := newTestServerWithLogger(lb)
	body := strings.NewReader(`{"level":"WARN"}`)
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/logs/level", body)
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	s.Logger.Info("should be suppressed")
	entries := s.LogBuffer.Get()
	for _, e := range entries {
		if e.Message == "should be suppressed" {
			t.Fatal("info log was not suppressed after setting log level to warn")
		}
	}
}

func TestAPILogsDownload(t *testing.T) {
	lb := newLogBuf()
	s := newTestServerWithLogger(lb)
	s.Logger.Info("downloadable", "key", "val")
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/logs/download", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("content-type: got %q", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(body, "downloadable") {
		t.Fatal("download body should contain log message")
	}
	if !strings.Contains(body, "key") || !strings.Contains(body, "val") {
		t.Fatal("download body should contain log attributes")
	}
}

func TestAPIConfigReloadWithoutCallback(t *testing.T) {
	s := newTestServer()
	s.ConfigReload = nil
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/config/reload", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without callback, got %d", w.Code)
	}
}

func TestAPIConfigReloadSuccess(t *testing.T) {
	s := newTestServer()
	reloaded := false
	s.ConfigReload = func() error {
		reloaded = true
		return nil
	}
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/config/reload", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !reloaded {
		t.Fatal("reload callback was not called")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	s := newTestServer()
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/metrics", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "switchd_http_requests_total") {
		t.Fatal("metrics should contain switchd_http_requests_total")
	}
	if !strings.Contains(body, "go_goroutines") {
		t.Fatal("metrics should contain go_goroutines (runtime metric)")
	}
}

func TestMetricsRecordedAfterRequest(t *testing.T) {
	s := newTestServer()
	// Make a request that goes through the metrics middleware
	w1, r1 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches", nil)
	s.Router.ServeHTTP(w1, r1)

	// Check metrics output includes the request
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/metrics", nil)
	s.Router.ServeHTTP(w2, r2)
	body := w2.Body.String()
	if !strings.Contains(body, `switchd_http_requests_total{method="GET",path="/api/switches",status="200"}`) {
		t.Fatal("metrics should include the recorded API request")
	}
}
