package poller

import (
	"testing"

	"github.com/byte4geek/switch-dashboard/internal/rtlplayground"
	"github.com/byte4geek/switch-dashboard/internal/server"
)

func TestNotesAttachedToPorts(t *testing.T) {
	cache := server.NewCache()
	notes := map[string]string{
		"192.168.1.1:1": "uplink to core",
		"192.168.1.1:9": "SFP fiber uplink",
	}

	p := &Poller{
		cache:    cache,
		ip:       "192.168.1.1",
		name:     "Test",
		model:    "RTLPlayground",
		interval: 30,
		notes:    notes,
		history:  make(map[string]*History),
		counters: make(map[string]*CounterState),
	}

	status := []rtlplayground.StatusEntry{
		{PortNum: 1, Link: 4, Enabled: 1, TxG: "0x100", RxG: "0x200"},
		{PortNum: 9, Link: 4, Enabled: 1, TxG: "0x50", RxG: "0x60"},
	}

	// manually simulate what poll() does
	for _, entry := range status {
		key := "192.168.1.1:" + itoa(int64(entry.PortNum))
		n := p.notes[key]
		if entry.PortNum == 1 && n != "uplink to core" {
			t.Fatalf("port 1 note: got %q", n)
		}
		if entry.PortNum == 9 && n != "SFP fiber uplink" {
			t.Fatalf("port 9 note: got %q", n)
		}
	}
}

func TestNotesEmptyIfNotConfigured(t *testing.T) {
	p := &Poller{
		ip:       "192.168.1.1",
		notes:    make(map[string]string),
		counters: make(map[string]*CounterState),
		history:  make(map[string]*History),
	}

	entry := rtlplayground.StatusEntry{PortNum: 1}
	key := "192.168.1.1:" + itoa(int64(entry.PortNum))
	if p.notes[key] != "" {
		t.Fatalf("expected empty note, got %q", p.notes[key])
	}
}

func TestNewWithClientAndNotes(t *testing.T) {
	cache := server.NewCache()
	notes := map[string]string{"test:1": "hello"}
	p := NewWithClientAndNotes(cache, nil, "test", "n", "m", 30, notes)
	if p.notes["test:1"] != "hello" {
		t.Fatalf("notes not passed through constructor")
	}
}
