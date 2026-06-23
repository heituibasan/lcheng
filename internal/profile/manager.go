package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"vpn_clash/internal/paths"
)

type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Filename  string `json:"filename"`
	SourceURL string `json:"sourceUrl,omitempty"`
	UpdatedAt int64  `json:"updatedAt"`
}

type Manager struct {
	mu       sync.RWMutex
	dir      string
	metaPath string
	items    []Profile
}

func NewManager() (*Manager, error) {
	dir, err := paths.AppDataDir()
	if err != nil {
		return nil, err
	}
	profileDir := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return nil, err
	}

	m := &Manager{
		dir:      profileDir,
		metaPath: filepath.Join(dir, "profiles.json"),
	}
	if err := m.load(); err != nil {
		return nil, err
	}
	if len(m.items) == 0 {
		if err := m.ensureDefault(); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) load() error {
	m.items = []Profile{}
	raw, err := os.ReadFile(m.metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(raw, &m.items)
}

func (m *Manager) save() error {
	raw, err := json.MarshalIndent(m.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.metaPath, raw, 0o644)
}

func (m *Manager) ensureDefault() error {
	id := "default"
	filename := "default.yaml"
	path := filepath.Join(m.dir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			return err
		}
	}
	m.items = []Profile{{
		ID:        id,
		Name:      "默认配置",
		Filename:  filename,
		UpdatedAt: time.Now().Unix(),
	}}
	return m.save()
}

func (m *Manager) List() []Profile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Profile, len(m.items))
	copy(out, m.items)
	return out
}

func (m *Manager) CreateWithSource(name, sourceURL, content string) (Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sourceURL != "" {
		for i, item := range m.items {
			if item.SourceURL == sourceURL {
				if err := os.WriteFile(filepath.Join(m.dir, item.Filename), []byte(content), 0o644); err != nil {
					return Profile{}, err
				}
				m.items[i].UpdatedAt = time.Now().Unix()
				if name != "" {
					m.items[i].Name = name
				}
				if err := m.save(); err != nil {
					return Profile{}, err
				}
				return m.items[i], nil
			}
		}
	}

	id := fmt.Sprintf("profile_%d", time.Now().UnixNano())
	filename := id + ".yaml"
	if err := os.WriteFile(filepath.Join(m.dir, filename), []byte(content), 0o644); err != nil {
		return Profile{}, err
	}
	item := Profile{
		ID:        id,
		Name:      name,
		Filename:  filename,
		SourceURL: sourceURL,
		UpdatedAt: time.Now().Unix(),
	}
	m.items = append(m.items, item)
	if err := m.save(); err != nil {
		return Profile{}, err
	}
	return item, nil
}

func (m *Manager) Create(name, content string) (Profile, error) {
	return m.CreateWithSource(name, "", content)
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	next := make([]Profile, 0, len(m.items))
	for _, item := range m.items {
		if item.ID == id {
			_ = os.Remove(filepath.Join(m.dir, item.Filename))
			continue
		}
		next = append(next, item)
	}
	m.items = next
	return m.save()
}

func (m *Manager) Read(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.items {
		if item.ID == id {
			data, err := os.ReadFile(filepath.Join(m.dir, item.Filename))
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("profile not found")
}

func (m *Manager) Write(id, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.items {
		if item.ID == id {
			if err := os.WriteFile(filepath.Join(m.dir, item.Filename), []byte(content), 0o644); err != nil {
				return err
			}
			m.items[i].UpdatedAt = time.Now().Unix()
			return m.save()
		}
	}
	return fmt.Errorf("profile not found")
}

func (m *Manager) Rename(id, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.items {
		if item.ID == id {
			m.items[i].Name = name
			m.items[i].UpdatedAt = time.Now().Unix()
			return m.save()
		}
	}
	return fmt.Errorf("profile not found")
}
