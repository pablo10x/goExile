# Production Readiness Checklist

## Overview
This checklist tracks what's needed to move the Game Server Registry from development to production.

**Current Status**: Development-ready, needs security & persistence hardening for production

---

## 🔴 CRITICAL (Must Have Before Production)

### [ ] 1. API Key Authentication
**Why**: Anyone on the internet can register/delete fake servers, DoS the registry  
**Effort**: 30 minutes  
**Impact**: CRITICAL  
**Status**: NOT STARTED

Tasks:
- [ ] Add `API_KEY` environment variable
- [ ] Create auth middleware to check `X-API-Key` header
- [ ] Apply auth to POST/DELETE endpoints (register, heartbeat, delete)
- [ ] Keep GET endpoints public (for clients to discover)
- [ ] Add auth test with cURL
- [ ] Document in README how to set API_KEY

**File**: `server/main.go` → Add `checkAPIKey()` function, wrap handlers

---

### [ ] 2. Data Persistence (SQLite)
**Why**: All servers lost on restart/crash → bad UX  
**Effort**: 45 minutes  
**Impact**: HIGH  
**Status**: NOT STARTED

Tasks:
- [ ] Add `github.com/mattn/go-sqlite3` dependency
- [ ] Create `db.go` with init/create tables
- [ ] Replace in-memory map with DB queries
- [ ] Load servers from DB on startup
- [ ] Save on register/delete/heartbeat
- [ ] Add migration for schema
- [ ] Test persistence across restarts
- [ ] Add `registry.db` to `.gitignore`

**Files**: `server/db.go` (new), `server/main.go` (modify Registry), `server/go.mod` (add dep)

---

### [ ] 3. HTTPS/TLS Support
**Why**: Credentials/data sent in plaintext over HTTP → man-in-the-middle attacks  
**Effort**: 20 minutes  
**Impact**: HIGH  
**Status**: NOT STARTED

