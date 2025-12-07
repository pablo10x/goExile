# Implementation Tasks (Prioritized)

## Quick Reference

**Phase 1 (Security - 2 hours):**
- [ ] Task 1: Add API Key Authentication
- [ ] Task 2: Add HTTPS/TLS Support
- [ ] Task 3: Add Request Logging

**Phase 2 (Reliability - 2-3 hours):**
- [ ] Task 4: Add SQLite Persistence
- [ ] Task 5: Add Graceful Shutdown
- [ ] Task 6: Add Rate Limiting

**Phase 3 (Observability - 1+ hours):**
- [ ] Task 7: Add Prometheus Metrics
- [ ] Task 8: Improve Request Validation

**Phase 4 (Testing & DevOps - 2-3 hours):**
- [ ] Task 9: Docker Support
- [ ] Task 10: Unit Tests
- [ ] Task 11: Integration Tests

**Phase 5 (Polish - optional):**
- [ ] Task 12: API Documentation (OpenAPI)
- [ ] Task 13: Redis for Horizontal Scaling

---

## Task 1: Add API Key Authentication ⏱️ 30 min

**Priority**: CRITICAL  
**Files to modify**: `server/main.go`  
**Files to create**: None  
**Dependencies to add**: None

### What You're Adding
- Require `X-API-Key` header on registration/deletion endpoints
- Keep GET endpoints public (for clients)
- Read API key from `API_KEY` environment variable

### Implementation Checklist
- [ ] Add `checkAPIKey(r *http.Request) error` function
- [ ] Wrap `RegisterServer` with auth check
- [ ] Wrap `HeartbeatServer` with auth check
- [ ] Wrap `DeleteServer` with auth check
- [ ] Keep `ListServers` and `GetServer` public (no auth)
- [ ] Return 401 Unauthorized on bad key
- [ ] Test with cURL (with and without key)
- [ ] Update README with auth example

### Example Usage (After Implementation)
```bash
# Set API key
export API_KEY="my-secret-key-12345"
go run main.go

# Register (with auth)
curl -X POST http://localhost:8081/api/servers \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-secret-key-12345" \
  -d '{"name":"room-1","host":"1.2.3.4","port":7777,"max_players":8}'

# List (no auth needed)
curl http://localhost:8081/api/servers
```

### Code Snippet to Add
```go
func checkAPIKey(r *http.Request) error {
    expected := os.Getenv("API_KEY")
    if expected == "" {
        return fmt.Errorf("API_KEY not configured")
    }
    actual := r.Header.Get("X-API-Key")
    if actual != expected {
        return fmt.Errorf("unauthorized")
    }
    return nil
}
```

---

## Task 2: Add HTTPS/TLS Support ⏱️ 20 min

**Priority**: CRITICAL  
**Files to modify**: `server/main.go`  
**Files to create**: None (certs provided by deployment env)  
**Dependencies to add**: None

### What You're Adding
- Listen on HTTPS instead of HTTP
- Load cert/key from environment variables or files
- Generate self-signed cert for testing

