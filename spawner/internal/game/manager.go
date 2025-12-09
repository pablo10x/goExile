package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"spawner/internal/config"
	"sync"
	"time"
)

// Instance represents a running game server.
type Instance struct {
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	ProcessID int       `json:"pid"`
	Status    string    `json:"status"` // "Running", "Stopped", "Error"
	Region    string    `json:"region"`
	StartTime time.Time `json:"start_time"`
	Path      string    `json:"path"` // Path to this instance's directory

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
	m := &Manager{
		cfg:       cfg,
		instances: make(map[string]*Instance),
		logger:    logger,
	}
	return m
}

// RestoreInstances loads state from disk and restarts servers that should be running.
func (m *Manager) RestoreInstances() error {
	if err := m.LoadState(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, inst := range m.instances {
		if inst.Status == "Running" {
			m.logger.Info("Restoring instance", "id", id, "port", inst.Port)
			if err := m.startProcess(inst); err != nil {
				m.logger.Error("Failed to restore instance", "id", id, "error", err)
				inst.Status = "Error"
			}
		}
	}
	return nil
}

// Spawn starts a new game server instance.
// It initializes the instance record and starts the provisioning process in the background.
func (m *Manager) Spawn(ctx context.Context) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	port, err := m.findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	id := fmt.Sprintf("%s-%d", m.cfg.Region, port)
	instanceDir := filepath.Join(m.cfg.InstancesDir, id)

	instance := &Instance{
		ID:        id,
		Port:      port,
		Status:    "Provisioning",
		Region:    m.cfg.Region,
		StartTime: time.Now(),
		Path:      instanceDir,
	}

	m.instances[id] = instance
	if err := m.SaveState(); err != nil {
		m.logger.Error("Failed to save state during spawn", "error", err)
	}

	m.logger.Info("Starting async provisioning for new instance", "id", id, "port", port)
	
	// Run provisioning in background to avoid blocking the API request (and avoiding timeouts)
	go m.provisionAndStart(instance)

	return instance, nil
}

// provisionAndStart handles the heavy lifting of copying files and starting the process.
func (m *Manager) provisionAndStart(inst *Instance) {
	// Ensure instances directory exists
	if err := os.MkdirAll(m.cfg.InstancesDir, 0755); err != nil {
		m.setErrorState(inst, fmt.Errorf("failed to create instances directory: %w", err))
		return
	}

	// Copy game files to new instance directory
	m.logger.Info("Copying game files...", "id", inst.ID, "dir", inst.Path)
	if err := copyDir(m.cfg.GameInstallDir, inst.Path); err != nil {
		m.setErrorState(inst, fmt.Errorf("failed to copy game files: %w", err))
		return
	}

	// Start the process
	if err := m.startProcess(inst); err != nil {
		m.setErrorState(inst, fmt.Errorf("failed to start process: %w", err))
		return
	}

	m.logger.Info("Instance provisioning complete and started", "id", inst.ID)
	// startProcess sets Status to Running and saves state
}

func (m *Manager) setErrorState(inst *Instance, err error) {
	m.logger.Error("Provisioning failed", "id", inst.ID, "error", err)
	m.mu.Lock()
	inst.Status = "Error"
	m.mu.Unlock()
	m.SaveState()
}

// startProcess constructs the command and starts the process for an instance.
// It assumes the instance directory and files are already set up.
func (m *Manager) startProcess(inst *Instance) error {
	// Construct absolute path to the binary within the instance directory
	// We treat m.cfg.GameBinaryPath as relative to the instance root
	binaryPath := filepath.Join(inst.Path, m.cfg.GameBinaryPath)
	
	// Resolve absolute path for the command execution to be safe
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute binary path: %w", err)
	}

	// Ensure binary is executable (no-op on Windows usually, but good practice)
	_ = os.Chmod(absBinaryPath, 0755)

	cmd := exec.Command(absBinaryPath,
		"-batchmode",
		"-nographics",
		"-mode", "server",
		"-port", fmt.Sprintf("%d", inst.Port),
	)
	
	cmd.Dir = inst.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process %s: %w", absBinaryPath, err)
	}

	m.mu.Lock()
	inst.cmd = cmd
	inst.ProcessID = cmd.Process.Pid
	inst.Status = "Running"
	m.mu.Unlock()
	
	m.SaveState()
	
	// Monitor in background
	go m.monitorInstance(inst.ID, cmd)
	
	return nil
}

// monitorInstance waits for the process to exit and cleans up.
func (m *Manager) monitorInstance(id string, cmd *exec.Cmd) {
	err := cmd.Wait()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if instance, exists := m.instances[id]; exists {
		// Only update if the command matches (handling restarts/race conditions)
		if instance.cmd == cmd {
			instance.Status = "Stopped"
			instance.ProcessID = 0
			instance.cmd = nil
			
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

			if err := m.SaveState(); err != nil {
				m.logger.Error("Failed to save state after instance stop", "error", err)
			}
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

	if instance.Status != "Running" {
		return fmt.Errorf("instance is not running")
	}

	if instance.cmd != nil && instance.cmd.Process != nil {
		if err := instance.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	
	// monitorInstance will update status and save state
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
	// We don't clear map here, so state persists
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
func (m *Manager) findAvailablePort() (int, error) {
	for port := m.cfg.MinGamePort; port <= m.cfg.MaxGamePort; port++ {
		// 1. Check internal tracking (including stopped instances to reserve port)
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
