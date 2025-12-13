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

### Instance Management via WebSocket

1. ✅ **Add list_instances command handler** - Added `list_instances` handler in `spawner/internal/ws/client.go` to return instances via WebSocket
2. ✅ **Migrate ListSpawnerInstances to WebSocket** - Updated `ListSpawnerInstances` in `server/handlers.go` to use WebSocket instead of HTTP (fixes 401 error)
3. ✅ **Verify SpawnInstance uses WebSocket** - Confirmed `SpawnInstance` already uses WebSocket correctly

### Complete Handler Migration to WebSocket

1. ✅ **Add all WebSocket handlers in spawner** - Added handlers for: restart_instance, get_instance_stats, get_instance_history, update_instance, rename_instance, remove_instance, backup_instance, restore_instance, list_backups, delete_backup, update_template, get_logs, clear_logs, get_instance_logs, clear_instance_logs
2. ✅ **Migrate RestartSpawnerInstance to WebSocket** - Updated to use WebSocket
3. ✅ **Migrate GetInstanceStats to WebSocket** - Updated to use WebSocket
4. ✅ **Migrate GetInstanceHistory to WebSocket** - Updated to use WebSocket
5. ✅ **Migrate UpdateSpawnerInstance to WebSocket** - Updated to use WebSocket
6. ✅ **Migrate RenameSpawnerInstance to WebSocket** - Updated to use WebSocket
7. ✅ **Migrate RemoveSpawnerInstance to WebSocket** - Updated to use WebSocket
8. ✅ **Migrate BackupSpawnerInstance to WebSocket** - Updated to use WebSocket
9. ✅ **Migrate RestoreSpawnerInstance to WebSocket** - Updated to use WebSocket
10. ✅ **Migrate ListSpawnerBackups to WebSocket** - Updated to use WebSocket
11. ✅ **Migrate DeleteSpawnerBackup to WebSocket** - Updated to use WebSocket
12. ✅ **Migrate UpdateSpawnerTemplate to WebSocket** - Updated to use WebSocket
13. ✅ **Migrate GetSpawnerLogs to WebSocket** - Updated to use WebSocket
14. ✅ **Migrate ClearSpawnerLogs to WebSocket** - Updated to use WebSocket
15. ✅ **Migrate GetInstanceLogs to WebSocket** - Updated to use WebSocket (returns full content, streaming can be added later)
16. ✅ **Migrate ClearInstanceLogs to WebSocket** - Updated to use WebSocket
17. ✅ **Clean up unused imports** - Removed unused `io` import from handlers.go
