// Package state handles persistent editor state (history, sessions, etc.)
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/wx13/sith/config"
)

const (
	maxHistoryEntries = 100
	historyFile       = "history.json"
)

// History holds search/replace/goto history across sessions.
type History struct {
	Search  []string `json:"search"`
	Replace []string `json:"replace"`
	Goto    []string `json:"goto"`

	mu   sync.Mutex
	path string
}

// NewHistory creates a new History, loading from disk if available.
func NewHistory() *History {
	h := &History{
		Search:  []string{},
		Replace: []string{},
		Goto:    []string{},
	}

	configDir := config.ConfigDir()
	if configDir == "" {
		return h
	}

	h.path = filepath.Join(configDir, historyFile)
	h.load()
	return h
}

// load reads history from disk.
func (h *History) load() {
	data, err := os.ReadFile(h.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, h)
}

// Save writes history to disk.
func (h *History) Save() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.path == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(h.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0644)
}

// AddSearch adds a search term to history.
func (h *History) AddSearch(term string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Search = addToHistory(h.Search, term)
}

// AddReplace adds a replace term to history.
func (h *History) AddReplace(term string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Replace = addToHistory(h.Replace, term)
}

// AddGoto adds a goto term to history.
func (h *History) AddGoto(term string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Goto = addToHistory(h.Goto, term)
}

// GetSearch returns search history (most recent first).
func (h *History) GetSearch() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return reverseStrings(h.Search)
}

// GetReplace returns replace history (most recent first).
func (h *History) GetReplace() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return reverseStrings(h.Replace)
}

// GetGoto returns goto history (most recent first).
func (h *History) GetGoto() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return reverseStrings(h.Goto)
}

// addToHistory adds a term to a history slice, avoiding duplicates and
// enforcing max size. Returns the updated slice.
func addToHistory(history []string, term string) []string {
	if term == "" {
		return history
	}

	// Remove existing occurrence if present
	for i, t := range history {
		if t == term {
			history = append(history[:i], history[i+1:]...)
			break
		}
	}

	// Add to end (most recent)
	history = append(history, term)

	// Trim to max size
	if len(history) > maxHistoryEntries {
		history = history[len(history)-maxHistoryEntries:]
	}

	return history
}

// reverseStrings returns a reversed copy of the slice.
func reverseStrings(s []string) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}
