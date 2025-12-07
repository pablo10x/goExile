package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"spawner/internal/config"
	"sync"
	"time"
)

// Instance represents a running game server.
type Instance struct {
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	ProcessID int       `json:"pid"`
	Status    string    `json:"status"`
	Region    string    `json:"region"`
	StartTime time.Time `json:"start_time"`

	cmd *exec.Cmd // Private: command handle for process management
}

// Manager handles the lifecycle of game server processes.
type Manager struct {
	cfg       *config.Config
	instances map[string]*Instance // Key is ID (string)
	mu        sync.RWMutex         // RWMutex for better concurrency
	logger    *slog.Logger
}

// NewManager creates a new game process manager.
func NewManager(cfg *config.Config, logger *slog.Logger) *Manager {
	return &Manager{
		cfg:       cfg,
		instances: make(map[string]*Instance),
		logger:    logger,
	}
}

// Spawn starts a new game server instance.
func (m *Manager) Spawn(ctx context.Context) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	port, err := m.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	// Use CommandContext for better control (cancellation/timeout)
	cmd := exec.CommandContext(ctx, m.cfg.GameBinaryPath,
		"-batchmode",
		"-nographics",
		"-mode", "server",
		"-port", fmt.Sprintf("%d", port),
	)
	
	// Redirect stdout/stderr for debugging (or pipe to logger in future)
	// For production, you might want to pipe this to a file or log collector.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	id := fmt.Sprintf("%s-%d", m.cfg.Region, port)
	instance := &Instance{
		ID:        id,
		Port:      port,
		ProcessID: cmd.Process.Pid,
		Status:    "Running",
		Region:    m.cfg.Region,
		StartTime: time.Now(),
		cmd:       cmd,
	}

	m.instances[id] = instance
	m.logger.Info("Spawned new game server", "id", id, "port", port, "pid", instance.ProcessID)

	// Monitor process in background
	go m.monitorInstance(id, cmd)

	return instance, nil
}

// monitorInstance waits for the process to exit and cleans up.
func (m *Manager) monitorInstance(id string, cmd *exec.Cmd) {
	err := cmd.Wait()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if instance, exists := m.instances[id]; exists {
		instance.Status = "Stopped"
		// We keep the instance in the map with 'Stopped' status or delete it?
		// Usually, we might want to keep history or just delete. 
		// For this simple spawner, let's remove it to free the port logically.
		delete(m.instances, id)
		
		exitCode := 0
		if err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				exitCode = exitError.ExitCode()
			}
			m.logger.Warn("Game server exited with error", "id", id, "error", err, "exit_code", exitCode)
		} else {
			m.logger.Info("Game server stopped normally", "id", id)
		}
	}
}

// StopInstance kills a specific game server.
func (m *Manager) StopInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	if instance.cmd != nil && instance.cmd.Process != nil {
		if err := instance.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	
	// monitorInstance will handle the cleanup
	return nil
}

// Shutdown stops all running instances gracefully.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Shutting down manager, stopping all instances...")
	for id, instance := range m.instances {
		if instance.cmd != nil && instance.cmd.Process != nil {
			m.logger.Debug("Killing instance", "id", id, "pid", instance.ProcessID)
			_ = instance.cmd.Process.Kill()
		}
	}
	// Clear map
	m.instances = make(map[string]*Instance)
}

// GetInstance returns a copy of the instance status.
func (m *Manager) GetInstance(id string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	inst, ok := m.instances[id]
	if !ok {
		return nil, false
	}
	// Return copy to avoid race conditions on pointer
	val := *inst
	val.cmd = nil // hide internal cmd
	return &val, true
}

// ListInstances returns all running instances.
func (m *Manager) ListInstances() []Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		val := *inst
		val.cmd = nil
		list = append(list, val)
	}
	return list
}

// findAvailablePort scans for an open TCP port.
// Note: There is still a theoretical race condition here between finding and binding.
func (m *Manager) findAvailablePort() (int, error) {
	for port := m.cfg.MinGamePort; port <= m.cfg.MaxGamePort; port++ {
		// 1. Check internal tracking
		portInUse := false
		for _, inst := range m.instances {
			if inst.Port == port {
				portInUse = true
				break
			}
		}
		if portInUse {
			continue
		}

		// 2. Check OS availability
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", m.cfg.MinGamePort, m.cfg.MaxGamePort)
}