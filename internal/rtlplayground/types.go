package rtlplayground

type Information struct {
	IPAddress     string `json:"ip_address"`
	IPGateway     string `json:"ip_gateway,omitempty"`
	IPNetmask     string `json:"ip_netmask,omitempty"`
	TelnetEnabled string `json:"telnet_enabled"`
	WebEnabled    string `json:"web_enabled"`
	MACAddress    string `json:"mac_address"`
	SwVer         string `json:"sw_ver"`
	BuildDate     string `json:"build_date,omitempty"`
	HWVer         string `json:"hw_ver"`
	FlashSize     string `json:"flash_size,omitempty"`
	Hostname      string `json:"hostname,omitempty"`
	SFPSlot0      string `json:"sfp_slot_0,omitempty"`
	SFPSlot1      string `json:"sfp_slot_1,omitempty"`
}

type StatusEntry struct {
	PortNum   int    `json:"portNum"`
	LogPort   int    `json:"logPort"`
	Name      string `json:"name,omitempty"`
	IsSFP     int    `json:"isSFP"`
	Enabled   int    `json:"enabled"`
	SFPVendor string `json:"sfp_vendor,omitempty"`
	SFPModel  string `json:"sfp_model,omitempty"`
	SFPSerial string `json:"sfp_serial,omitempty"`
	SFPLos    any    `json:"sfp_los,omitempty"`
	Adv       string `json:"adv,omitempty"`
	Link      int    `json:"link"`
	TxG       string `json:"txG"`
	TxB       string `json:"txB"`
	RxG       string `json:"rxG"`
	RxB       string `json:"rxB"`
}

func (s *StatusEntry) HasSFP() bool { return s.IsSFP != 0 }

type SFPDiagEntry struct {
	PortNum    int    `json:"portNum"`
	SFPOptions string `json:"sfp_options"`
	SFPTemp    string `json:"sfp_temp,omitempty"`
	SFPVCC     string `json:"sfp_vcc,omitempty"`
	SFPTXBias  string `json:"sfp_txbias,omitempty"`
	SFPTXPower string `json:"sfp_txpower,omitempty"`
	SFPRXPower string `json:"sfp_rxpower,omitempty"`
	SFPState   string `json:"sfp_state,omitempty"`
}

type SFPEEPROM struct {
	Slot int    `json:"slot"`
	Data string `json:"data"`
}

type L2Entry struct {
	MAC   string `json:"mac"`
	VLAN  string `json:"vlan"`
	Type  string `json:"type"`
	Port  any    `json:"port"`
	Index string `json:"idx"`
}

func (e *L2Entry) PortStr() string {
	switch v := e.Port.(type) {
	case string:
		return v
	case float64:
		return itoa64(int64(v))
	default:
		return ""
	}
}

type L2DeleteResult struct {
	Result int `json:"result"`
}

type VLANEntry struct {
	Members string `json:"members"`
	Name    string `json:"name,omitempty"`
	PVID    string `json:"pvid,omitempty"`
}

type VLANListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EEEEntry struct {
	PortNum int    `json:"portNum"`
	IsSFP   int    `json:"isSFP"`
	EEE     string `json:"eee,omitempty"`
	EEELP   string `json:"eee_lp,omitempty"`
	Active  int    `json:"active"`
}

type BandwidthEntry struct {
	PortNum   int    `json:"portNum"`
	ILimited  int    `json:"iLimited"`
	IBW       string `json:"iBW"`
	IFC       int    `json:"iFC"`
	ELimited  int    `json:"eLimited"`
	EBW       string `json:"eBW"`
}

type MirrorConfig struct {
	Enabled  int    `json:"enabled"`
	MPort    int    `json:"mPort"`
	MirrorRX string `json:"mirror_rx"`
	MirrorTX string `json:"mirror_tx"`
}

type LAGEntry struct {
	LAGNum  int    `json:"lagNum"`
	Members string `json:"members"`
	Hash    string `json:"hash"`
}

type MTUEntry struct {
	PortNum int    `json:"portNum"`
	MTU     string `json:"mtu"`
}

type PortInfo struct {
	PortNum   int    `json:"portNum"`
	Name      string `json:"name,omitempty"`
}

func itoa64(v int64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := v < 0
	if neg {
		v = -v
	}
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
