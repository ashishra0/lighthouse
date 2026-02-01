# Lighthouse

A network device discovery tool for scanning and tracking devices on your local network.

## Overview

Lighthouse scans your local network using nmap to discover connected devices. It provides both a command-line interface and a web dashboard for viewing and managing discovered devices.

## Features

- Network scanning using nmap
- Device discovery and tracking
- SQLite database for persistence
- Command-line interface
- Web dashboard for visualization
- First seen / last seen timestamps
- MAC address and vendor identification (with sudo)

## Requirements

- Go 1.19 or higher
- Ruby (system default is fine)
- nmap

### Installing nmap

**macOS:**
```bash
brew install nmap
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install nmap
```

## Installation

### From Source

```bash
git clone <your-repo-url>
cd lighthouse
go build -o lighthouse ./cmd/lighthouse
```

## Usage

### Scanning Networks

Scan your network to discover devices:

```bash
# Scan default network (192.168.68.0/22)
./lighthouse scan

# Scan specific network
./lighthouse scan 192.168.1.0/24

# Scan with sudo to get MAC addresses and vendor info
sudo ./lighthouse scan 192.168.1.0/24
```

### Listing Devices

View discovered devices in your terminal:

```bash
./lighthouse list
```

### Web Dashboard

Start the web server and view devices in your browser:

```bash
./lighthouse serve
```

Then open http://localhost:8080 in your browser.

## Why sudo?

Running scans with sudo provides additional information:

- MAC addresses of discovered devices
- Vendor identification from MAC address
- More accurate device information

Without sudo, you'll still see IP addresses and hostnames, but MAC addresses and vendors will be missing.

## Project Structure

```
lighthouse/
├── cmd/lighthouse/          # Main application entry point
├── internal/
│   ├── scanner/            # Network scanning logic
│   └── storage/            # SQLite database operations
├── scripts/
│   └── scanner.rb          # Ruby wrapper for nmap
├── web/
│   └── index.html          # Web dashboard
└── data/
    └── lighthouse.db       # SQLite database (created on first run)
```

## How It Works

1. **Ruby Scanner**: Executes nmap with XML output format
2. **XML Parsing**: Extracts device information (IP, MAC, hostname, vendor)
3. **JSON Conversion**: Converts data to JSON for Go to consume
4. **Go Processing**: Saves devices to SQLite database
5. **Web API**: Serves device data via HTTP API
6. **Dashboard**: Displays devices in browser with auto-refresh

## Database Schema

The SQLite database stores:

- IP address (unique identifier)
- MAC address
- Hostname
- Vendor (from MAC OUI lookup)
- First seen timestamp
- Last seen timestamp

Devices are automatically updated when rescanned (UPSERT logic).

## Learning Resources

This project demonstrates:

- **Networking**: IP addresses, CIDR notation, subnets, MAC addresses, ARP
- **Go**: CLI with Cobra, process execution, SQLite integration, HTTP server
- **Ruby**: System commands, XML parsing, JSON generation
- **SQLite**: Schema design, UPSERT operations, queries

## Networking Concepts

### IP Addresses
Logical addresses assigned to devices on a network (e.g., 192.168.1.100)

### CIDR Notation
Compact way to represent network ranges (e.g., 192.168.1.0/24 = 256 addresses)

### MAC Addresses
Physical hardware addresses burned into network cards (e.g., b8:27:eb:75:0e:74)

### ARP Table
Maps IP addresses to MAC addresses for local network communication

### nmap
Network scanning tool that actively probes IP addresses to discover devices

## Technical Details

### nmap Flags Used

- `-sn`: Host discovery only (no port scan) - faster and less intrusive
- `-T4`: Aggressive timing template for faster scanning
- `-oX -`: Output XML to stdout for parsing

### Data Flow

```
User Command
    ↓
Go CLI (Cobra)
    ↓
Ruby Scanner Script
    ↓
nmap (System)
    ↓
XML Output
    ↓
Ruby Parser
    ↓
JSON Output
    ↓
Go Scanner Wrapper
    ↓
SQLite Database
    ↓
Web API / CLI Output
```

## Examples

### Scan a /24 network (256 addresses)
```bash
sudo ./lighthouse scan 192.168.1.0/24
```

### Scan a /22 network (1024 addresses)
```bash
sudo ./lighthouse scan 192.168.0.0/22
```

### Scan just a small range for testing
```bash
./lighthouse scan 192.168.1.0/28  # Only 16 addresses
```

## Troubleshooting

### "nmap: command not found"
Install nmap using your package manager (see Requirements section)

### No MAC addresses showing
Run the scan with sudo: `sudo ./lighthouse scan <network>`

### Database errors
Ensure the `data/` directory exists and is writable

### Web server won't start
Check if port 8080 is already in use

## Development

### Build
```bash
go build -o lighthouse ./cmd/lighthouse
```

### Test Ruby scanner directly
```bash
ruby scripts/scanner.rb 192.168.1.0/24
```

### Test Go modules
```bash
go test ./...
```

## License

MIT

## Author

Built as a learning project for understanding networking concepts, Go development, and network scanning tools.
