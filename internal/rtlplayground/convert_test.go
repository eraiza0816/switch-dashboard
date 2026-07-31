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

func TestLinkSpeedString(t *testing.T) {
	tests := []struct {
		link int
		want string
	}{
		{0, ""},
		{1, "100M"},
		{2, "1G"},
		{3, "2.5G"},
		{4, "10G"},
		{5, "Auto"},
		{99, "Auto"},
	}
	for _, tc := range tests {
		got := LinkSpeedString(tc.link)
		if got != tc.want {
			t.Errorf("LinkSpeedString(%d) = %q, want %q", tc.link, got, tc.want)
		}
	}
}
