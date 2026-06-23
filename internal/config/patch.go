package config

import (
	"gopkg.in/yaml.v3"
)

func PatchYAML(content string, patches map[string]any) (string, error) {
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return "", err
	}
	for k, v := range patches {
		raw[k] = v
	}
	out, err := yaml.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func PatchNested(content, key string, patch map[string]any) (string, error) {
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return "", err
	}
	current, _ := raw[key].(map[string]any)
	if current == nil {
		current = map[string]any{}
	}
	for k, v := range patch {
		current[k] = v
	}
	raw[key] = current
	out, err := yaml.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
