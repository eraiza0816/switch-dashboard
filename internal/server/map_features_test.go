package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eraiza0816/switch-dashboard/internal/config"
	"github.com/eraiza0816/switch-dashboard/internal/logbuf"
	"github.com/eraiza0816/switch-dashboard/internal/oui"
)

type mapTestConfig struct {
	switches  []config.SwitchConfig
	infra     []config.InfraDevice
	unmanaged []config.UnmanagedSwitch
	ignored   []string
}

func (m mapTestConfig) Title() string                { return "Map Test" }
func (m mapTestConfig) RefreshInterval() int         { return 30 }
func (m mapTestConfig) EnabledColumns() []string     { return []string{"port", "status", "speed"} }
func (m mapTestConfig) GridColumns() string          { return "auto" }
func (m mapTestConfig) PortsWrapThreshold() int      { return 0 }
func (m mapTestConfig) ColumnWidths() map[string]int { return nil }
func (m mapTestConfig) ColumnOrder() []string        { return nil }
func (m mapTestConfig) Version() string              { return "test" }
func (m mapTestConfig) Switches() []config.SwitchConfig {
	return m.switches
}
func (m mapTestConfig) InfrastructureDevices() []config.InfraDevice { return m.infra }
func (m mapTestConfig) UnmanagedSwitches() []config.UnmanagedSwitch { return m.unmanaged }
func (m mapTestConfig) IgnoredMACs() []string                       { return m.ignored }

func newMapTestServer(t *testing.T, cfg mapTestConfig, sw ...*SwitchData) *Server {
	t.Helper()
	cache := NewCache()
	for _, s := range sw {
		cache.UpdateSwitch(s.IP, s)
	}
	srv := NewServer(cache, cfg, testLogger, newLogBuf(), nil, nil)
	srv.OUI = oui.New()
	return srv
}

func TestAPIDeviceTypesDefault(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{})
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/device_types", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var types map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &types); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(types) == 0 {
		t.Fatal("expected default device types")
	}
	if _, ok := types["laptop"]; !ok {
		t.Fatal("expected 'laptop' device type")
	}
}

func TestAPIDeviceTypesRawSave(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{})
	s.DataDir = t.TempDir()
	content := "gadget:\n  label: Gadget\n  icon: mdi:devices\n"
	body := `{"content":"` + strings.ReplaceAll(content, "\n", "\\n") + `"}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/device_types/raw", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("save: expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	// Verify persisted
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/device_types", nil)
	s.Router.ServeHTTP(w2, r2)
	var types map[string]map[string]any
	json.Unmarshal(w2.Body.Bytes(), &types)
	if _, ok := types["gadget"]; !ok {
		t.Fatal("expected saved 'gadget' device type")
	}
}

func TestAPIUpdateType(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FF",
		Status: "online",
		MACTable: []MACEntry{{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "1"}},
	})
	body := `{"mac":"AA:BB:CC:DD:EE:01","type":"nas"}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/clients/update_type", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Topology should include device_type
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w2, r2)
	var topo Topology
	json.Unmarshal(w2.Body.Bytes(), &topo)
	found := false
	for _, n := range topo.Nodes {
		if n.MAC == "AA:BB:CC:DD:EE:01" {
			found = true
			if n.DeviceType != "nas" {
				t.Fatalf("expected device_type nas, got %q", n.DeviceType)
			}
		}
	}
	if !found {
		t.Fatal("client node not found")
	}
}

func TestAPIClientDelete(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{})
	s.clientsMu.Lock()
	s.Clients["AABBCCDDEE01"] = config.ClientEntry{MAC: "AA:BB:CC:DD:EE:01", Host: "Keep", Status: "offline"}
	s.Clients["AABBCCDDEE02"] = config.ClientEntry{MAC: "AA:BB:CC:DD:EE:02", Host: "Remove", Status: "offline"}
	s.clientsMu.Unlock()

	body := `{"mac":"AA:BB:CC:DD:EE:02"}`
	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/clients/delete", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if _, ok := s.Clients["AABBCCDDEE02"]; ok {
		t.Fatal("client should have been deleted")
	}
	if _, ok := s.Clients["AABBCCDDEE01"]; !ok {
		t.Fatal("client should remain")
	}
}

