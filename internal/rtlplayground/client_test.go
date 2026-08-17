package rtlplayground

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
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

// The firmware answers every login with a 302: valid password redirects to
// index.html with a session cookie, invalid to login.html without one.  A
// wrong password must be detected even though the HTTP status is the same.
func TestLoginWrongPasswordDetected(t *testing.T) {
	loginHits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHits++
		if r.FormValue("pwd") == "correct" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
			w.Header().Set("Location", "index.html")
		} else {
			w.Header().Set("Location", "login.html")
		}
		w.WriteHeader(http.StatusFound)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	if _, err := New(ts.Listener.Addr().String(), "wrong"); err == nil {
		t.Fatal("expected login failure for wrong password")
	}

	// A 302 without a session cookie is also a failure (e.g. stale backend).
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "index.html")
		w.WriteHeader(http.StatusFound)
	}))
	defer ts2.Close()
	if _, err := New(ts2.Listener.Addr().String(), "correct"); err == nil {
		t.Fatal("expected login failure without session cookie")
	}

	c, err := New(ts.Listener.Addr().String(), "correct")
	if err != nil {
		t.Fatalf("login with correct password: %v", err)
	}
	if c.password != "correct" {
		t.Fatalf("password not stored: %q", c.password)
	}
	if loginHits != 2 {
		t.Fatalf("login hits = %d, want 2 (wrong + correct)", loginHits)
	}
}

