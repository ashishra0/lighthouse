package network

import (
	"fmt"
	"net"
	"strings"
)

type Network struct {
	CIDR      string
	Interface string
	IP        string
}

func DetectNetworks() ([]Network, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	var networks []Network
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			if ipNet.IP.To4() == nil {
				continue
			}

			if skipInterface(iface.Name) {
				continue
			}

			ones, _ := ipNet.Mask.Size()
			networkIP := ipNet.IP.Mask(ipNet.Mask)
			cidr := fmt.Sprintf("%s/%d", networkIP.String(), ones)

			network := Network{
				CIDR:      cidr,
				Interface: iface.Name,
				IP:        ipNet.IP.String(),
			}
			networks = append(networks, network)
		}
	}

	return networks, nil
}

func GetPrimaryNetwork() (Network, error) {
	networks, err := DetectNetworks()
	if err != nil {
		return Network{}, err
	}

	if len(networks) == 0 {
		return Network{}, fmt.Errorf("no networks detected")
	}

	for _, net := range networks {
		if net.Interface == "en0" {
			return net, nil
		}
	}

	for _, net := range networks {
		if strings.HasPrefix(net.Interface, "wlan") {
			return net, nil
		}
	}

	for _, net := range networks {
		if strings.HasPrefix(net.Interface, "eth") || strings.HasPrefix(net.Interface, "en") {
			return net, nil
		}
	}

	return networks[0], nil
}

func skipInterface(name string) bool {
	skipPrefixes := []string{
		"awdl",      // Apple Wireless Direct Link
		"llw",       // Low Latency WLAN
		"utun",      // VPN tunnels
		"bridge",    // Bridges
		"docker",    // Docker
		"veth",      // Virtual Ethernet
		"virbr",     // Virtual Bridge
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
