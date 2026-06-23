package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub"
	"github.com/metacubex/mihomo/hub/executor"

	appconfig "vpn_clash/internal/config"
	"vpn_clash/internal/paths"
)

var (
	ErrAlreadyRunning = errors.New("core is already running")
	ErrNotRunning     = errors.New("core is not running")
)

type Manager struct {
	mu      sync.Mutex
	running bool
	store   *appconfig.Store
}

func NewManager(store *appconfig.Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return ErrAlreadyRunning
	}

	homeDir, err := paths.HomeDir()
	if err != nil {
		return err
	}
	C.SetHomeDir(homeDir)

	configPath, err := paths.ConfigFile()
	if err != nil {
		return err
	}
	C.SetConfig(configPath)

	if err := config.Init(homeDir); err != nil {
		return fmt.Errorf("init config directory: %w", err)
	}

	content := m.store.Get()
	if err := hub.Parse([]byte(content)); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	m.running = true
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return ErrNotRunning
	}

	executor.Shutdown()
	m.running = false
	return nil
}

func (m *Manager) Reload(content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return ErrNotRunning
	}

	cfg, err := executor.ParseWithBytes([]byte(content))
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	hub.ApplyConfig(cfg)
	return nil
}

func (m *Manager) TestConfig(content string) error {
	homeDir, err := paths.HomeDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(homeDir, "rules"), 0o755); err != nil {
		return err
	}

	prevHome := C.Path.HomeDir()
	C.SetHomeDir(homeDir)
	defer C.SetHomeDir(prevHome)

	if err := config.Init(homeDir); err != nil {
		return err
	}

	_, err = executor.ParseWithBytes([]byte(content))
	return err
}
