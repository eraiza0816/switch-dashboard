package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultDeviceTypesYAML = `laptop:
  label: Laptop
  icon: mdi:laptop
smartphone:
  label: Smartphone
  icon: mdi:cellphone
server:
  label: Server
  icon: mdi:server
pc_desktop:
  label: PC Desktop
  icon: mdi:desktop-tower-monitor
nas:
  label: NAS
  icon: mdi:nas
ipcam:
  label: IpCam
  icon: mdi:cctv
tv:
  label: TV
  icon: mdi:television
nvr:
  label: NVR
  icon: mdi:video
smart_switch:
  label: Smart Switch
  icon: mdi:power-socket-eu
smart_plug:
  label: Smart Plug
  icon: mdi:power-plug
sensore:
  label: Sensore
  icon: mdi:access-point
audiovideo:
  label: AudioVideo
  icon: mdi:speaker
vacuum_robot:
  label: Vacuum Robot
  icon: mdi:robot-vacuum
air_conditioner:
  label: Air Conditioner
  icon: mdi:air-conditioner
dehumidifier:
  label: Dehumidifier
  icon: mdi:air-humidifier
three_d_printer:
  label: 3D Printer
  icon: mdi:printer-3d
dryer:
  label: Dryer
  icon: mdi:tumble-dryer
router:
  label: Router
  icon: mdi:router-wireless
repeater:
  label: Repeater
  icon: mdi:access-point
switch:
  label: Switch
  icon: mdi:switch
internet:
  label: Internet
  icon: mdi:earth
unmanaged_switch:
  label: Unmanaged Switch
  icon: mdi:switch
`

type deviceTypeDef struct {
	Label string `json:"label" yaml:"label"`
	Icon  string `json:"icon,omitempty" yaml:"icon,omitempty"`
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
}

func (s *Server) deviceTypesPath() string {
	if s.DeviceTypesPath != "" {
		return s.DeviceTypesPath
	}
	return filepath.Join(s.DataDir, "device_types.yaml")
}

// handleAPIGetDeviceTypes returns the parsed device type definitions as JSON.
func (s *Server) handleAPIGetDeviceTypes(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(s.deviceTypesPath())
	if err != nil || len(content) == 0 {
		content = []byte(defaultDeviceTypesYAML)
	}
	var parsed map[string]deviceTypeDef
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parsed)
}

// handleAPIDeviceTypesRaw serves and updates the raw device_types YAML file.
func (s *Server) handleAPIDeviceTypesRaw(w http.ResponseWriter, r *http.Request) {
	path := s.deviceTypesPath()

	if r.Method == http.MethodPost {
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		if len(req.Content) == 0 {
			http.Error(w, `{"error":"content cannot be empty"}`, http.StatusBadRequest)
			return
		}
		var parsed map[string]deviceTypeDef
		if err := yaml.Unmarshal([]byte(req.Content), &parsed); err != nil {
			http.Error(w, `{"error":"invalid yaml"}`, http.StatusBadRequest)
			return
		}
		for k, v := range parsed {
			if v.Label == "" {
				http.Error(w, `{"error":"entry '`+k+`' must contain a 'label' key"}`, http.StatusBadRequest)
				return
			}
		}
		if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		content = []byte(defaultDeviceTypesYAML)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": string(content)})
}
