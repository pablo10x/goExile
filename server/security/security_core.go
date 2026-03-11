package security

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/registry"
	"exile/server/sse"
	"exile/server/utils"

	"github.com/jmoiron/sqlx"
)

// =================================================================================
// SECURITY: ADVANCED THREAT INTELLIGENCE & FIREWALL
// =================================================================================

const (
	SignalTypeTraffic   = "TRAFFIC"
	SignalTypeBlock     = "BLOCK"
	SignalTypeRateLimit = "RATE_LIMIT"
	SignalTypeReport    = "REPORT"
	SignalTypeAuthFail  = "AUTH_FAIL"
)

var (
	// State Caches
	BannedIPCache = make(map[string]bool)
	RuleCache     = []models.SecurityRule{}
	BanCacheMu    sync.RWMutex

	// System Status
	SecurityActive = false
	SecurityError  = ""

	// High-Performance Event Bus
	signalChan = make(chan SecuritySignal, 10000)

	// Config Cache
	configMu       sync.RWMutex
	autoBanEnabled = true
	banThreshold   = 100
	systemMode     = "ENFORCEMENT" // "ENFORCEMENT" or "SIMULATION"
	strictMode     = false
	decayRate      = 10 // minutes
	whitelist      = make(map[string]bool)
	geoIPEnabled   = false
	allowedCountries = make(map[string]bool)

	// GeoIP Cache (In-Memory)
	geoIPCache = make(map[string]string) // IP -> ISO Code
	geoMu      sync.RWMutex

	// GeoIP Circuit Breaker
	geoErrorCount int
	geoLastAttempt time.Time
	geoTripped     bool

	// IP Reputation Tracker (In-Memory)
	ipScores = make(map[string]int)
	scoreMu  sync.RWMutex

	// Rate Limiters
	limiters = make(map[string]*rateLimiter)
	limitMu  sync.RWMutex

	// Lifecycle Management
	done = make(chan struct{})
	wg   sync.WaitGroup

	// Deduplication for ban operations
	banningIPs sync.Map
)

// SecuritySignal represents a distinct event in the system
type SecuritySignal struct {
	IP        string
	Type      string
	Severity  int // 0-100 impact on reputation
	Details   string
	Timestamp time.Time
	ReqInfo   *RequestInfo
}

type RequestInfo struct {
	Method string
	Path   string
	Port   int
}

// CheckSystemRequirements checks if ufw is installed and accessible.
func CheckSystemRequirements() error {
	path, err := exec.LookPath("ufw")
	if err != nil {
		return fmt.Errorf("ufw not installed")
	}
	_ = path
	cmd := exec.Command("ufw", "status")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("sudo", "-n", "ufw", "status")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("permission denied (requires root/sudo)")
		}
	}
	return nil
}

// IngestSignal is the public API to feed data into Security Engine
func IngestSignal(ip string, signalType string, severity int, details string) {
	// Avoid panic if sending to closed channel during shutdown
	select {
	case <-done:
		return
	default:
	}

	select {
	case signalChan <- SecuritySignal{
		IP:        ip,
		Type:      signalType,
		Severity:  severity,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}:
	default:
		// Drop if full to prevent blocking
	}
}

// StartSecurityBackground initializes the security engine
func StartSecurityBackground(db *sqlx.DB) {
	if db == nil {
		return
	}

	if err := CheckSystemRequirements(); err != nil {
		SecurityError = err.Error()
		SecurityActive = false
		log.Printf("Security: System requirements failed: %v", err)
	} else {
		SecurityActive = true
		log.Println("Security: System Active. Security Engine Started.")
	}

	RefreshBanCache(db)
	syncConfig(db)

	go analysisLoop(db)
	go maintenanceLoop(db)
}

// StopSecurity gracefully shuts down the engine and flushes logs
func StopSecurity() {
	if SecurityActive {
		log.Println("Security: Stopping Security Engine...")
		close(done)
		wg.Wait()
		SecurityActive = false
		log.Println("Security: Stopped.")
	}
}

