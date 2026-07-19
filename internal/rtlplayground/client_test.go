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
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/information.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Information{
			IP:       "192.168.10.247",
			GW:       "192.168.10.1",
			Mask:     "255.255.255.0",
			MAC:      "AA:BB:CC:DD:EE:FF",
			SwVer:    "v0.1.0",
			Build:    "2024-01-01",
			HWVer:    "HC-SWTGW218AS",
			Flash:    2097152,
			Hostname: "test-switch",
			Telnet:   0,
			Web:      1,
			SFP0:     "Lightron Inc.",
			SFP1:     "",
		})
	})

	mux.HandleFunc("/status.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]StatusEntry{
			{PortNum: 1, Name: "port1", Link: 4, Enabled: 1, TxG: "0x100", RxG: "0x200", TxB: "0x0", RxB: "0x0"},
			{PortNum: 2, Name: "port2", Link: 0, Enabled: 1, TxG: "0x0", RxG: "0x0", TxB: "0x0", RxB: "0x0"},
			{PortNum: 9, Name: "sfp1", Link: 4, Enabled: 1, SFP: "Lightron Inc. WSPXG-ES3LC-IHA", TxG: "0x10", RxG: "0x20", TxB: "0x0", RxB: "0x0"},
		})
	})

	mux.HandleFunc("/sfp_diag.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SFPDiagEntry{
			{Port: 9, Options: 3, Temperature: "0x1a-0x00", VCC: "0x20-0x00", TXBias: "0x30-0x00", TXPower: "0x40-0x00", RXPower: "0x50-0x00", Laser: 1},
		})
	})

	mux.HandleFunc("/l2.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]L2Entry{
			{MAC: "AA:BB:CC:DD:EE:01", VLAN: "1", Type: "l", Port: "5", Index: 0},
			{MAC: "AA:BB:CC:DD:EE:02", VLAN: "1", Type: "l", Port: "7", Index: 1},
			{MAC: "AA:BB:CC:DD:EE:03", VLAN: "10", Type: "s", Port: "1", Index: 2},
		})
	})

	mux.HandleFunc("/eee.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]EEEEntry{
			{Port: 1, IsSFP: false, Adv2500: true, Adv1000: true, Active: true},
			{Port: 2, IsSFP: false, Adv1000: true, Adv100: true, Active: false},
		})
	})

	mux.HandleFunc("/mtu.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]MTUEntry{
			{Port: 1, MTU: "0x2400"},
			{Port: 2, MTU: "0x05DC"},
		})
	})

	mux.HandleFunc("/vlanlist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]VLANListItem{
			{ID: 1, Name: "default"},
			{ID: 10, Name: "iot"},
		})
	})

	mux.HandleFunc("/bandwidth.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]BandwidthEntry{
			{Port: 1, IngLimit: 0, EgrLimit: 0},
		})
	})

	return httptest.NewServer(mux)
}

func newAuthenticatedClient(ts *httptest.Server) *Client {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{Jar: jar}
	return NewWithClient(ts.Listener.Addr().String(), httpClient)
}

func TestScrapeInformation(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	// Manually login
	client.login("test")

	info, err := client.ScrapeInformation()
	if err != nil {
		t.Fatalf("ScrapeInformation failed: %v", err)
	}
	if info.IP != "192.168.10.247" {
		t.Fatalf("expected IP 192.168.10.247, got %s", info.IP)
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("expected MAC AA:BB:CC:DD:EE:FF, got %s", info.MAC)
	}
	if info.Hostname != "test-switch" {
		t.Fatalf("expected hostname test-switch, got %s", info.Hostname)
	}
}

func TestScrapeStatus(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	status, err := client.ScrapeStatus()
	if err != nil {
		t.Fatalf("ScrapeStatus failed: %v", err)
	}
	if len(status) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(status))
	}
	if status[0].PortNum != 1 {
		t.Fatalf("expected port 1, got %d", status[0].PortNum)
	}
	if status[0].Link != 4 {
		t.Fatalf("expected link=4 (10G), got %d", status[0].Link)
	}
	if status[2].SFP != "Lightron Inc. WSPXG-ES3LC-IHA" {
		t.Fatalf("unexpected SFP string: %s", status[2].SFP)
	}
}

func TestScrapeSFPDiag(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	diag, err := client.ScrapeSFPDiag()
	if err != nil {
		t.Fatalf("ScrapeSFPDiag failed: %v", err)
	}
	if len(diag) != 1 {
		t.Fatalf("expected 1 SFP diagnostic entry, got %d", len(diag))
	}
	if diag[0].Port != 9 {
		t.Fatalf("expected port 9, got %d", diag[0].Port)
	}
}

func TestScrapeMACTable(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	entries, err := client.ScrapeMACTable(0)
	if err != nil {
		t.Fatalf("ScrapeMACTable failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 MAC entries, got %d", len(entries))
	}
	if entries[0].MAC != "AA:BB:CC:DD:EE:01" {
		t.Fatalf("expected MAC AA:BB:CC:DD:EE:01, got %s", entries[0].MAC)
	}
	if entries[0].Port != "5" {
		t.Fatalf("expected port 5, got %s", entries[0].Port)
	}
}

func TestScrapeEEE(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	eee, err := client.ScrapeEEE()
	if err != nil {
		t.Fatalf("ScrapeEEE failed: %v", err)
	}
	if len(eee) != 2 {
		t.Fatalf("expected 2 EEE entries, got %d", len(eee))
	}
	if !eee[0].Active {
		t.Fatal("expected port 1 EEE active")
	}
}

func TestScrapeMTU(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	mtu, err := client.ScrapeMTU()
	if err != nil {
		t.Fatalf("ScrapeMTU failed: %v", err)
	}
	if len(mtu) != 2 {
		t.Fatalf("expected 2 MTU entries, got %d", len(mtu))
	}
}

func TestScrapeVLANList(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	list, err := client.ScrapeVLANList()
	if err != nil {
		t.Fatalf("ScrapeVLANList failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 VLANs, got %d", len(list))
	}
	if list[1].Name != "iot" {
		t.Fatalf("expected VLAN 10 name 'iot', got %s", list[1].Name)
	}
}

func TestScrapeAllMACTable(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	client := newAuthenticatedClient(ts)
	client.login("test")

	entries, err := client.ScrapeAllMACTable()
	if err != nil {
		t.Fatalf("ScrapeAllMACTable failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 total MAC entries, got %d", len(entries))
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
		t.Fatal("expected login failure, got nil")
	}
}
