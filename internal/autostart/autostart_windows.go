//go:build windows

package autostart

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName = "绿橙"
)

func SetEnabled(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if !enabled {
		err = key.DeleteValue(appName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	return key.SetStringValue(appName, `"`+exe+`"`)
}

func IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
