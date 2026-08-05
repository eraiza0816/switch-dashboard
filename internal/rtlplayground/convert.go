package rtlplayground

import "strconv"

// ParseHex converts a "0x..." packet counter string to an int64. It returns 0
// for malformed input.
func ParseHex(s string) int64 {
	if len(s) < 3 || s[:2] != "0x" {
		return 0
	}
	var v int64
	for _, c := range s[2:] {
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= int64(c - '0')
		case c >= 'a' && c <= 'f':
			v |= int64(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v |= int64(c - 'A' + 10)
		default:
			return 0
		}
	}
	return v
}

// ParseMTUHex converts the /mtu.json max-frame-length value (a "0x" prefixed
// hex string, e.g. "0x05ee" = 1518) to decimal. It returns 0 for malformed
// input.
func ParseMTUHex(s string) int64 {
	if len(s) < 3 || s[:2] != "0x" {
		return 0
	}
	v, err := strconv.ParseInt(s[2:], 16, 32)
	if err != nil {
		return 0
	}
	return v
}

// LinkSpeedString converts the /status.json link code to a display speed
// string. The firmware derives the code as (RTL837X_REG_LINKS speed field + 1):
//
//	1=10M, 2=100M, 3=1G, 5=10G, 6=2.5G, 7=5G; 0=down, 4=undefined.
//
// (see RTLPlayground rtl837x_port.c port_stats_print and
// tools/rtlplayground_exporter linkSpeedToBPS).
func LinkSpeedString(link int) string {
	switch link {
	case 0:
		return ""
	case 1:
		return "10M"
	case 2:
		return "100M"
	case 3:
		return "1G"
	case 5:
		return "10G"
	case 6:
		return "2.5G"
	case 7:
		return "5G"
	default:
		return ""
	}
}