### Implementation Checklist
- [ ] Modify `run()` to use `http.ListenAndServeTLS()`
- [ ] Read `TLS_CERT` and `TLS_KEY` env vars
- [ ] Generate self-signed cert (command provided below)
- [ ] Test with curl using `--insecure` flag
- [ ] Update README with cert generation example
- [ ] Document how to get real certs (Let's Encrypt)

### Generate Self-Signed Cert (for testing)
```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout server/key.pem \
  -out server/cert.pem \
  -days 365 \
  -nodes \
  -subj "/CN=localhost"
```

### Example Usage (After Implementation)
```bash
export TLS_CERT="./cert.pem"
export TLS_KEY="./key.pem"
go run main.go

# Test with curl (ignore cert validation for self-signed)
curl -k --cacert ./cert.pem https://localhost:8081/health
```

### Code Snippet to Add
```go
func run() error {
    // ... router setup ...
    
    certFile := os.Getenv("TLS_CERT")
    keyFile := os.Getenv("TLS_KEY")
    
    if certFile == "" || keyFile == "" {
        log.Println("TLS_CERT and TLS_KEY not set, using HTTP (insecure!)")
        return http.ListenAndServe(":8081", router)
    }
    
    log.Println("Starting on HTTPS with TLS")
    return http.ListenAndServeTLS(":8081", certFile, keyFile, router)
}
```

---

## Task 3: Add Request Logging ⏱️ 20 min

**Priority**: HIGH  
**Files to modify**: `server/main.go`  
**Files to create**: None  
**Dependencies to add**: None (use stdlib `log/slog`)

### What You're Adding
- Structured JSON logging for all important events
- Log who did what (register, delete, etc.) with timestamp
- Log auth failures
- Log cleanup events

### Implementation Checklist
- [ ] Import `log/slog`
- [ ] Create logger in `main()`
- [ ] Log on RegisterServer: `{ action: "register", server_id: 1, name: "room-1", ip: "1.2.3.4" }`
- [ ] Log on DeleteServer: `{ action: "delete", server_id: 1, ip: "1.2.3.4" }`
- [ ] Log on Cleanup: `{ action: "cleanup", server_id: 1, reason: "expired" }`
- [ ] Log auth failures: `{ action: "auth_failed", endpoint: "/api/servers", ip: "1.2.3.4" }`
- [ ] Optional: Add LOG_FILE env var to write to file
- [ ] Test and verify output

### Example Usage (After Implementation)
```bash
# Log to stdout (default)
go run main.go

# Output:
# {"time":"2025-12-05T17:30:00Z","level":"INFO","msg":"registered server","server_id":1,"name":"room-1"}
# {"time":"2025-12-05T17:30:05Z","level":"INFO","msg":"cleanup removed server","server_id":1}

# Log to file
export LOG_FILE="./registry.log"
go run main.go
```

---

## Task 4: Add SQLite Persistence ⏱️ 45 min

**Priority**: CRITICAL  
**Files to modify**: `server/main.go`, `server/go.mod`  
**Files to create**: `server/db.go`  
**Dependencies to add**: `github.com/mattn/go-sqlite3`

### What You're Adding
- SQLite database to persist servers across restarts
- Auto-load servers on startup
- Auto-save on register/delete
- Simple schema (one `servers` table)

### Implementation Checklist
- [ ] Add `github.com/mattn/go-sqlite3` to `go.mod`
- [ ] Create `db.go` with:
  - [ ] `InitDB(dbPath string) (*sql.DB, error)` function
  - [ ] `CreateTables(db *sql.DB) error` function
  - [ ] `SaveServer(db *sql.DB, s *Server) error` function
  - [ ] `LoadServers(db *sql.DB) ([]Server, error)` function
  - [ ] `DeleteServer(db *sql.DB, id int) error` function
- [ ] Modify Registry to use DB instead of map
- [ ] Load servers from DB in `main()` before starting server
- [ ] Test restart persistence

### Database Schema
```sql
CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    max_players INTEGER NOT NULL,
    current_players INTEGER NOT NULL,
    region TEXT,
    status TEXT,
    last_seen DATETIME NOT NULL
);
```

### Example Usage (After Implementation)
```bash
go run main.go

# Registry logs: "Loaded 3 servers from database"
# Register a server
curl -X POST http://localhost:8081/api/servers \
  -H "X-API-Key: test" \
  -d '{"name":"room-1",...}'

# Kill the server (Ctrl+C)
# Restart
go run main.go

# Registry logs: "Loaded 4 servers from database"
# Server 1 still there!
```

---

## Task 5: Add Graceful Shutdown ⏱️ 15 min

**Priority**: HIGH  
**Files to modify**: `server/main.go`  
**Files to create**: None  
**Dependencies to add**: None

### What You're Adding
- Catch OS signals (Ctrl+C, systemd stop, etc.)
- Stop accepting new requests
- Wait for in-flight requests to finish (timeout 30s)
- Close database cleanly
- Exit with code 0

### Implementation Checklist
- [ ] Import `os` and `os/signal`
- [ ] Create signal channel
- [ ] Create context with timeout
- [ ] Catch SIGTERM, SIGINT
- [ ] Call `server.Shutdown(ctx)` on http.Server
- [ ] Close database on exit
- [ ] Test: Start server → Ctrl+C → verify clean exit in logs

### Example Usage (After Implementation)
```bash
go run main.go
# In another terminal:
pkill -TERM go  # or Ctrl+C

# Output:
# Received shutdown signal
# Waiting for in-flight requests to complete (timeout 30s)
# Database closed
# Shutdown complete
```

---

## Task 6: Add Rate Limiting ⏱️ 25 min

**Priority**: HIGH  
**Files to modify**: `server/main.go`, `server/go.mod`  
**Files to create**: None  
**Dependencies to add**: `golang.org/x/time/rate`

### What You're Adding
- Per-IP rate limiting (e.g., 100 requests/sec)
- Return 429 (Too Many Requests) when exceeded
- Simple token bucket algorithm

### Implementation Checklist
- [ ] Add `golang.org/x/time/rate` to `go.mod`
- [ ] Create rate limiter map: `map[string]*rate.Limiter`
- [ ] Create middleware to check IP + limit
- [ ] Return 429 on limited
- [ ] Apply middleware to all endpoints
- [ ] Test with `ab` or `wrk` benchmark tool

### Example Usage (After Implementation)
```bash
# Hammer endpoint
ab -n 1000 -c 100 http://localhost:8081/health

# Some requests get 429 Too Many Requests
```

---

## Task 7: Add Prometheus Metrics ⏱️ 45 min

**Priority**: NICE-TO-HAVE  
**Files to modify**: `server/main.go`, `server/go.mod`  
**Files to create**: `server/metrics.go`  
**Dependencies to add**: `github.com/prometheus/client_golang`

### What You're Adding
- Prometheus metrics endpoint
- Counter for servers registered/deleted
- Gauge for active server count
- Request latency histogram

### Metrics to Add
- `servers_registered_total` (counter)
- `servers_deleted_total` (counter)
- `active_servers` (gauge)
- `request_duration_seconds` (histogram)
- `heartbeats_total` (counter)

### Example Usage (After Implementation)
```bash
# Scrape metrics
curl http://localhost:8081/metrics

# Output:
# # HELP servers_registered_total Total servers registered
# # TYPE servers_registered_total counter
# servers_registered_total 42
#
# # HELP active_servers Current active servers
# # TYPE active_servers gauge
# active_servers 37
```

---

## Task 8: Improve Request Validation ⏱️ 30 min

**Priority**: NICE-TO-HAVE  
**Files to modify**: `server/main.go`  
**Files to create**: None  
**Dependencies to add**: None

### What You're Adding
- IP address validation (valid IP or hostname)
- Region validation (whitelist allowed regions)
- Name length validation (3-64 chars)
- More comprehensive error messages

### Validation Rules
- `Name`: 3-64 chars, alphanumeric + hyphens
- `Host`: Valid IPv4, IPv6, or hostname
- `Port`: 1-65535
- `MaxPlayers`: 1-10000
- `Region`: Must be one of: "us-west", "us-east", "eu", "asia-pacific"
- `CurrentPlayers`: 0-MaxPlayers

### Example Usage (After Implementation)
```bash
# Invalid region
curl -X POST http://localhost:8081/api/servers \
  -d '{"name":"room-1","host":"1.2.3.4","port":7777,"max_players":8,"region":"invalid"}'

# Error: "invalid region: must be one of us-west, us-east, eu, asia-pacific"
```

---

## Task 9: Docker Support ⏱️ 20 min

**Priority**: NICE-TO-HAVE  
**Files to create**: `Dockerfile`, `.dockerignore`, `docker-compose.yml`  
**Files to modify**: None  
**Dependencies to add**: None

### What You're Adding
- Dockerfile (multi-stage build for small image)
- docker-compose.yml for local development
- .dockerignore to exclude unnecessary files

### Files to Create

**Dockerfile:**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o registry main.go db.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/registry /registry
EXPOSE 8081
CMD ["/registry"]
```

**docker-compose.yml:**
```yaml
version: '3'
services:
  registry:
    build: .
    ports:
      - "8081:8081"
    environment:
      API_KEY: dev-key
    volumes:
      - ./registry.db:/registry.db
```

### Example Usage
```bash
docker build -t game-registry .
docker run -p 8081:8081 -e API_KEY=test game-registry

# Or with compose
docker-compose up
```

---

## Task 10: Unit Tests ⏱️ 1-2 hours

**Priority**: MEDIUM  
**Files to create**: `server/main_test.go`  
**Files to modify**: None  
**Dependencies to add**: None

### Test Cases to Add
- [ ] RegisterServer: valid input, invalid input, duplicate
- [ ] HeartbeatServer: exists, not exists
- [ ] ListServers: empty, multiple
- [ ] GetServer: exists, not exists
- [ ] DeleteServer: exists, not exists
- [ ] Cleanup: removes expired servers
- [ ] Authentication: valid key, invalid key, missing key
- [ ] Validation: bad host, bad port, bad region

### Example Test
```go
func TestRegisterServer(t *testing.T) {
    // Setup
    router := mux.NewRouter()
    // ... register handlers ...
    
    // Test valid registration
    body := `{"name":"room-1","host":"1.2.3.4","port":7777,"max_players":8}`
    req := httptest.NewRequest("POST", "/api/servers", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "test-key")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }
}
```

---

## Task 11: Integration Tests ⏱️ 1+ hours

**Priority**: MEDIUM  
**Files to create**: `server/integration_test.go`  
**Files to modify**: None  
**Dependencies to add**: None

### Test Scenarios
- [ ] Full workflow: register → heartbeat → list
- [ ] Cleanup removes expired servers
- [ ] Concurrent registrations don't race
- [ ] Persistence: register → restart → verify
- [ ] Auth failures on protected endpoints

---

## Task 12: API Documentation (OpenAPI/Swagger) ⏱️ 30-45 min

**Priority**: NICE-TO-HAVE  
**Files to create**: `docs/swagger.yaml` (or auto-generated)  
**Files to modify**: `server/main.go` (add annotations)  
**Dependencies to add**: `github.com/swaggo/swag`

### What You're Adding
- Swagger/OpenAPI documentation
- Interactive API docs at `/swagger/index.html`
- Auto-generated client SDKs (optional)

---

## Task 13: Redis for Horizontal Scaling ⏱️ 2+ hours

**Priority**: LOW (only if needed)  
**Files to create**: `server/registry_redis.go`  
**Files to modify**: `server/main.go`, `server/go.mod`  
**Dependencies to add**: `github.com/redis/go-redis/v9`

### What You're Adding
- Replace in-memory map with Redis
- Support multiple registry instances
- Shared server list across instances

---

## Notes

- **Bold text** = Most critical
- ⏱️ = Estimated time to complete
- Start with Phase 1 (Security) before going to production
- Phase 2 & 3 recommended within first month
- Phase 4+ can be done gradually

---

**Last Updated**: 2025-12-05
