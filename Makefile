.PHONY: all start-server start-spawner start-web-dashboard \
        stop-server stop-spawner stop-web-dashboard \
        restart-server restart-spawner restart-web-dashboard \
        stop-all clean build-server build-spawner build-web-dashboard

SHELL := powershell.exe

# Variables for PID files
SERVER_PID_FILE := server_pid.txt
SPAWNER_PID_FILE := spawner_pid.txt
WEB_DASHBOARD_PID_FILE := web_dashboard_pid.txt

# --- Build Targets ---
build-server:
	@Write-Host "Building server..."
	pushd server; go build -o server.exe .; popd
	@Write-Host "Server build complete."

build-spawner:
	@Write-Host "Building spawner..."
	pushd spawner; go build -o spawner.exe .; popd
	@Write-Host "Spawner build complete."

build-web-dashboard:
	@Write-Host "Ensuring web-dashboard dependencies are installed..."
	pushd web-dashboard; npm install; popd
	@Write-Host "Web-dashboard dependencies checked."

# --- Start Targets (Background) ---
start-server: build-server
	@Write-Host "Starting server..."
	Start-Process -FilePath ".\server\server.exe" -NoNewWindow -PassThru | ForEach-Object { $$_.Id | Out-File -FilePath "$(SERVER_PID_FILE)" }
	@Write-Host "Server started. PID saved to $(SERVER_PID_FILE)"

start-spawner: build-spawner
	@Write-Host "Starting spawner..."
	Start-Process -FilePath ".\spawner\spawner.exe" -NoNewWindow -PassThru | ForEach-Object { $$_.Id | Out-File -FilePath "$(SPAWNER_PID_FILE)" }
	@Write-Host "Spawner started. PID saved to $(SPAWNER_PID_FILE)"

start-web-dashboard: build-web-dashboard
	@Write-Host "Starting web-dashboard..."
	Start-Process -FilePath "npm.cmd" -ArgumentList "run dev" -WorkingDirectory ".\web-dashboard" -NoNewWindow -PassThru | ForEach-Object { $$_.Id | Out-File -FilePath "$(WEB_DASHBOARD_PID_FILE)" }
	@Write-Host "Web-dashboard started. PID saved to $(WEB_DASHBOARD_PID_FILE)"

# --- Stop Targets ---
stop-server:
	@Write-Host "Stopping server..."
	if (Test-Path "$(SERVER_PID_FILE)") { ^
		$$pid = Get-Content -Path "$(SERVER_PID_FILE)"; ^
		Stop-Process -Id $$pid -Force -ErrorAction SilentlyContinue; ^
		Remove-Item "$(SERVER_PID_FILE)"; ^
		Write-Host "Server process $$pid stopped."; ^
	} else { ^
		Write-Host "Server PID file not found. Is the server running?"; ^
	}

stop-spawner:
	@Write-Host "Stopping spawner..."
	if (Test-Path "$(SPAWNER_PID_FILE)") { ^
		$$pid = Get-Content -Path "$(SPAWNER_PID_FILE)"; ^
		Stop-Process -Id $$pid -Force -ErrorAction SilentlyContinue; ^
		Remove-Item "$(SPAWNER_PID_FILE)"; ^
		Write-Host "Spawner process $$pid stopped."; ^
	} else { ^
		Write-Host "Spawner PID file not found. Is the spawner running?"; ^
	}

stop-web-dashboard:
	@Write-Host "Stopping web-dashboard..."
	if (Test-Path "$(WEB_DASHBOARD_PID_FILE)") { ^
		$$pid = Get-Content -Path "$(WEB_DASHBOARD_PID_FILE)"; ^
		Stop-Process -Id $$pid -Force -ErrorAction SilentlyContinue; ^
		Remove-Item "$(WEB_DASHBOARD_PID_FILE)"; ^
		Write-Host "Web-dashboard process $$pid stopped."; ^
	} else { ^
		Write-Host "Web-dashboard PID file not found. Is the web-dashboard running?"; ^
	}

# --- Restart Targets ---
restart-server: stop-server start-server
	@Write-Host "Server restarted."

restart-spawner: stop-spawner start-spawner
	@Write-Host "Spawner restarted."

restart-web-dashboard: stop-web-dashboard start-web-dashboard
	@Write-Host "Web-dashboard restarted."

# --- Combined Targets ---
all: start-server start-spawner start-web-dashboard
	@Write-Host "All services started."

stop-all: stop-server stop-spawner stop-web-dashboard
	@Write-Host "All services stopped."

clean:
	@Write-Host "Cleaning up build artifacts and PID files..."
	Remove-Item "$(SERVER_PID_FILE)" -ErrorAction SilentlyContinue
	Remove-Item "$(SPAWNER_PID_FILE)" -ErrorAction SilentlyContinue
	Remove-Item "$(WEB_DASHBOARD_PID_FILE)" -ErrorAction SilentlyContinue
	Remove-Item "server\server.exe" -ErrorAction SilentlyContinue
	Remove-Item "spawner\spawner.exe" -ErrorAction SilentlyContinue
	@Write-Host "Cleanup complete."
