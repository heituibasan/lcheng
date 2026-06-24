package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseURL = "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest"

var files = []string{
	"geoip.metadb",
	"geosite.dat",
	"geoip.dat",
}

func main() {
	root, err := filepath.Abs(filepath.Join("internal", "geobundle", "data"))
	if err != nil {
		exitErr(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		exitErr(err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	for _, name := range files {
		if err := download(client, baseURL+"/"+name, filepath.Join(root, name)); err != nil {
			exitErr(fmt.Errorf("%s: %w", name, err))
		}
	}

	fmt.Println("Geo databases updated in", root)
}

func download(client *http.Client, url, dest string) error {
	fmt.Println("Downloading", url)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	written, err := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	if written < 1024 {
		_ = os.Remove(tmp)
		return fmt.Errorf("downloaded file too small (%d bytes)", written)
	}

	_ = os.Remove(dest)
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	fmt.Printf("  -> %s (%d bytes)\n", dest, written)
	return nil
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
