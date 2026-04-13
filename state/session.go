package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/wx13/sith/config"
)

const (
	sessionsDir       = "sessions"
	maxSessionAge     = 30 * 24 * time.Hour // 30 days
	maxSessionCount   = 50
)

// FileState holds the state of a single open file.
type FileState struct {
	Path   string `json:"path"`
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Active bool   `json:"active,omitempty"`
}

// Session holds the state of an editor session.
type Session struct {
	Files     []FileState `json:"files"`
	Timestamp time.Time   `json:"timestamp"`

	path string
}

// NewSession creates a new Session for the current working directory.
func NewSession() *Session {
	s := &Session{
		Files:     []FileState{},
		Timestamp: time.Now(),
	}

	configDir := config.ConfigDir()
	if configDir == "" {
		return s
	}

	cwd, err := os.Getwd()
	if err != nil {
		return s
	}

	// Create a hash of the cwd for the session filename
	hash := sha256.Sum256([]byte(cwd))
	hashStr := hex.EncodeToString(hash[:8]) // Use first 8 bytes (16 hex chars)

	s.path = filepath.Join(configDir, sessionsDir, hashStr+".json")
	return s
}

// Load reads the session from disk.
func (s *Session) Load() error {
	if s.path == "" {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}

// Save writes the session to disk.
func (s *Session) Save() error {
	if s.path == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	s.Timestamp = time.Now()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(s.path, data, 0644)
	if err != nil {
		return err
	}

	// Cleanup old sessions
	CleanupSessions()

	return nil
}

// HasSession returns true if a saved session exists for the current directory.
func (s *Session) HasSession() bool {
	if s.path == "" {
		return false
	}
	_, err := os.Stat(s.path)
	return err == nil
}

// Clear removes the session file.
func (s *Session) Clear() error {
	if s.path == "" {
		return nil
	}
	return os.Remove(s.path)
}

// AddFile adds a file to the session.
func (s *Session) AddFile(path string, row, col int, active bool) {
	s.Files = append(s.Files, FileState{
		Path:   path,
		Row:    row,
		Col:    col,
		Active: active,
	})
}

// GetFiles returns the list of files in the session.
func (s *Session) GetFiles() []FileState {
	return s.Files
}

// Age returns how long ago the session was saved.
func (s *Session) Age() time.Duration {
	return time.Since(s.Timestamp)
}

// CleanupSessions removes old session files.
// Called automatically when saving a session.
func CleanupSessions() {
	configDir := config.ConfigDir()
	if configDir == "" {
		return
	}

	sessDir := filepath.Join(configDir, sessionsDir)
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return
	}

	type sessionFile struct {
		path    string
		modTime time.Time
	}

	var sessions []sessionFile
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(sessDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Delete if older than maxSessionAge
		if now.Sub(info.ModTime()) > maxSessionAge {
			os.Remove(path)
			continue
		}

		sessions = append(sessions, sessionFile{
			path:    path,
			modTime: info.ModTime(),
		})
	}

	// If still over limit, delete oldest
	if len(sessions) > maxSessionCount {
		// Sort by mod time, oldest first
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].modTime.Before(sessions[j].modTime)
		})

		// Remove oldest until we're under the limit
		for i := 0; i < len(sessions)-maxSessionCount; i++ {
			os.Remove(sessions[i].path)
		}
	}
}
