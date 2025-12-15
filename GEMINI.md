# Exile Project - Gemini Context

This document provides a comprehensive overview of the **Exile** project structure, architecture, and development workflows. Use this context to assist with code navigation, refactoring, and feature implementation.

## 📂 Project Structure

The project is a multi-component system for managing game server instances, consisting of a central Master Server (`server`), distributed Spawners (`spawner`), and a Web Dashboard (`web-dashboard`).

```
/mnt/c/Users/pab/Desktop/goExile/
├── server/                 # Master Server (Registry, API, Auth, WebSocket)
│   ├── database/           # SQLite database files
│   ├── files/              # File storage for game server binaries
│   ├── main.go             # Entry point
│   └── ...                 # Handlers, Middleware, Auth logic
├── spawner/                # Spawner Service (Manages Game Instances)
│   ├── api/                # API handlers for Spawner-Server comms
│   ├── game_server/        # Local game server binaries
│   ├── internal/           # Internal logic (Game Manager, Updater)
│   ├── main.go             # Entry point
│   └── ...
├── web-dashboard/          # SvelteKit Frontend
│   ├── src/                # Source code
│   │   ├── lib/            # Shared components and utilities
│   │   └── routes/         # SvelteKit pages/routes
│   ├── package.json        # Frontend dependencies
│   └── ...
└── game_server/            # (Optional) Standalone game server data
```

## 🏗️ Architecture

### 1. Master Server (`server`)
*   **Role:** Central registry and orchestrator.
*   **Tech Stack:** Go (Standard Library + `gorilla/mux` likely), SQLite.
*   **Key Responsibilities:**
    *   **Registry:** Tracks active Spawners via HTTP/WebSocket registration and heartbeats.
    *   **Authentication:** Handles Admin login (Email/Password + TOTP 2FA) and Spawner auth (`X-API-Key`).
    *   **Dashboard API:** Provides REST endpoints and SSE (Server-Sent Events) for real-time frontend updates.
    *   **File Serving:** Hosts `game_server.zip` for Spawners to download.
    *   **WebSocket:** Manages real-time communication with Spawners and potentially the frontend.

### 2. Spawner (`spawner`)
*   **Role:** Node agent that runs on remote machines to spawn game servers.
*   **Tech Stack:** Go.
*   **Key Responsibilities:**
    *   **Instance Management:** Starts, stops, and monitors game server processes.
    *   **Communication:** Registers with Master Server, sends heartbeats, and receives spawn commands via WebSocket/HTTP.
    *   **Self-Update:** Downloads and updates the game server binary from the Master Server.
    *   **Resource Monitoring:** Tracks CPU/RAM usage to report to Master.

### 3. Web Dashboard (`web-dashboard`)
*   **Role:** Admin UI for monitoring and management.
*   **Tech Stack:** Svelte 5, SvelteKit, TypeScript, TailwindCSS, Vite.
*   **Key Features:**
    *   Real-time stats (SSE).
    *   Spawner management (list, status).
    *   Instance control (Start/Stop/Restart/Logs).
    *   Authentication flows (Login, 2FA).

## 🚀 Building & Running

### Server (Master)
```bash
cd server
# Setup .env (see server/.env.example)
go run .
# Runs on http://localhost:8081 (default)
```

### Spawner
```bash
cd spawner
# Setup config/env
go run .
```

### Web Dashboard
```bash
cd web-dashboard
pnpm install  # or npm install
npm run dev   # Starts Vite dev server
# Runs on http://localhost:5173 (proxies to backend likely configured in vite.config.ts)
```

## 🧪 Development Conventions

*   **Go:** Follow standard Go idioms. Use `go fmt`.
    *   Error handling: Return errors, don't panic (except in main startup).
    *   Logging: Use structured logging (`slog`) where possible or standard `log`.
*   **Frontend (Svelte):**
    *   Use TypeScript for all new components (`.svelte` script lang="ts").
    *   Styling: TailwindCSS utility classes.
    *   State Management: Svelte Stores (`src/lib/stores.ts`) or Runes (Svelte 5).
    *   Components: Located in `src/lib/components`.
*   **Testing:**
    *   Go: `go test ./...`
    *   Frontend: `npm run check` (Svelte Check), `npm run test` (Vitest).

## 🔑 Key Configuration

*   **`server/.env`**: Controls Master Server port, database path, admin credentials, and 2FA secrets.
*   **`X-API-Key`**: Critical security header for Spawner <-> Server communication.
*   **`spawner/internal/ws/client.go`**: Handles Spawner's WebSocket connection logic (Heartbeats, Command handling).
*   **`server/dashboard.go` & `server/auth.go`**: Core logic for Dashboard API and Authentication.

## 📝 Recent Context
*   **Fixed Spawner Panic:** Resolved concurrent write panic in `spawner/internal/ws/client.go` by implementing a `writePump`.
*   **Fixed Heartbeat:** Improved `heartbeatLoop` resilience in Spawner.
*   **Stability:** Increased Server heartbeat timeout thresholds and ensured connection keep-alive on any message.
*   **Status Logic:** Disabled time-based status degradation ("Unresponsive"). Status is now strictly "Online" (connected) or "Offline" (disconnected).
*   **Dashboard UI:** Fixed Spawner status mismatch ("active" vs "Online") enabling the spawn button; removed verbose heartbeat logging.
*   **Logging:** Implemented log rotation (max 5MB, keep 20 lines) in Spawner, filtered noisy log types, and added file size display in Dashboard LogViewer.
*   **Performance:** Refactored Spawner metrics collection to be asynchronous, preventing I/O blocks from delaying heartbeats.

## 🔒 Security Note
*   **Transport Encryption:** The Master Server runs on plain HTTP (`:8081`). For production security, you **MUST** run it behind a Reverse Proxy (Nginx, Caddy, Cloudflare) that handles SSL/TLS termination. This ensures the WebSocket connection and API Key are encrypted.
*   **Authentication:** All sensitive endpoints are protected by `X-API-Key` (Spawners) or Session Cookie (Dashboard).
*   **Security:** Removed insecure logging of 2FA secrets in `server`.
*   **Dashboard Updates:** Added "Start" button, "Node Logs" tab (embedded `LogViewer`), and switched Console logs to polling in `InstanceManagerModal`.
