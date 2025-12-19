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
*   **Database:** Integrated PostgreSQL support (via `pgx` driver) alongside SQLite. Added "Databases" tab to Dashboard for table listing. Use `DB_DRIVER=pgx` for Postgres. Added 5s timeout to DB init to prevent startup hangs. Enhanced Database Management: Full UI for Overview, Table Browser/Editor (CRUD), Internal Backups (Create/Restore/Download), and Config viewing.
*   **Dashboard Stability:** Fixed Svelte 5 reactivity issues in Layout and Dashboard Page (Runes migration) resolving navigation bugs and blank screen issues. Fixed accessibility warnings and template syntax errors.
*   **Performance:** Refactored Spawner metrics collection to be asynchronous, preventing I/O blocks from delaying heartbeats.

## 🔒 Security Note
*   **Transport Encryption:** The Master Server runs on plain HTTP (`:8081`). For production security, you **MUST** run it behind a Reverse Proxy (Nginx, Caddy, Cloudflare) that handles SSL/TLS termination. This ensures the WebSocket connection and API Key are encrypted.
*   **Authentication:** All sensitive endpoints are protected by `X-API-Key` (Spawners) or Session Cookie (Dashboard).
*   **Security:** Removed insecure logging of 2FA secrets in `server`.
*   **Dashboard Updates:** Added "Start" button, "Node Logs" tab (embedded `LogViewer`), and switched Console logs to polling in `InstanceManagerModal`.

## 🛡️ Security Hardening (Recent)
The following security improvements have been implemented:

### SQL Injection Prevention
*   **Input Validation:** All SQL identifiers (schema, table, column, function names) are validated via `security.go` helpers.
*   **Type Whitelist:** SQL types are validated against a strict whitelist of allowed PostgreSQL types.
*   **Parameterized Queries:** Function execution uses parameterized queries (`$1`, `$2`, etc.) instead of string concatenation.
*   **Role Creation:** Uses PostgreSQL's `format()` function with `%I` (identifier) and `%L` (literal) for safe escaping.
*   **Read-Only SQL:** The `ExecuteSQLHandler` now enforces read-only queries (SELECT, WITH, EXPLAIN only).

### WebSocket Security
*   **Origin Validation:** WebSocket connections now validate the `Origin` header against an allowlist.
*   **Custom Origins:** Set `ALLOWED_ORIGINS` environment variable for production domains (comma-separated).

### XSS Prevention
*   **HTML Sanitization:** Frontend components using `{@html}` now sanitize input, allowing only safe formatting tags.

### Rate Limiting & IP Handling
*   **Proxy Support:** Rate limiting now uses `X-Forwarded-For` and `X-Real-IP` headers when behind a reverse proxy.
*   **IP Validation:** Firewall blocking validates IPs to prevent blocking localhost or critical addresses.

### Path Traversal Prevention
*   **Filename Validation:** All file operations validate filenames for `..`, `/`, `\`, and shell metacharacters.

### Error Sanitization
*   **Safe Errors:** Database errors are sanitized to prevent leaking sensitive information (paths, credentials).

### Password Security
*   **Complexity Requirements:** Role passwords require 12+ characters with uppercase, lowercase, digits, and special characters.
*   **Dangerous Options Blocked:** `SUPERUSER`, `REPLICATION`, and `BYPASSRLS` role options are blocked.
*   **Security Definer Blocked:** Functions cannot use `SECURITY DEFINER` to prevent privilege escalation.

### Frontend Cleanup
*   **Removed Component:** Removed `LucideIconWrapper.svelte` as it was redundant.
*   **LogViewer Refactor:** Updated `LogViewer.svelte` to use Svelte 5 `{@const Icon = tab.icon}` syntax for dynamic icon rendering, fixing deprecation warnings.

### 🔒 Security Enhancements (Post-Audit)
*   **Spawner RCE Patch:** Implemented strict ID validation in `spawner/internal/game/manager.go` to prevent path traversal vulnerability in `RenameInstance` and `Spawn`.
*   **Security Headers:** Added `SecurityHeadersMiddleware` in Server to enforce `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, and `Content-Security-Policy`.
*   **Server Binding:** Updated `server/main.go` to bind to `127.0.0.1` by default (configurable via `SERVER_HOST`).
*   **Read-Only Database:** Added support for a separate `READONLY_DB_DSN` in `server/db.go` and `server/main.go`, used by `ExecuteSQLHandler` for safer ad-hoc queries.
*   **Frontend Sanitization:** Integrated `DOMPurify` in `web-dashboard` (`StatsCard.svelte`, `ConfirmDialog.svelte`) to replace regex-based sanitization.