func analysisLoop(db *sqlx.DB) {
	wg.Add(1)
	defer wg.Done()

	logBuffer := make([]models.SecurityLog, 0, 100)
	logTicker := time.NewTicker(2 * time.Second)

	flushLogs := func() {
		if len(logBuffer) == 0 {
			return
		}
		for _, l := range logBuffer {
			_ = database.SaveSecurityLog(db, &l)
		}
		logBuffer = logBuffer[:0]
	}

	defer logTicker.Stop()

	for {
		select {
		case <-done:
			flushLogs()
			return

		case sig := <-signalChan:
			configMu.RLock()
			limit := banThreshold
			enabled := autoBanEnabled
			isStrict := strictMode
			mode := systemMode
			configMu.RUnlock()

			severity := sig.Severity
			if isStrict {
				severity *= 2
			}

			newScore := updateScore(sig.IP, severity)

			if enabled && newScore >= limit && mode == "ENFORCEMENT" {
				reason := fmt.Sprintf("Auto-ban: Reputation %d exceeded threshold. Trigger: %s", newScore, sig.Type)
				go banIP(db, sig.IP, reason)

				// Broadcast to Dashboard for Native Notifications
				sse.GlobalHub.Broadcast(sse.Event{
					Type: "security_alert",
					Payload: map[string]interface{}{
						"title":    "Security Enforcement",
						"message":  fmt.Sprintf("IP Address %s has been restricted due to %s", sig.IP, sig.Type),
						"severity": "critical",
					},
				})
			}

			if sig.Type != SignalTypeTraffic || sig.Severity > 0 {
				l := models.SecurityLog{
					SourceIP:  sig.IP,
					Action:    sig.Type,
					Protocol:  "TCP",
					Timestamp: sig.Timestamp,
					Details:   sig.Details,
				}
				if sig.ReqInfo != nil {
					l.DestPort = sig.ReqInfo.Port
					l.Protocol = sig.ReqInfo.Method
				}
				logBuffer = append(logBuffer, l)
			}

			if sig.Type == SignalTypeBlock {
				registry.GlobalStats.RecordSecurityBlock()
			} else if sig.Type == SignalTypeRateLimit {
				registry.GlobalStats.RecordSecurityRateLimit()
			}

		case <-logTicker.C:
			flushLogs()
		}
	}
}

func updateScore(ip string, change int) int {
	scoreMu.Lock()
	defer scoreMu.Unlock()
	if ip == "127.0.0.1" || ip == "::1" {
		return 0
	}
	ipScores[ip] += change
	return ipScores[ip]
}

func maintenanceLoop(db *sqlx.DB) {
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	syncTicker := time.NewTicker(30 * time.Second)
	defer syncTicker.Stop()

	minutes := 0

	for {
		select {
		case <-done:
			return

		case <-ticker.C:
			minutes++

			configMu.RLock()
			currentDecay := decayRate
			configMu.RUnlock()

			if minutes >= currentDecay {
				minutes = 0
				scoreMu.Lock()
				for ip, score := range ipScores {
					if score > 0 {
						ipScores[ip] = score / 2
						if ipScores[ip] < 5 {
							delete(ipScores, ip)
						}
					}
				}
				scoreMu.Unlock()

				// Clear GeoIP cache every hour to prevent memory leaks
				geoMu.Lock()
				if len(geoIPCache) > 1000 {
					geoIPCache = make(map[string]string)
				}
				geoMu.Unlock()
			}
			cleanupLimiters()

		case <-syncTicker.C:
			RefreshBanCache(db)
			syncConfig(db)
		}
	}
}

func RefreshBanCache(db *sqlx.DB) {
	ips, err := database.GetBannedIPList(db)
	if err != nil {
		// log.Printf("Security: Failed to refresh ban cache: %v", err)
		return
	}
	rules, err := database.GetSecurityRules(db)
	if err != nil {
		// log.Printf("Security: Failed to refresh rule cache: %v", err)
	}

	newCache := make(map[string]bool)
	for _, ip := range ips {
		newCache[ip] = true
	}

	BanCacheMu.Lock()
	BannedIPCache = newCache
	if err == nil {
		RuleCache = rules
	}
	BanCacheMu.Unlock()

	registry.GlobalStats.UpdateSecurityActiveBans(len(ips))
}

