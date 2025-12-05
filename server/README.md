# Game Server Registry API

A lightweight, thread-safe HTTP API service that manages registration, discovery, and health monitoring of multiplayer game server instances. Unity game server instances register themselves, send periodic heartbeats, and clients can discover available servers.

---

## Overview

This is a **server registry microservice** designed for multiplayer game backends. It:

1. **Accepts server registrations** from running Unity game server instances
2. **Stores server metadata** (host, port, region, max players, etc.) in memory
3. **Tracks server health** via periodic heartbeat pings from each server
4. **Allows clients to discover servers** via a list/get API
5. **Auto-cleans expired servers** that stop sending heartbeats (assume they crashed)
6. **Validates all input** to prevent bad data and attacks
7. **Runs a background cleanup goroutine** that periodically removes stale servers

---

## Architecture

### Data Flow

```
Unity Server 1 ──────────┐
                         ├──> [Server Registry API] ──> [In-Memory Map]
Unity Server 2 ──────────┤         (Go HTTP Server)      (Thread-Safe)
Unity Server 3 ──────────┘
                                        ↓
                                   
Game Client ──────────────────────> [Discover Servers]
                                   [Connect to Best One]
```

### Core Components

#### 1. **Server Struct**
```go
type Server struct {
    ID             int       // Unique server ID (auto-assigned by registry)
    Name           string    // Human-readable name (e.g., "Room-A")
    Host           string    // Public IP or hostname
    Port           int       // Port number (1-65535)
    MaxPlayers     int       // Server capacity
    CurrentPlayers int       // How many players connected now
    Region         string    // Geographic region (e.g., "eu", "us-west")
    Status         string    // "active" or "expired"
    LastSeen       time.Time // UTC timestamp of last heartbeat
}
```

#### 2. **Registry (Thread-Safe Storage)**
```go
type Registry struct {
    mu     sync.RWMutex        // Ensures thread-safe access
    nextID int                 // Counter for auto-increment ID
    items  map[int]*Server     // In-memory key-value store
}
```

**Why `sync.RWMutex`?**
- Multiple goroutines (one per HTTP request) access the registry simultaneously
- Read locks allow many readers (list/get) to work at once
- Write locks ensure register/delete/heartbeat don't race

#### 3. **Background Cleanup Goroutine**
- Runs every 30 seconds (configurable via `cleanupInterval`)
- Checks each server's `LastSeen` timestamp
- If older than 60 seconds (configurable via `serverTTL`), **removes it**
- Logs when a server is removed
- **Prevents stale servers** from piling up if a server crashes without graceful shutdown

---

## Endpoints

### 1. Register a Server
**Request:**
```
POST /api/servers
Content-Type: application/json

{
  "name": "room-a",
  "host": "192.168.1.100",
  "port": 7777,
  "max_players": 8,
  "current_players": 0,
  "region": "eu"
}
```

**Response (201 Created):**
```json
{
  "id": 1
}
```

**What happens:**
- Validates host and port (required, port must be 1-65535)
- Validates max_players (must be 1-10000)
- Auto-assigns unique ID (starting from 1, increments)
- Sets status to "active"
- Records current UTC timestamp in LastSeen
- Stores server in memory map
- Returns the server ID for future requests

**Errors:**
- `400 Bad Request`: Missing/invalid host or port
- `400 Bad Request`: Invalid max_players value
- `400 Bad Request`: Malformed JSON or unknown fields

---

