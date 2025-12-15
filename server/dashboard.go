package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var (
	loginAttempts   = make(map[string]time.Time)
	loginAttemptsMu sync.Mutex
)

// ErrorLog represents a single error event.
type ErrorLog struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Message   string    `json:"message"`
	ClientIP  string    `json:"client_ip"`
}

// DashboardStats holds metrics for display.
type DashboardStats struct {
	mu              sync.RWMutex
	TotalRequests   int64
	TotalErrors     int64
	ActiveSpawners  int
	BytesSent       int64
	BytesReceived   int64
	MemUsage        uint64 // in bytes
	LastRequestTime time.Time
	DBConnected     bool
	Uptime          time.Duration
	StartTime       time.Time
	ErrorLogs       []ErrorLog
}

// GlobalStats is the global dashboard stats instance.
var GlobalStats = &DashboardStats{
	StartTime: time.Now(),
	ErrorLogs: make([]ErrorLog, 0),
}

// RecordRequest increments the request counter and bandwidth stats.
func (ds *DashboardStats) RecordRequest(statusCode int, bytesIn, bytesOut int64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.TotalRequests++
	ds.BytesReceived += bytesIn
	ds.BytesSent += bytesOut
	if statusCode >= 400 {
		ds.TotalErrors++
	}
	ds.LastRequestTime = time.Now()
}

// RecordError adds an error to the log.
func (ds *DashboardStats) RecordError(path string, status int, message string, clientIP string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	logEntry := ErrorLog{
		Timestamp: time.Now(),
		Path:      path,
		Status:    status,
		Message:   message,
		ClientIP:  clientIP,
	}

	// Prepend and keep last 50
	ds.ErrorLogs = append([]ErrorLog{logEntry}, ds.ErrorLogs...)
	if len(ds.ErrorLogs) > 50 {
		ds.ErrorLogs = ds.ErrorLogs[:50]
	}
}

// UpdateMemoryStats updates the memory usage stat.
func (ds *DashboardStats) UpdateMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ds.mu.Lock()
	ds.MemUsage = m.Alloc
	ds.mu.Unlock()
}

// GetStats returns a snapshot of current stats.
func (ds *DashboardStats) GetStats() (totalReq int64, totalErr int64, active int, dbOK bool, uptime time.Duration, mem uint64, tx, rx int64) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	totalReq = ds.TotalRequests
	totalErr = ds.TotalErrors
	active = ds.ActiveSpawners
	dbOK = ds.DBConnected
	uptime = time.Since(ds.StartTime)
	mem = ds.MemUsage
	tx = ds.BytesSent
	rx = ds.BytesReceived
	return
}

// GetErrors returns the list of recent errors.
func (ds *DashboardStats) GetErrors() []ErrorLog {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	// Return a copy to avoid races
	logs := make([]ErrorLog, len(ds.ErrorLogs))
	copy(logs, ds.ErrorLogs)
	return logs
}

// ClearErrors empties the error log and resets the error count.
func (ds *DashboardStats) ClearErrors() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ErrorLogs = make([]ErrorLog, 0)
	ds.TotalErrors = 0 // Reset the error counter
}

// UpdateActiveServers updates the active server count.
func (ds *DashboardStats) UpdateActiveServers(count int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.ActiveSpawners = count
}

// SetDBConnected updates DB connection status.
func (ds *DashboardStats) SetDBConnected(connected bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.DBConnected = connected
}

