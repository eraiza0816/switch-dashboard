package rtlplayground

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

// LinkSpeedString converts a link number to a display speed string.
func LinkSpeedString(link int) string {
	switch link {
	case 0:
		return ""
	case 1:
		return "100M"
	case 2:
		return "1G"
	case 3:
		return "2.5G"
	case 4:
		return "10G"
	default:
		return "Auto"
	}
}
