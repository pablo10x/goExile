.PHONY: all build build-server build-spawner run-server-bg clean

# Directories
SERVER_DIR := server
SPAWNER_DIR := spawner

# Output binaries
SERVER_BIN := server
SPAWNER_BIN := spawner

all: build

build: build-server build-spawner

build-server:
	@echo "Building Master Server..."
	cd $(SERVER_DIR) && go build -o $(SERVER_BIN) .

build-spawner:
	@echo "Building Spawner..."
	cd $(SPAWNER_DIR) && go build -o $(SPAWNER_BIN) .

run-server-bg: build-server
	@echo "Starting Master Server in background..."
	cd $(SERVER_DIR) && (nohup ./$(SERVER_BIN) > server.log 2>&1 & echo $$! > server.pid)
	@echo "Master Server running. PID stored in $(SERVER_DIR)/server.pid. Logs: $(SERVER_DIR)/server.log"

stop-server:
	@if [ -f $(SERVER_DIR)/server.pid ]; then \
		kill `cat $(SERVER_DIR)/server.pid` && rm $(SERVER_DIR)/server.pid; \
		echo "Master Server stopped."; \
	else \
		echo "No PID file found for Master Server."; \
	fi

clean:
	rm -f $(SERVER_DIR)/$(SERVER_BIN) $(SERVER_DIR)/server.log $(SERVER_DIR)/server.pid
	rm -f $(SPAWNER_DIR)/$(SPAWNER_BIN)