//go:build !windows

package sysproxy

func SetEnabled(host string, port int) error {
	return nil
}

func SetDisabled() error {
	return nil
}

func IsEnabled() (bool, string, error) {
	return false, "", nil
}

func Refresh() {}