func banIP(db *sqlx.DB, ip string, reason string) {
	// Deduplication: Check if already processing
	if _, loaded := banningIPs.LoadOrStore(ip, true); loaded {
		return
	}
	defer banningIPs.Delete(ip)

	BanCacheMu.RLock()
	if BannedIPCache[ip] {
		BanCacheMu.RUnlock()
		return
	}
	BanCacheMu.RUnlock()

	log.Printf("Security: RESTRICTING ACCESS %s - %s", ip, reason)

	if err := utils.BlockIPSystem(ip); err != nil {
		// log.Printf("Security: Failed to execute OS block: %v", err)
	}

	rule := &models.SecurityRule{
		Name:      fmt.Sprintf("Auto-Restriction %s", ip),
		CIDR:      ip,
		Port:      "*",
		Protocol:  "ANY",
		Action:    "DENY",
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}
	database.CreateSecurityRule(db, rule)

	rep := &models.IPReputation{
		IP:              ip,
		ReputationScore: 100,
		TotalEvents:     1,
		LastSeen:        time.Now().UTC(),
		IsBanned:        true,
		BanReason:       reason,
	}
	database.UpdateIPReputation(db, rep)

	RefreshBanCache(db)
}

func syncConfig(db *sqlx.DB) {
	// Auto Ban
	cfg, err := database.GetConfigByKey(db, "security.auto_ban_enabled")
	enabled := true
	if err == nil && cfg != nil && cfg.Value == "false" {
		enabled = false
	}

	// Threshold
	threshold := 100
	cfgThresh, err := database.GetConfigByKey(db, "security.auto_ban_threshold")
	if err == nil && cfgThresh != nil {
		if val, err := strconv.Atoi(cfgThresh.Value); err == nil && val > 0 {
			threshold = val
		}
	}

	// Mode
	mode := "ENFORCEMENT"
	cfgMode, err := database.GetConfigByKey(db, "security.mode")
	if err == nil && cfgMode != nil {
		mode = strings.ToUpper(cfgMode.Value)
	}

	// Strict Mode
	isStrict := false
	cfgStrict, err := database.GetConfigByKey(db, "security.strict_mode")
	if err == nil && cfgStrict != nil && cfgStrict.Value == "true" {
		isStrict = true
	}

	// Decay Rate
	decay := 10
	cfgDecay, err := database.GetConfigByKey(db, "security.decay_rate")
	if err == nil && cfgDecay != nil {
		if val, err := strconv.Atoi(cfgDecay.Value); err == nil && val > 0 {
			decay = val
		}
	}

	// Whitelist
	newWhitelist := make(map[string]bool)
	cfgWhite, err := database.GetConfigByKey(db, "security.whitelist_ips")
	if err == nil && cfgWhite != nil {
		ips := strings.Split(cfgWhite.Value, ",")
		for _, ip := range ips {
			newWhitelist[strings.TrimSpace(ip)] = true
		}
	}

	// GeoIP
	isGeoEnabled := false
	cfgGeo, err := database.GetConfigByKey(db, "security.geoip_enabled")
	if err == nil && cfgGeo != nil && cfgGeo.Value == "true" {
		isGeoEnabled = true
	}

	// Allowed Countries
	newAllowedCountries := make(map[string]bool)
	cfgCountries, err := database.GetConfigByKey(db, "security.allowed_countries")
	if err == nil && cfgCountries != nil {
		codes := strings.Split(cfgCountries.Value, ",")
		for _, code := range codes {
			code = strings.ToUpper(strings.TrimSpace(code))
			if code != "" {
				newAllowedCountries[code] = true
			}
		}
	}

	configMu.Lock()
	autoBanEnabled = enabled
	banThreshold = threshold
	systemMode = mode
	strictMode = isStrict
	decayRate = decay
	whitelist = newWhitelist
	geoIPEnabled = isGeoEnabled
	allowedCountries = newAllowedCountries
	configMu.Unlock()
}

