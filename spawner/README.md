# 🎮 Game Server Spawner

A production-ready, lightweight, and extensible service written in Go (Golang) designed to launch, monitor, and manage dedicated game server instances on a host machine. It communicates with a central Spawner Registry (Master Server) to report its status and request new game instances.

---

## 🚀 Features

*   **Master Server Integration:**
    *   Registers itself with the Master Server upon startup.
    *   Sends periodic heartbeats with its current instance load and status.
    *   Authenticates with the Master using an `X-API-Key` header.
    *   Automatically downloads missing game server build files from the Master Server.
*   **Dynamic Game Instance Management:**
    *   Receives commands from the Master Server to spawn new game server instances.
    *   Automatically finds and assigns available TCP ports within a configured range for each game instance.
    *   Launches game servers as child processes and monitors their lifecycle.
    *   Provides an API to list and stop individual game instances.
*   **Robust Logging:** Uses structured logging (`log/slog`) with output directed to a `spawner.log` file, suitable for production environments.
*   **Graceful Shutdown:** Ensures all running game server instances are terminated cleanly when the Spawner service itself shuts down.

---

## 🛠️ Getting Started

### Prerequisites

*   Go (version 1.20+)
*   Git
*   A compiled dedicated game server executable (e.g., a Unity headless build).

### 1. Clone the Repository

```bash
git clone https://github.com/your-repo/goExile.git
cd goExile
```

### 2. Prepare Game Server Files

*   Create a directory named `game_server` inside the `spawner/` folder.
*   Place your compiled game server executable and all its required files (data, libraries, configs) into `spawner/game_server/`.

### 3. Configuration (`spawner/.env`)

Create a `.env` file in the `spawner/` directory. This file will hold environment-specific settings crucial for the Spawner's operation.

```ini
# Spawner Host Region (e.g., "US-East", "Europe-West")
REGION=Your_Region

# Path to your Game Server Executable (relative to spawner/ or absolute)
# Example for Windows: ./game_server/DedicatedServer.exe
# Example for Linux:   ./game_server/DedicatedServer.x86_64
GAME_BINARY_PATH=./game_server/DedicatedServer.exe

# Directory where game server files are installed/extracted.
# This should match where you placed your game server files.
GAME_INSTALL_DIR=./game_server

# Port for the Spawner's own internal API (used by Master to send commands)
SPAWNER_PORT=8080

# Port range for game server instances launched by this Spawner
MIN_GAME_PORT=7777
MAX_GAME_PORT=8000

# Master Server URL and API Key
# The Spawner will register and communicate with this Master.
MASTER_URL=http://localhost:8081
MASTER_API_KEY=your_very_secret_master_api_key_here
```

### 4. Build and Run

Navigate to the `spawner/` directory and run:

```bash
cd spawner
go mod tidy          # Download dependencies
go run ./cmd/server  # Run the Spawner service
```

**Important:** For the Spawner to function correctly, ensure the Master Server is already running and accessible at the configured `MASTER_URL`.

---

## 🏗️ Architecture & Workflow

### Startup Sequence

1.  **Loads Configuration**: Reads settings from `.env` and environment variables.
2.  **Validates Game Binary**: Checks if the `GAME_BINARY_PATH` exists.
3.  **Automatic Download**: If the game binary is not found locally, the Spawner attempts to download `game_server.zip` from the configured `MASTER_URL/api/spawners/download` endpoint, using the `MASTER_API_KEY` for authentication.
4.  **Registers with Master**: Sends a `POST` request to `MASTER_URL/api/spawners` to register itself, including its host, port, region, and instance capacity.
5.  **Starts Heartbeat**: Initiates a background routine to send periodic heartbeats (`POST MASTER_URL/api/spawners/{id}/heartbeat`) to keep its registration active and update its current instance load.
6.  **Starts API Server**: The Spawner's own API server begins listening on `SPAWNER_PORT` for commands from the Master.

### Game Server Spawning

