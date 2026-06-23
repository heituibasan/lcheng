package settings

import (
	"encoding/json"
	"os"
	"sync"

	"vpn_clash/internal/paths"
)

type Settings struct {
	AutoStartCore         bool   `json:"autoStartCore"`
	AutoSystemProxy       bool   `json:"autoSystemProxy"`
	SystemProxyEnabled    bool   `json:"systemProxyEnabled"`
	MixedPort             int    `json:"mixedPort"`
	TunEnabled            bool   `json:"tunEnabled"`
	AllowLan              bool   `json:"allowLan"`
	LogLevel              string `json:"logLevel"`
	SubscriptionUserAgent string `json:"subscriptionUserAgent"`
	ActiveProfileID       string                         `json:"activeProfileId"`
	ProxySelections       map[string]map[string]string   `json:"proxySelections"`
	StartMinimized        bool                           `json:"startMinimized"`
	LaunchAtLogin         bool                           `json:"launchAtLogin"`
}

func Default() Settings {
	return Settings{
		AutoStartCore:         false,
		AutoSystemProxy:       true,
		SystemProxyEnabled:    false,
		MixedPort:             7890,
		TunEnabled:            false,
		AllowLan:              false,
		LogLevel:              "info",
		SubscriptionUserAgent: "clash-verge/v1.0",
		ActiveProfileID:       "default",
		ProxySelections:       make(map[string]map[string]string),
		StartMinimized:        false,
		LaunchAtLogin:         false,
	}
}

type Store struct {
	mu   sync.RWMutex
	path string
	data Settings
}

func NewStore() (*Store, error) {
	dir, err := paths.AppDataDir()
	if err != nil {
		return nil, err
	}
	s := &Store{path: dir + string(os.PathSeparator) + "settings.json"}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	s.data = Default()
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Save(s.data)
		}
		return err
	}
	return json.Unmarshal(raw, &s.data)
}

func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *Store) Save(data Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data.MixedPort <= 0 {
		data.MixedPort = 7890
	}
	if data.LogLevel == "" {
		data.LogLevel = "info"
	}
	if data.SubscriptionUserAgent == "" {
		data.SubscriptionUserAgent = "clash-verge/v1.0"
	}
	s.data = data
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}