// PrintBanner prints a stylish ASCII banner.
func PrintBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════╗
║                 🎮 SPAWNER REGISTRY 🎮                        ║
║              xx             ║
╚═══════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// PrintStatus prints real-time server status.
func PrintStatus(w http.ResponseWriter, r *http.Request) {
	// Simple record for status page itself
	GlobalStats.RecordRequest(http.StatusOK, 0, 0)

	totalReq, totalErr, active, dbOK, uptime, mem, tx, rx := GlobalStats.GetStats()

	dbStatus := "✓ Connected"
	if !dbOK {
		dbStatus = "✗ Disconnected"
	}

	status := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                    📊 REGISTRY STATUS 📊                       ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  🕐 Uptime:            %-42s║
║  🖥️  Active Spawners:   %-42d║
║  📡 Total Requests:    %-42d║
║  ⚠️  Errors:           %-42d║
║  💾 Memory Usage:      %-42s║
║  ⬆️  Tx:               %-42s║
║  ⬇️  Rx:               %-42s║
║  🗄️  Database:         %-42s║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`,
		formatDuration(uptime),
		active,
		totalReq,
		totalErr,
		formatBytes(mem),
		formatBytes(uint64(tx)),
		formatBytes(uint64(rx)),
		dbStatus,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, status)
}

// formatBytes formats bytes to human readable string
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration in a readable way.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// PrintSpawnerList prints formatted spawner list.
func PrintSpawnerList(spawners []*Spawner) string {
	if len(spawners) == 0 {
		return `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SPAWNERS 📦                            ║
╠════════════════════════════════════════════════════════════════╣
║  No spawners registered yet.                                   ║
╚════════════════════════════════════════════════════════════════╝
`
	}

	header := `
╔════════════════════════════════════════════════════════════════╗
║                      📦 SPAWNERS 📦                            ║
╠════════════════════════════════════════════════════════════════╣
║ ID   │ Region           │ Host:Port         │ Status   │ Inst.  ║
╠════════════════════════════════════════════════════════════════╣
`

	body := ""
	for _, s := range spawners {
		status := "🟢 Online"
		if s.Status == "offline" {
			status = "🔴 Offline"
		}

		body += fmt.Sprintf("║ %-4d │ %-16s │ %-17s │ %-8s │ %d/%d  ║\n",
			s.ID,
			truncate(s.Region, 16),
			fmt.Sprintf("%s:%d", s.Host, s.Port),
			status,
			s.CurrentInstances,
			s.MaxInstances,
		)
	}

	footer := `╚════════════════════════════════════════════════════════════════╝
`

	return header + body + footer
}

// truncate truncates a string to max length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}

// PrintConfig prints server configuration.
func PrintConfig(port string, dbPath string) {
	config := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║                   ⚙️  CONFIGURATION ⚙️                        ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  🌐 HTTP Port:         %-42s║
║  🗄️  Database Path:     %-42s║
║  🔧 Server TTL:        %-42ds║
║  🧹 Cleanup Interval:  %-42ds║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝
`,
		port,
		dbPath,
		serverTTL,
		cleanupInterval,
	)
	fmt.Println(config)
}

// SendEmail sends a verification code via SMTP.
func SendEmail(authConfig AuthConfig, code string) error {
	auth := smtp.PlainAuth("", authConfig.SMTPUser, authConfig.SMTPPass, authConfig.SMTPHost)
	
	to := []string{authConfig.Email}
	msg := []byte("To: " + authConfig.Email + "\r\n" +
		"Subject: Dashboard Verification Code\r\n" +
		"\r\n" +
		"Your verification code is: " + code + "\r\n")
		
	addr := fmt.Sprintf("%s:%s", authConfig.SMTPHost, authConfig.SMTPPort)
	return smtp.SendMail(addr, auth, authConfig.SMTPFrom, to, msg)
}

func generateSecureCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = '0' + (b[i] % 10)
	}
	return string(b)
}

