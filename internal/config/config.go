package config

import (
	"encoding/json"
	"os"
	"sync"
)

type SwitchConfig struct {
	Name       string `json:"name"`
	IP         string `json:"ip"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Model      string `json:"model"`
	PortCount  int    `json:"port_count"`
	Enabled    bool   `json:"enabled"`
	ParentIP   string `json:"parent_ip,omitempty"`
	ParentPort string `json:"parent_port,omitempty"`
	UplinkPort string `json:"uplink_port,omitempty"`
	Bridge     string `json:"bridge,omitempty"`
}

type Settings struct {
	IgnoredMACs []string `json:"ignored_macs,omitempty"`
	LogLevel    string   `json:"log_level,omitempty"`
}

type Config struct {
	mu sync.RWMutex

	Title                 string                  `json:"title"`
	RefreshInterval       int                     `json:"refresh_interval"`
	MACRefreshMultiplier  int                     `json:"mac_refresh_multiplier"`
	MaxRequestRetries     int                     `json:"max_request_retries"`
	PortsWrapThreshold    int                     `json:"ports_wrap_threshold"`
	GridColumns           string                  `json:"grid_columns"`
	EnabledColumns        []string                `json:"enabled_columns"`
	ColumnWidths          map[string]int          `json:"column_widths,omitempty"`
	ColumnOrder           []string                `json:"column_order,omitempty"`
	MapPositions          map[string]MapPosition  `json:"map_positions,omitempty"`
	Switches              []SwitchConfig          `json:"switches"`
	InfrastructureDevices []InfraDevice           `json:"infrastructure_devices,omitempty"`
	UnmanagedSwitches     []UnmanagedSwitch       `json:"unmanaged_switches,omitempty"`
	Clients               map[string]ClientEntry  `json:"clients,omitempty"`
	Notes                 map[string]string       `json:"notes,omitempty"`
	Settings              Settings                `json:"settings,omitempty"`
	ScannerEnabled        bool                    `json:"scanner_enabled"`
	TelemetryEnabled      bool                    `json:"telemetry_enabled"`
	path                  string
}

type MapPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type InfraDevice struct {
	Name string `json:"name"`
	MAC  string `json:"mac"`
	Type string `json:"type"`
}

type UnmanagedSwitch struct {
	Name      string `json:"name"`
	ParentIP  string `json:"parent_ip"`
	ParentPort string `json:"parent_port"`
}

type ClientEntry struct {
	MAC            string `json:"mac,omitempty"`
	Host           string `json:"host,omitempty"`
	IP             string `json:"ip,omitempty"`
	Port           string `json:"port,omitempty"`
	VLAN           string `json:"vlan,omitempty"`
	Status         string `json:"status,omitempty"`
	LastSeen       int64  `json:"last_seen,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
	ScannerIP      string `json:"scanner_ip,omitempty"`
	Vendor         string `json:"vendor,omitempty"`
	ScannerStatus  string `json:"scanner_status,omitempty"`
	ScannerDetected bool `json:"scanner_detected,omitempty"`
	KnownHost      int    `json:"known_host,omitempty"`
	Note           string `json:"note,omitempty"`
	FirstSeen      string `json:"first_seen,omitempty"`
	LastSeenOnline string `json:"last_seen_online,omitempty"`
	LastUpdated    string `json:"last_updated,omitempty"`
}

func Load(path string) (*Config, error) {
	c := &Config{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	c.SetDefaults()
	return c, nil
}

func (c *Config) SetDefaults() {
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = 30
	}
	if c.MACRefreshMultiplier <= 0 {
		c.MACRefreshMultiplier = 5
	}
	if c.MaxRequestRetries <= 0 {
		c.MaxRequestRetries = 5
	}
	if len(c.EnabledColumns) == 0 {
		c.EnabledColumns = []string{"port", "status", "speed", "packets", "bytes", "info", "notes"}
	}
	if c.GridColumns == "" {
		c.GridColumns = "auto"
	}
	if c.Clients == nil {
		c.Clients = make(map[string]ClientEntry)
	}
	if c.Notes == nil {
		c.Notes = make(map[string]string)
	}
	if c.ColumnWidths == nil {
		c.ColumnWidths = make(map[string]int)
	}
	if c.MapPositions == nil {
		c.MapPositions = make(map[string]MapPosition)
	}
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

func (c *Config) ActiveSwitches() []SwitchConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var active []SwitchConfig
	for _, sw := range c.Switches {
		if sw.Enabled {
			active = append(active, sw)
		}
	}
	return active
}
