# Tasks

## TODO

## In Progress

## Done

### WebSocket Migration (from template.md: "change node communication with master server to websocket")

1. ✅ **Remove HTTP registration loop** - Replaced `registerLoop()` in `node/main.go` with WebSocket-based registration
2. ✅ **Update WebSocket client registration** - Modified `node/internal/ws/client.go` to handle initial registration with full node metadata (region, host, port, max_instances, etc.)
3. ✅ **Remove unused HTTP heartbeat function** - Deleted `heartbeatLoop()` function from `node/main.go` (replaced by WebSocket heartbeat)
4. ✅ **Handle node ID assignment via WebSocket** - Master server now assigns and returns node ID through WebSocket registration response
5. ⚠️ **Test WebSocket connection lifecycle** - Manual testing recommended to verify reconnection, registration retry, and error handling
6. ✅ **Update master server WebSocket handler** - Updated `server/ws_manager.go` to handle initial registration with full metadata and return assigned ID
7. ✅ **Remove HTTP dependencies** - Cleaned up unused HTTP client code and imports (bytes, strconv, cpu, disk, mem, filepath) from `node/main.go`

### Instance Management via WebSocket

1. ✅ **Add list_instances command handler** - Added `list_instances` handler in `node/internal/ws/client.go` to return instances via WebSocket
2. ✅ **Migrate ListNodeInstances to WebSocket** - Updated `ListNodeInstances` in `server/handlers.go` to use WebSocket instead of HTTP (fixes 401 error)
3. ✅ **Verify SpawnInstance uses WebSocket** - Confirmed `SpawnInstance` already uses WebSocket correctly

### Complete Handler Migration to WebSocket

1. ✅ **Add all WebSocket handlers in node** - Added handlers for: restart_instance, get_instance_stats, get_instance_history, update_instance, rename_instance, remove_instance, backup_instance, restore_instance, list_backups, delete_backup, update_template, get_logs, clear_logs, get_instance_logs, clear_instance_logs
2. ✅ **Migrate RestartNodeInstance to WebSocket** - Updated to use WebSocket
3. ✅ **Migrate GetInstanceStats to WebSocket** - Updated to use WebSocket
4. ✅ **Migrate GetInstanceHistory to WebSocket** - Updated to use WebSocket
5. ✅ **Migrate UpdateNodeInstance to WebSocket** - Updated to use WebSocket
6. ✅ **Migrate RenameNodeInstance to WebSocket** - Updated to use WebSocket
7. ✅ **Migrate RemoveNodeInstance to WebSocket** - Updated to use WebSocket
8. ✅ **Migrate BackupNodeInstance to WebSocket** - Updated to use WebSocket
9. ✅ **Migrate RestoreNodeInstance to WebSocket** - Updated to use WebSocket
10. ✅ **Migrate ListNodeBackups to WebSocket** - Updated to use WebSocket
11. ✅ **Migrate DeleteNodeBackup to WebSocket** - Updated to use WebSocket
12. ✅ **Migrate UpdateNodeTemplate to WebSocket** - Updated to use WebSocket
13. ✅ **Migrate GetNodeLogs to WebSocket** - Updated to use WebSocket
14. ✅ **Migrate ClearNodeLogs to WebSocket** - Updated to use WebSocket
15. ✅ **Migrate GetInstanceLogs to WebSocket** - Updated to use WebSocket (returns full content, streaming can be added later)
16. ✅ **Migrate ClearInstanceLogs to WebSocket** - Updated to use WebSocket
17. ✅ **Clean up unused imports** - Removed unused `io` import from handlers.go

### UI/UX Animations & Enhancements

1. ✅ **Fix event handler syntax consistency** - Updated all `on:click` to `onclick` in server page for Svelte 5 compatibility
2. ✅ **Enhance dashboard with particle animations** - Added floating background particles, grid patterns, and gradient effects
3. ✅ **Implement staggered card animations** - Sequential loading animations for dashboard stat cards with 0.1s intervals
4. ✅ **Add 3D tilting effects to server cards** - Implemented perspective transforms and rotation on hover for version history cards
5. ✅ **Create floating action buttons** - Added animate-in activate/delete buttons that appear on card hover
6. ✅ **Enhance sidebar navigation animations** - Added icon animations, gradient backgrounds, and smooth transitions
7. ✅ **Implement glassmorphism effects** - Added backdrop blur and transparent gradient overlays throughout UI
8. ✅ **Add tech-inspired backgrounds** - Created data grid patterns and animated gradient orbs
9. ✅ **Fix PostCSS @apply errors** - Replaced @apply directives with pure Tailwind CSS classes
10. ✅ **Optimize animation performance** - Used transform-gpu and proper z-index layering for smooth animations
