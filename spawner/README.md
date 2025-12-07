# Game Server Spawner

A production-ready, lightweight, distributed infrastructure tool written in Go (Golang) to spawn "headless" Unity game server instances on demand.

## Features

- **Dynamic Spawning:** Launches Unity server instances on request.
- **Port Management:** Automatically finds and assigns free TCP ports.
- **Process Management:** Tracks running instances, supports graceful shutdown, and monitors process health.
- **API Security:** Simple API Key authentication.
- **Structured Logging:** Uses `log/slog` for JSON-formatted logs.
- **Graceful Shutdown:** cleanly terminates all child game servers when the spawner stops.

## Project Structure

- `cmd/server/`: Entry point.
- `internal/config/`: Configuration loading and validation.
- `internal/game/`: Core logic for process management (spawn, stop, monitor).
- `api/`: HTTP handlers and routing.

## Configuration

The application is configured via environment variables or a `.env` file.

| Variable | Description | Default |
|----------|-------------|---------|
| `REGION` | (Required) The region identifier (e.g., "EU", "US-East"). | - |
| `GAME_BINARY_PATH` | (Required) Absolute path to the Unity server build. | - |
| `GAME_DOWNLOAD_URL` | (Optional) URL to a ZIP file containing the game server. Used if binary is missing. | - |
| `GAME_DOWNLOAD_TOKEN` | (Optional) Bearer token for downloading the game server (if URL is protected). | - |
| `GAME_INSTALL_DIR` | (Optional) Directory to extract the downloaded ZIP to. | `./game_server` |
| `SPAWNER_PORT` | The port this API listens on. | 8080 |
| `MIN_GAME_PORT` | Start of the port range for game servers. | 7777 |
| `MAX_GAME_PORT` | End of the port range. | 8000 |

### Example `.env`

```ini
REGION=US_West
GAME_BINARY_PATH=C:\Builds\MyGame\MyGame.exe
SPAWNER_PORT=8080
# Windows Example:
# GAME_BINARY_PATH=C:/Builds/MyGame/MyGame.exe
# Linux Example:
# GAME_BINARY_PATH=/home/user/game/MyGame.x86_64
# New Fields
GAME_DOWNLOAD_URL=https://example.com/builds/latest_server.zip
GAME_INSTALL_DIR=./game_server
MIN_GAME_PORT=9000
MAX_GAME_PORT=9100
```

## API Usage

### Authentication

If `SPAWNER_API_KEY` is set, include the header:
`X-API-Key: secret_password_123`

### Endpoints

#### 1. Check Health
`GET /health`
Returns the status and region of the spawner.

#### 2. Spawn Server
`POST /spawn`
Starts a new game server instance.
**Response:**
```json
{
  "id": "US_West-9001",
  "port": 9001,
  "pid": 12345,
  "status": "Running",
  "region": "US_West",
  "start_time": "2023-10-27T10:00:00Z"
}
```

#### 3. List Instances
`GET /instances`
Returns a list of all currently running servers.

#### 4. Stop Instance
`DELETE /instance/:id`
Stops the specified game server.

## Development

### Running locally
```bash
go run cmd/server/main.go
```

### Running Tests
```bash
go test ./...
```

## Unity Integration

Ensure your Unity Server build reads the `-port` command line argument.
See `internal/game/manager.go` for the exact arguments passed:
`-batchmode -nographics -mode server -port <port>`
