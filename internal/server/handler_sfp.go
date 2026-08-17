package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPISwitchSFP(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		s.writeError(w, http.StatusBadRequest, "missing ip")
		return
	}
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		s.writeError(w, http.StatusNotFound, "switch not found")
		return
	}

	// Build transceiver data from cached switch data
	sfpInfo := parseSFPFromSwitchData(sw)

	if sfpInfo == nil {
		s.writeJSON(w, http.StatusOK, map[string]string{"error": "No SFP module detected"})
		return
	}
	s.writeJSON(w, http.StatusOK, sfpInfo)
}

type sfpTransceiverData struct {
	VendorName  string `json:"vendor_name,omitempty"`
	VendorPN    string `json:"vendor_pn,omitempty"`
	VendorSN    string `json:"vendor_sn,omitempty"`
	VendorRev   string `json:"vendor_revision,omitempty"`
	Type        string `json:"transceiver_type,omitempty"`
	Connector   string `json:"connector_type,omitempty"`
	Compliance  string `json:"eth_compliance_codes,omitempty"`
	Wavelength  string `json:"wavelength,omitempty"`
	Bitrate     string `json:"bitrate,omitempty"`
	Temperature string `json:"temperature,omitempty"`
	Voltage     string `json:"voltage,omitempty"`
	Current     string `json:"current,omitempty"`
	TXPower     string `json:"tx_power,omitempty"`
	RXPower     string `json:"rx_power,omitempty"`
	OEPresent   string `json:"oe_present,omitempty"`
	LOS         string `json:"loss_of_signal,omitempty"`
	DDMIEnabled string `json:"ddmi_enabled,omitempty"`
}

// parseSFPFromSwitchData builds the transceiver response from the data the
// poller scraped from the switch (/status.json SFP fields + /sfp_diag.json).
// The firmware does not expose EEPROM fields like vendor OUI, revision or
// type/connector codes, so those stay empty.
func parseSFPFromSwitchData(sw *SwitchData) *sfpTransceiverData {
	for _, p := range sw.Ports {
		if !p.IsSFP {
			continue
		}
		if p.SFPVendor == "" && p.SFPModel == "" {
			continue
		}
		info := &sfpTransceiverData{
			VendorName: p.SFPVendor,
			VendorPN:   p.SFPModel,
			VendorSN:   p.SFPSerial,
			OEPresent:  "1",
			LOS:        boolStr(p.SFPLos),
		}
		for _, d := range sw.SFPDiag {
			if d.Port == parsePortNum(p.Port) {
				info.Temperature = d.Temp
				info.Voltage = d.VCC
				info.Current = d.Bias
				info.TXPower = d.TXPower
				info.RXPower = d.RXPower
				info.DDMIEnabled = boolStr(d.HasDDMI)
			}
		}
		return info
	}
	return nil
}

func boolStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func parsePortNum(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
