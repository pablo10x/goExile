package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// -- Caching --

var (
	BannedIPCache = make(map[string]bool)
	BanCacheMu    sync.RWMutex
)

// StartRedEyeBackground initializes background tasks for RedEye.
func StartRedEyeBackground(db *sqlx.DB) {
	if db == nil {
		return
	}
	
	// Initial load
	RefreshBanCache(db)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		cleanupTicker := time.NewTicker(10 * time.Minute)
		anomalyTicker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		defer cleanupTicker.Stop()
		defer anomalyTicker.Stop()

		for {
			select {
			case <-ticker.C:
				RefreshBanCache(db)
			case <-cleanupTicker.C:
				cleanupLimiters()
			case <-anomalyTicker.C:
				RunAnomalyDetection(db)
			}
		}
	}()
}

func RefreshBanCache(db *sqlx.DB) {
	ips, err := GetBannedIPList(db)
	if err != nil {
		log.Printf("RedEye: Failed to refresh ban cache: %v", err)
		return
	}

	newCache := make(map[string]bool)
	for _, ip := range ips {
		newCache[ip] = true
	}

	BanCacheMu.Lock()
	BannedIPCache = newCache
	BanCacheMu.Unlock()
}

// RunAnomalyDetection scans for high-frequency blocks and auto-bans IPs.
func RunAnomalyDetection(db *sqlx.DB) {
	// Check if enabled
	cfg, err := GetConfigByKey(db, "redeye.auto_ban_enabled")
	if err != nil || cfg == nil || cfg.Value != "true" {
		return
	}

	// Get threshold
	threshold := 100
	cfgThresh, err := GetConfigByKey(db, "redeye.auto_ban_threshold")
	if err == nil && cfgThresh != nil {
		if val, err := strconv.Atoi(cfgThresh.Value); err == nil && val > 0 {
			threshold = val
		}
	}

	// Scan logs from last minute
	lastMinute := time.Now().Add(-1 * time.Minute).Unix()
	query := `SELECT source_ip, COUNT(*) as count 
              FROM redeye_logs 
              WHERE timestamp > $1 AND (action = 'DENY' OR action = 'RATE_LIMIT_HIT') 
              GROUP BY source_ip 
              HAVING count > $2`
	
	rows, err := db.Query(query, lastMinute, threshold)
	if err != nil {
		log.Printf("RedEye Anomaly: Scan failed: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ip string
		var count int
		if err := rows.Scan(&ip, &count); err != nil {
			continue
		}

		// Check if already banned (via cache to save DB hits)
		BanCacheMu.RLock()
		isBanned := BannedIPCache[ip]
		BanCacheMu.RUnlock()

		if isBanned {
			continue
		}

		log.Printf("RedEye Anomaly: Auto-banning IP %s (Events: %d)", ip, count)

		// Ban logic
		rep := &RedEyeIPReputation{
			IP:              ip,
			ReputationScore: 100,
			TotalEvents:     count,
			LastSeen:        time.Now().UTC(),
			IsBanned:        true,
			BanReason:       fmt.Sprintf("Auto-ban: High frequency blocks (%d/min)", count),
		}
		
		// Use a transaction or just sequenced calls
		if err := UpdateIPReputation(db, rep); err != nil {
			log.Printf("RedEye Anomaly: Failed to update reputation for %s: %v", ip, err)
			continue
		}

		// Create Rule
		rule := &RedEyeRule{
			Name:      fmt.Sprintf("Auto-Ban %s", ip),
			CIDR:      ip,
			Port:      "*",
			Protocol:  "ANY",
			Action:    "DENY",
			Enabled:   true,
			CreatedAt: time.Now().UTC(),
		}
		if _, err := CreateRedEyeRule(db, rule); err != nil {
			log.Printf("RedEye Anomaly: Failed to create ban rule for %s: %v", ip, err)
		}
	}
	
	// Refresh cache if we banned anyone (checking if rows iterated? simplified: just refresh)
	RefreshBanCache(db)
}

// RedEyeMiddleware serves as the guardian for the backend.
// It handles IP blocking, Rule enforcement (Allow/Deny), and Rate Limiting.
func RedEyeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dbConn == nil {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := GetClientIP(r)

		// Bypass for localhost
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Check IP Reputation / Ban List (Cached)
		BanCacheMu.RLock()
		isBanned := BannedIPCache[clientIP]
		BanCacheMu.RUnlock()

		if isBanned {
			http.Error(w, "Access Denied (Banned)", http.StatusForbidden)
			return
		}

		// 2. Rules
		rules, err := GetRedEyeRules(dbConn)
		if err != nil {
			log.Printf("RedEye error fetching rules: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		blocked := false
		rateLimited := false
		var matchedRule *RedEyeRule

		for i := range rules {
			rule := &rules[i]
			if !rule.Enabled {
				continue
			}

			if ipMatch(clientIP, rule.CIDR) {
				matchedRule = rule
				
				if rule.Action == "DENY" {
					blocked = true
					break
				}
				
				if rule.Action == "RATE_LIMIT" {
					if !checkRateLimit(clientIP, rule.RateLimit, rule.Burst) {
						rateLimited = true
						break
					}
					break 
				}

				if rule.Action == "ALLOW" {
					break // Explicit allow
				}
			}
		}

		// Log significant events
		if matchedRule != nil && (blocked || rateLimited) {
			action := matchedRule.Action
			if rateLimited {
				action = "RATE_LIMIT_HIT"
			}
			go func(ruleID int, act string) {
				l := &RedEyeLog{
					RuleID:    &ruleID,
					SourceIP:  clientIP,
					DestPort:  getPort(r),
					Protocol:  r.Method,
					Action:    act,
					Timestamp: time.Now().UTC(),
				}
				SaveRedEyeLog(dbConn, l)
			}(matchedRule.ID, action)
		}

		if blocked {
			http.Error(w, "Access Denied (RedEye)", http.StatusForbidden)
			return
		}

		if rateLimited {
			http.Error(w, "Rate Limit Exceeded (RedEye)", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// -- Rate Limiting Implementation --

type rateLimiter struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

var (
	limiters = make(map[string]*rateLimiter)
	limitMu  sync.RWMutex
)

// checkRateLimit returns true if request is allowed, false if limit exceeded.
func checkRateLimit(ip string, limit int, burst int) bool {
	if limit <= 0 {
		return true // No limit
	}
	if burst <= 0 {
		burst = limit // Default burst to limit
	}

	limitMu.RLock()
	lim, exists := limiters[ip]
	limitMu.RUnlock()

	if !exists {
		limitMu.Lock()
		// Double check
		if lim, exists = limiters[ip]; !exists {
			lim = &rateLimiter{
				tokens:     float64(burst),
				lastUpdate: time.Now(),
			}
			limiters[ip] = lim
		}
		limitMu.Unlock()
	}

	lim.mu.Lock()
	defer lim.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lim.lastUpdate).Seconds()
	
	// Refill tokens
	lim.tokens += elapsed * float64(limit)
	if lim.tokens > float64(burst) {
		lim.tokens = float64(burst)
	}
	lim.lastUpdate = now

	if lim.tokens >= 1 {
		lim.tokens--
		return true
	}

	return false
}

// cleanupLimiters cleans up old entries to prevent memory leak
func cleanupLimiters() {
	limitMu.Lock()
	for ip, lim := range limiters {
		if time.Since(lim.lastUpdate) > 10*time.Minute {
			delete(limiters, ip)
		}
	}
	limitMu.Unlock()
}

// -- Helpers --

// BlockIPSystem executes OS-level blocking (UFW).
func BlockIPSystem(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP: %s", ip)
	}
	ipStr := parsedIP.String()

	if err := ValidateIP(ipStr); err != nil {
		return err
	}

	// Windows support? 
	// For now, only Linux UFW is implemented.
	// We could add netsh/PowerShell for Windows but that's complex.
	// We'll keep UFW logic for Linux hosts.
	
	cmd := exec.Command("ufw", "deny", "from", ipStr, "to", "any")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Try sudo
		if strings.Contains(string(output), "root") || strings.Contains(string(output), "permission") {
			cmd = exec.Command("sudo", "ufw", "deny", "from", ipStr, "to", "any")
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("ufw failed: %v %s", err, string(out))
			}
		} else {
			return fmt.Errorf("ufw failed: %v %s", err, string(output))
		}
	}
	
	return nil
}

func ipMatch(ip, cidr string) bool {
	if ip == cidr {
		return true
	}
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	parsedIP := net.ParseIP(ip)
	return ipNet.Contains(parsedIP)
}

func getPort(r *http.Request) int {
	_, portStr, err := net.SplitHostPort(r.Host)
	if err != nil {
		if r.TLS != nil {
			return 443
		}
		return 80
	}
	p, _ := strconv.Atoi(portStr)
	return p
}