### 2. Heartbeat (Keep-Alive)
**Request:**
```
POST /api/servers/{id}/heartbeat
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

**What happens:**
- Server sends this every ~10-30 seconds to stay registered
- Updates the `LastSeen` timestamp to **now (UTC)**
- Marks status as "active"
- If server doesn't exist, returns 404

**Why?**
- Tells the registry "I'm still alive and running"
- If server doesn't send heartbeat for 60 seconds, cleanup removes it
- Simple but effective crash detection (no active health check needed)

**Errors:**
- `404 Not Found`: Server ID doesn't exist
- `400 Bad Request`: Invalid server ID format

---

### 3. List All Servers
**Request:**
```
GET /api/servers
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "room-a",
    "host": "192.168.1.100",
    "port": 7777,
    "max_players": 8,
    "current_players": 3,
    "region": "eu",
    "status": "active",
    "last_seen": "2025-12-05T17:25:12.345Z"
  },
  {
    "id": 2,
    "name": "room-b",
    "host": "192.168.1.101",
    "port": 7777,
    "max_players": 8,
    "current_players": 6,
    "region": "eu",
    "status": "active",
    "last_seen": "2025-12-05T17:25:05.123Z"
  }
]
```

**What happens:**
- Returns **all active servers** as a JSON array
- Each server includes all metadata
- Useful for game clients to discover available game rooms
- Empty array if no servers registered

---

### 4. Get Single Server
**Request:**
```
GET /api/servers/{id}
```

**Example:**
```
GET /api/servers/1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "room-a",
  "host": "192.168.1.100",
  "port": 7777,
  "max_players": 8,
  "current_players": 3,
  "region": "eu",
  "status": "active",
  "last_seen": "2025-12-05T17:25:12.345Z"
}
```

**What happens:**
- Returns metadata for a specific server
- Used when a client wants details before joining
- Can check player count, region, etc.

**Errors:**
- `404 Not Found`: Server ID doesn't exist
- `400 Bad Request`: Invalid server ID format

---

### 5. Delete/Unregister a Server
**Request:**
```
DELETE /api/servers/{id}
```

**Response (200 OK):**
```json
{
  "message": "deleted"
}
```

**What happens:**
- Removes server from registry immediately
- Used for graceful shutdown (server tells registry it's going down)
- Prevents cleanup goroutine from having to wait 60 seconds

**Errors:**
- `404 Not Found`: Server ID doesn't exist
- `400 Bad Request`: Invalid server ID format

---

### 6. Health Check
**Request:**
```
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy"
}
```

**What happens:**
- Simple endpoint to verify the registry service is running
- Used by load balancers, monitoring systems, Docker health checks

---

## Safety Features

### 1. **Input Validation**
- **JSON parsing**: Uses `decoder.DisallowUnknownFields()` to reject unknown fields
- **Request body size limit**: Max 1MB per request (prevents DoS)
- **Host/Port validation**: Host must be non-empty, port must be 1-65535
- **Max players validation**: Must be 1-10000
- **ID validation**: Must be 1-1000000

### 2. **Thread Safety**
- All registry operations use `sync.RWMutex`
- Read operations (List, Get) use read locks (many can run concurrently)
- Write operations (Register, Heartbeat, Delete) use write locks (exclusive)
- No data races or concurrent access bugs

### 3. **Proper Error Handling**
- All errors are caught and returned as JSON responses
- Proper HTTP status codes (201, 400, 404, 500)
- Client gets clear error messages

### 4. **Resource Management**
- Cleanup goroutine runs in background, removes stale servers automatically
- Memory map never grows unbounded (expired servers removed)
- Request bodies limited to 1MB

---

## Constants & Configuration

```go
const (
    maxBodySize     = 1 << 20        // 1 MB - max request body size
    maxIDValue      = 1000000        // Max server ID allowed
    serverTTL       = 60 * time.Second      // Server expires if no heartbeat for 60s
    cleanupInterval = 30 * time.Second     // Cleanup runs every 30s
)
```

**To change these:**
- Increase `serverTTL` if servers need longer between heartbeats (e.g., 120 seconds for unstable networks)
- Decrease `cleanupInterval` if you want faster detection of dead servers (e.g., 15 seconds)
- Increase `maxBodySize` if servers send large metadata

---

## Execution Flow (Step-by-Step)

### Server Startup
1. `main()` calls `run()`
2. `run()` creates HTTP router (Gorilla Mux)
3. Registers 6 route handlers (register, list, get, delete, heartbeat, health)
4. **Spawns background cleanup goroutine** with `go registry.Cleanup(...)`
5. Starts HTTP server on `:8081`
6. Cleanup goroutine runs forever, checking every 30 seconds

### When Unity Server Registers
1. POST request arrives at `/api/servers`
2. Handler decodes JSON body
3. Validates host, port, max_players
4. Lock registry write lock
5. Assign new ID from `nextID` counter
6. Create Server struct with ID, current time as LastSeen, status = "active"
7. Add to `registry.items[id]`
8. Unlock
9. Return `{"id": 1}`

### When Heartbeat Arrives
1. POST request to `/api/servers/{id}/heartbeat`
2. Parse ID from URL
3. Lock registry write lock
4. Find server by ID
5. Update `LastSeen = time.Now().UTC()`
6. Unlock
7. Return `{"status": "ok"}`

### When Client Lists Servers
1. GET request to `/api/servers`
2. Lock registry read lock
3. Iterate all servers, copy to slice
4. Unlock
5. Return JSON array

### Every 30 Seconds (Cleanup)
1. Calculate cutoff = now - 60 seconds
2. Lock registry write lock
3. For each server:
   - If LastSeen < cutoff, remove it (server didn't heartbeat)
4. Unlock
5. Sleep 30 seconds, repeat

---

## Usage Examples

### Quick Start
```bash
cd server
go mod tidy
go run main.go
```

### Test with cURL

**Register server:**
```bash
curl -X POST http://localhost:8081/api/servers \
  -H "Content-Type: application/json" \
  -d '{"name":"room-1","host":"192.168.1.10","port":7777,"max_players":8,"current_players":0,"region":"eu"}'

