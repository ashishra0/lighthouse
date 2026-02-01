# Lighthouse

Network device discovery tool for scanning and tracking devices on your local network.

## Quick Start

### Requirements

- Go 1.19+
- Ruby
- nmap: `brew install nmap`

### Installation

```bash
git clone <repo-url>
cd lighthouse
go build -o lighthouse ./cmd/lighthouse
```

### Usage

```bash
# Scan your network (auto-detects network)
sudo ./lighthouse scan

# Scan specific network
sudo ./lighthouse scan 192.168.1.0/24

./lighthouse list

./lighthouse serve

./lighthouse networks
```

## Why sudo?

Running with sudo allows nmap to read the ARP table, which provides:
- MAC addresses
- Device vendor identification
- More accurate device information

Without sudo, you'll see IP addresses and hostnames, but MAC addresses will be missing.

## Web Dashboard

The dashboard shows:
- Device counts (total, online, offline)
- Detected networks with status
- Device list with online/offline status
- Auto-refresh every 30 seconds

Devices are marked offline if not seen in the last 10 minutes.

```

## How It Works

1. Ruby script wraps nmap and outputs JSON
2. Go executes Ruby script and parses results
3. Devices saved to SQLite database
4. Web dashboard fetches data via REST API
5. Auto-refresh keeps data current

## API Endpoints

- `GET /api/devices` - List all devices
- `GET /api/stats` - Device statistics
- `GET /api/networks` - Detected networks

## Database

SQLite database stores:
- IP address, MAC address, hostname, vendor
- Online/offline status
- First seen and last seen timestamps

## Troubleshooting

**Database permissions error:**
```bash
sudo chown $USER data/lighthouse.db
# Or delete and rescan:
rm -rf data/lighthouse.db
```

**No MAC addresses:**
Run scan with sudo: `sudo ./lighthouse scan`

**Port 8080 in use:**
The web server uses port 8080 by default
