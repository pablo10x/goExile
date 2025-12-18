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

	// 2. Security check: Prevent blocking critical IPs
	if err := ValidateIP(ipStr); err != nil {
		log.Printf("⚠️ Refusing to block IP %s: %v", ipStr, err)
		return fmt.Errorf("cannot block IP %s: %w", ipStr, err)
	}

	// 3. Additional safety: Don't block private network ranges in production
	// (they might be legitimate internal traffic)
	if parsedIP.IsPrivate() {
		log.Printf("⚠️ Warning: Blocking private IP %s - verify this is intentional", ipStr)
	}

	log.Printf("🔥 BLOCKING IP via UFW: %s", ipStr)

	// 4. Execute UFW command
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
