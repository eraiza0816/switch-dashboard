package poller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eraiza0816/switch-dashboard/internal/history"
	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
	"github.com/eraiza0816/switch-dashboard/internal/server"
)

// fakeSwitchServer mimics the RTLPlayground JSON API (page_impl.c).  It
// serves all endpoints the poller scrapes, with the field formats the
// firmware actually emits (link codes, hex counters, "0x" MTU values).
// When breakCounters is set, /counters.json returns 404.
func fakeSwitchServer(breakCounters bool) (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess"})
		w.Header().Set("Location", "index.html")
		w.WriteHeader(http.StatusFound)
	})

	mux.HandleFunc("/status.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]rtlplayground.StatusEntry{
			{PortNum: 1, LogPort: 4, IsSFP: 0, Enabled: 1, Link: 5, TxG: "0x100", RxG: "0x200", TxB: "0x0", RxB: "0x0"}, // 10G
			{PortNum: 2, LogPort: 5, IsSFP: 0, Enabled: 1, Link: 3, TxG: "0x10", RxG: "0x20", TxB: "0x0", RxB: "0x0"},   // 1G
			{PortNum: 6, LogPort: 8, IsSFP: 1, Enabled: 1, Link: 5, SFPVendor: "OEM", SFPModel: "10G-SFP+", SFPSerial: "12345678", TxG: "0x1", RxG: "0x2", TxB: "0x0", RxB: "0x0"},
		})
	})

	mux.HandleFunc("/information.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rtlplayground.Information{
			IPAddress:  "10.0.0.1",
			MACAddress: "1c:2a:a3:23:00:02",
			SwVer:      "v0.2.22",
			Hostname:   "rtlplayground",
		})
	})

	mux.HandleFunc("/sfp_diag.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]rtlplayground.SFPDiagEntry{
			{PortNum: 6, SFPOptions: "0x68", SFPTemp: "0x2a67", SFPVCC: "0x7f01", SFPTXBias: "0x0d72", SFPTXPower: "0x150c", SFPRXPower: "0x0000", SFPState: "0x00"},
		})
	})

	mux.HandleFunc("/l2.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	})

	mux.HandleFunc("/eee.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	})

	mux.HandleFunc("/vlanlist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	})

	mux.HandleFunc("/lag.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	})

	mux.HandleFunc("/mirror.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"enabled":0,"mPort":1,"mirror_rx":"0000000000000000","mirror_tx":"0000000000000000"}`))
	})

	mux.HandleFunc("/bandwidth.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	})

	// One jumbo (9000) and one standard (1518) port.
	mux.HandleFunc("/mtu.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"portNum":1,"mtu":"0x2328"},{"portNum":2,"mtu":"0x05ee"}]`))
	})

	// MIB counters: index 0 = In Octets (RX), index 1 = Out Octets (TX).
	mux.HandleFunc("/counters.json", func(w http.ResponseWriter, r *http.Request) {
		if breakCounters {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`["0x0000000000001000","0x0000000000002000"]`))
	})

	return httptest.NewServer(mux), mux
}

func TestPollScrapesRealData(t *testing.T) {
	ts, _ := fakeSwitchServer(false)
	defer ts.Close()

	cache := server.NewCache()
	client := rtlplayground.NewWithClient(ts.Listener.Addr().String(), ts.Client())
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)
	p.poll()

	sw := cache.GetSwitch("10.0.0.1")
	if sw == nil {
		t.Fatal("switch not in cache")
	}
	if sw.Hostname != "rtlplayground" || sw.Firmware != "v0.2.22" {
		t.Fatalf("info fields: hostname=%q fw=%q", sw.Hostname, sw.Firmware)
	}

	// Link codes follow the firmware mapping (1=10M, 3=1G, 5=10G).
	byPort := map[string]server.PortState{}
	for _, p := range sw.Ports {
		byPort[p.Port] = p
	}
	if byPort["1"].Speed != "10G" {
		t.Errorf("port 1 speed = %q, want 10G (link code 5)", byPort["1"].Speed)
	}
	if byPort["2"].Speed != "1G" {
		t.Errorf("port 2 speed = %q, want 1G (link code 3)", byPort["2"].Speed)
	}

	// Real byte counters from /counters.json (0x2000 = 8192 TX, 0x1000 = 4096 RX).
	if byPort["1"].TXBytes != 8192 || byPort["1"].RXBytes != 4096 {
		t.Errorf("port 1 bytes: tx=%d rx=%d, want 8192/4096", byPort["1"].TXBytes, byPort["1"].RXBytes)
	}
	if byPort["1"].TXPackets != 0x100 || byPort["1"].RXPackets != 0x200 {
		t.Errorf("port 1 packets: tx=%d rx=%d, want 256/512", byPort["1"].TXPackets, byPort["1"].RXPackets)
	}

	// Jumbo only when the MTU exceeds 1518; size is the decimal value.
	if !sw.Jumbo.Enabled || sw.Jumbo.Size != "9000" {
		t.Errorf("jumbo = %+v, want enabled with size 9000", sw.Jumbo)
	}

	// SFP port carries vendor/serial/LOS, diag is stored formatted.
	if byPort["6"].SFPVendor != "OEM" || byPort["6"].SFPSerial != "12345678" {
		t.Errorf("sfp fields: %+v", byPort["6"])
	}
	if len(sw.SFPDiag) != 1 {
		t.Fatalf("sfp diag entries = %d, want 1", len(sw.SFPDiag))
	}
	d := sw.SFPDiag[0]
	if d.Port != 6 || d.Temp != "42.4 C" || d.VCC != "3.25 V" || d.TXPower != "-2.69 dBm" || !d.HasDDMI {
		t.Errorf("sfp diag = %+v", d)
	}
}

