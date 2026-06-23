//go:build windows

package sysproxy

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	internetOptionSettingsChanged = 39
	internetOptionRefresh         = 37
)

var (
	wininet           = syscall.NewLazyDLL("wininet.dll")
	internetSetOption = wininet.NewProc("InternetSetOptionW")
)

func notifyProxyChange() {
	internetSetOption.Call(0, internetOptionSettingsChanged, 0, 0)
	internetSetOption.Call(0, internetOptionRefresh, 0, 0)
}

func SetEnabled(host string, port int) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	server := fmt.Sprintf("%s:%d", host, port)
	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}
	if err := key.SetStringValue("ProxyServer", server); err != nil {
		return err
	}
	notifyProxyChange()
	return nil
}

func SetDisabled() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetDWordValue("ProxyEnable", 0); err != nil {
		return err
	}
	notifyProxyChange()
	return nil
}

func IsEnabled() (bool, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return false, "", err
	}
	defer key.Close()

	enabled, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil {
		return false, "", err
	}
	server, _, err := key.GetStringValue("ProxyServer")
	if err != nil {
		server = ""
	}
	return enabled == 1, server, nil
}
