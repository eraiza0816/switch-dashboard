package rtlplayground

import "testing"

func TestParseHex(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"0x0", 0},
		{"0x10", 16},
		{"0xff", 255},
		{"0x100", 256},
		{"0x8f0d1800", 2400000000},
		{"0x000000008f0d1800", 2400000000},
	}
	for _, tc := range tests {
		got := ParseHex(tc.input)
		if got != tc.want {
			t.Errorf("ParseHex(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseHexEdge(t *testing.T) {
	if got := ParseHex(""); got != 0 {
		t.Fatalf("empty string: got %d", got)
	}
	if got := ParseHex("0x"); got != 0 {
		t.Fatalf("0x only: got %d", got)
	}
	if got := ParseHex("invalid"); got != 0 {
		t.Fatalf("invalid hex: got %d", got)
	}
}

// TestParseMTUHex verifies the /mtu.json hex values (lower- and upper-case)
// parse to decimal, matching the firmware's byte_to_html output.
func TestParseMTUHex(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"0x05ee", 1518},
		{"0x05DC", 1500},
		{"0x2328", 9000},
		{"0x3fff", 16383},
		{"0x0", 0},
		{"", 0},
		{"05ee", 0},
	}
	for _, tc := range tests {
		got := ParseMTUHex(tc.input)
		if got != tc.want {
			t.Errorf("ParseMTUHex(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestLinkSpeedString verifies the firmware link-code mapping: the code is
// (RTL837X_REG_LINKS speed field + 1), 1=10M, 2=100M, 3=1G, 5=10G, 6=2.5G,
// 7=5G; 0=down, 4=undefined (see RTLPlayground rtl837x_port.c
// port_stats_print and the exporter's linkSpeedToBPS).
func TestLinkSpeedString(t *testing.T) {
	tests := []struct {
		link int
		want string
	}{
		{0, ""},
		{1, "10M"},
		{2, "100M"},
		{3, "1G"},
		{4, ""}, // undefined speed code
		{5, "10G"},
		{6, "2.5G"},
		{7, "5G"},
		{99, ""},
	}
	for _, tc := range tests {
		got := LinkSpeedString(tc.link)
		if got != tc.want {
			t.Errorf("LinkSpeedString(%d) = %q, want %q", tc.link, got, tc.want)
		}
	}
}
