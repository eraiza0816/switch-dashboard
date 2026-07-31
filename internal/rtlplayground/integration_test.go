//go:build integration

package rtlplayground

import (
	"os"
	"testing"
)

func TestIntegrationAgainstSimulator(t *testing.T) {
	ip := os.Getenv("SWITCH_IP")
	if ip == "" {
		ip = "127.0.0.1:8080"
	}
	password := os.Getenv("SWITCH_PASSWORD")
	if password == "" {
		password = "1234"
	}

	client, err := New(ip, password)
	if err != nil {
		t.Fatalf("New client failed: %v", err)
	}

	info, err := client.ScrapeInformation()
	if err != nil {
		t.Fatalf("ScrapeInformation: %v", err)
	}
	t.Logf("Switch: %s %s (IP: %s, MAC: %s)", info.HWVer, info.SwVer, info.IPAddress, info.MACAddress)

	status, err := client.ScrapeStatus()
	if err != nil {
		t.Fatalf("ScrapeStatus: %v", err)
	}
	t.Logf("Ports: %d", len(status))
	for _, p := range status {
		t.Logf("  Port %d: link=%d enabled=%d txG=%s", p.PortNum, p.Link, p.Enabled, p.TxG)
	}

	diag, err := client.ScrapeSFPDiag()
	if err != nil {
		t.Logf("ScrapeSFPDiag (may be empty): %v", err)
	} else {
		t.Logf("SFP diag entries: %d", len(diag))
	}

	macTable, err := client.ScrapeAllMACTable()
	if err != nil {
		t.Fatalf("ScrapeAllMACTable: %v", err)
	}
	t.Logf("MAC table entries: %d", len(macTable))
	for _, e := range macTable {
		t.Logf("  %s -> port %s (VLAN %s, type %s)", e.MAC, e.Port, e.VLAN, e.Type)
	}

	eee, err := client.ScrapeEEE()
	if err != nil {
		t.Fatalf("ScrapeEEE: %v", err)
	}
	t.Logf("EEE entries: %d", len(eee))

	mtu, err := client.ScrapeMTU()
	if err != nil {
		t.Fatalf("ScrapeMTU: %v", err)
	}
	t.Logf("MTU entries: %d", len(mtu))

	vlanList, err := client.ScrapeVLANList()
	if err != nil {
		t.Fatalf("ScrapeVLANList: %v", err)
	}
	t.Logf("VLANs: %d", len(vlanList))
	for _, v := range vlanList {
		t.Logf("  VLAN %d: %s", v.ID, v.Name)
	}

	bw, err := client.ScrapeBandwidth()
	if err != nil {
		t.Fatalf("ScrapeBandwidth: %v", err)
	}
	t.Logf("Bandwidth entries: %d", len(bw))

	mirror, err := client.ScrapeMirror()
	if err != nil {
		t.Fatalf("ScrapeMirror: %v", err)
	}
	t.Logf("Mirror: enabled=%d monitorPort=%d", mirror.Enabled, mirror.MPort)

	lag, err := client.ScrapeLAG()
	if err != nil {
		t.Fatalf("ScrapeLAG: %v", err)
	}
	t.Logf("LAG groups: %d", len(lag))
}
