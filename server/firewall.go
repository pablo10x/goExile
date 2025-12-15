package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

// BlockIP blocks an IP address using UFW (Uncomplicated Firewall).
// It requires the application to have root privileges (or passwordless sudo).
func BlockIP(ip string) error {
	// 1. Validate IP to prevent command injection
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	ipStr := parsedIP.String()

	// 2. Check if already blocked (optional, optimization)
	// We can skip this and let UFW handle it, but checking prevents log spam.
	// For simplicity and robustness, we'll just try to deny.

	log.Printf("🔥 BLOCKING IP via UFW: %s", ipStr)

	// 3. Execute UFW command
	// Command: ufw deny from <ip> to any
	// We use "sudo" just in case the binary isn't running as root, 
	// but the user environment must allow this.
	// If running as root (docker/systemd), sudo might not be needed or installed.
	// We'll try running "ufw" directly first.
	
	cmd := exec.Command("ufw", "deny", "from", ipStr, "to", "any")
	output, err := cmd.CombinedOutput()
	
	// If failed, maybe try with sudo?
	if err != nil {
		// Check if "command not found" vs "permission denied"
		// If ufw not in path, we can't do much.
		log.Printf("UFW direct execution failed: %v. Output: %s", err, string(output))
		
		if strings.Contains(string(output), "permission") || strings.Contains(string(output), "root") {
			log.Println("Attempting with sudo...")
			cmd = exec.Command("sudo", "ufw", "deny", "from", ipStr, "to", "any")
			output, err = cmd.CombinedOutput()
		}
	}

	if err != nil {
		return fmt.Errorf("failed to block IP %s: %v. Output: %s", ipStr, err, string(output))
	}

	log.Printf("✅ Successfully blocked IP %s", ipStr)
	return nil
}