# Returns: {"id":1}
```

**Send heartbeat:**
```bash
curl -X POST http://localhost:8081/api/servers/1/heartbeat

# Returns: {"status":"ok"}
```

**List all servers:**
```bash
curl http://localhost:8081/api/servers

# Returns: [{"id":1, "name":"room-1", ...}]
```

**Get specific server:**
```bash
curl http://localhost:8081/api/servers/1

# Returns: {"id":1, "name":"room-1", ...}
```

**Delete server:**
```bash
curl -X DELETE http://localhost:8081/api/servers/1

# Returns: {"message":"deleted"}
```

**Health check:**
```bash
curl http://localhost:8081/health

# Returns: {"status":"healthy"}
```

---

## Typical Workflow

### Game Server Instance (Unity)
1. Game server starts
2. **Register** with API: `POST /api/servers` → gets ID = 5
3. Enter game loop
4. **Every 10-30 seconds**: Send `POST /api/servers/5/heartbeat`
5. Game server shutsdown gracefully: `DELETE /api/servers/5`
6. OR crash → no heartbeat → cleanup removes it after 60 seconds

### Game Client
1. Launches game
2. **Discover servers**: `GET /api/servers`
3. Shows list to player (room name, players, region, etc.)
4. Player clicks "Join Room 5"
5. Client connects directly to server at `192.168.1.10:7777`
6. Done! Registry job is complete

---

## Key Design Decisions

| Decision | Why |
|----------|-----|
| **In-memory storage** | Fast, simple for low player counts. No DB latency. |
| **Thread-safe with RWMutex** | Handles concurrent HTTP requests safely. |
| **Heartbeat-based health** | No need for active health checks; simple and reliable. |
| **Auto-cleanup goroutine** | Servers can crash; we don't want zombie entries forever. |
| **JSON validation** | Prevents malformed data and accidental field typos. |
| **Request size limits** | Prevents memory exhaustion attacks. |
| **Proper HTTP codes** | Clients can distinguish 404 (not found) from 400 (bad request). |

---

## Limitations & Future Improvements

### Current Limitations
- **In-memory only**: Restarts lose all server data
- **Single instance**: No horizontal scaling
- **No authentication**: Any client can register/delete servers
- **No filtering**: Clients see all servers; no region/capacity filtering yet

### Next Steps (Optional)
1. **Add persistence**: Use SQLite or Redis to survive restarts
2. **Add authentication**: API key or OAuth for server registration
3. **Add filtering**: Query params like `?region=eu&min_players=2`
4. **Add metrics**: Prometheus counters for registrations, heartbeats
5. **Horizontal scale**: Run multiple instances behind a load balancer, use Redis for shared registry

---

## Summary

This server registry is a **microservice that solves multiplayer server discovery**. It's:

- ✅ **Fast** (in-memory, no database)
- ✅ **Safe** (validated inputs, thread-safe)
- ✅ **Reliable** (auto-cleanup, heartbeat-based health)
- ✅ **Simple** (~250 lines of code)
- ✅ **Production-ready** (proper error handling, logging, HTTP standards)

Perfect for indie game backends, game jams, or learning multiplayer architectures!