Tasks:
- [ ] Add TLS cert/key path env vars (`TLS_CERT`, `TLS_KEY`)
- [ ] Switch from `http.ListenAndServe` to `http.ListenAndServeTLS`
- [ ] Generate self-signed cert for testing: `openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes`
- [ ] Update connection examples in README
- [ ] Add deployment docs (how to get real certs from Let's Encrypt)

**File**: `server/main.go` → Modify `run()` function

---

## 🟡 HIGH PRIORITY (Strongly Recommended)

### [ ] 4. Rate Limiting
**Why**: Prevent abuse, DoS attacks  
**Effort**: 25 minutes  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Add `golang.org/x/time/rate` dependency
- [ ] Create per-IP rate limiter (e.g., 100 req/sec)
- [ ] Add middleware to check rate limit
- [ ] Return 429 (Too Many Requests) when limit exceeded
- [ ] Test with `ab` or `wrk` benchmark
- [ ] Document rate limits in README

**File**: `server/main.go` → Add `rateLimitMiddleware()`

---

### [ ] 5. Request Logging & Audit Trail
**Why**: Track who registered/deleted servers, debug issues  
**Effort**: 20 minutes  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Add structured JSON logging (use `log/slog` or `logrus`)
- [ ] Log all registrations: `{ action: "register", server_id: 1, ip: "1.2.3.4", timestamp: "..." }`
- [ ] Log all deletions: `{ action: "delete", server_id: 1, ip: "1.2.3.4" }`
- [ ] Log heartbeats (optional, verbose)
- [ ] Log cleanup removals: `{ action: "cleanup", server_id: 1, reason: "expired" }`
- [ ] Add log file output (e.g., `LOG_FILE` env var)
- [ ] Rotate logs (e.g., daily)

**File**: `server/main.go` → Replace `log.Printf` with structured logger

---

### [ ] 6. Graceful Shutdown
**Why**: Don't kill requests mid-flight; save state cleanly  
**Effort**: 15 minutes  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Catch OS signals (SIGTERM, SIGINT)
- [ ] Stop accepting new requests
- [ ] Wait for in-flight requests to complete (timeout 30s)
- [ ] Flush logs
- [ ] Close database connection
- [ ] Exit cleanly

**File**: `server/main.go` → Add signal handler in `run()`

---

## 🟢 NICE-TO-HAVE (Optional for MVP)

### [ ] 7. Prometheus Metrics
**Why**: Monitor registry health (registrations/sec, server count, etc.)  
**Effort**: 45 minutes  
**Impact**: LOW-MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Add `prometheus/client_golang` dependency
- [ ] Create counters: `servers_registered_total`, `servers_deleted_total`, `heartbeats_total`
- [ ] Create gauges: `active_servers`, `cleanup_removed_total`
- [ ] Track request latency
- [ ] Expose `/metrics` endpoint
- [ ] Add Prometheus scrape config example in README

**File**: `server/metrics.go` (new), `server/main.go` (integrate)

---

### [ ] 8. Request Validation Improvements
**Why**: Catch more bad input early  
**Effort**: 30 minutes  
**Impact**: MEDIUM  
**Status**: PARTIAL (basic validation exists)

Tasks:
- [ ] Add IP validation (validate `Host` is valid IP or hostname)
- [ ] Add region validation (whitelist: "us-west", "us-east", "eu", etc.)
- [ ] Add name length limits (min 3, max 64 chars)
- [ ] Add custom validation function
- [ ] Write unit tests for validation

**File**: `server/main.go` → Enhance `RegisterServer()` validation

---

### [ ] 9. Horizontal Scaling (Redis)
**Why**: Scale beyond single instance  
**Effort**: 2+ hours  
**Impact**: LOW (unless you need it now)  
**Status**: NOT STARTED

Tasks:
- [ ] Add `redis` dependency
- [ ] Replace in-memory map with Redis hash
- [ ] Use Redis pub/sub for cleanup notifications
- [ ] Add connection pooling
- [ ] Handle Redis failures gracefully (fallback to memory)
- [ ] Update deployment docs

**Files**: `server/registry_redis.go` (new), `server/main.go` (modify)

**Note**: Only if you plan >1 registry instance

---

### [ ] 10. Docker Support
**Why**: Easy deployment, consistent environment  
**Effort**: 20 minutes  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Create `Dockerfile` (multi-stage build)
- [ ] Create `.dockerignore`
- [ ] Create `docker-compose.yml` (app + sqlite)
- [ ] Test docker build & run
- [ ] Add docker run command to README

**Files**: `Dockerfile`, `.dockerignore`, `docker-compose.yml`

---

### [ ] 11. Unit Tests
**Why**: Catch bugs, regression testing  
**Effort**: 1-2 hours  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Test RegisterServer (valid, invalid, duplicate ID)
- [ ] Test heartbeat (exists, not exists)
- [ ] Test ListServers (empty, multiple)
- [ ] Test DeleteServer (exists, not exists)
- [ ] Test cleanup (TTL expiration)
- [ ] Test authentication (valid, invalid key)
- [ ] Aim for >80% coverage

**File**: `server/main_test.go` (new)

---

### [ ] 12. Integration Tests
**Why**: Test full workflows  
**Effort**: 1+ hours  
**Impact**: MEDIUM  
**Status**: NOT STARTED

Tasks:
- [ ] Test register → heartbeat → list flow
- [ ] Test register → delete flow
- [ ] Test cleanup removes expired servers
- [ ] Test concurrent requests
- [ ] Use HTTP client, not mocks

**File**: `server/integration_test.go` (new)

---

### [ ] 13. API Documentation (OpenAPI/Swagger)
**Why**: Auto-generate client SDKs, interactive docs  
**Effort**: 30-45 minutes  
**Impact**: LOW (nice to have)  
**Status**: NOT STARTED

Tasks:
- [ ] Add swagger annotations to handlers
- [ ] Add `swaggo/swag` dependency
- [ ] Generate `docs/swagger.json`
- [ ] Expose `/swagger/index.html` endpoint
- [ ] Test interactive docs

**Files**: `docs/swagger.json` (generated), `server/main.go` (annotations)

---

## 📋 Recommended Implementation Order

**Phase 1: Security (2 hours)**
1. API Key Authentication
2. HTTPS/TLS
3. Request Logging

**Phase 2: Reliability (2-3 hours)**
4. Data Persistence (SQLite)
5. Graceful Shutdown
6. Rate Limiting

**Phase 3: Observability (1+ hours)**
7. Prometheus Metrics
8. Request Validation Improvements

**Phase 4: DevOps & Testing (2-3 hours)**
9. Docker Support
10. Unit Tests
11. Integration Tests

**Phase 5: Polish (optional)**
12. API Documentation
13. Horizontal Scaling (Redis)

---

## Testing Before Production

### Manual Testing Checklist
- [ ] Start server
- [ ] Register a server (with API key)
- [ ] Verify server appears in list
- [ ] Send heartbeat
- [ ] Stop sending heartbeats for >60s
- [ ] Verify cleanup removes server
- [ ] Try register without API key (should fail)
- [ ] Kill server process
- [ ] Restart server
- [ ] Verify registered servers persist
- [ ] Try DELETE (should require auth)
- [ ] Monitor logs during operations

### Load Testing
- [ ] Test with 100 concurrent registrations
- [ ] Test with 1000 heartbeats/sec
- [ ] Monitor memory usage
- [ ] Verify no goroutine leaks

### Security Testing
- [ ] Try to register with malicious JSON (oversized arrays, deep nesting)
- [ ] Try to bypass auth with modified headers
- [ ] Verify request body size limit works
- [ ] Test with SQL injection strings (if using DB)

---

## Deployment Checklist

Before going live:
- [ ] All CRITICAL items complete
- [ ] All HIGH items complete
- [ ] Tests pass (unit + integration)
- [ ] Load testing passed
- [ ] Security audit done
- [ ] Documentation updated
- [ ] Monitoring/alerting set up
- [ ] Rollback plan defined
- [ ] On-call rotation defined
- [ ] Incident runbook created

---

## Post-Deployment Monitoring

- [ ] Check error rates (should be ~0%)
- [ ] Check response times (<100ms p99)
- [ ] Check active server count
- [ ] Check cleanup removing stale servers
- [ ] Check auth failures (should be low)
- [ ] Monitor database size (if using SQLite)
- [ ] Monitor memory usage
- [ ] Review logs daily for first week

---

## Rollback Plan

If issues occur:
1. Stop accepting new registrations (deployment gate)
2. Switch traffic to previous version
3. Diagnose issue in logs/metrics
4. Fix code
5. Test thoroughly
6. Re-deploy

---

## Questions to Answer Before Production

- [ ] Who are the game servers? (trusted internal, untrusted external?)
- [ ] How many servers expect to register? (10, 100, 10,000?)
- [ ] How often do servers heartbeat? (10s, 30s, 60s?)
- [ ] What's acceptable downtime? (0s, 5m, 1h?)
- [ ] Do servers need to survive a restart? (yes/no)
- [ ] Where will this be deployed? (cloud, on-prem, edge?)
- [ ] Who monitors it? (on-call engineer?)
- [ ] What's the SLA? (99%, 99.9%, 99.99%)

---

**Last Updated**: 2025-12-05  
**Owner**: DevOps / Backend Team  
**Status**: In Progress
