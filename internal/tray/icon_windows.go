//go:build windows

package tray

import _ "embed"

//go:embed assets/tray.ico
var iconData []byte