func TestAPIClientImportCSV(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{})
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "clients.csv")
	fw.Write([]byte("hostname,mac\nMyPC,AA:BB:CC:DD:EE:01\nNAS,11:22:33:44:55:66\nbad,zzz\n"))
	mw.Close()

	w, r := httptest.NewRecorder(), httptest.NewRequest("POST", "/api/clients/import_csv", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["imported"] != float64(2) {
		t.Fatalf("expected 2 imported, got %v", resp["imported"])
	}
	if s.Clients["AABBCCDDEE01"].Host != "MyPC" {
		t.Fatalf("expected MyPC, got %q", s.Clients["AABBCCDDEE01"].Host)
	}
}

func TestTopologySwitchLinks(t *testing.T) {
	cfg := mapTestConfig{switches: []config.SwitchConfig{
		{Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", Enabled: true},
		{Name: "Edge", IP: "192.168.1.2", Model: "RTLPlayground", Enabled: true,
			ParentIP: "192.168.1.1", ParentPort: "8"},
	}}
	core := &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FF",
		Status: "online",
		Ports: []PortState{
			{Port: "8", Status: "up", Speed: "2.5G", SpeedTX: 1000, SpeedRX: 2000},
		},
		MACTable: []MACEntry{
			{MAC: "AA:BB:CC:DD:EE:10", Type: "l", Port: "8", VLAN: "1"},
			{MAC: "AA:BB:CC:DD:EE:11", Type: "l", Port: "1", VLAN: "1"},
		},
	}
	edge := &SwitchData{
		Name: "Edge", IP: "192.168.1.2", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FE",
		Status: "online",
		MACTable: []MACEntry{
			{MAC: "AA:BB:CC:DD:EE:10", Type: "l", Port: "5", VLAN: "1"},
			{MAC: "AA:BB:CC:DD:EE:20", Type: "l", Port: "1", VLAN: "1"},
		},
	}
	s := newMapTestServer(t, cfg, core, edge)
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w, r)
	var topo Topology
	json.Unmarshal(w.Body.Bytes(), &topo)

	foundLink := false
	for _, l := range topo.Links {
		if l.Type == "uplink" && l.Source == "192.168.1.1" && l.Target == "192.168.1.2" {
			foundLink = true
			if l.Speed != "2.5G" {
				t.Fatalf("expected speed 2.5G on uplink, got %q", l.Speed)
			}
			if l.TXBPS != 1000 {
				t.Fatalf("expected tx 1000, got %d", l.TXBPS)
			}
		}
	}
	if !foundLink {
		t.Fatal("expected static uplink link between core and edge")
	}

	// The MAC AA:BB:CC:DD:EE:10 is learned on core's uplink port 8 AND on edge
	// port 5. It must attach to the edge switch (its real parent), not to core.
	var clientParent string
	for _, l := range topo.Links {
		if l.Target == "AABBCCDDEE10" {
			clientParent = l.Source
		}
	}
	if clientParent != "192.168.1.2" {
		t.Fatalf("expected client parent edge switch, got %q", clientParent)
	}
}

