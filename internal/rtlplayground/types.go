package rtlplayground

type Information struct {
	IP       string `json:"ip"`
	GW       string `json:"gw"`
	Mask     string `json:"mask"`
	MAC      string `json:"mac"`
	SwVer    string `json:"sw_ver"`
	Build    string `json:"build"`
	HWVer    string `json:"hw_ver"`
	Flash    int64  `json:"flash"`
	Hostname string `json:"hostname"`
	Telnet   int    `json:"telnet"`
	Web      int    `json:"web"`
	SFP0     string `json:"sfp0"`
	SFP1     string `json:"sfp1"`
}

type StatusEntry struct {
	PortNum int    `json:"portNum"`
	Name    string `json:"name"`
	Link    int    `json:"link"`
	Enabled int    `json:"enabled"`
	SFP     string `json:"sfp,omitempty"`
	Adv     string `json:"adv"`
	TxG     string `json:"txG"`
	RxG     string `json:"rxG"`
	TxB     string `json:"txB"`
	RxB     string `json:"rxB"`
}

type SFPDiagEntry struct {
	Port        int    `json:"port"`
	Options     int    `json:"options"`
	Temperature string `json:"temperature"`
	VCC         string `json:"vcc"`
	TXBias      string `json:"tx_bias"`
	TXPower     string `json:"tx_power"`
	RXPower     string `json:"rx_power"`
	Laser       int    `json:"laser"`
}

type SFPEEPROM struct {
	Slot int    `json:"slot"`
	Data string `json:"data"`
}

type L2Entry struct {
	MAC   string `json:"mac"`
	VLAN  string `json:"vlan"`
	Type  string `json:"type"`
	Port  string `json:"port"`
	Index int    `json:"idx,omitempty"`
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
	Port       int  `json:"port"`
	IsSFP      bool `json:"isSFP"`
	Adv2500    bool `json:"adv2500"`
	Adv1000    bool `json:"adv1000"`
	Adv100     bool `json:"adv100"`
	LP2500     bool `json:"lp2500"`
	LP1000     bool `json:"lp1000"`
	LP100      bool `json:"lp100"`
	Active     bool `json:"active"`
}

type BandwidthEntry struct {
	Port     int    `json:"port"`
	IngLimit int    `json:"ing_limit"`
	IngBW    string `json:"ing_bw"`
	IngFC    int    `json:"ing_fc"`
	EgrLimit int    `json:"egr_limit"`
	EgrBW    string `json:"egr_bw"`
}

type MirrorConfig struct {
	Enabled  int    `json:"enabled"`
	MonPort  int    `json:"mon_port"`
	MirrorRX string `json:"mirror_rx"`
	MirrorTX string `json:"mirror_tx"`
}

type LAGEntry struct {
	Group  int    `json:"group"`
	Ports  string `json:"ports"`
	Hash   string `json:"hash"`
}

type MTUEntry struct {
	Port int    `json:"port"`
	MTU  string `json:"mtu"`
}

type CounterEntry struct {
	Port int    `json:"port"`
	Name string `json:"name,omitempty"`
	Val  string `json:"val,omitempty"`
}
