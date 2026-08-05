package rtlplayground

import (
	"fmt"
	"math"
)

// SFP DDMI scaling factors, matching the firmware WebUI (html/main.js
// pollSfpDiag): temperature is a signed 16-bit value in units of 1/256 °C,
// supply voltage in units of 0.1 mV, bias current in units of 2 µA, and the
// optical power readings in units of 0.1 µW (mW * 1e-4).
const (
	sfpTempScale  = 1.0 / 256.0
	sfpVCCScale   = 0.0001
	sfpBiasScale  = 0.002
	sfpPowerScale = 0.0001 // mW per LSB

	sfpOptionDDMI = 0x40
)

// SFPDiagValue is a sfp_diag.json entry converted to display-ready units.
type SFPDiagValue struct {
	Port    int
	Options string
	Temp    string
	VCC     string
	Bias    string
	TXPower string
	RXPower string
	State   string
	HasDDMI bool
}

// FormatSFPDiag converts the raw hex strings of one /sfp_diag.json entry
// into human-readable values (°C, V, mA, dBm). Zero readings are reported as
// empty strings, matching the firmware WebUI which skips 0x0000 values.
func FormatSFPDiag(e SFPDiagEntry) SFPDiagValue {
	v := SFPDiagValue{
		Port:    e.PortNum,
		Options: e.SFPOptions,
		State:   e.SFPState,
		HasDDMI: ParseHex(e.SFPOptions)&sfpOptionDDMI != 0,
	}
	if t := parseSFPSigned16(e.SFPTemp); t != 0 {
		v.Temp = fmt.Sprintf("%.1f C", float64(t)*sfpTempScale)
	}
	if c := ParseHex(e.SFPVCC); c != 0 {
		v.VCC = fmt.Sprintf("%.2f V", float64(c)*sfpVCCScale)
	}
	if b := ParseHex(e.SFPTXBias); b != 0 {
		v.Bias = fmt.Sprintf("%.2f mA", float64(b)*sfpBiasScale)
	}
	if mw := float64(ParseHex(e.SFPTXPower)) * sfpPowerScale; mw != 0 {
		v.TXPower = fmt.Sprintf("%.2f dBm", 10*math.Log10(mw))
	}
	if mw := float64(ParseHex(e.SFPRXPower)) * sfpPowerScale; mw != 0 {
		v.RXPower = fmt.Sprintf("%.2f dBm", 10*math.Log10(mw))
	}
	return v
}

// parseSFPSigned16 interprets a "0x" hex string as a signed 16-bit value
// (temperature is two's-complement).
func parseSFPSigned16(s string) int16 {
	v := ParseHex(s)
	if v > 0x7fff {
		v -= 0x10000
	}
	return int16(v)
}
