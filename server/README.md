# 🎮 Spawner Registry (Master Server)

A central HTTP API service for registering, managing, and monitoring game Spawner instances. It also provides a real-time dashboard for operational insights and facilitates the deployment of game server files to Spawners.

---

## 🚀 Features

*   **Spawner Management:** Spawners register with the Master, send periodic heartbeats, and are tracked in memory (with optional SQLite persistence).
*   **Real-time Dashboard:** A comprehensive web dashboard (powered by Server-Sent Events - SSE) displays:
    *   System statistics (uptime, requests, errors, process memory, bandwidth).
    *   A list of active Spawners, their regions, host/port, and current instance counts.
    *   Detailed error logs with timestamps, paths, and client IPs.
*   **Game Server Deployment:** Hosts and serves game server build packages (`g ame_server.zip`) to Spawners for automatic deployment.
*   **Authentication & Security:**
    *   Dashboard access is secured with username/password authentication.
    *   API Key authentication (`X-API-Key` header) is used to secure Spawner-Master communication (registration, heartbeats, downloads).
*   **Persistence:** Optionally uses an SQLite database (`registry.db`) to store Spawner registrations across restarts.
*   **Cleanup:** A background goroutine automatically removes inactive Spawners that fail to send heartbeats within a configured TTL.
*   **Error Reporting:** Catches and logs API errors with contextual information (path, status, client IP).

---

## 🛠️ Getting Started

### Prerequisites

*   Go (version 1.20+)
*   Git

### 1. Clone the Repository

```bash
git clone https://github.com/your-repo/goExile.git
cd goExile
```

### 2. Configuration (`server/.env`)

Create a `.env` file in the `server/` directory. This file will hold environment-specific settings.

```ini
# Database Path (Optional)
# If not set, defaults to "database/registry.db"
DB_PATH=database/registry.db

# Dashboard Administrator Credentials
# Essential for logging into the dashboard. Change these from defaults!
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin123

# API Key for Spawner Authentication (REQUIRED for secure Spawner communication)
# Spawners must use this exact key in their X-API-Key header.
MASTER_API_KEY=your_very_secret_master_api_key_here
```

### 3. Game Server Package for Spawners

*   Place your compiled game server build (as a `.zip` archive) in the `server/files/` directory.
*   The archive **must be named `game_server.zip`**.
*   The contents of `game_server.zip` should be structured such that the `GAME_BINARY_PATH` configured in the Spawner's `.env` is correct after extraction.

### 4. Build and Run

Navigate to the `server/` directory and run:

```bash
cd server
go mod tidy          # Download dependencies
go run .             # Run the Master Server
```

The server will start on `http://localhost:8081`.

---

## 🌐 Accessing the Dashboard

Once the Master Server is running, open your web browser and navigate to:
`http://localhost:8081`

You will be prompted to log in using the `ADMIN_EMAIL` and `ADMIN_PASSWORD` configured in your `server/.env` file.

The dashboard will display real-time statistics and a list of all registered Spawners. You can also click the "Errors" card to view recent application errors.

---

## 🔑 API Endpoints

The Master Server exposes various API endpoints. Spawner-related endpoints (`/api/spawners/*`) are protected by `X-API-Key` authentication if `MASTER_API_KEY` is set in the environment.

### Spawner Registration & Management (via `X-API-Key` header)

*   `POST /api/spawners`: Register a new Spawner.
*   `POST /api/spawners/{id}/heartbeat`: Send a heartbeat to keep a Spawner active and update its stats.
*   `POST /api/spawners/{id}/spawn`: Triggers a new game instance on the specified Spawner (proxies request).
*   `GET /api/spawners`: List all registered Spawners.
*   `GET /api/spawners/{id}`: Get details of a specific Spawner.
*   `DELETE /api/spawners/{id}`: Deregister a Spawner.
*   `GET /api/spawners/download`: Download the `game_server.zip` package.

### Dashboard Data & Other

*   `GET /api/stats`: Get current Master Server statistics (for polling fallback).
*   `GET /api/errors`: Get recent application error logs.
*   `GET /events`: Server-Sent Events endpoint for real-time dashboard updates.
*   `GET /health`: Basic health check endpoint.
*   `GET /login`, `POST /login`, `GET /logout`: Dashboard authentication endpoints.
*   `GET /dashboard`, `GET /errors`, `GET /users`: Dashboard UI pages.

---

## 🏗️ Architecture & Workflow

### Spawner Lifecycle

1.  A Spawner instance starts up.
2.  It attempts to **register** itself with the Master Server by sending a `POST /api/spawners` request, including its region, host, port, and maximum instance capacity, along with the `X-API-Key` header.
3.  If successful, the Master Server assigns it a unique ID.
4.  The Spawner then begins sending **periodic heartbeats** (`POST /api/spawners/{id}/heartbeat`) to keep its registration active and update its current instance count.
5.  If the Spawner's `GAME_BINARY_PATH` is missing, it will automatically **download** `game_server.zip` from the Master Server's `/api/spawners/download` endpoint, extracting it to `GAME_INSTALL_DIR`.

### Dashboard Real-time Updates

1.  When a user accesses the dashboard, it establishes an **Server-Sent Events (SSE) connection** to `/events`.
2.  The Master Server continuously pushes updates:
    *   **Every 1 second:** Current Master Server statistics (`stats` event).
    *   **Every 2 seconds:** The full list of registered Spawners (`spawners` event).
3.  If the SSE connection fails, the dashboard automatically falls back to **polling** (`/api/stats` and `/api/spawners`) every 3 seconds to maintain updates.

---

## 🧪 Testing

To run tests for the Master Server:

```bash
cd server
go test ./... -v
```

---

## ⛔ Limitations & Future Enhancements

*   **No HTTPS:** All communication is currently over HTTP. **HTTPS is critical for production deployments.**
*   **Single Point of Failure:** The Master Server is currently a single instance. High availability would require a distributed setup.
*   **System-Level Metrics:** The dashboard shows Go process memory/bandwidth. Full system-level resource monitoring (CPU, total RAM, disk I/O, network interface traffic) would require external libraries or OS-specific integrations.
*   **Game Instance Details:** The Master only knows about Spawners. It doesn't track individual game instances running on a Spawner.
*   **Richer Authentication/Authorization:** For more complex scenarios, OAuth2, JWTs, and role-based access control (RBAC) could be implemented.
*   **Centralized Logging/Alerting:** Integrating with external log aggregators (e.g., ELK stack) and alerting systems for critical events.

---

This `README.md` provides a comprehensive overview of the Master Server.

Now I will create `spawner/README.md`.