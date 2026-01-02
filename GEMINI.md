# Exile Project - Gemini Context

This document provides a comprehensive overview of the **Exile** project structure, architecture, and development workflows. Use this context to assist with code navigation, refactoring, and feature implementation.

## 📂 Project Structure

The project is a multi-component system for managing game server instances, consisting of a central Master Server (`server`), distributed Nodes (`node`), and a Web Dashboard (`web-dashboard`).

```
/mnt/c/Users/pab/Desktop/goExile/
├── server/                 # Master Server (Registry, API, Auth, WebSocket)
│   ├── database/           # SQLite database files
│   ├── files/              # File storage for game server binaries
│   ├── main.go             # Entry point
│   └── ...                 # Handlers, Middleware, Auth logic
├── node/                # Node Service (Manages Game Instances)
│   ├── api/                # API handlers for Node-Server comms
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

### 🛠️ Unity Game Server Packer (`packer/`)
*   **Role:** High-performance build utility for packaging Unity Game Servers.
*   **Tech Stack:** Go, Bubble Tea (TUI).
*   **Key Features:**
    *   **Modern TUI:** Real-time progress bars, spinners, and build summaries.
    *   **Manifest Generation:** Automatically creates a `manifest.json` with SHA256 hashes for integrity verification and differential updates.
    *   **Smart Versioning:** Auto-detects version from `build_info.json`.
    *   **Dry Run Mode:** Simulates packing to verify `.packerignore` rules.
    *   **Performance:** Streaming SHA256 hashing and buffered ZIP writing for maximum I/O efficiency.
    *   **Customizable:** Supports `.packerignore` files and granular CLI flags (`-out`, `-quiet`, `-compression`).

## 🏗️ Architecture

### 1. Master Server (`server`)
*   **Role:** Central registry and orchestrator.
*   **Tech Stack:** Go (Standard Library + `gorilla/mux` likely), SQLite.
*   **Key Responsibilities:**
    *   **Registry:** Tracks active Nodes via HTTP/WebSocket registration and heartbeats.
    *   **Authentication:** Handles Admin login (Email/Password + TOTP 2FA) and Node auth (`X-API-Key`).
    *   **Dashboard API:** Provides REST endpoints and SSE (Server-Sent Events) for real-time frontend updates.
    *   **File Serving:** Hosts `game_server.zip` for Nodes to download.
    *   **WebSocket:** Manages real-time communication with Nodes and potentially the frontend.

### 2. Node (`node`)
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
    *   Node management (list, status).
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

### Node
```bash
cd node
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
*   **Workflow Requirements:**
    *   Always update `GEMINI.md` with significant changes.
    *   Commit changes only if: tests pass, `npm run check` passes in `web-dashboard`, and server/node build without errors.
    *   **Client API Enforcement:** Whenever making changes that affect the game client (API endpoints, WebSocket protocol, data models), you **MUST** update `server/docs/CLIENT_API.md` and provide/update C# example scripts in `server/docs/unity/` to ensure frontend/unity alignment.

## 🔑 Key Configuration

*   **`server/.env`**: Controls Master Server port, database path, admin credentials, and 2FA secrets.
*   **`PRODUCTION_MODE`**: Set to `true` to enable TOTP 2FA. In development (default), TOTP is bypassed for faster login.
*   **`X-API-Key`**: Critical security header for Node <-> Server communication.
*   **`X-Game-API-Key`**: Header required for all Game Client -> Server API requests.
*   **`node/internal/ws/client.go`**: Handles Node's WebSocket connection logic (Heartbeats, Command handling).
*   **`server/dashboard.go` & `server/auth.go`**: Core logic for Dashboard API and Authentication.

*   **Developer Experience:**
    *   Completely overhauled the **Makefile** for better DX:
        *   Added colorized output for better readability.
        *   Implemented a comprehensive `help` command.
        *   Added `dev-*` commands to quickly start components.
        *   Streamlined build, test, and linting workflows for all components.
        *   Added a `check-all` command for full pre-commit verification.
