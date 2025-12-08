package game

import (
	"encoding/json"
	"os"
)

// SaveState writes the current instances to a JSON file.
func (m *Manager) SaveState() error {
	// We only need read lock to marshal, but since we might be writing to a file,
	// let's just ensure we have a stable view.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert map to slice or just save the map directly
	// We need to exclude the 'cmd' field which is not serializable.
	// The Instance struct already has json tags.

	data, err := json.MarshalIndent(m.instances, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.cfg.StateFilePath, data, 0644)
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
