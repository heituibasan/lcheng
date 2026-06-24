package geobundle

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/metacubex/mihomo/component/mmdb"

	"vpn_clash/internal/paths"
)

//go:embed data/*
var embeddedData embed.FS

// Ensure extracts bundled Geo databases into the mihomo home directory when
// local copies are missing or invalid, so the core never needs to download
// them on first launch.
func Ensure() error {
	homeDir, err := paths.HomeDir()
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(embeddedData, "data")
	if err != nil {
		return fmt.Errorf("read embedded geodata: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := ensureFile(homeDir, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func ensureFile(homeDir, name string) error {
	target := filepath.Join(homeDir, name)
	if isValid(name, target) {
		return nil
	}

	raw, err := embeddedData.ReadFile("data/" + name)
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", name, err)
	}
	if len(raw) == 0 {
		return fmt.Errorf("embedded %s is empty", name)
	}

	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	_ = os.Remove(target)
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("install %s: %w", name, err)
	}
	return nil
}

func isValid(name, path string) bool {
	nameLower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(nameLower, ".metadb"), strings.HasSuffix(nameLower, ".mmdb"):
		return mmdb.Verify(path)
	case strings.HasSuffix(nameLower, ".dat"):
		fi, err := os.Stat(path)
		if err != nil {
			return false
		}
		return fi.Size() > 100*1024
	default:
		fi, err := os.Stat(path)
		return err == nil && fi.Size() > 0
	}
}
