//go:build !windows

package autostart

import "fmt"

func SetEnabled(enabled bool) error {
	if enabled {
		return fmt.Errorf("launch at login is only supported on Windows")
	}
	return nil
}

func IsEnabled() (bool, error) {
	return false, nil
}