func getCountryCode(ip string) string {
	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") {
		return "US"
	}

	geoMu.RLock()
	code, exists := geoIPCache[ip]
	tripped := geoTripped
	lastAttempt := geoLastAttempt
	geoMu.RUnlock()

	if exists {
		return code
	}

	// Circuit Breaker Logic: If API failed 5 times recently, bypass for 5 minutes
	if tripped {
		if time.Since(lastAttempt) > 5*time.Minute {
			geoMu.Lock()
			geoTripped = false
			geoErrorCount = 0
			geoMu.Unlock()
		} else {
			return "XX" // Bypassing due to API instability
		}
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://ipapi.co/%s/country/", ip), nil)
	req.Header.Set("User-Agent", "goExile-Security-Engine/1.0")

	client := http.Client{Timeout: 2 * time.Second} // Tight timeout
	resp, err := client.Do(req)
	
	if err != nil {
		geoMu.Lock()
		geoErrorCount++
		if geoErrorCount >= 5 {
			geoTripped = true
			geoLastAttempt = time.Now()
			log.Printf("Security: GeoIP API failure threshold reached. Circuit Tripped (Bypass Active).")
		}
		geoMu.Unlock()
		return "XX"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	countryCode := strings.ToUpper(strings.TrimSpace(string(body)))

	// If the API returns an error message or rate limit notice, it won't be 2 chars
	if len(countryCode) == 2 && countryCode != "XX" {
		geoMu.Lock()
		geoIPCache[ip] = countryCode
		geoMu.Unlock()
		return countryCode
	}

	return "XX"
}

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if database.DBConn == nil {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := utils.GetClientIP(r)
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			next.ServeHTTP(w, r)
			return
		}

		configMu.RLock()
		isWhitelisted := whitelist[clientIP]
		mode := systemMode
		isGeoEnabled := geoIPEnabled
		allowed := allowedCountries
		configMu.RUnlock()

		if isWhitelisted {
			next.ServeHTTP(w, r)
			return
		}

		// 1. GeoIP Enforcement
		if isGeoEnabled && len(allowed) > 0 && !allowed["ALL"] {
			country := getCountryCode(clientIP)
			// If we can't determine the country (XX), we allow it to avoid false positives
			// unless 'XX' is explicitly in the disallowed list (not supported here yet)
			if country != "XX" && !allowed[country] {
				IngestSignal(clientIP, SignalTypeBlock, 5, "GeoIP: Unauthorized Country ("+country+")")
				if mode == "ENFORCEMENT" {
					http.Error(w, "Access Denied (Geographic Restriction)", http.StatusForbidden)
					return
				}
			}
		}

		BanCacheMu.RLock()
		isBanned := BannedIPCache[clientIP]
		rules := RuleCache
		BanCacheMu.RUnlock()

		if isBanned {
			IngestSignal(clientIP, SignalTypeBlock, 0, "Blocked (Cached)")
			if mode == "ENFORCEMENT" {
				http.Error(w, "Access Denied (Security Engine)", http.StatusForbidden)
				return
			}
		}

		blocked := false
		rateLimited := false
		var matchedRule *models.SecurityRule

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
					break
				}
			}
		}

		if blocked {
			IngestSignal(clientIP, SignalTypeBlock, 10, "Rule: "+matchedRule.Name)
			if mode == "ENFORCEMENT" {
				http.Error(w, "Access Denied (Security Engine)", http.StatusForbidden)
				return
			}
		}

		if rateLimited {
			IngestSignal(clientIP, SignalTypeRateLimit, 5, "Rate Limit: "+matchedRule.Name)
			if mode == "ENFORCEMENT" {
				http.Error(w, "Rate Limit Exceeded (Security Engine)", http.StatusTooManyRequests)
				return
			}
		}

		// Log generic traffic for statistics
		IngestSignal(clientIP, SignalTypeTraffic, 0, r.Method+" "+r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// Rate Limiting Logic
type rateLimiter struct {
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func checkRateLimit(ip string, limit int, burst int) bool {
	if limit <= 0 {
		return true
	}
	if burst <= 0 {
		burst = limit
	}

	limitMu.RLock()
	lim, exists := limiters[ip]
	limitMu.RUnlock()

	if !exists {
		limitMu.Lock()
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

func cleanupLimiters() {
	limitMu.Lock()
	for ip, lim := range limiters {
		if time.Since(lim.lastUpdate) > 10*time.Minute {
			delete(limiters, ip)
		}
	}
	limitMu.Unlock()
}

// Helpers
func ipMatch(ip, cidr string) bool {
	if cidr == "*" || cidr == "0.0.0.0/0" || cidr == "::/0" {
		return true
	}
	if !strings.Contains(cidr, "/") {
		return ip == cidr
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ipnet.Contains(net.ParseIP(ip))
}

func getPort(r *http.Request) int {
	_, portStr, err := net.SplitHostPort(r.Host)
	if err != nil {
		return 80
	}
	p, _ := strconv.Atoi(portStr)
	return p
}

// GetEngineStats returns real-time metrics from the memory engine
func GetEngineStats() map[string]interface{} {
	scoreMu.RLock()
	activeTrackers := len(ipScores)
	scoreMu.RUnlock()

	BanCacheMu.RLock()
	activeRules := len(RuleCache)
	cachedBans := len(BannedIPCache)
	BanCacheMu.RUnlock()

	return map[string]interface{}{
		"active_trackers": activeTrackers,
		"cached_rules":    activeRules,
		"cached_bans":     cachedBans,
		"queue_depth":     len(signalChan),
	}
}
