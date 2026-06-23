//go:build windows

package tray

import "encoding/base64"

func loadIconData() []byte {
	const encoded = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAASElEQVR42u3XsQ0AIAwDsXn/n3EYgQqJkJLu3Q0k2Q0AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADgXw0BAAAB0QABGk8k6QAAAABJRU5ErkJggg=="
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}
	return data
}

var iconData = loadIconData()