// The firmware session expires after 200s (SESSION_TIMEOUT in httpd.c) and
// JSON API requests never refresh it, so a 401 must trigger a re-login and a
// retry of the request.
func TestSessionExpiryReLogin(t *testing.T) {
	loginHits := 0
	statusHits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHits++
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess"})
		w.Header().Set("Location", "index.html")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/status.json", func(w http.ResponseWriter, r *http.Request) {
		statusHits++
		if statusHits == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"portNum":1,"link":5,"enabled":1}]`))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, err := New(ts.Listener.Addr().String(), "pw")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := c.ScrapeStatus(); err != nil {
		t.Fatalf("ScrapeStatus after re-login: %v", err)
	}
	if loginHits != 2 {
		t.Fatalf("login hits = %d, want 2 (initial + re-login)", loginHits)
	}
	if statusHits != 2 {
		t.Fatalf("status hits = %d, want 2 (401 + retry)", statusHits)
	}
}

// POST /cmd must also survive a session expiry (the body is regenerated).
func TestCommandReLogin(t *testing.T) {
	loginHits := 0
	cmdHits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHits++
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess"})
		w.Header().Set("Location", "index.html")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
		cmdHits++
		if cmdHits == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, err := New(ts.Listener.Addr().String(), "pw")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if err := c.ExecuteCommand("hostname test"); err != nil {
		t.Fatalf("ExecuteCommand after re-login: %v", err)
	}
	if loginHits != 2 {
		t.Fatalf("login hits = %d, want 2", loginHits)
	}
	if cmdHits != 2 {
		t.Fatalf("cmd hits = %d, want 2 (401 + retry)", cmdHits)
	}
}

// Reboot is executed as the "reset" CLI command, since the firmware has no
// /reset HTTP endpoint.
func TestRebootSendsResetCommand(t *testing.T) {
	var mu sync.Mutex
	var cmds []string
	mux := http.NewServeMux()
	mux.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		cmds = append(cmds, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), ts.Client())
	if err := c.Reboot(); err != nil {
		t.Fatalf("Reboot: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(cmds) != 1 || cmds[0] != "reset" {
		t.Fatalf("commands = %v, want [reset]", cmds)
	}
}

// The firmware parses the raw POST /cmd body as the command text (see
// httpd.c execute_commands), so it must not be form-encoded.
func TestExecuteCommandRawBody(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	mux := http.NewServeMux()
	mux.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), ts.Client())
	if err := c.ExecuteCommand("port 3 name uplink"); err != nil {
		t.Fatalf("ExecuteCommand: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 1 || bodies[0] != "port 3 name uplink" {
		t.Fatalf("body = %q, want raw command text", bodies[0])
	}
}

func TestScrapeCounters(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/counters.json", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("port") != "3" {
			t.Errorf("port query = %q, want 3", r.URL.Query().Get("port"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`["0x0000000000001000","0x0000000000002000","0x0000000000003000"]`))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), ts.Client())
	counters, err := c.ScrapeCounters(3)
	if err != nil {
		t.Fatalf("ScrapeCounters: %v", err)
	}
	if len(counters) != 3 {
		t.Fatalf("expected 3 counters, got %d", len(counters))
	}
	if ParseHex(counters[0]) != 0x1000 || ParseHex(counters[1]) != 0x2000 {
		t.Fatalf("unexpected counters: %v", counters)
	}
}

// FormatSFPDiag converts raw sfp_diag.json hex into display units using the
// same scaling as the firmware WebUI (main.js pollSfpDiag).
func TestFormatSFPDiag(t *testing.T) {
	v := FormatSFPDiag(SFPDiagEntry{
		PortNum:    6,
		SFPOptions: "0x68",
		SFPTemp:    "0x2a67", // 10855 / 256 = 42.4 C
		SFPVCC:     "0x7f01", // 32513 * 0.0001 = 3.25 V
		SFPTXBias:  "0x0d72", // 3442 * 0.002 = 6.88 mA
		SFPTXPower: "0x150c", // 5388 * 0.0001 mW = -2.69 dBm
		SFPRXPower: "0x0000",
		SFPState:   "0x00",
	})
	if v.Port != 6 || !v.HasDDMI {
		t.Fatalf("port/ddmi: %+v", v)
	}
	if v.Temp != "42.4 C" {
		t.Errorf("temp = %q, want 42.4 C", v.Temp)
	}
	if v.VCC != "3.25 V" {
		t.Errorf("vcc = %q, want 3.25 V", v.VCC)
	}
	if v.Bias != "6.88 mA" {
		t.Errorf("bias = %q, want 6.88 mA", v.Bias)
	}
	if v.TXPower != "-2.69 dBm" {
		t.Errorf("tx power = %q, want -2.69 dBm", v.TXPower)
	}
	if v.RXPower != "" {
		t.Errorf("rx power = %q, want empty (zero reading)", v.RXPower)
	}
}

func TestFormatSFPDiagNegativeTemp(t *testing.T) {
	v := FormatSFPDiag(SFPDiagEntry{
		PortNum:    9,
		SFPOptions: "0x00",
		SFPTemp:    "0xff00", // -256 / 256 = -1.0 C
	})
	if v.Temp != "-1.0 C" {
		t.Errorf("temp = %q, want -1.0 C", v.Temp)
	}
	if v.HasDDMI {
		t.Error("ddmi should be false without the 0x40 option bit")
	}
}

func TestLOS(t *testing.T) {
	los1 := StatusEntry{SFPLos: float64(1)}
	if !los1.LOS() {
		t.Error("LOS(1) should be true")
	}
	los0 := StatusEntry{SFPLos: float64(0)}
	if los0.LOS() {
		t.Error("LOS(0) should be false")
	}
	losNil := StatusEntry{}
	if losNil.LOS() {
		t.Error("LOS(null) should be false")
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

// encTestServer mocks the firmware /enc endpoint: it decrypts the request
// body (nonce[12] || ct || tag[16]), records the plaintext command and
// responds with an encrypted fixed body.
func encTestServer(key []byte, respPlain string) (*httptest.Server, *sync.Mutex, *string) {
	lastCmd := ""
	var mu sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/enc" {
			http.NotFound(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) < aeadNonceLen+aeadTagLen+1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctLen := len(body) - aeadNonceLen - aeadTagLen
		pt, err := aeadDecrypt(key, body[:aeadNonceLen], body[aeadNonceLen:aeadNonceLen+ctLen], body[len(body)-aeadTagLen:])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		mu.Lock()
		lastCmd = string(pt)
		mu.Unlock()
		respNonce := bytes.Repeat([]byte{0x22}, 12)
		pkt, _ := aeadEncrypt(key, respNonce, []byte(respPlain))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(pkt)
	}))
	return ts, &mu, &lastCmd
}

func TestPostEnc(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)
	ts, mu, lastCmd := encTestServer(key, `{"result":"ok"}`)
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), &http.Client{})
	if err := c.SetPSK(keyHex); err != nil {
		t.Fatal(err)
	}
	resp, err := c.PostEnc("hostname test")
	if err != nil {
		t.Fatalf("PostEnc: %v", err)
	}
	if resp != `{"result":"ok"}` {
		t.Fatalf("unexpected response: %q", resp)
	}
	mu.Lock()
	got := *lastCmd
	mu.Unlock()
	if got != "hostname test" {
		t.Fatalf("server received %q, want %q", got, "hostname test")
	}
}

func TestPostEncNoPSK(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)
	ts, _, _ := encTestServer(key, "{}")
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), &http.Client{})
	if _, err := c.PostEnc("hostname test"); err == nil {
		t.Fatal("PostEnc without PSK should fail")
	}
}

func TestEncAPI(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)
	ts, mu, lastCmd := encTestServer(key, `{"name":"test"}`)
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), &http.Client{})
	if err := c.SetPSK(keyHex); err != nil {
		t.Fatal(err)
	}
	body, err := c.EncAPI("/status.json")
	if err != nil {
		t.Fatalf("EncAPI: %v", err)
	}
	if string(body) != `{"name":"test"}` {
		t.Fatalf("unexpected body: %q", body)
	}
	mu.Lock()
	got := *lastCmd
	mu.Unlock()
	if got != "api /status.json" {
		t.Fatalf("server received %q, want %q", got, "api /status.json")
	}
}

func TestExecuteCommand(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)

	var mu sync.Mutex
	cmdHits, encHits := 0, 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cmd":
			mu.Lock()
			cmdHits++
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		case "/enc":
			body, _ := io.ReadAll(r.Body)
			ctLen := len(body) - aeadNonceLen - aeadTagLen
			pt, err := aeadDecrypt(key, body[:aeadNonceLen], body[aeadNonceLen:aeadNonceLen+ctLen], body[len(body)-aeadTagLen:])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			mu.Lock()
			encHits++
			mu.Unlock()
			respNonce := bytes.Repeat([]byte{0x33}, 12)
			pkt, _ := aeadEncrypt(key, respNonce, []byte(`{"result":"ok"}`))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(pkt)
			_ = pt
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), &http.Client{})

	// no PSK: command goes to plaintext /cmd
	if err := c.ExecuteCommand("hostname switch-1"); err != nil {
		t.Fatalf("ExecuteCommand (plain): %v", err)
	}
	if cmdHits != 1 || encHits != 0 {
		t.Fatalf("plain routing: cmd=%d enc=%d, want cmd=1 enc=0", cmdHits, encHits)
	}

	// PSK configured: command goes to encrypted /enc
	if err := c.SetPSK(keyHex); err != nil {
		t.Fatal(err)
	}
	if err := c.ExecuteCommand("hostname switch-1"); err != nil {
		t.Fatalf("ExecuteCommand (enc): %v", err)
	}
	if cmdHits != 1 || encHits != 1 {
		t.Fatalf("enc routing: cmd=%d enc=%d, want cmd=1 enc=1", cmdHits, encHits)
	}

	// invalid command is rejected before anything is sent
	if err := c.ExecuteCommand("hostname bad..name"); err == nil {
		t.Fatal("ExecuteCommand accepted invalid command")
	}
	if cmdHits != 1 || encHits != 1 {
		t.Fatalf("invalid command was sent: cmd=%d enc=%d", cmdHits, encHits)
	}
}

// The PSK login challenge must decrypt back to the fixed challenge string
// with the same key.
func TestPSKLoginChallenge(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)

	c := NewWithClient("127.0.0.1:1", &http.Client{})
	if err := c.SetPSK(keyHex); err != nil {
		t.Fatal(err)
	}
	challenge, err := c.pskLoginChallenge()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := hex.DecodeString(challenge)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != aeadNonceLen+len("RTLP-LOGIN-1")+aeadTagLen {
		t.Fatalf("challenge length = %d, want %d", len(raw), aeadNonceLen+len("RTLP-LOGIN-1")+aeadTagLen)
	}
	pt, err := aeadDecrypt(key, raw[:aeadNonceLen], raw[aeadNonceLen:len(raw)-aeadTagLen], raw[len(raw)-aeadTagLen:])
	if err != nil {
		t.Fatal(err)
	}
	if string(pt) != "RTLP-LOGIN-1" {
		t.Fatalf("challenge plaintext = %q, want RTLP-LOGIN-1", pt)
	}
}

// In PSK mode the firmware rejects password logins, so NewWithPSK must send
// the encrypted enc= challenge instead of pwd=, and scraping must work with
// the resulting session.
func TestNewWithPSKLogsInWithEncChallenge(t *testing.T) {
	keyHex := strings.Repeat("42", 32)
	key, _ := hex.DecodeString(keyHex)

	var mu sync.Mutex
	loginForm := ""
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		loginForm = string(body)
		mu.Unlock()
		ch, err := hex.DecodeString(strings.TrimPrefix(string(body), "enc="))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if len(ch) < aeadNonceLen+aeadTagLen {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		pt, err := aeadDecrypt(key, ch[:aeadNonceLen], ch[aeadNonceLen:len(ch)-aeadTagLen], ch[len(ch)-aeadTagLen:])
		if err != nil || string(pt) != "RTLP-LOGIN-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "psk-session"})
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/information.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Information{HWVer: "PSK-SWITCH", SwVer: "v0.2.23"})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c, err := NewWithPSK(ts.Listener.Addr().String(), "ignored", keyHex)
	if err != nil {
		t.Fatalf("NewWithPSK: %v", err)
	}
	mu.Lock()
	form := loginForm
	mu.Unlock()
	if !strings.HasPrefix(form, "enc=") {
		t.Fatalf("login form = %q, want enc= challenge", form)
	}
	if strings.Contains(form, "pwd=") {
		t.Fatalf("login form must not contain pwd=: %q", form)
	}

	info, err := c.ScrapeInformation()
	if err != nil {
		t.Fatalf("scrape after psk login: %v", err)
	}
	if info.HWVer != "PSK-SWITCH" {
		t.Fatalf("hw ver = %q, want PSK-SWITCH", info.HWVer)
	}
}

// The firmware signals a completed configuration upload by closing the
// connection, so an EOF on /config must be treated as success.
func TestConfigUploadEOFIsSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config" {
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("hijack unsupported")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatal(err)
			}
			io.Copy(io.Discard, r.Body)
			conn.Close()
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c := NewWithClient(ts.Listener.Addr().String(), &http.Client{})
	if err := c.UploadConfig("hostname test\n"); err != nil {
		t.Fatalf("UploadConfig: %v", err)
	}
}