1.  The Master Server (e.g., via the dashboard) sends a `POST` request to `SPAWNER_HOST:SPAWNER_PORT/spawn`.
2.  The Spawner receives the request and:
    *   Finds an available TCP port within its `MIN_GAME_PORT`-`MAX_GAME_PORT` range.
    *   Launches your game server executable (`GAME_BINARY_PATH`) as a child process.
    *   Passes command-line arguments, including the assigned port (e.g., `-port 7777`).
    *   Registers the new game instance internally.
3.  The Spawner's heartbeat automatically updates the Master Server with the new count of running game instances.

---

## 🔑 Spawner's Internal API

The Spawner hosts its own API that the Master Server can call to manage game instances.

### Endpoints

*   `POST /spawn`: Starts a new game server instance.
*   `GET /instances`: Lists all game server instances currently running on this Spawner.
*   `DELETE /instance/:id`: Stops a specific game server instance by its ID.
*   `GET /health`: Checks the liveness of the Spawner service.

---

## 🎮 Unity Game Server Integration (FishNet Example)

For your Unity game server to use the dynamic port assigned by the Spawner, its networking solution must be configured to read a command-line argument.

**Example for FishNet (C# Script):**

1.  Create a C# script (e.g., `SpawnerPortSetter.cs`) in your Unity project.
2.  Attach this script to a GameObject in your server build scene.
3.  Ensure your FishNet `NetworkManager` and its `Transport` are referenced in the script's Inspector fields.

```csharp
using FishNet.Managing;
using FishNet.Transporting;
using UnityEngine;
using System;

public class SpawnerPortSetter : MonoBehaviour
{
    [SerializeField] private NetworkManager _networkManager;
    [SerializeField] private Transport _transport;
    [SerializeField] private ushort _defaultPort = 7770; // Fallback port

    void Awake()
    {
        if (_networkManager == null) _networkManager = FindObjectOfType<NetworkManager>();
        if (_networkManager == null) { Debug.LogError("[SpawnerPortSetter] NetworkManager not found."); return; }

        if (_transport == null) _transport = _networkManager.TransportManager.GetTransport(0);
        if (_transport == null) { Debug.LogError("[SpawnerPortSetter] FishNet Transport not found."); return; }

        int port = GetPortFromCommandLine(_defaultPort);
        Debug.Log($"[SpawnerPortSetter] Setting Transport Port to: {port}");
        _transport.SetPort((ushort)port); // Or cast and set directly: ((ToggledTransport)_transport).Port = (ushort)port;
    }

    private int GetPortFromCommandLine(ushort defaultPort)
    {
        string[] args = Environment.GetCommandLineArgs();
        for (int i = 0; i < args.Length; i++)
        {
            if (args[i].ToLower() == "-port" && i + 1 < args.Length)
            {
                if (int.TryParse(args[i + 1], out int parsedPort))
                {
                    Debug.Log($"[SpawnerPortSetter] -port argument found: {parsedPort}");
                    return parsedPort;
                }
            }
        }
        Debug.Log($"[SpawnerPortSetter] -port argument not found. Using default: {defaultPort}");
        return defaultPort;
    }
}
```

The Spawner launches your game server with:
`[GAME_BINARY_PATH] -batchmode -nographics -mode server -port [assigned_port]`

---

## 🧪 Testing

To run tests for the Spawner module:

```bash
cd spawner
go test ./... -v
```

---

## ⛔ Limitations & Future Enhancements

*   **No HTTPS:** All API communication is currently over HTTP. Implementing HTTPS for the Spawner's own API would be crucial for production security if exposed outside localhost.
*   **Game Instance Details:** The Spawner's API provides basic instance details. More granular information (e.g., current player count per instance) would require deeper integration with the game server itself.
*   **Resource Limits:** No CPU/memory limits are imposed on launched game server processes.
*   **Advanced Orchestration:** Features like instance auto-scaling based on load, instance migration, or rolling updates are not implemented.

---

This `README.md` provides a comprehensive overview of the Game Server Spawner.