// HandleLogin processes login requests.
func HandleLogin(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate Limiting: 5 attempts per 15 minutes per IP
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	allowed, count := LoginRateLimiter.Allow(ip)
	if !allowed {
		log.Printf("Login rate limited for IP: %s (Count: %d)", ip, count)

		// Trigger firewall block on first failure
		if count == LoginRateLimiter.maxAttempts+1 {
			go func(targetIP string) {
				if err := BlockIP(targetIP); err != nil {
					log.Printf("Failed to block IP %s: %v", targetIP, err)
				}
			}(ip)
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Verify email and hashed password
	if email == authConfig.Email {
		err := bcrypt.CompareHashAndPassword([]byte(authConfig.HashedPassword), []byte(password))
		if err == nil { // Password matches
			// Reset rate limit on success
			LoginRateLimiter.Reset(ip)

			// Determine Initial Step
			initialStep := AuthStepAuthenticated
			if authConfig.TOTPSecret != "" {
				initialStep = AuthStepTOTP
			}

			sessionID, err := sessionStore.CreateSession(initialStep)
			if err != nil {
				http.Error(w, "Failed to create session", http.StatusInternalServerError)
				return
			}

			// Secure attributes based on production mode
			sameSite := http.SameSiteLaxMode
			if authConfig.IsProduction {
				sameSite = http.SameSiteStrictMode
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   authConfig.IsProduction, // Only true in production
				SameSite: sameSite,
				MaxAge:   86400,
			})

			if initialStep == AuthStepTOTP {
				http.Redirect(w, r, "/login/2fa", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			return
		}
	}

	// Authentication failed
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Handle2FAVerify processes TOTP verification requests.
func Handle2FAVerify(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Validate current session state
	isValid, authStep := sessionStore.ValidateSession(cookie.Value)
	if !isValid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if authStep == AuthStepAuthenticated {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	
	if authStep != AuthStepTOTP {
		// Wrong step (maybe stuck in email step?)
		// Just let them try email verify if they are there, but here we expect TOTP.
		writeError(w, r, http.StatusBadRequest, "Invalid auth step")
		return
	}

	// Rate Limiting: 3 attempts per 15 minutes per Session
	allowed, _ := TwoFactorRateLimiter.Allow(cookie.Value)
	if !allowed {
		log.Printf("2FA rate limit exceeded for session: %s", cookie.Value)
		sessionStore.RevokeSession(cookie.Value)
		writeError(w, r, http.StatusTooManyRequests, "Too many failed attempts. Session revoked.")
		return
	}

	// Parse TOTP Code
	var code string
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			code = req.Code
		}
	} else {
		code = r.FormValue("code")
	}

	code = strings.TrimSpace(code)

	if code == "" {
		writeError(w, r, http.StatusBadRequest, "Code required")
		return
	}

	// Verify Code
	if authConfig.TOTPSecret == "" {
		sessionStore.MarkSessionAuthenticated(cookie.Value)
		writeJSON(w, http.StatusOK, map[string]string{"message": "authenticated"})
		return
	}

	// Validate with Skew (1 step = 30s) to allow for time drift
	valid, _ := totp.ValidateCustom(code, authConfig.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if valid {
		// Reset rate limit on success
		TwoFactorRateLimiter.Reset(cookie.Value)
		
		// 3-Step Auth: Email Verification
		if authConfig.SMTPHost != "" {
			emailCode := generateSecureCode()
			sessionStore.SetEmailCode(cookie.Value, emailCode)
			sessionStore.SetSessionStep(cookie.Value, AuthStepEmail)
			
			if err := SendEmail(authConfig, emailCode); err != nil {
				log.Printf("Failed to send email: %v", err)
				writeError(w, r, http.StatusInternalServerError, "Failed to send verification email")
				return
			}
			
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "email_step_required",
				"next_step": "email",
			})
			return
		}

		sessionStore.MarkSessionAuthenticated(cookie.Value)
		writeJSON(w, http.StatusOK, map[string]string{"message": "authenticated"})
		return
	}

	log.Printf("2FA Failed. Session: %s... CodeLen: %d Content-Type: %s", 
		cookie.Value[:8], len(code), r.Header.Get("Content-Type"))
	writeError(w, r, http.StatusUnauthorized, "Invalid code")
}

// HandleEmailVerify processes Email verification requests.
func HandleEmailVerify(w http.ResponseWriter, r *http.Request, authConfig AuthConfig, sessionStore *SessionStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	isValid, authStep := sessionStore.ValidateSession(cookie.Value)
	if !isValid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if authStep != AuthStepEmail {
		writeError(w, r, http.StatusBadRequest, "Invalid auth step")
		return
	}

	allowed, _ := EmailCodeRateLimiter.Allow(cookie.Value)
	if !allowed {
		sessionStore.RevokeSession(cookie.Value)
		writeError(w, r, http.StatusTooManyRequests, "Too many failed attempts. Session revoked.")
		return
	}

	var code string
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			code = req.Code
		}
	} else {
		code = r.FormValue("code")
	}

	if sessionStore.VerifyEmailCode(cookie.Value, code) {
		EmailCodeRateLimiter.Reset(cookie.Value)
		sessionStore.MarkSessionAuthenticated(cookie.Value)
		writeJSON(w, http.StatusOK, map[string]string{"message": "authenticated"})
		return
	}

	writeError(w, r, http.StatusUnauthorized, "Invalid email code")
}

// HandleLogout revokes the session.
func HandleLogout(w http.ResponseWriter, r *http.Request, sessionStore *SessionStore) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionStore.RevokeSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}