*   **Dashboard UI:**
    *   Enhanced **SystemTopology** component:
        *   Resized RedEye and Database nodes to be smaller than the Master node for better visual hierarchy.
        *   Replaced hexagonal frames with custom SVG "Cyber Frames" featuring corner brackets.
        *   Upgraded interception animation from plasma bolts to realistic rockets with flickering engine flames and smoke trails.
        *   Added a cinematic "destruction" sequence where the attacker icon shatters and fades upon impact.
*   **Bug Fixes:**
    *   Fixed a syntax error and missing `fmt` import in `server/auth/auth.go`.
    *   Removed an unreachable and incorrect `os.Chdir()` call in `node/main.go`.
    *   **Master Server Tests:** Fixed failing tests in `server` by adding SQLite support for in-memory testing and initializing the database in `main_test.go`.
    *   **Registry Sync:** Fixed `RegisterNode` handler to correctly sync with the in-memory registry, resolving 404 errors in tests.
*   **Node Linting & Robustness:**        *   Resolved `copylocks` issues in `Instance` struct by implementing a `clone()` method.
        *   Fixed shadowing of built-in functions and improved variable naming (e.g., `CPUUsage`).
        *   Added comprehensive error handling for `json.Unmarshal`, `os.Chmod`, and connection closing.
        *   Suppressed noise from non-critical error returns in `defer` and background tasks.
        *   Fixed multiple syntax errors and assignment mismatches introduced during automated refactoring.
        *   Added missing package-level documentation and method comments across all node modules.
        *   Renamed types to avoid stuttering (`Result`, `Request`, `Message`).
        *   Restricted file permissions for sensitive files (`.env`, state files, logs, version files) to `0600`.
        *   Aligned node to Go 1.24.0 for toolchain consistency.
*   **Theme Engine:** Implemented a full-featured Light/Dark mode system with glassmorphism effects.
    *   Default mode is Dark.
    *   Background animation (particle canvas) is preserved across both modes.
    *   Added theme toggle buttons to desktop sidebar and mobile header.
*   **Logging Cleanup:** Removed verbose `fmt.Println` and debug logs from the server codebase (`players.go`, `handlers_upload.go`, `main.go`) to ensure clean production-ready output.
*   **Authentication Improvements:**
    *   `AuthenticatePlayerHandler` now supports both JSON and Unity's `WWWForm` (form-data) for better compatibility.
    *   Disabled TOTP requirement in non-production environments for developer convenience.
*   **Documentation:** Updated `server/docs/CLIENT_API.md` and created `server/docs/unity/AuthenticationManager.cs` to reflect the latest authentication logic and provide a drop-in Unity implementation.
*   **Fixed Node Panic:** Resolved concurrent write panic in `node/internal/ws/client.go` by implementing a `writePump`.
*   **Fixed Heartbeat:** Improved `heartbeatLoop` resilience in Node.
*   **Stability:** Increased Server heartbeat timeout thresholds and ensured connection keep-alive on any message.
*   **Status Logic:** Disabled time-based status degradation ("Unresponsive"). Status is now strictly "Online" (connected) or "Offline" (disconnected).
*   **Dashboard UI:** Fixed Node status mismatch ("active" vs "Online") enabling the spawn button; removed verbose heartbeat logging.
*   **Logging:** Implemented log rotation (max 5MB, keep 20 lines) in Node, filtered noisy log types, and added file size display in Dashboard LogViewer.
*   **Database:** Integrated PostgreSQL support (via `pgx` driver) alongside SQLite. Added "Databases" tab to Dashboard for table listing. Use `DB_DRIVER=pgx` for Postgres. Added 5s timeout to DB init to prevent startup hangs. Enhanced Database Management: Full UI for Overview, Table Browser/Editor (CRUD), Internal Backups (Create/Restore/Download), and Config viewing.
*   **Dashboard Stability:** Fixed Svelte 5 reactivity issues in Layout and Dashboard Page (Runes migration) resolving navigation bugs and blank screen issues. Fixed accessibility warnings and template syntax errors.
*   **Theme & Aesthetic Migration:** Completed the migration to the tactical Rust/Stone "Exile" terminal aesthetic.
    *   Replaced all remaining hardcoded blue, indigo, and cyan accents with Rust theme colors across all major routes (`/dashboard`, `/config`, `/logs`, `/database`, `/notes`, `/users`).
    *   Implemented full `industrial_styling` support: all cards and modals now respect the sharp-corner/heavy-border toggle.
    *   Updated `NotificationBell` and `TaskItem` components to match the new industrial aesthetic.
    *   Fixed multiple Svelte 5 syntax errors and tag nesting issues in `config/+page.svelte`.
    *   Optimized Three.js lifecycle management: `SectionBackground`, `NavbarParticles`, and `GlobalSmoke` now properly stop their animation loops when disabled or in Low Power Mode.
    *   Implemented persistence for aesthetic settings: `theme`, `backgroundConfig`, and `siteSettings` now automatically sync with `localStorage`.
