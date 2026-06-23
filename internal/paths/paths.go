package paths

import (
	"os"
	"path/filepath"
)

func AppDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "vpn_clash")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func ConfigFile() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func HomeDir() (string, error) {
	dir, err := AppDataDir()
	if err != nil {
		return "", err
	}
	home := filepath.Join(dir, "mihomo")
	if err := os.MkdirAll(home, 0o755); err != nil {
		return "", err
	}
	return home, nil
}
