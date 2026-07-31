package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   "Switch Dashboard API",
			"version": Version,
			"description": "REST API for the RTLPlayground Switch Dashboard.\n\n" +
				"The dashboard connects to RTLPlayground-powered switches via HTTP and provides " +
				"real-time port status, SFP diagnostics, MAC forwarding tables, bandwidth history, " +
				"and management actions (backup, reboot, firmware upgrade).",
		},
		"servers": []map[string]any{
			{"url": "/", "description": "Local dashboard server"},
		},
		"paths": map[string]any{
			"/api/switches": map[string]any{
				"get": map[string]any{
					"summary":     "List all switches with live status",
					"description": "Returns port states, counters, MAC table, SFP info, EEE, VLAN, LAG for all configured switches.",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Array of switch data objects",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type":  "array",
										"items": map[string]any{"$ref": "#/components/schemas/SwitchData"},
									},
								},
							},
						},
					},
				},
			},
			"/api/switches/{ip}/sfp": map[string]any{
				"get": map[string]any{
					"summary":     "SFP+ DDMI diagnostics",
					"description": "Returns temperature, voltage, bias current, TX/RX power for SFP+ modules.",
					"parameters": []map[string]any{
						{"name": "ip", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "SFP diagnostic data"},
					},
				},
			},
			"/api/switches/{ip}/transceiver": map[string]any{
				"get": map[string]any{
					"summary":     "SFP EEPROM information",
					"parameters": []map[string]any{
						{"name": "ip", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "SFP module information"},
					},
				},
			},
			"/api/switches/{ip}/refresh_mac": map[string]any{
				"post": map[string]any{
					"summary":     "Refresh MAC forwarding table",
					"parameters": []map[string]any{
						{"name": "ip", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "MAC table refreshed"},
					},
				},
			},
			"/api/switches/{ip}/backup": map[string]any{
				"post": map[string]any{
					"summary":     "Download switch configuration backup",
					"parameters": []map[string]any{
						{"name": "ip", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Backup saved to server"},
					},
				},
			},
			"/api/switches/{ip}/reboot": map[string]any{
				"post": map[string]any{
					"summary":     "Reboot switch",
					"parameters": []map[string]any{
						{"name": "ip", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Reboot command sent"},
					},
				},
			},
			"/api/speeds": map[string]any{
				"get": map[string]any{
					"summary":     "Real-time port speeds",
					"description": "Returns per-port TX/RX speeds in bits per second.",
					"responses": map[string]any{
						"200": map[string]any{"description": "Speed data"},
					},
				},
			},
			"/api/history": map[string]any{
				"get": map[string]any{
					"summary":     "Bandwidth history",
					"description": "Returns timestamped TX/RX data points for a port. Ranges: live, 1h, 24h.",
					"parameters": []map[string]any{
						{"name": "ip", "in": "query", "schema": map[string]any{"type": "string"}},
						{"name": "port", "in": "query", "schema": map[string]any{"type": "string"}},
						{"name": "range", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"live", "1h", "24h"}}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "History data"},
					},
				},
			},
			"/api/topology": map[string]any{
				"get": map[string]any{
					"summary":     "Network topology graph",
					"description": "Returns nodes and links derived from MAC forwarding tables.",
					"responses": map[string]any{
						"200": map[string]any{"description": "Topology graph"},
					},
				},
			},
			"/api/settings": map[string]any{
				"get": map[string]any{
					"summary": "Get UI settings",
					"responses": map[string]any{
						"200": map[string]any{"description": "Settings object"},
					},
				},
				"post": map[string]any{
					"summary": "Update UI settings",
					"responses": map[string]any{
						"200": map[string]any{"description": "Settings updated"},
					},
				},
			},
			"/api/notes": map[string]any{
				"post": map[string]any{
					"summary": "Save port note",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"key":  map[string]any{"type": "string"},
										"note": map[string]any{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Note saved"},
					},
				},
			},
			"/api/reset": map[string]any{
				"post": map[string]any{
					"summary":     "Reset cumulative counters",
					"responses": map[string]any{
						"200": map[string]any{"description": "Counters reset"},
					},
				},
			},
			"/api/logs": map[string]any{
				"get": map[string]any{
					"summary":     "Get server log lines",
					"responses": map[string]any{
						"200": map[string]any{"description": "Log lines array"},
					},
				},
			},
			"/api/logs/level": map[string]any{
				"post": map[string]any{
					"summary": "Change log level",
					"responses": map[string]any{
						"200": map[string]any{"description": "Log level changed"},
					},
				},
			},
			"/api/logs/clear": map[string]any{
				"post": map[string]any{
					"summary":     "Clear server logs",
					"responses": map[string]any{
						"200": map[string]any{"description": "Logs cleared"},
					},
				},
			},
			"/api/backups": map[string]any{
				"get": map[string]any{
					"summary":     "List configuration backups",
					"responses": map[string]any{
						"200": map[string]any{"description": "Backup list"},
					},
				},
			},
			"/api/backups/{filename}/download": map[string]any{
				"get": map[string]any{
					"summary":     "Download a backup file",
					"parameters": []map[string]any{
						{"name": "filename", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Binary file"},
					},
				},
			},
			"/api/backups/{filename}": map[string]any{
				"delete": map[string]any{
					"summary":     "Delete a backup file",
					"parameters": []map[string]any{
						{"name": "filename", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Backup deleted"},
					},
				},
			},
			"/api/vendors": map[string]any{
				"get": map[string]any{
					"summary":     "Get custom MAC vendor mappings",
					"responses": map[string]any{
						"200": map[string]any{"description": "Vendor mappings"},
					},
				},
				"post": map[string]any{
					"summary":     "Save custom MAC vendor mappings",
					"responses": map[string]any{
						"200": map[string]any{"description": "Vendor mappings saved"},
					},
				},
			},
			"/api/vendors/update_oui": map[string]any{
				"post": map[string]any{
					"summary":     "Download IEEE OUI database",
					"responses": map[string]any{
						"200": map[string]any{"description": "OUI database updated"},
					},
				},
			},
			"/api/clients/update_host": map[string]any{
				"post": map[string]any{
					"summary": "Override client hostname in topology",
					"description": "Persists a MAC-to-hostname mapping so the topology graph shows a custom nickname for a client device.",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"mac":  map[string]any{"type": "string"},
										"host": map[string]any{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Hostname override saved"},
					},
				},
			},
			"/api/layout_positions": map[string]any{
				"get": map[string]any{
					"summary":     "Get topology node layout positions",
					"description": "Returns per-node x/y positions saved by the interactive map editor.",
					"responses": map[string]any{
						"200": map[string]any{"description": "Position map"},
					},
				},
				"post": map[string]any{
					"summary": "Save topology node layout positions",
					"description": "Persists the current interactive map node positions to the server.",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"additionalProperties": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"x": map[string]any{"type": "number"},
											"y": map[string]any{"type": "number"},
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Positions saved"},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"SwitchData": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":     map[string]any{"type": "string"},
						"ip":       map[string]any{"type": "string"},
						"model":    map[string]any{"type": "string"},
						"mac":      map[string]any{"type": "string"},
						"hostname": map[string]any{"type": "string"},
						"firmware": map[string]any{"type": "string"},
						"uptime":   map[string]any{"type": "string"},
						"status":   map[string]any{"type": "string"},
						"ports": map[string]any{
							"type":  "array",
							"items": map[string]any{"$ref": "#/components/schemas/PortState"},
						},
					},
				},
				"PortState": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"port":    map[string]any{"type": "string"},
						"status":  map[string]any{"type": "string"},
						"link":    map[string]any{"type": "string"},
						"speed":   map[string]any{"type": "string"},
						"tx_bytes":  map[string]any{"type": "integer"},
						"rx_bytes":  map[string]any{"type": "integer"},
						"cum_tx":    map[string]any{"type": "integer"},
						"cum_rx":    map[string]any{"type": "integer"},
						"speed_tx_bps": map[string]any{"type": "integer"},
						"speed_rx_bps": map[string]any{"type": "integer"},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}
