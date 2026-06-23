package config

import (
	"errors"
	"os"
	"sync"

	"gopkg.in/yaml.v3"

	"vpn_clash/internal/paths"
)

type Store struct {
	mu       sync.RWMutex
	filePath string
	content  string
}

func NewStore() (*Store, error) {
	filePath, err := paths.ConfigFile()
	if err != nil {
		return nil, err
	}

	store := &Store{filePath: filePath}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.content = DefaultYAML
			return s.Save(s.content)
		}
		return err
	}
	s.content = string(data)
	return nil
}

func (s *Store) Path() string {
	return s.filePath
}

func (s *Store) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.content
}

func (s *Store) Save(content string) error {
	if err := Validate(content); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.WriteFile(s.filePath, []byte(content), 0o644); err != nil {
		return err
	}
	s.content = content
	return nil
}

func Validate(content string) error {
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return err
	}
	return nil
}

func (s *Store) Controller() (address string, secret string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(s.content), &raw); err != nil {
		return "127.0.0.1:9090", ""
	}

	if v, ok := raw["external-controller"].(string); ok && v != "" {
		address = v
	} else {
		address = "127.0.0.1:9090"
	}

	if v, ok := raw["secret"].(string); ok {
		secret = v
	}
	return address, secret
}

func parsePort(raw map[string]any, key string, fallback int) int {
	v, ok := raw[key]
	if !ok {
		return fallback
	}
	switch n := v.(type) {
	case int:
		if n > 0 {
			return n
		}
	case int64:
		if n > 0 {
			return int(n)
		}
	case float64:
		if n > 0 {
			return int(n)
		}
	}
	return fallback
}

func (s *Store) MixedPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(s.content), &raw); err != nil {
		return 7890
	}
	return parsePort(raw, "mixed-port", 7890)
}