func TestTopologyInfraAndUnmanaged(t *testing.T) {
	cfg := mapTestConfig{
		infra: []config.InfraDevice{
			{Name: "Router", MAC: "AA:AA:AA:AA:AA:01", Type: "router"},
		},
		unmanaged: []config.UnmanagedSwitch{
			{Name: "Small Switch", ParentIP: "192.168.1.1", ParentPort: "3"},
		},
	}
	sw := &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FF",
		Status: "online",
		MACTable: []MACEntry{
			{MAC: "AA:AA:AA:AA:AA:01", Type: "l", Port: "2", VLAN: "1"},
		},
	}
	s := newMapTestServer(t, cfg, sw)
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w, r)
	var topo Topology
	json.Unmarshal(w.Body.Bytes(), &topo)

	var router *TopologyNode
	var unmanaged *TopologyNode
	for i := range topo.Nodes {
		if topo.Nodes[i].Type == "router" {
			router = &topo.Nodes[i]
		}
		if topo.Nodes[i].Type == "unmanaged_switch" {
			unmanaged = &topo.Nodes[i]
		}
	}
	if router == nil {
		t.Fatal("expected router infra node")
	}
	if router.Status != "online" {
		t.Fatalf("expected router online, got %q", router.Status)
	}
	if unmanaged == nil {
		t.Fatal("expected unmanaged switch node")
	}

	foundUnmanagedLink := false
	for _, l := range topo.Links {
		if l.Source == "192.168.1.1" && l.Type == "infra" && strings.HasPrefix(l.Target, "unmanaged_") {
			foundUnmanagedLink = true
		}
	}
	if !foundUnmanagedLink {
		t.Fatal("expected link to unmanaged switch")
	}
}

func TestTopologyOfflineClientPersistence(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FF",
		Status: "online",
		MACTable: []MACEntry{
			{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "1", VLAN: "1"},
		},
	})
	// First request seeds the client as online
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w, r)

	// MAC no longer present -> should be marked offline on next request
	s.Cache.GetSwitch("192.168.1.1").MACTable = []MACEntry{}
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w2, r2)
	var topo Topology
	json.Unmarshal(w2.Body.Bytes(), &topo)

	var client *TopologyNode
	for i := range topo.Nodes {
		if topo.Nodes[i].MAC == "AA:BB:CC:DD:EE:01" {
			client = &topo.Nodes[i]
		}
	}
	if client == nil {
		t.Fatal("expected offline client node to persist")
	}
	if client.Status != "offline" {
		t.Fatalf("expected offline status, got %q", client.Status)
	}
}

func TestIgnoredMACFiltering(t *testing.T) {
	cfg := mapTestConfig{ignored: []string{"AA:BB:CC:00:00:*"}}
	sw := &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", MAC: "AA:BB:CC:DD:EE:FF",
		Status: "online",
		MACTable: []MACEntry{
			{MAC: "AA:BB:CC:00:00:01", Type: "l", Port: "1", VLAN: "1"},
			{MAC: "AA:BB:CC:DD:EE:01", Type: "l", Port: "2", VLAN: "1"},
		},
	}
	s := newMapTestServer(t, cfg, sw)

	// /api/switches should filter ignored MACs
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches", nil)
	s.Router.ServeHTTP(w, r)
	var data []map[string]any
	json.Unmarshal(w.Body.Bytes(), &data)
	macTable := data[0]["mac_table"].([]any)
	if len(macTable) != 1 {
		t.Fatalf("expected 1 MAC after ignore filter, got %d", len(macTable))
	}

	// topology should also filter
	w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/topology", nil)
	s.Router.ServeHTTP(w2, r2)
	var topo Topology
	json.Unmarshal(w2.Body.Bytes(), &topo)
	for _, n := range topo.Nodes {
		if n.MAC == "AA:BB:CC:00:00:01" {
			t.Fatal("ignored MAC should not appear in topology")
		}
	}
}

func TestAPISwitchImage(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", Status: "online",
	})
	s.DataDir = t.TempDir()
	s.DeviceTemplatesDir = filepath.Join(s.DataDir, "device-templates")
	os.MkdirAll(s.DeviceTemplatesDir, 0755)
	os.WriteFile(filepath.Join(s.DeviceTemplatesDir, "RTLPlayground.png"), []byte("PNGDATA"), 0644)

	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches/192.168.1.1/image", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "PNGDATA" {
		t.Fatalf("expected PNG data, got %q", w.Body.String())
	}
}

func TestAPISwitchImageFallback404(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground", Status: "online",
	})
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/api/switches/192.168.1.1/image", nil)
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Fatalf("expected 200 or 404, got %d", w.Code)
	}
}

var _ = logbuf.New
