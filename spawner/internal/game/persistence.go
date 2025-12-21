package game

import (
	"encoding/json"
	"os"
)

// SaveState writes the current instances to a JSON file.
// It acquires a read lock.
func (m *Manager) SaveState() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.saveStateInternal()
}

// saveStateInternal writes state to disk without locking.
// Caller MUST hold at least a read lock.
func (m *Manager) saveStateInternal() error {
	data, err := json.MarshalIndent(m.instances, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.cfg.StateFilePath, data, 0600)
}

// LoadState reads instances from the JSON file.
func (m *Manager) LoadState() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.cfg.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet, start fresh
		}
		return err
	}

	var loadedInstances map[string]*Instance
	if err := json.Unmarshal(data, &loadedInstances); err != nil {
		return err
	}

	m.instances = loadedInstances
	return nil
}
