package rtlplayground

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		pwd := r.FormValue("pwd")
		if pwd == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test"})
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/information.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Information{
			IPAddress:     "192.168.10.247",
			IPGateway:     "192.168.10.1",
			IPNetmask:     "255.255.255.0",
			TelnetEnabled: "0",
			WebEnabled:    "1",
			MACAddress:    "AA:BB:CC:DD:EE:FF",
			SwVer:         "v0.1.0",
			BuildDate:     "2024-01-01",
			HWVer:         "HC-SWTGW218AS",
			FlashSize:     "2MB",
			Hostname:      "test-switch",
		})
	})

	mux.HandleFunc("/status.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]StatusEntry{
			{PortNum: 1, LogPort: 4, Name: "port1", IsSFP: 0, Enabled: 1, Link: 4, Adv: "100000", TxG: "0x100", RxG: "0x200", TxB: "0x0", RxB: "0x0"},
			{PortNum: 2, LogPort: 5, IsSFP: 0, Enabled: 1, Link: 0, Adv: "000011", TxG: "0x0", RxG: "0x0", TxB: "0x0", RxB: "0x0"},
			{PortNum: 9, LogPort: 8, IsSFP: 1, Enabled: 1, SFPVendor: "Lightron Inc.", SFPModel: "WSPXG-ES3LC-IHA", Link: 4, TxG: "0x10", RxG: "0x20", TxB: "0x0", RxB: "0x0"},
		})
	})

	mux.HandleFunc("/sfp_diag.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SFPDiagEntry{
			{PortNum: 9, SFPOptions: "0x43", SFPTemp: "0x1a00", SFPVCC: "0x2000", SFPTXBias: "0x3000", SFPTXPower: "0x4000", SFPRXPower: "0x5000", SFPState: "0x01"},
		})
	})

	mux.HandleFunc("/l2.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]L2Entry{
			{MAC: "AA:BB:CC:DD:EE:01", VLAN: "001", Type: "l", Port: "5", Index: "0000"},
			{MAC: "AA:BB:CC:DD:EE:02", VLAN: "001", Type: "l", Port: "7", Index: "0001"},
			{MAC: "AA:BB:CC:DD:EE:03", VLAN: "001", Type: "s", Port: "1", Index: "0002"},
		})
	})

	mux.HandleFunc("/eee.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]EEEEntry{
			{PortNum: 1, IsSFP: 0, EEE: "110", EEELP: "010", Active: 1},
			{PortNum: 2, IsSFP: 0, EEE: "000", EEELP: "000", Active: 0},
		})
	})

	mux.HandleFunc("/mtu.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]MTUEntry{
			{PortNum: 1, MTU: "0x2400"},
			{PortNum: 2, MTU: "0x05DC"},
		})
	})

	mux.HandleFunc("/vlanlist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]VLANListItem{
			{ID: 1, Name: "default"},
			{ID: 10, Name: "iot"},
		})
	})

	return httptest.NewServer(mux)
}

func newClient(ts *httptest.Server) *Client {
	jar, _ := cookiejar.New(nil)
	return NewWithClient(ts.Listener.Addr().String(), &http.Client{Jar: jar})
}

func TestScrapeInformation(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	info, err := c.ScrapeInformation()
	if err != nil {
		t.Fatalf("ScrapeInformation: %v", err)
	}
	if info.IPAddress != "192.168.10.247" {
		t.Fatalf("IP: got %q", info.IPAddress)
	}
	if info.MACAddress != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("MAC: got %q", info.MACAddress)
	}
	if info.Hostname != "test-switch" {
		t.Fatalf("Hostname: got %q", info.Hostname)
	}
	if info.SwVer != "v0.1.0" {
		t.Fatalf("SwVer: got %q", info.SwVer)
	}
	if info.TelnetEnabled != "0" || info.WebEnabled != "1" {
		t.Fatalf("Telnet/Web: got %q/%q", info.TelnetEnabled, info.WebEnabled)
	}
}

func TestScrapeStatus(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	status, err := c.ScrapeStatus()
	if err != nil {
		t.Fatalf("ScrapeStatus: %v", err)
	}
	if len(status) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(status))
	}
	if status[0].PortNum != 1 {
		t.Fatalf("portNum: got %d", status[0].PortNum)
	}
	if status[0].LogPort != 4 {
		t.Fatalf("logPort: got %d", status[0].LogPort)
	}
	if status[2].IsSFP != 1 {
		t.Fatalf("port 9 should be SFP")
	}
	if status[2].SFPVendor != "Lightron Inc." {
		t.Fatalf("SFP vendor: got %q", status[2].SFPVendor)
	}
}

func TestScrapeSFPDiag(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	diag, err := c.ScrapeSFPDiag()
	if err != nil {
		t.Fatalf("ScrapeSFPDiag: %v", err)
	}
	if len(diag) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(diag))
	}
	if diag[0].PortNum != 9 {
		t.Fatalf("portNum: got %d", diag[0].PortNum)
	}
	if diag[0].SFPOptions != "0x43" {
		t.Fatalf("sfp_options: got %q", diag[0].SFPOptions)
	}
}

func TestScrapeMACTable(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	entries, err := c.ScrapeMACTable(0)
	if err != nil {
		t.Fatalf("ScrapeMACTable: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3, got %d", len(entries))
	}
	if entries[0].MAC != "AA:BB:CC:DD:EE:01" {
		t.Fatalf("MAC: got %q", entries[0].MAC)
	}
	if entries[0].VLAN != "001" {
		t.Fatalf("VLAN: got %q", entries[0].VLAN)
	}
	if entries[0].Type != "l" {
		t.Fatalf("type: got %q", entries[0].Type)
	}
}

func TestScrapeAllMACTable(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	entries, err := c.ScrapeAllMACTable()
	if err != nil {
		t.Fatalf("ScrapeAllMACTable: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3, got %d", len(entries))
	}
}

func TestScrapeEEE(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	eee, err := c.ScrapeEEE()
	if err != nil {
		t.Fatalf("ScrapeEEE: %v", err)
	}
	if len(eee) != 2 {
		t.Fatalf("expected 2, got %d", len(eee))
	}
	if eee[0].Active != 1 {
		t.Fatal("port 1 EEE should be active")
	}
}

func TestScrapeMTU(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	mtu, err := c.ScrapeMTU()
	if err != nil {
		t.Fatalf("ScrapeMTU: %v", err)
	}
	if len(mtu) != 2 {
		t.Fatalf("expected 2, got %d", len(mtu))
	}
}

func TestScrapeVLANList(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newClient(ts)
	c.login("test")

	list, err := c.ScrapeVLANList()
	if err != nil {
		t.Fatalf("ScrapeVLANList: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[1].Name != "iot" {
		t.Fatalf("expected 'iot', got %q", list[1].Name)
	}
}

func TestLoginFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	_, err := New(ts.Listener.Addr().String(), "wrong")
	if err == nil {
		t.Fatal("expected login failure")
	}
}

func TestParseHexIndex(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"0000", 0},
		{"0001", 1},
		{"0012", 18},
		{"00ff", 255},
		{"0100", 256},
		{"0abc", 2748},
	}
	for _, tc := range tests {
		got := parseHexIndex(tc.s)
		if got != tc.want {
			t.Errorf("parseHexIndex(%q) = %d, want %d", tc.s, got, tc.want)
		}
	}
}
