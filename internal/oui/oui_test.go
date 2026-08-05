package oui

import (
	"strings"
	"testing"
)

func TestParseCSVAcceptsAllRegistries(t *testing.T) {
	csv := `Registry,Assignment,Organization Name,Organization Address
MA-L,0012AB,ExampleCorp L,addr
MA-M,0012AB5,ExampleCorp M,addr
MA-S,0012AB511,ExampleCorp S,addr
MA-S,9ABCDEF01,ExampleCorp S2,addr
`
	db := &DB{entries: map[string]string{}}
	entries, err := parseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	db.entries = entries

	cases := []struct {
		mac  string
		want string
	}{
		{"00:12:AB:12:34:56", "ExampleCorp L"},           // MA-L 6 hex
		{"00:12:AB:5C:77:33:44", "ExampleCorp M"},        // MA-M 7 hex
		{"00:12:AB:51:1A:2B:3C:4D", "ExampleCorp S"},     // MA-S 9 hex
		{"9A:BC:DE:F0:12:34:56:78:9A", "ExampleCorp S2"}, // MA-S full 9 hex
		{"AA:BB:CC:DD:EE:01", ""},                        // unregistered
	}
	for _, c := range cases {
		if got := db.Lookup(c.mac); got != c.want {
			t.Errorf("Lookup(%q) = %q, want %q", c.mac, got, c.want)
		}
	}
}

func TestLookupLongestPrefixWins(t *testing.T) {
	db := &DB{entries: map[string]string{
		"0012AB":  "L-vendor",
		"0012AB5": "M-vendor",
	}}
	// 6-hex match only
	if got := db.Lookup("0012AB123456"); got != "L-vendor" {
		t.Errorf("short match = %q, want L-vendor", got)
	}
	// 8-hex MAC resolves via 7-hex MA-M prefix
	if got := db.Lookup("00:12:AB:5C:77"); got != "M-vendor" {
		t.Errorf("long match = %q, want M-vendor", got)
	}
}

func TestCustomOverrideBeatsRegistry(t *testing.T) {
	db := &DB{
		entries: map[string]string{"0012AB": "L-vendor"},
		custom:  map[string]string{"0012AB": "My Override"},
	}
	if got := db.Lookup("00:12:AB:12:34:56"); got != "My Override" {
		t.Errorf("custom lookup = %q, want My Override", got)
	}
}

func TestNormalizeAssignmentKeepsLength(t *testing.T) {
	if got := normalizeAssignment("00-12-AB-5C"); got != "0012AB5C" {
		t.Errorf("normalizeAssignment = %q, want 0012AB5C", got)
	}
	if got := normalizeAssignment("12:34:56"); got != "123456" {
		t.Errorf("normalizeAssignment = %q, want 123456", got)
	}
}
