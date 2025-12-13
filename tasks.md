# Tasks

## TODO

## In Progress

## Done

### WebSocket Migration (from template.md: "change spawner communication with master server to websocket")

1. ✅ **Remove HTTP registration loop** - Replaced `registerLoop()` in `spawner/main.go` with WebSocket-based registration
2. ✅ **Update WebSocket client registration** - Modified `spawner/internal/ws/client.go` to handle initial registration with full spawner metadata (region, host, port, max_instances, etc.)
3. ✅ **Remove unused HTTP heartbeat function** - Deleted `heartbeatLoop()` function from `spawner/main.go` (replaced by WebSocket heartbeat)
4. ✅ **Handle spawner ID assignment via WebSocket** - Master server now assigns and returns spawner ID through WebSocket registration response
5. ⚠️ **Test WebSocket connection lifecycle** - Manual testing recommended to verify reconnection, registration retry, and error handling
6. ✅ **Update master server WebSocket handler** - Updated `server/ws_manager.go` to handle initial registration with full metadata and return assigned ID
7. ✅ **Remove HTTP dependencies** - Cleaned up unused HTTP client code and imports (bytes, strconv, cpu, disk, mem, filepath) from `spawner/main.go`
