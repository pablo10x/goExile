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
	"spawner/internal/updater"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// InstanceStats holds resource usage information.
type InstanceStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage uint64  `json:"memory_usage"`
	DiskUsage   uint64  `json:"disk_usage"`
}

// Instance represents a running game server.
type Instance struct {
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	ProcessID int       `json:"pid"`
	Status    string    `json:"status"` // "Running", "Stopped", "Error"
	Region    string    `json:"region"`
	Version   string    `json:"version"`
	StartTime time.Time `json:"start_time"`
	Path      string    `json:"path"` // Path to this instance's directory

	cmd *exec.Cmd // Private: command handle for process management
}

// Manager handles the lifecycle of game server processes.
type Manager struct {
	cfg       *config.Config
	instances map[string]*Instance // Key is ID (string)
	mu        sync.RWMutex         // RWMutex for better concurrency
	busy      bool                 // If true, manager is performing a global operation (update)
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

// IsBusy returns true if the manager is currently performing a blocking operation.
func (m *Manager) IsBusy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.busy
}

// UpdateTemplate wraps the updater logic and sets the busy state.
func (m *Manager) UpdateTemplate() (string, error) {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return "", fmt.Errorf("spawner is already busy")
	}
	m.busy = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.busy = false
		m.mu.Unlock()
	}()

	// Call updater (this is blocking and time consuming)
	return updater.UpdateTemplate(m.cfg, m.logger)
}

// RestoreInstances loads state from disk and restarts servers that should be running.
func (m *Manager) RestoreInstances() error {
	if err := m.LoadState(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, inst := range m.instances {
		// Ensure version is loaded from disk if missing
		if inst.Version == "" {
			inst.Version = m.readVersionFile(inst.Path)
		}

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

	if m.busy {
		return nil, fmt.Errorf("spawner is busy updating")
	}

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
	if err := m.saveStateInternal(); err != nil {
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

	// Read version from the copied files
	inst.Version = m.readVersionFile(inst.Path)

	// Start the process (requires lock)
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.startProcess(inst); err != nil {
		// If start fails, we are already locked, so we can set error state directly or call internal helper
		m.logger.Error("Provisioning failed", "id", inst.ID, "error", err)
		inst.Status = "Error"
		m.saveStateInternal()
		return
	}

	m.logger.Info("Instance provisioning complete and started", "id", inst.ID)
	m.saveStateInternal()
}

// UpdateInstance stops the instance if running and re-copies the game files.
func (m *Manager) UpdateInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	// Stop if running
	if inst.Status == "Running" && inst.cmd != nil && inst.cmd.Process != nil {
		m.logger.Info("Stopping instance for update", "id", id)
		_ = inst.cmd.Process.Kill()
		inst.cmd.Wait() // Wait for exit
		inst.Status = "Stopped"
		inst.ProcessID = 0
		inst.cmd = nil
	}

	// Update local template if needed
	// NOTE: We don't call UpdateTemplate here because it would trigger busy state recursively if we implemented it that way.
	// But updater.UpdateTemplate is safe to call if we are already holding lock? No, updater.UpdateTemplate takes time.
	// We should probably rely on the global template being updated separately, OR call it here but we are holding the lock.
	// If we hold the lock during download, EVERYTHING freezes.
	// So we should NOT call updater.UpdateTemplate inside UpdateInstance if possible, or we accept the freeze.
	// For now, keeping original logic but without setting global busy flag since this is a specific instance update.
	// However, if we want to ensure template is fresh, we might need to. 
	// Let's assume UpdateInstance uses the *current* template. 
	// The original code called updater.UpdateTemplate. I will keep it but be aware it blocks.
	
	if _, err := updater.UpdateTemplate(m.cfg, m.logger); err != nil {
		m.logger.Warn("Failed to update template from master", "error", err)
		inst.Status = "Error"
		m.saveStateInternal()
		return fmt.Errorf("failed to pull update from master: %w", err)
	}

	m.logger.Info("Updating instance files", "id", id)
	if err := copyDir(m.cfg.GameInstallDir, inst.Path); err != nil {
		inst.Status = "Error"
		m.saveStateInternal()
		return fmt.Errorf("failed to update game files: %w", err)
	}

	inst.Version = m.readVersionFile(inst.Path)

	m.logger.Info("Instance updated successfully", "id", id)
	return m.saveStateInternal()
}

// RenameInstance renames an instance ID and its directory.
func (m *Manager) RenameInstance(id string, newID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	if _, exists := m.instances[newID]; exists {
		return fmt.Errorf("instance with ID %s already exists", newID)
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	// Stop if running
	if inst.Status == "Running" && inst.cmd != nil && inst.cmd.Process != nil {
		m.logger.Info("Stopping instance for rename", "id", id)
		_ = inst.cmd.Process.Kill()
		inst.cmd.Wait()
		inst.Status = "Stopped"
		inst.ProcessID = 0
		inst.cmd = nil
	}

	oldPath := inst.Path
	newPath := filepath.Join(m.cfg.InstancesDir, newID)

	m.logger.Info("Renaming instance", "old_id", id, "new_id", newID, "old_path", oldPath, "new_path", newPath)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory: %w", err)
	}

	// Update Instance
	inst.ID = newID
	inst.Path = newPath

	// Update Map
	delete(m.instances, id)
	m.instances[newID] = inst

	return m.saveStateInternal()
}

func (m *Manager) setErrorState(inst *Instance, err error) {
	m.logger.Error("Provisioning failed", "id", inst.ID, "error", err)
	m.mu.Lock()
	inst.Status = "Error"
	m.saveStateInternal()
	m.mu.Unlock()
}

// startProcess constructs the command and starts the process for an instance.
// It assumes the instance directory and files are already set up.
// Caller MUST hold the lock.
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

	// Prepare arguments for the game server binary
	args := []string{
		"-batchmode",
		"-nographics",
		"-mode", "server",
		"-port", fmt.Sprintf("%d", inst.Port),
	}

	logFilePath := filepath.Join(inst.Path, "gameserver.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
	}
	defer logFile.Close() // Close parent's handle, child inherits it

	// Open firewall port
	m.openFirewallPort(inst.Port)

	cmd := newGameCmd(absBinaryPath, args, logFile)
	cmd.Dir = inst.Path

	if err := cmd.Start(); err != nil {
		// Failed to start, close port to be clean
		m.closeFirewallPort(inst.Port)
		return fmt.Errorf("failed to start process %s: %w", absBinaryPath, err)
	}

	inst.cmd = cmd
	inst.ProcessID = cmd.Process.Pid
	inst.Status = "Running"

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

			// Close firewall port
			m.closeFirewallPort(instance.Port)

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

			if err := m.saveStateInternal(); err != nil {
				m.logger.Error("Failed to save state after instance stop", "error", err)
			}
		}
	}
}

