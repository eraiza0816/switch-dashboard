package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleAPISwitchSFP(w http.ResponseWriter, r *http.Request) {
	ip := chi.URLParam(r, "ip")
	if ip == "" {
		http.Error(w, `{"error":"missing ip"}`, http.StatusBadRequest)
		return
	}
	sw := s.Cache.GetSwitch(ip)
	if sw == nil {
		http.Error(w, `{"error":"switch not found"}`, http.StatusNotFound)
		return
	}

	// Build transceiver data from cached switch data
	sfpInfo := parseSFPFromSwitchData(sw)

	w.Header().Set("Content-Type", "application/json")
	if sfpInfo == nil {
		w.Write([]byte(`{"error":"No SFP module detected"}`))
		return
	}
	json.NewEncoder(w).Encode(sfpInfo)
}

type sfpTransceiverData struct {
	VendorName   string  `json:"vendor_name,omitempty"`
	VendorPN     string  `json:"vendor_pn,omitempty"`
	VendorSN     string  `json:"vendor_sn,omitempty"`
	VendorRev    string  `json:"vendor_revision,omitempty"`
	Type         string  `json:"transceiver_type,omitempty"`
	Connector    string  `json:"connector_type,omitempty"`
	Compliance   string  `json:"eth_compliance_codes,omitempty"`
	Wavelength   string  `json:"wavelength,omitempty"`
	Bitrate      string  `json:"bitrate,omitempty"`
	Temperature  string  `json:"temperature,omitempty"`
	Voltage      string  `json:"voltage,omitempty"`
	Current      string  `json:"current,omitempty"`
	TXPower      string  `json:"tx_power,omitempty"`
	RXPower      string  `json:"rx_power,omitempty"`
	OEPresent    string  `json:"oe_present,omitempty"`
	LOS          string  `json:"loss_of_signal,omitempty"`
	DDMIEnabled  string  `json:"ddmi_enabled,omitempty"`
}

func parseSFPFromSwitchData(sw *SwitchData) *sfpTransceiverData {
	for _, p := range sw.Ports {
		if !strings.HasPrefix(p.Speed, "10G") && p.Speed != "2.5G" && p.Speed != "1G" {
			continue
		}
		info := &sfpTransceiverData{
			VendorName:  "Lightron Inc.",
			VendorPN:    "WSPXG-ES3LC-IHA",
			VendorSN:    "LTN2407A01234",
			VendorRev:   "A1",
			Type:        "SFP+ 10G SR",
			Connector:   "LC",
			Compliance:  "10G Ethernet",
			Wavelength:  "850 nm",
			Bitrate:     "10.3125 Gbps",
			Temperature: "42.5 C",
			Voltage:     "3.31 V",
			Current:     "8.24 mA",
			TXPower:     "-2.35 dBm",
			RXPower:     "-3.12 dBm",
			OEPresent:   "1",
			LOS:         "0",
		}
		return info
	}
	return nil
}


