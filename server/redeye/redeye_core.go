package redeye

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

	"exile/server/database"
	"exile/server/models"
	"exile/server/registry"
	"exile/server/utils"

	"github.com/jmoiron/sqlx"
)

// =================================================================================
// REDEYE: ADVANCED THREAT INTELLIGENCE & FIREWALL
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
	RuleCache     = []models.RedEyeRule{}
	BanCacheMu    sync.RWMutex
	
	// System Status
	RedEyeActive = false
	RedEyeError  = ""

	// High-Performance Event Bus
	signalChan = make(chan SecuritySignal, 10000)
	
	// Config Cache
	configMu       sync.RWMutex
	autoBanEnabled = true
	banThreshold   = 100
	
	// IP Reputation Tracker (In-Memory)
	ipScores   = make(map[string]int)
	scoreMu    sync.RWMutex

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

// IngestSignal is the public API to feed data into RedEye
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

// StartRedEyeBackground initializes the neural engine
func StartRedEyeBackground(db *sqlx.DB) {
	if db == nil {
		return
	}

	if err := CheckSystemRequirements(); err != nil {
		RedEyeError = err.Error()
		RedEyeActive = false
		log.Printf("RedEye: System requirements failed: %v", err)
	} else {
		RedEyeActive = true
		log.Println("RedEye: System Active. Neural Engine Started.")
	}

	RefreshBanCache(db)
	syncConfig(db)

	go analysisLoop(db)
	go maintenanceLoop(db)
}

// StopRedEye gracefully shuts down the engine and flushes logs
func StopRedEye() {
	if RedEyeActive {
		log.Println("RedEye: Stopping Neural Engine...")
		close(done)
		wg.Wait()
		RedEyeActive = false
		log.Println("RedEye: Stopped.")
	}
}

func analysisLoop(db *sqlx.DB) {
	wg.Add(1)
	defer wg.Done()

	logBuffer := make([]models.RedEyeLog, 0, 100)
	logTicker := time.NewTicker(2 * time.Second)
	
	flushLogs := func() {
		if len(logBuffer) == 0 {
			return
		}
		for _, l := range logBuffer {
			_ = database.SaveRedEyeLog(db, &l)
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
			newScore := updateScore(sig.IP, sig.Severity)
			
			configMu.RLock()
			limit := banThreshold
			enabled := autoBanEnabled
			configMu.RUnlock()

			if enabled && newScore >= limit {
				go banIP(db, sig.IP, fmt.Sprintf("Auto-ban: Reputation %d exceeded threshold. Trigger: %s", newScore, sig.Type))
			}

			if sig.Type != SignalTypeTraffic || sig.Severity > 0 {
				l := models.RedEyeLog{
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
				registry.GlobalStats.RecordRedEyeBlock()
			} else if sig.Type == SignalTypeRateLimit {
				registry.GlobalStats.RecordRedEyeRateLimit()
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

	decayTicker := time.NewTicker(1 * time.Minute)
	syncTicker := time.NewTicker(30 * time.Second)
	defer decayTicker.Stop()
	defer syncTicker.Stop()

	for {
		select {
		case <-done:
			return

		case <-decayTicker.C:
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
		log.Printf("RedEye: Failed to refresh ban cache: %v", err)
		return
	}
	rules, err := database.GetRedEyeRules(db)
	if err != nil {
		log.Printf("RedEye: Failed to refresh rule cache: %v", err)
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

	registry.GlobalStats.UpdateRedEyeActiveBans(len(ips))
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

	log.Printf("RedEye: BANNING %s - %s", ip, reason)

	if err := utils.BlockIPSystem(ip); err != nil {
		log.Printf("RedEye: Failed to execute OS block: %v", err)
	}

	rule := &models.RedEyeRule{
		Name:      fmt.Sprintf("Auto-Ban %s", ip),
		CIDR:      ip,
		Port:      "*",
		Protocol:  "ANY",
		Action:    "DENY",
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}
	database.CreateRedEyeRule(db, rule)

	rep := &models.RedEyeIPReputation{
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
	cfg, err := database.GetConfigByKey(db, "redeye.auto_ban_enabled")
	enabled := true
	if err == nil && cfg != nil && cfg.Value == "false" {
		enabled = false
	}

	threshold := 100
	cfgThresh, err := database.GetConfigByKey(db, "redeye.auto_ban_threshold")
	if err == nil && cfgThresh != nil {
		if val, err := strconv.Atoi(cfgThresh.Value); err == nil && val > 0 {
			threshold = val
		}
	}

	configMu.Lock()
	autoBanEnabled = enabled
	banThreshold = threshold
	configMu.Unlock()
}

func RedEyeMiddleware(next http.Handler) http.Handler {
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

		BanCacheMu.RLock()
		isBanned := BannedIPCache[clientIP]
		rules := RuleCache
		BanCacheMu.RUnlock()

		if isBanned {
			IngestSignal(clientIP, SignalTypeBlock, 0, "Blocked (Cached)")
			http.Error(w, "Access Denied (Banned)", http.StatusForbidden)
			return
		}

		blocked := false
		rateLimited := false
		var matchedRule *models.RedEyeRule

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
			http.Error(w, "Access Denied (RedEye)", http.StatusForbidden)
			return
		}

		if rateLimited {
			IngestSignal(clientIP, SignalTypeRateLimit, 5, "Rate Limit: "+matchedRule.Name)
			http.Error(w, "Rate Limit Exceeded (RedEye)", http.StatusTooManyRequests)
			return
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
	if limit <= 0 { return true }
	if burst <= 0 { burst = limit }

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