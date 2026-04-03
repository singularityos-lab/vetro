package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	Metadata *Metadata
	CacheDir string
}

func NewManager() (*Manager, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	cacheDir := filepath.Join(dataHome, "vetro", "metadata")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		CacheDir: cacheDir,
	}, nil
}

func (m *Manager) LoadGtkMetadata(girPath string, force bool) error {
	return m.loadGIRFile(girPath, force)
}

func (m *Manager) loadGIRFile(girPath string, force bool) error {
	hash, err := fileHash(girPath)
	if err != nil {
		return fmt.Errorf("failed to hash GIR: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(girPath), ".gir")
	cacheFile := filepath.Join(m.CacheDir, fmt.Sprintf("gir-%s-%s.json", baseName, hash))

	var meta *Metadata

	if !force {
		if data, err := os.ReadFile(cacheFile); err == nil {
			var cached Metadata
			if err := json.Unmarshal(data, &cached); err == nil {
				meta = &cached
			}
		}
	}

	if meta == nil {
		parsed, err := ParseGIR(girPath)
		if err != nil {
			return err
		}
		data, err := json.Marshal(parsed)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cacheFile, data, 0644); err != nil {
			return err
		}
		meta = parsed
	}

	if m.Metadata == nil {
		m.Metadata = meta
	} else {
		m.Metadata.Merge(meta)
	}

	return nil
}

// LoadAllGIRs discovers and loads all .gir files found in standard paths.
func (m *Manager) LoadAllGIRs(force bool) error {
	paths := FindAllGIRs()
	if len(paths) == 0 {
		return nil
	}
	if m.Metadata == nil {
		m.Metadata = &Metadata{Classes: make(map[string]*ClassMetadata)}
	}
	for _, p := range paths {
		if err := m.loadGIRFile(p, force); err != nil {
			// Skip GIRs that fail to parse (e.g., malformed third-party files)
			continue
		}
	}
	return nil
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// FindGIR tries to locate the Gtk-4.0.gir file.
func FindGIR() string {
	out, err := exec.Command("pkg-config", "--variable=girdir", "gobject-introspection-1.0").Output()
	if err == nil {
		girDir := strings.TrimSpace(string(out))
		girPath := filepath.Join(girDir, "Gtk-4.0.gir")
		if _, err := os.Stat(girPath); err == nil {
			return girPath
		}
	}

	paths := []string{
		"/usr/share/gir-1.0/Gtk-4.0.gir",
		"/usr/local/share/gir-1.0/Gtk-4.0.gir",
		"/app/share/gir-1.0/Gtk-4.0.gir",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// FindAllGIRs returns paths to all .gir files found in standard directories,
// including user-local paths (~/.local/share/gir-1.0).
func FindAllGIRs() []string {
	girDirs := []string{
		"/usr/share/gir-1.0",
		"/usr/local/share/gir-1.0",
		"/app/share/gir-1.0",
	}

	// User-local GIR directories (higher priority — prepended)
	if home, err := os.UserHomeDir(); err == nil {
		userGirDirs := []string{
			filepath.Join(home, ".local", "share", "gir-1.0"),
			filepath.Join(home, ".local", "lib", "girepository-1.0"),
		}
		// Also respect XDG_DATA_HOME
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			userGirDirs = append([]string{filepath.Join(xdgData, "gir-1.0")}, userGirDirs...)
		}
		girDirs = append(userGirDirs, girDirs...)
	}

	if out, err := exec.Command("pkg-config", "--variable=girdir", "gobject-introspection-1.0").Output(); err == nil {
		if dir := strings.TrimSpace(string(out)); dir != "" {
			girDirs = append([]string{dir}, girDirs...)
		}
	}

	seen := make(map[string]bool)
	var result []string

	for _, dir := range girDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".gir") {
				continue
			}
			full := filepath.Join(dir, e.Name())
			if !seen[full] {
				seen[full] = true
				result = append(result, full)
			}
		}
	}

	return result
}