// StopInstance kills a specific game server and waits for it to stop.
func (m *Manager) StopInstance(id string) error {
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return fmt.Errorf("spawner is busy updating")
	}

	instance, exists := m.instances[id]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("instance not found")
	}

	if instance.Status != "Running" {
		m.mu.Unlock()
		return fmt.Errorf("instance is not running")
	}

	if instance.cmd != nil && instance.cmd.Process != nil {
		if err := instance.cmd.Process.Kill(); err != nil {
			m.mu.Unlock()
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	m.mu.Unlock()

	// Wait for monitorInstance to update status
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for instance to stop")
		case <-ticker.C:
			m.mu.RLock()
			inst, ok := m.instances[id]
			status := ""
			if ok {
				status = inst.Status
			}
			m.mu.RUnlock()
			
			if !ok || status == "Stopped" || status == "Error" {
				return nil
			}
		}
	}
}

// RemoveInstance deletes a stopped instance from disk and registry.
func (m *Manager) RemoveInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	if inst.Status == "Running" {
		return fmt.Errorf("instance is running, stop it first")
	}

	m.logger.Info("Removing instance", "id", id, "path", inst.Path)

	// Remove files
	if err := os.RemoveAll(inst.Path); err != nil {
		m.logger.Error("Failed to remove instance files", "id", id, "error", err)
		// We continue to remove from memory/state even if file deletion fails
	}

	delete(m.instances, id)
	return m.saveStateInternal()
}

// StartInstance starts a previously stopped or errored game server instance.
func (m *Manager) StartInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	instance, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance with ID %s not found", id)
	}

	if instance.Status == "Running" {
		return fmt.Errorf("instance with ID %s is already running", id)
	}

	m.logger.Info("Attempting to start instance", "id", id, "status", instance.Status)

	// The startProcess function expects the instance directory to be set up
	// and will handle setting the status to "Running" and launching monitorInstance.
	if err := m.startProcess(instance); err != nil {
		m.logger.Error("Failed to start process for instance", "id", id, "error", err)
		instance.Status = "Error" // Mark as error if starting fails
		m.saveStateInternal()
		return fmt.Errorf("failed to start instance %s: %w", id, err)
	}

	// Instance status is now "Running" from startProcess.
	// Save the updated state to disk.
	if err := m.saveStateInternal(); err != nil {
		m.logger.Error("Failed to save state after starting instance", "id", id, "error", err)
		// This is a warning, the instance is running, but state might be inconsistent
	}

	m.logger.Info("Instance started successfully", "id", id, "port", instance.Port)
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

// GetInstanceLogPath returns the absolute path to the log file for an instance.
func (m *Manager) GetInstanceLogPath(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.instances[id]
	if !exists {
		return "", fmt.Errorf("instance not found")
	}

	return filepath.Join(inst.Path, "gameserver.log"), nil
}

