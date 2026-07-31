package server

import (
	"testing"

	"github.com/eraiza0816/switch-dashboard/internal/config"
)

const (
	topoCoreMAC = "AA:BB:CC:DD:EE:FF"
	topoEdgeMAC = "AA:BB:CC:DD:EE:99"
	topoCliMAC  = "AA:BB:CC:DD:EE:01"
	topoInfMAC  = "AA:BB:CC:DD:EE:02"
)

func topoCoreSwitch() *SwitchData {
	return &SwitchData{
		Name: "Core", IP: "192.168.1.1", Model: "RTLPlayground",
		MAC: topoCoreMAC, Status: "online",
		MACTable: []MACEntry{
			{MAC: topoEdgeMAC, Type: "s", Port: "1", VLAN: "1"},         // another switch
			{MAC: topoCliMAC, Type: "l", Port: "2", VLAN: "1"},          // client
			{MAC: topoInfMAC, Type: "l", Port: "3", VLAN: "1"},          // infra device
			{MAC: "AA:BB:CC:DD:EE:03", Type: "l", Port: "9", VLAN: "1"}, // CPU port, must be skipped
		},
	}
}

func topoEdgeSwitch() *SwitchData {
	return &SwitchData{
		Name: "Edge", IP: "192.168.1.2", Model: "RTLPlayground",
		MAC: topoEdgeMAC, Status: "online",
	}
}

func TestTopologyStateMaps(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{
		switches: []config.SwitchConfig{
			{Name: "Edge", IP: "192.168.1.2", Enabled: true, ParentIP: "192.168.1.1", ParentPort: "1", UplinkPort: "24"},
		},
	}, topoCoreSwitch(), topoEdgeSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())

	if _, ok := st.switchesByIP["192.168.1.1"]; !ok {
		t.Fatal("missing core switch in switchesByIP")
	}
	if ip, ok := st.switchMACToIP["AABBCCDDEEFF"]; !ok || ip != "192.168.1.1" {
		t.Fatalf("switchMACToIP = %q, %v; want 192.168.1.1", ip, ok)
	}
	if !st.switchPorts["192.168.1.1:1"] {
		t.Fatal("port 1 should be marked as carrying a switch")
	}
	if st.switchPorts["192.168.1.1:2"] {
		t.Fatal("port 2 should not be marked as carrying a switch")
	}
	if !st.uplinkPorts["192.168.1.1:1"] {
		t.Fatal("port 1 should be a configured uplink")
	}
	if _, ok := st.swPortMACs["192.168.1.1"]["9"]; ok {
		t.Fatal("CPU port 9 must be excluded from swPortMACs")
	}
	if _, ok := st.swPortMACs["192.168.1.1"]["2"]; !ok {
		t.Fatal("port 2 should be present in swPortMACs")
	}
}

func TestTopologyStateActiveClients(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{
		switches: []config.SwitchConfig{
			{Name: "Edge", IP: "192.168.1.2", Enabled: true, ParentIP: "192.168.1.1", ParentPort: "1", UplinkPort: "24"},
		},
		infra: []config.InfraDevice{
			{Name: "NAS", MAC: topoInfMAC, Type: "nas"},
		},
	}, topoCoreSwitch(), topoEdgeSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())
	active := st.activeClients()

	if len(active) != 1 {
		t.Fatalf("activeClients = %d entries, want 1", len(active))
	}
	e, ok := active["AABBCCDDEE01"]
	if !ok {
		t.Fatalf("expected client %s in activeClients, got %v", topoCliMAC, active)
	}
	if e.Port != "2" || e.IP != "192.168.1.1" {
		t.Fatalf("client entry = %+v, want ip 192.168.1.1 port 2", e)
	}
}

func TestTopologyStateResolveSource(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{
		unmanaged: []config.UnmanagedSwitch{
			{Name: "Under Desk", ParentIP: "192.168.1.1", ParentPort: "3"},
		},
	}, topoCoreSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())

	source, sourcePort := st.resolveSource("192.168.1.1", "3")
	if source != "unmanaged_192.168.1.1:3" || sourcePort != "" {
		t.Fatalf("resolveSource(unmanaged) = %q, %q; want unmanaged id", source, sourcePort)
	}

	source, sourcePort = st.resolveSource("192.168.1.1", "2")
	if source != "192.168.1.1" || sourcePort != "Port 2" {
		t.Fatalf("resolveSource(normal) = %q, %q; want 192.168.1.1 / Port 2", source, sourcePort)
	}
}

func TestTopologyStateSkipClient(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{
		infra:   []config.InfraDevice{{Name: "NAS", MAC: topoInfMAC, Type: "nas"}},
		ignored: []string{"AA:BB:CC:DD:EE:0A"},
	}, topoCoreSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())

	cases := []struct {
		raw  string
		want bool
	}{
		{raw: "", want: true},                   // empty
		{raw: topoCoreMAC, want: true},          // switch MAC
		{raw: topoInfMAC, want: true},           // infra device
		{raw: "AA:BB:CC:DD:EE:0A", want: true},  // ignored
		{raw: "AA:BB:CC:DD:EE:0B", want: false}, // normal client
	}
	for _, tc := range cases {
		if got := st.skipClient(tc.raw, NormalizeMAC(tc.raw)); got != tc.want {
			t.Errorf("skipClient(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

func TestTopologyStateFindInfraConn(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, topoCoreSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())

	ip, port, ok := st.findInfraConn("AABBCCDDEE02")
	if !ok || ip != "192.168.1.1" || port != "3" {
		t.Fatalf("findInfraConn = %q, %q, %v; want 192.168.1.1 / 3 / true", ip, port, ok)
	}
	if _, _, ok := st.findInfraConn("AABBCCDDEEFF"); ok {
		t.Fatal("findInfraConn should not match a switch MAC")
	}
}

func TestTopologyStateBuildClientNode(t *testing.T) {
	s := newMapTestServer(t, mapTestConfig{}, topoCoreSwitch())

	st := s.newTopologyState(s.Cache.GetSwitches())
	active := st.activeClients()
	clientSet := map[string]bool{}
	for mac := range active {
		clientSet[mac] = true
	}

	c := config.ClientEntry{MAC: topoCliMAC, Status: "offline"}
	node, link, ok := st.buildClientNode(c, clientSet, active)
	if !ok {
		t.Fatal("expected client node to be built")
	}
	if node.Status != "online" {
		t.Fatalf("node status = %q, want online (active client)", node.Status)
	}
	if node.LastSeenPort != "2" {
		t.Fatalf("node last_seen_port = %q, want 2", node.LastSeenPort)
	}
	if link.Source != "192.168.1.1" || link.SourcePort != "Port 2" {
		t.Fatalf("link = %+v, want source 192.168.1.1 Port 2", link)
	}
}