*   **Performance:** Refactored Node metrics collection to be asynchronous, preventing I/O blocks from delaying heartbeats. Added a "Low Power Mode" toggle to disable resource-intensive background animations.

## 🔒 Security Note
*   **Transport Encryption:** The Master Server runs on plain HTTP (`:8081`). For production security, you **MUST** run it behind a Reverse Proxy (Nginx, Caddy, Cloudflare) that handles SSL/TLS termination. This ensures the WebSocket connection and API Key are encrypted.
*   **Authentication:** All sensitive endpoints are protected by `X-API-Key` (Nodes) or Session Cookie (Dashboard).
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
*   **Node RCE Patch:** Implemented strict ID validation in `node/internal/game/manager.go` to prevent path traversal vulnerability in `RenameInstance` and `Spawn`.
*   **Security Headers:** Added `SecurityHeadersMiddleware` in Server to enforce `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, and `Content-Security-Policy`.
*   **Server Binding:** Updated `server/main.go` to bind to `127.0.0.1` by default (configurable via `SERVER_HOST`).
*   **Read-Only Database:** Added support for a separate `READONLY_DB_DSN` in `server/db.go` and `server/main.go`, used by `ExecuteSQLHandler` for safer ad-hoc queries.
*   **Frontend Sanitization:** Integrated `DOMPurify` in `web-dashboard` (`StatsCard.svelte`, `ConfirmDialog.svelte`) to replace regex-based sanitization.

### Node Authentication
*   **Two-Stage Enrollment:** Nodes now undergo a two-stage registration:
    1.  **Handshake:** Node connects with an enrollment key and announces its network presence.
    2.  **Configuration:** Admin configures the node (Region, Capacity) from the Dashboard and approves registration.
*   **Unique API Keys:** Nodes receive a unique, persistent API key only after admin approval.
*   **Simplified CLI:** Node CLI is simplified to `-m <master_url>` and `-key <enrollment_key>`. All other settings are managed centrally.
*   **Persistence:** Configuration (Region, MaxInstances, APIKey) is saved to the node's local `.env` upon successful enrollment.

### Remote Configuration & Management
*   **Real-Time Sync:** Settings updated in the Dashboard (Region, Max Instances) are instantly pushed to connected Nodes via WebSocket (`update_config` command).
*   **Persistence:** Nodes automatically save updated configurations to their local `.env` file, ensuring changes persist across restarts.
*   **Drain Mode:** (In Progress) Support for enabling "Maintenance Mode" to prevent new spawns while allowing existing games to finish.

### System Logging & Error Reporting
*   **Architecture:** Implemented a new persistent, categorized logging system.
*   **Database:** Added `system_logs` table to store detailed logs.
*   **RedEye Performance:** Added a new "RedEye Guardian" card and detailed metrics modal to the Performance tab, tracking real-time blocks, rate-limiting, and active bans.
### Game Player System
*   **Schema:** Implemented a new `player_system` database schema to isolate player data.
*   **Entities:** Added `players`, `friendships`, and `friend_requests` tables. Added `uid` column to `players` table for Firebase integration.
*   **Authentication & Security**:
    *   `Auth_GameMiddleware`: A dedicated middleware for game clients that enforces authentication via `X-Game-API-Key`.
    *   `POST /api/game/auth`: Authenticates a player via Firebase ID Token using the **Firebase Admin SDK** for robust production-ready verification. It links UID, returns full player profile, and provides a temporary `ws_auth_key`.
    *   `GET /api/game/ws`: WebSocket endpoint for real-time player communication (Game Client).
    *   `GET /api/game/players`: List all players (Dashboard - Session Protected).
    *   `GET /api/game/players/{id}`: Get player details (Dashboard - Session Protected).
    *   **Note:** All other player interactions (friends, reports, etc.) are handled via WebSocket messages.


### Firebase Remote Config
*   **Feature:** Implemented full CRUD support for Firebase Remote Config from the dashboard. Users can now create, update, delete, and sync parameters directly via the `/config` page.

### Backend
*   **Created `server/logging.go` service for structured logging.

*   **Backend:**
    *   Created `server/logging.go` service for structured logging.
    *   Updated `StatsMiddleware` to capture and categorize errors (Internal, Node, Security).
    *   Updated `GlobalStats` to only count "Internal" API errors in `TotalErrors` metric, improving "Performance" tab accuracy.
*   **Frontend:**
    *   Created new **System Logs** page (`/logs`) with filtering by category and detailed inspection.
    *   Updated Dashboard "Total Errors" card to link to the new logs page.

### UI Improvements
*   **System Calibration Interface (Theme Lab Overhaul):**
    *   Redesigned the Theme Lab into a comprehensive calibration workbench with four primary subsystems: Chromatic Matrix, Atmospheric Core, Geometric Logic, and Structural Engine.
    *   **Advanced Typography Engine:** Implemented granular control over global letter-spacing, line-height, paragraph spacing, and weighted typeface levels (Heading vs. Base).
    *   **Kinetic Physics System:** Added global calibration for transition velocities, hover expansion depth, and button-press depth across the entire UI.
    *   **Atmospheric Effects:** Introduced advanced controls for glitch frequency, lens fringing (chromatic aberration), flicker intensity, and grid substrate opacity.
    *   **Preset Engine:** Added industrial theme presets (DEEP_COMMAND, MERCURY_PROTO, SOLAR_FLARE, VOID_WALKER) for instant system-wide visual reconfiguration.
*   **Layout & CSS Audit:**
    *   **Z-Index Normalization:** Re-stratified UI layers to ensure the Command Module (Sidebar) and System Tickersit above all atmospheric overlays.
    *   **Navigation Centering:** Fixed alignment bugs in the collapsed sidebar state, ensuring perfectly centered tactical icons.
    *   **Typography Harmonization:** Migrated all text blocks, code segments, and headers to a unified CSS variable registry for real-time recalibration.
*   **Player Account Page Overhaul:**
    *   Refactored the layout to include a strategic summary section with StatsCards (Total Subjects, Online Count, Cumulative XP, Active Reports).
    *   Implemented a new sorting engine allowing administrators to sort by ID, Name, XP, or Last Seen.
    *   Upgraded Player and Report cards with "Tactical Corners", monospace typography for identifiers/values, and dynamic theme-aware rounding.
    *   Introduced "Ban/Restore" functionality with real-time UI feedback and backend integration.
    *   Modernized the `EditPlayerModal` to match the industrial aesthetic, featuring monospace input fields and theme-calibrated accents.
*   **Performance Page Overhaul:**
    *   Refactored the layout to introduce a new "Network Operations" card alongside the "Memory Matrix", optimizing screen real estate usage.
    *   Implemented detailed network metrics display: Active Links, Inbound/Outbound traffic stats, and Error Vector visualization.
    *   Fixed a TypeScript type mismatch error in the RedEye Sentinel Core component.
    *   Resolved missing properties in `theme/+page.svelte` to satisfy `npm run check`.
*   **Notes & Tasks**: Refactored `web-dashboard/src/routes/notes/+page.svelte` to feature a 2-column layout (Tasks sidebar, Notes grid).
*   **Tasks**: Added `TaskItem` component with cleaner styling and animations.
*   **Notes**: Enhanced `NoteCard` visuals with gradients, shadow effects, and better status indicators.

### 🏗️ Architectural Refactor (Post-Audit)
*   **Package Reorganization**: Reorganized the entire `server` codebase into a clean, modular package structure:
    *   `auth`: Authentication logic, session management, and Firebase Remote Config.
    *   `config`: Configuration management.
    *   `database`: Persistence layer, migrations, and database administration.
    *   `handlers`: Core API handlers (Nodes, Instances, Notes, Tasks).
    *   `logging`: Persistent system logging service.
    *   `metrics`: Performance and resource monitoring.
    *   `middleware`: Security and orchestration middleware.
    *   `models`: Shared data structures.
    *   `redeye`: Security and traffic management.
    *   `registry`: Central node and instance registry.
    *   `sse`: Real-time events hub.
    *   `utils`: Shared utility and UI functions.
*   **Import Cycle Resolution**: Fixed multiple import cycles by extracting shared logic into `utils` and qualifying symbol references correctly.
*   **Build Stability**: Resolved all build and test failures caused by the reorganization.

### ✅ Top-Tier Tasks System
*   **Hierarchical Structure**: Implemented recursive sub-tasks support in both backend and frontend.
*   **Discussion System**: Added persistent comments for each task.
*   **Metadata**: Added `in_progress` status and `deadline` support.
*   **UI/UX**: 
    *   Recursive rendering in `TaskItem.svelte`.
    *   Pulsing "In Progress" badges and overdue deadline indicators.
    *   In-line sub-task and comment creation.
*   **Database**: Implemented schema migrations to automatically add new columns to existing databases.

### 👁️ RedEye Visual Overhaul
*   **Robotic Aesthetic**: Implemented a modern, high-contrast cyber-theme for the RedEye Guardian panel.
*   **Animations**: Added a grid background and scanning line animation for a "neural core" feel.
*   **Technical Detail**: Enhanced data display with monospaced typography and technical metadata (CRC, entropy tracking).
*   **Modern Components**: Refactored the dashboard to use Svelte 5 best practices, including `$props`, `$state`, and the new dynamic component syntax.
*   **Type Safety**: Resolved all TypeScript indexing and component type errors across modernized components.

### 🐛 Bug Fixes & Stability
*   **Firebase Config Status**: Fixed a JSON tag casing issue (`Connected` vs `connected`) in the `/api/config/firebase/status` endpoint that caused the dashboard to incorrectly report "Not Configured".
*   **Database Scanning**: Fixed a `Scan` error in `GetTodos` where `todo_comments`' `created_at` (INTEGER) could not be automatically scanned into `time.Time`. Implemented manual scanning for comments to correctly handle Unix timestamp conversion.
*   **Frontend Safety**: Added null/undefined checks for `toLocaleString` and `filter` across the dashboard to prevent runtime crashes.
*   **SQL Repair**: Fixed corrupted SQL queries in the persistence layer.
*   **JSON Handling**: Ensured API endpoints return empty slices instead of `null` for better frontend compatibility.
*   **Config API Routes**: Fixed `404 Not Found` errors when updating configuration keys containing dots (e.g., `site.settings`) by updating the `gorilla/mux` route pattern to `{key:.+}` and ensuring correct precedence over specific routes.
*   **Database Scanning**: Fixed `500 Internal Server Error` caused by `sql: Scan error on column ...: converting NULL to string is unsupported`. Implemented `sql.NullString` handling for nullable columns in `server_config`, `notes`, `instance_actions`, and `redeye_ip_reputation` tables.
*   **Database Performance**: Optimized query speeds by implementing critical indexes on `system_logs`, `instance_actions`, `redeye_logs`, and `redeye_anticheat_events`.
*   **RedEye Optimization**: Implemented an in-memory `RuleCache` for the RedEye security middleware, eliminating redundant database lookups for every incoming request and reducing overall system latency.
*   **Three.js Lifecycle**: Optimized background animations (`GlobalSmoke`, `NavbarParticles`) with throttled 30FPS loops and immediate resource disposal when `Low Power Mode` is enabled, significantly reducing CPU/GPU overhead.
*   **Frontend Synchronization**: Upgraded persistent stores with a 500ms debounced save mechanism to prevent database write-spam during rapid UI interactions.
*   **Backend Resilience**: Hardened the HTTP server with production-grade read/write timeouts and context-aware background goroutines for graceful shutdown.
*   **Network Compression**: Implemented a native Gzip middleware in the Go backend. All JSON responses (logs, configs, stats) are now compressed during transit, reducing bandwidth usage by up to 80% and improving load times for remote operators.
*   **Intelligent Hibernation**: Integrated the **Page Visibility API** into the dashboard. Real-time data streams (SSE) are now automatically paused when the browser tab is hidden and resumed when brought to the focus, significantly reducing background CPU and network overhead.
*   **Compositor Isolation**: Applied `isolation: isolate` and `contain: content` to the System Topology SVG. This prevents node-level animations from triggering expensive full-screen layout recalculations.
*   **Reactivity Throttling**: Refactored the topology heartbeat logic using Svelte 5's `untrack`. Heartbeat pulses are now processed in an isolated context, preventing redundant array scans and keeping the main thread free for high-priority UI interactions.
*   **Log Buffer Windowing**: Implemented an automated rendering cap for the Log Viewer. The system now only renders the most recent 200 log lines by default, eliminating the browser "hang" typical when loading massive operational logs while still keeping the full dataset available in memory.
*   **Tactical Command Palette**: Implemented a global **Smart-Uplink** hub accessible via `Ctrl+K` or `Cmd+K`. This enables instant fuzzy-search navigation across all system modules and the execution of high-level system commands (e.g., reboots, theme switching) without leaving the current view.
*   **Keyboard-First Navigation**: Added global shortcut listeners for "Quick Jump" navigation (`G+D` for Dashboard, `G+L` for Logs, etc.) and full keyboard support for all interactive dialogs.
*   **A11y Hardening**: Performed a comprehensive accessibility pass. All custom industrial modals and components now include standard ARIA roles, high-contrast focus indicators, and semantic labels for screen-reader compatibility.
- Renamed "Spawner" to "Node" across the project and implemented "Drain Mode" for maintenance. Updated the frontend to Svelte 5 Runes and fixed multiple stability issues.
- Shifted aesthetic from "Military Industrial" to "Modern Tech" (Darker "20% Industrial" Variant).
    - Palette: Deep Slate/Black (#020617) backgrounds, Blue 500 primary, Amber 500 accent.
    - Style: Rounded corners (xl/lg), glassmorphism, softer borders.
    - Background: New `MotherboardBackground` component with circuit path animations.
- Cleaned up all components to remove `siteSettings.aesthetic` dependencies, ensuring consistent styling and better performance.
- Removed "hover card animation" (CardHoverOverlay) and CRT scanline/noise effects for a cleaner, static industrial look.
- Fixed a major issue where full-screen atmospheric overlays (clouds, rain) were blocking clicks on interactive elements by adding `pointer-events-none`.
- Fixed a runtime error in `SystemTopology.svelte` by removing the `topology_blobs` reference, which was causing a crash after the theme system removal.
- Fixed several TypeScript and accessibility issues during the UI refactor.

### 📦 Game Server Management
*   **Smart Uploads**: Integrated `JSZip` into the dashboard to automatically extract version metadata from `game_server.zip` uploads (via `manifest.json`), streamlining the release process.
*   **Layout Fix**: Resolved a missing `cubicOut` import in `+layout.svelte` to satisfy type checking.

### 🎨 Sidebar & UI Refinement
*   **Icon Refresh**: Updated sidebar icons to better match the "Tactical" aesthetic:
    *   Dashboard -> `Gauge`
    *   Notes -> `FileText`
    *   Config -> `Sliders`
*   **Scrollbar**: Hidden the default scrollbar on the sidebar (`no-scrollbar`) for a cleaner look while maintaining scrollability.
*   **Active State**: Enhanced the `.nav-link-industrial` active state with a left border indicator and gradient background.

## 🔴 RedEye System Architecture Analysis

**Question:** Should the RedEye system in the backend be its own service?

**Conclusion:** **No, it is not recommended to separate the RedEye system into its own service at this time.**

**Reasoning:**

The RedEye system, as implemented in `server/redeye_core.go` and `server/handlers_redeye.go`, is an intrinsic and deeply integrated part of the Master Server. Key factors influencing this conclusion include:

1.  **Tight Database Coupling:** RedEye relies heavily on the Master Server's `dbConn` for storing and retrieving rules, logs, statistics, configuration, and IP reputation data. Separating it would necessitate either:
    *   A dedicated RedEye database, leading to data duplication and synchronization challenges.
    *   An additional network hop for the RedEye service to communicate with the Master Server's database, introducing latency and complexity.

2.  **Core Middleware Functionality:** The `RedEyeMiddleware` is a critical security and traffic management layer. It intercepts all incoming HTTP requests to the Master Server to perform IP banning, rule enforcement, and rate limiting *before* requests reach other handlers. Extracting this into a separate service would mean:
    *   Either duplicating the middleware logic within the Master Server, negating the purpose of separation.
    *   Or routing *all* Master Server traffic through the RedEye service, adding significant overhead, increasing latency, and creating a single point of failure.

3.  **Real-time Interaction:** Features such as real-time auto-banning and immediate cache refreshing for IP bans and rules require minimal latency to be effective. An additional service would introduce communication delays, potentially impacting the responsiveness of security measures.

4.  **Shared Configuration:** RedEye utilizes the Master Server's existing configuration mechanisms (`GetConfigByKey`, `UpdateConfig`), indicating a shared operational context.

5.  **Anti-Cheat Event Reporting:** The `ReportAnticheatEventHandler` suggests that other parts of the Master Server (or external components interacting with the Master Server) report events directly to this endpoint. This tight integration ensures timely updates to RedEye's reputation system.

While separating services can offer benefits like independent scalability and improved modularity, these advantages would likely require a substantial re-architecture of RedEye (e.g., transitioning to an event-driven model, introducing a dedicated, synchronized database) to overcome the challenges posed by its current deep integration. The current implementation does not indicate that RedEye's resource consumption or operational independence is a significant bottleneck that would justify such a complex undertaking.

Therefore, for the current architecture, RedEye is most appropriately managed as an internal component of the Master Server.
## 🎨 Style Guide

### Terminal Theme
*   **Background:** Deep charcoal (#0a0a0a) or tactical gray (#121212) with radial vignette overlays.
*   **Typography:** 
    *   Primary: Bold, uppercase sans-serif (Inter) for headlines and high-level subjects.
    *   Secondary: Strict monospace (JetBrains Mono) for technical data, logs, and metadata.
*   **Palette:**
    *   Primary: Tactical Amber/Orange (#f97316).
    *   Secondary: Alert Red (#ef4444) for faults and critical actions.
    *   Neutral: Slate Gray (#a0a0a0) and Deep Steel (#1a1a1a).
*   **UI Structure:** Horizontally aligned data cards resembling intelligence briefings or intercepted transmissions. Minimal 1px borders and low-opacity separators.
*   **Atmosphere:** Subtle CRT scanlines (low opacity), soft radial vignette, and cinematic warm-up animations. No noise texture for maximum readability.
*   **Feel:** Professional command-center, surveillance-driven, dystopian, and authoritative. AAA game cinematic intelligence interface.
