# Switch Dashboard

A real-time monitoring dashboard for switches running **RTLPlayground** firmware (RTL8372/RTL8373 based 2.5GbE switches). Written in Go, served as a single binary.

![Dashboard](https://raw.githubusercontent.com/byte4geek/switch-dashboard/refs/heads/main/images/dashboard.png)

## Features

- **Real-time port status**: link state, speed, duplex, TX/RX counters
- **Bandwidth charts**: Live / 1-hour / 24-hour rolling history (Chart.js)
- **SFP+ DDMI**: temperature, voltage, bias current, TX/RX power telemetry
- **MAC forwarding table**: searchable, filterable, with vendor resolution
- **EEE / VLAN / LAG / MTU / Mirror / Bandwidth control**: status display
- **Config backup & restore**: download/upload configuration via web UI
- **Firmware update**: upload new firmware through the web interface
- **Remote reboot**: safe device reboot with confirmation
- **Network topology map**: visualize connected clients per port
- **Dark glassmorphic UI**: custom typography, frosted-glass components

## Supported Hardware

RTLPlayground firmware runs on 20+ device models from Ampcom, Davuaz, FOXNEO, Hisource, Horaco, KeepLink, LIANGUO, Mokerlink, Sodola, Steamemo, TrendNet, XikeStor, and others. See the [RTLPlayground supported devices](https://github.com/logicog/RTLPlayground/blob/main/doc/supported_devices.md) for the full list.

## Quick Start

```bash
# Build the binary
go build -o switch-dashboard ./cmd/switch-dashboard/

# Prepare config.json
cat > config.json << 'EOF'
{
  "title": "My Dashboard",
  "refresh_interval": 10,
  "switches": [
    {
      "name": "Core Switch",
      "ip": "192.168.10.247",
      "password": "1234",
      "model": "RTLPlayground",
      "port_count": 9,
      "enabled": true
    }
  ]
}
EOF

# Run
./switch-dashboard
```

Open http://localhost:8080

## Docker

```bash
docker build -t switch-dashboard .
docker run -d --name switch-dashboard -p 8080:8080 \
  -v $(pwd)/config.json:/config.json \
  switch-dashboard
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/switches` | Live port status, counters, MAC table, SFP, EEE, VLAN, LAG |
| GET | `/api/switches/:ip/sfp` | SFP+ DDMI diagnostics |
| GET | `/api/topology` | Network graph from MAC forwarding table |
| GET | `/api/speeds` | Real-time per-port bandwidth (bps) |
| GET | `/api/history?ip=...&port=...&range=live\|1h\|24h` | Bandwidth history |
| POST | `/api/notes` | Save port annotation |
| POST | `/api/reset` | Reset cumulative counters |
| POST | `/api/switches/:ip/backup` | Download config backup |
| POST | `/api/switches/:ip/reboot` | Reboot switch |
| POST | `/api/switches/:ip/upload` | Firmware upgrade |
| GET/POST | `/api/settings` | UI preferences |
| GET/POST | `/api/vendors` | Custom MAC vendor mappings |
| POST | `/api/vendors/update_oui` | Download IEEE OUI database |

## Architecture

```
User Browser (Vanilla JS + Chart.js)
        ↕  HTTP JSON API
Go Server (chi router, single binary)
        ↕  HTTP JSON API
RTLPlayground Switch (uIP embedded webserver)
```

- **Backend**: Go 1.22+, chi router, zero external dependencies beyond chi
- **Frontend**: Vanilla ES6 JS, Chart.js (CDN), inline CSS dark theme
- **Scraper**: Direct HTTP calls to RTLPlayground JSON endpoints (no HTML parsing)

## Project Structure

```
├── cmd/switch-dashboard/     # Entry point
├── internal/
│   ├── server/               # HTTP handlers, cache, templates
│   ├── rtlplayground/        # Switch HTTP client + JSON types
│   ├── poller/               # Background polling, counters, history
│   ├── config/               # config.json management
│   ├── store/                # Generic JSON file persistence
│   └── vendor/               # MAC OUI resolution
├── templates/                # Go html/templates
├── static/                   # Static assets
├── Dockerfile                # Multi-stage, scratch base (~8MB)
└── config.json               # Switch configuration
```

## License

MIT
