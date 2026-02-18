package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"lighthouse/internal/network"
	"lighthouse/internal/scanner"
	"lighthouse/internal/storage"
)

var db *storage.DB

func main() {
	// Initialize database
	var err error
	db, err = storage.NewDB("data/lighthouse.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	os.MkdirAll("data", 0755)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "lighthouse",
	Short: "Network device discovery tool",
	Long:  "Scan your local network and track devices",
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(networksCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan [network]",
	Short: "Scan network for devices",
	Long:  "Scan a network using nmap. Example: lighthouse scan 192.168.1.0/24\nIf no network is specified, auto-detects your primary network.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var networkCIDR string

		// Get network from args or auto-detect
		if len(args) > 0 {
			networkCIDR = args[0]
		} else {
			// Auto-detect primary network
			primaryNet, err := network.GetPrimaryNetwork()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to detect network: %v\n", err)
				fmt.Println("Please specify network manually: lighthouse scan 192.168.1.0/24")
				os.Exit(1)
			}
			networkCIDR = primaryNet.CIDR
			fmt.Printf("Auto-detected network: %s (interface: %s, your IP: %s)\n",
				primaryNet.CIDR, primaryNet.Interface, primaryNet.IP)
		}

		fmt.Printf("Scanning network %s...\n", networkCIDR)
		fmt.Println("   (This may take 30-60 seconds)")
		fmt.Println()

		// Create scanner and scan
		s := scanner.NewScanner()
		devices, err := s.ScanNetwork(networkCIDR)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d device(s)\n\n", len(devices))

		// Save each device
		saved := 0
		for _, device := range devices {
			err := db.SaveDevice(device.IP, device.MAC, device.Hostname, device.Vendor)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", device.IP, err)
			} else {
				saved++
				fmt.Printf("%-15s", device.IP)
				if device.MAC != "" {
					fmt.Printf(" | MAC: %-17s", device.MAC)
				}
				if device.Hostname != "" {
					fmt.Printf(" | %s", device.Hostname)
				}
				if device.Vendor != "" {
					fmt.Printf(" [%s]", device.Vendor)
				}
				fmt.Println()
			}
		}

		fmt.Printf("\nSaved %d devices to database\n", saved)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List discovered devices",
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := db.GetAllDevices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get devices: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("No devices found. Run 'lighthouse scan' first.")
			return
		}

		fmt.Printf("Found %d device(s):\n\n", len(devices))
		fmt.Println("IP Address       MAC Address        Vendor               Hostname")
		fmt.Println("───────────────────────────────────────────────────────────────────────────────")

		for _, d := range devices {
			fmt.Printf("%-15s  %-17s  %-20s %s\n",
				d.IP,
				formatString(d.MAC, 17),
				truncate(d.Vendor, 20),
				d.Hostname)
		}
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start web dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Lighthouse web server...")
		fmt.Println("Open your browser: http://localhost:8080")
		fmt.Println()
		startWebServer()
	},
}

var networksCmd = &cobra.Command{
	Use:   "networks",
	Short: "List detected networks",
	Run: func(cmd *cobra.Command, args []string) {
		networks, err := network.DetectNetworks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to detect networks: %v\n", err)
			os.Exit(1)
		}

		if len(networks) == 0 {
			fmt.Println("No networks detected")
			return
		}

		fmt.Printf("Detected %d network(s):\n\n", len(networks))
		fmt.Println("Interface    IP Address       Network CIDR")
		fmt.Println("──────────────────────────────────────────────────────")

		for _, net := range networks {
			fmt.Printf("%-12s %-15s %s\n", net.Interface, net.IP, net.CIDR)
		}

		fmt.Println("\nTo scan a specific network:")
		fmt.Println("  sudo ./lighthouse scan <CIDR>")
	},
}

func startWebServer() {
	// Serve static HTML file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	// API endpoint for devices
	http.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n=== /api/devices called ===")

		// Mark stale devices as offline (not seen in last 10 minutes)
		fmt.Println("→ Marking stale devices offline...")
		if err := db.MarkStaleDevicesOffline(10); err != nil {
			fmt.Printf("  ⚠️  Warning: %v\n", err)
		}

		fmt.Println("→ Fetching all devices from database...")
		devices, err := db.GetAllDevices()
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("  ✓ Found %d devices\n", len(devices))
		for i, d := range devices {
			status := "🟢 online"
			if !d.IsOnline {
				status = "⚪ offline"
			}
			fmt.Printf("    [%d] %s - %s %s\n", i+1, d.IP, d.Hostname, status)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(devices)
		fmt.Println("→ Response sent\n")
	})

	// API endpoint for stats
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n=== /api/stats called ===")

		// Mark stale devices offline first
		if err := db.MarkStaleDevicesOffline(10); err != nil {
			fmt.Printf("  ⚠️  Warning: %v\n", err)
		}

		fmt.Println("→ Getting statistics...")
		stats, err := db.GetStats()
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("  Total: %d | Online: %d | Offline: %d\n",
			stats.TotalDevices, stats.OnlineDevices, stats.OfflineDevices)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	// API endpoint for networks
	http.HandleFunc("/api/networks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n=== /api/networks called ===")

		fmt.Println("→ Detecting networks...")
		networks, err := network.DetectNetworks()
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("  ✓ Found %d network(s)\n", len(networks))
		for _, net := range networks {
			fmt.Printf("    - %s (%s) - %s\n", net.Interface, net.CIDR, net.IP)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(networks)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func formatString(s string, width int) string {
	if s == "" {
		return "-"
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	if s == "" {
		return "-"
	}
	return s
}
