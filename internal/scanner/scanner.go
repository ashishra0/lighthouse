package scanner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type Device struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
}

type Scanner struct {
	scriptPath string
}

func NewScanner() *Scanner {
	scriptPath, _ := filepath.Abs("scripts/scanner.rb")
	return &Scanner{scriptPath: scriptPath}
}

func (s *Scanner) ScanNetwork(cidr string) ([]Device, error) {
	cmd := exec.Command("ruby", s.scriptPath, cidr)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w\nOutput: %s", err, string(output))
	}

	var devices []Device
	if err := json.Unmarshal(output, &devices); err != nil {
		return nil, fmt.Errorf("failed to parse scan results: %w", err)
	}

	return devices, nil
}