// When /counters.json fails (or is unavailable), the poller falls back to
// the packet*800 estimate so bandwidth keeps working.
func TestPollCountersFallback(t *testing.T) {
	ts, _ := fakeSwitchServer(true)
	defer ts.Close()

	cache := server.NewCache()
	client := rtlplayground.NewWithClient(ts.Listener.Addr().String(), ts.Client())
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)
	p.poll()

	sw := cache.GetSwitch("10.0.0.1")
	if sw == nil {
		t.Fatal("switch not in cache")
	}
	for _, port := range sw.Ports {
		if port.Port == "1" {
			if port.TXBytes != 0x100*800 || port.RXBytes != 0x200*800 {
				t.Errorf("fallback bytes: tx=%d rx=%d", port.TXBytes, port.RXBytes)
			}
			if !port.Estimated {
				t.Error("fallback port must be flagged Estimated")
			}
		}
	}
}

// Estimated samples (packets*800 fallback) must never reach the history
// store, otherwise the bandwidth charts show fabricated values.
func TestPollEstimatedSamplesSkipHistory(t *testing.T) {
	ts, _ := fakeSwitchServer(true)
	defer ts.Close()

	histPath := t.TempDir() + "/hist.duckdb"
	store, err := history.NewStore(histPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	cache := server.NewCache()
	client := rtlplayground.NewWithClient(ts.Listener.Addr().String(), ts.Client())
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)
	p.SetHistoryStore(store)
	p.poll()

	points, err := store.QueryHistory("10.0.0.1", "1", "live")
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 0 {
		t.Fatalf("estimated samples must not be written to history, got %d", len(points))
	}
}

// In production mode (demo=false) the poller must never seed mock data.
func TestNoMockSeedWithoutDemo(t *testing.T) {
	cache := server.NewCache()
	p := New(cache, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)
	p.Start()
	defer p.Stop()

	if sw := cache.GetSwitch("10.0.0.1"); sw != nil {
		t.Fatalf("production poller seeded mock data: %+v", sw)
	}
}

// In demo mode mock data is seeded and flagged so the UI can label it.
func TestDemoSeedsMockFlagged(t *testing.T) {
	cache := server.NewCache()
	p := New(cache, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)
	p.SetDemo(true)
	p.Start()
	defer p.Stop()

	sw := cache.GetSwitch("10.0.0.1")
	if sw == nil {
		t.Fatal("demo poller did not seed mock data")
	}
	if !sw.Mock {
		t.Error("seeded data must be flagged Mock")
	}
	if sw.Status != "online" {
		t.Errorf("status = %q, want online", sw.Status)
	}
}

// After consecutive scrape failures the switch must be marked offline
// instead of staying "online" with stale or seeded data.
func TestMarkOfflineAfterConsecutiveFailures(t *testing.T) {
	cache := server.NewCache()
	cache.UpdateSwitch("10.0.0.1", &server.SwitchData{IP: "10.0.0.1", Status: "online"})

	client := rtlplayground.NewWithClient("127.0.0.1:1", &http.Client{Timeout: 200 * time.Millisecond})
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)

	for i := 0; i < maxScrapeFailures; i++ {
		p.poll()
	}

	sw := cache.GetSwitch("10.0.0.1")
	if sw == nil {
		t.Fatal("switch not in cache")
	}
	if sw.Status != "offline" {
		t.Fatalf("status = %q, want offline", sw.Status)
	}
	if sw.Error == "" {
		t.Error("offline switch must carry an error message")
	}
}

// An offline demo switch keeps its Mock flag: the data shown is still demo
// data and must remain labelled as such.
func TestOfflineDemoSwitchKeepsMockFlag(t *testing.T) {
	cache := server.NewCache()
	cache.UpdateSwitch("10.0.0.1", &server.SwitchData{IP: "10.0.0.1", Status: "online", Mock: true})

	client := rtlplayground.NewWithClient("127.0.0.1:1", &http.Client{Timeout: 200 * time.Millisecond})
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)

	for i := 0; i < maxScrapeFailures; i++ {
		p.poll()
	}

	sw := cache.GetSwitch("10.0.0.1")
	if sw.Status != "offline" {
		t.Fatalf("status = %q, want offline", sw.Status)
	}
	if !sw.Mock {
		t.Error("offline demo switch must keep Mock flag")
	}
}

// A single scrape failure must not yet flip the status to offline.
func TestSingleFailureKeepsStatus(t *testing.T) {
	cache := server.NewCache()
	cache.UpdateSwitch("10.0.0.1", &server.SwitchData{IP: "10.0.0.1", Status: "online"})

	client := rtlplayground.NewWithClient("127.0.0.1:1", &http.Client{Timeout: 200 * time.Millisecond})
	p := NewWithClient(cache, client, "10.0.0.1", "test-switch", "SWGT024", 30, testLogger)

	p.poll()

	sw := cache.GetSwitch("10.0.0.1")
	if sw.Status != "online" {
		t.Fatalf("status = %q, want online after single failure", sw.Status)
	}
}
