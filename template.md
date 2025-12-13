# Project Implementation Template

## Project Overview

This repository contains 3 interconnected projects:

1. **Master Server** (`server/`) - Central registry and API server
   - Manages spawner registration and tracking
   - Handles game instance spawning requests
   - Provides REST API and WebSocket endpoints
   - Optional SQLite persistence for spawner state
   - SSE (Server-Sent Events) for real-time dashboard updates
   - Authentication and session management

2. **Spawner** (`spawner/`) - Unity game server spawner and real-time manager
   - Spawns and manages multiple game server instances
   - Monitors resource usage (CPU, memory, disk)
   - Communicates with master server via WebSocket
   - Handles game server updates and version management
   - Provides local REST API for instance management
   - Persists instance state across restarts

3. **Web Dashboard** (`web-dashboard/`) - SvelteKit-based web dashboard
   - Real-time monitoring and control interface
   - Communicates with master server via REST API and SSE
   - Manages spawners, instances, and configurations
   - Performance metrics visualization
   - User authentication and session management

## Systems & Architecture

### Master Server Systems
- **Registry System**: In-memory spawner registry with optional SQLite persistence
- **WebSocket Manager**: Manages spawner connections, command routing, and responses
- **SSE Hub**: Real-time dashboard updates via Server-Sent Events
- **Authentication**: Session-based auth with API key support
- **Database**: Optional SQLite for spawner persistence and instance action logging
- **REST API**: HTTP endpoints for spawner management, instance spawning, config management
- **File Upload**: Game server version management and distribution

### Spawner Systems
- **Game Manager**: Lifecycle management of game server instances (spawn, start, stop, remove)
- **WebSocket Client**: Persistent connection to master server for registration, heartbeat, and commands
- **Resource Monitoring**: CPU, memory, disk usage tracking per instance
- **Update System**: Automatic game server version updates from master server
- **Persistence**: Instance state persistence across restarts
- **REST API**: Local HTTP API for instance management (protected by API key)

### Web Dashboard Systems
- **Real-time Updates**: SSE connection for live spawner/instance status
- **State Management**: Svelte stores for global state
- **Component Library**: Reusable UI components (charts, tables, modals, etc.)
- **Routing**: SvelteKit file-based routing
- **Authentication**: Login page with session management

## UI Components (web-dashboard)

### Core Components
- `ConfirmDialog.svelte` - Confirmation dialogs
- `Drawer.svelte` - Side drawer navigation
- `Dropdown.svelte` - Dropdown menus
- `StatsCard.svelte` - Statistics display cards
- `Terminal.svelte` - Terminal/console output viewer

### Spawner Management
- `SpawnerTable.svelte` - Main spawner listing table
- `InstanceManagerModal.svelte` - Instance management modal
- `InstanceRow.svelte` - Individual instance row component

### Monitoring & Metrics
- `ResourceMetricsPanel.svelte` - Resource metrics display
- `ResourceStatsCard.svelte` - Resource statistics cards
- `ResourceProgressBar.svelte` - Progress bars for resources
- `ResourceHistoryChart.svelte` - Historical resource usage charts
- `PlayersChart.svelte` - Player count charts
- `TopResourceConsumers.svelte` - Top resource consuming instances
- `LogViewer.svelte` - Log file viewer

## Packages & Dependencies

### Master Server (Go)
- `github.com/gorilla/mux` - HTTP router
- `github.com/gorilla/websocket` - WebSocket support
- `github.com/jmoiron/sqlx` - SQL database access
- `modernc.org/sqlite` - SQLite driver
- `golang.org/x/crypto` - Password hashing (bcrypt)

### Spawner (Go)
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/gorilla/websocket` - WebSocket client
- `github.com/joho/godotenv` - Environment variable loading
- `github.com/shirou/gopsutil/v3` - System and process utilities

### Web Dashboard (TypeScript/Svelte)
- `@sveltejs/kit` - SvelteKit framework
- `svelte` - Svelte UI framework
- `tailwindcss` - CSS framework
- `lucide-svelte` - Icon library
- `vitest` - Testing framework
- `typescript` - TypeScript support

## Routes & Pages (web-dashboard)

- `/` - Main dashboard page
- `/login` - Authentication page
- `/server` - Server management
- `/spawners` - Spawner listing
- `/spawners/[id]` - Individual spawner details
- `/config` - Configuration management
- `/config/[category]` - Category-specific config
- `/performance` - Performance metrics
- `/test` - Test page

## Features to Implement

## In Progress

## Completed

- ✅ **WebSocket Migration** - Changed spawner communication with master server from HTTP to WebSocket
  - Master server now accepts full registration metadata via WebSocket
  - Spawner registers with full metadata (region, host, port, max_instances, etc.)
  - Removed HTTP registration loop and heartbeat functions
  - WebSocket handles registration, heartbeat, and command routing

## Pre-Task Checklist

Before starting work on a new task, check:

1. **Read template.md** - Understand current project state and systems
2. **Read tasks.md** - Check existing tasks and priorities
3. **Check git status** - Ensure clean working directory
4. **Review related code** - Understand affected systems/components
5. **Check dependencies** - Verify required packages are available
6. **Test current state** - Ensure existing functionality works
7. **Update tasks.md** - Move task to "In Progress" when starting
8. **Update template.md** - Move feature to "In Progress" or "Completed" when done

## Notes