// ClearInstanceLogs truncates the log file for an instance.
func (m *Manager) ClearInstanceLogs(id string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	logPath := filepath.Join(inst.Path, "gameserver.log")
	if err := os.Truncate(logPath, 0); err != nil {
		return fmt.Errorf("failed to truncate log file: %w", err)
	}
	return nil
}

// GetInstanceStats returns resource usage statistics for an instance.
func (m *Manager) GetInstanceStats(id string) (*InstanceStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.instances[id]
	if !exists {
		return nil, fmt.Errorf("instance not found")
	}

	stats := &InstanceStats{}

	// Disk Usage
	var size uint64
	err := filepath.Walk(inst.Path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	if err != nil {
		m.logger.Warn("Failed to calculate disk usage", "id", id, "error", err)
	}
	stats.DiskUsage = size

	// Process Stats (if running)
	if inst.Status == "Running" && inst.ProcessID > 0 {
		p, err := process.NewProcess(int32(inst.ProcessID))
		if err == nil {
			cpu, _ := p.Percent(0)
			mem, _ := p.MemoryInfo()
			
			stats.CPUPercent = cpu
			if mem != nil {
				stats.MemoryUsage = mem.RSS
			}
		}
	}

	return stats, nil
}

// readVersionFile attempts to read the version string from version.txt in the instance directory.
func (m *Manager) readVersionFile(dir string) string {
	path := filepath.Join(dir, "version.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
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

func (m *Manager) openFirewallPort(port int) {
	cmd := exec.Command("ufw", "allow", fmt.Sprintf("%d", port))
	if out, err := cmd.CombinedOutput(); err != nil {
		m.logger.Error("Failed to open firewall port", "port", port, "error", err, "output", string(out))
	} else {
		m.logger.Info("Opened firewall port", "port", port)
	}
}

func (m *Manager) closeFirewallPort(port int) {
	cmd := exec.Command("ufw", "delete", "allow", fmt.Sprintf("%d", port))
	if out, err := cmd.CombinedOutput(); err != nil {
		m.logger.Error("Failed to close firewall port", "port", port, "error", err, "output", string(out))
	} else {
		m.logger.Info("Closed firewall port", "port", port)
	}
}

// BackupInstance creates a zip backup of the instance directory.
func (m *Manager) BackupInstance(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	if inst.Status == "Running" {
		return fmt.Errorf("instance must be stopped to backup")
	}

	backupDir := filepath.Join(inst.Path, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
    version := inst.Version
    if version == "" {
        version = "unknown"
    }
    // Clean version string
    version = strings.ReplaceAll(version, " ", "_")
	filename := fmt.Sprintf("backup_%s_v%s.zip", timestamp, version)
	backupPath := filepath.Join(backupDir, filename)

	m.logger.Info("Creating backup", "id", id, "file", filename)

	// Exclude "backups" folder and log files
	excludes := []string{"backups", "gameserver.log"}
	if err := zipDir(inst.Path, backupPath, excludes); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	return nil
}

// RestoreInstance restores a backup, overwriting current files.
func (m *Manager) RestoreInstance(id string, filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	if inst.Status == "Running" {
		return fmt.Errorf("instance must be stopped to restore")
	}

	backupDir := filepath.Join(inst.Path, "backups")
	backupPath := filepath.Join(backupDir, filename)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found")
	}

	m.logger.Info("Restoring backup", "id", id, "file", filename)

	// Wipe directory except backups
	entries, err := os.ReadDir(inst.Path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() == "backups" {
			continue
		}
		os.RemoveAll(filepath.Join(inst.Path, entry.Name()))
	}

	// Unzip
	if err := unzipDir(backupPath, inst.Path); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

    // Refresh version info from the restored files
    inst.Version = m.readVersionFile(inst.Path)
    m.saveStateInternal()

	return nil
}

// DeleteBackup deletes a specific backup file.
func (m *Manager) DeleteBackup(id string, filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.busy {
		return fmt.Errorf("spawner is busy updating")
	}

	inst, exists := m.instances[id]
	if !exists {
		return fmt.Errorf("instance not found")
	}

	backupPath := filepath.Join(inst.Path, "backups", filename)
	// Security check to prevent directory traversal
    if filepath.Dir(backupPath) != filepath.Join(inst.Path, "backups") {
         return fmt.Errorf("invalid filename")
    }

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found")
	}

	m.logger.Info("Deleting backup", "id", id, "file", filename)
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// ListBackups returns a list of backup files for an instance.
func (m *Manager) ListBackups(id string) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.instances[id]
	if !exists {
		return nil, fmt.Errorf("instance not found")
	}

	backupDir := filepath.Join(inst.Path, "backups")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	backups := []map[string]interface{}{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
			info, _ := entry.Info()
			backups = append(backups, map[string]interface{}{
				"filename": entry.Name(),
				"size":     info.Size(),
				"date":     info.ModTime(),
			})
		}
	}
	return backups, nil
}
