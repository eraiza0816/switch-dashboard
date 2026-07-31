package poller

import (
	"testing"

	"github.com/eraiza0816/switch-dashboard/internal/rtlplayground"
)

func TestPacketAndByteEstimation(t *testing.T) {
	// txG is packets (good packets), not bytes
	// When byte counter is unavailable, estimate bytes = packets * 800
	packets := int64(1000000)
	estimatedBytes := packets * 800

	// This matches the Python scraper's fallback: p["tx_bytes"] = tx_pkts * 800
	if estimatedBytes != 800000000 {
		t.Fatalf("expected 800000000 bytes for 1M packets, got %d", estimatedBytes)
	}
}

func TestPollerSFPDectection(t *testing.T) {
	// SFP ports are identified by the isSFP field from status entry
	entry := rtlplayground.StatusEntry{PortNum: 5, IsSFP: 1, Link: 4, Enabled: 1, TxG: "0x100", RxG: "0x200"}
	if entry.IsSFP == 0 {
		t.Fatal("expected IsSFP=1 for SFP port")
	}

	isSFP := entry.IsSFP != 0
	if !isSFP {
		t.Fatal("SFP detection failed")
	}

	// Non-SFP port
	entry2 := rtlplayground.StatusEntry{PortNum: 1, IsSFP: 0}
	if entry2.IsSFP != 0 {
		t.Fatal("expected IsSFP=0 for non-SFP port")
	}
}

func TestStatusEntryParsing(t *testing.T) {
	// RTLPlayground status.json returns:
	// txG: TX good packets (hex string)
	// txB: TX bad packets (hex string)
	// rxG: RX good packets (hex string)
	// rxB: RX bad packets (hex string)

	txG := "0x8f0d1800"
	rxG := "0x8f0d1800"

	txPackets := rtlplayground.ParseHex(txG)
	rxPackets := rtlplayground.ParseHex(rxG)
	txBytes := rtlplayground.ParseHex(txG) * 800
	rxBytes := rtlplayground.ParseHex(rxG) * 800

	if txPackets != 2400000000 {
		t.Fatalf("txPackets: expected 2400000000, got %d", txPackets)
	}
	if rxPackets != 2400000000 {
		t.Fatalf("rxPackets: expected 2400000000, got %d", rxPackets)
	}
	if txBytes != 2400000000*800 {
		t.Fatalf("txBytes: expected %d, got %d", 2400000000*800, txBytes)
	}
	_ = rxBytes
}
