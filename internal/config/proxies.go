package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProxyGroupPreview struct {
	Name string   `json:"name"`
	Type string   `json:"type"`
	Now  string   `json:"now"`
	All  []string `json:"all"`
}

type ProxyPreviewResult struct {
	Groups     []ProxyGroupPreview `json:"groups"`
	ProxyCount int                 `json:"proxyCount"`
	Mode       string              `json:"mode"`
}

func ParseProxyPreview(content string) (ProxyPreviewResult, error) {
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return ProxyPreviewResult{}, err
	}

	result := ProxyPreviewResult{Mode: "rule"}
	if mode, ok := raw["mode"].(string); ok {
		result.Mode = mode
	}

	if proxies, ok := raw["proxies"].([]any); ok {
		result.ProxyCount = len(proxies)
	}

	groupsRaw, _ := raw["proxy-groups"].([]any)
	for _, item := range groupsRaw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		groupType, _ := m["type"].(string)
		if name == "" {
			continue
		}
		all := toStringList(m["proxies"])
		now := ""
		if len(all) > 0 {
			now = all[0]
		}
		if use, ok := m["use"].([]any); ok && len(use) > 0 {
			// provider-based group, keep name visible
			_ = use
		}
		result.Groups = append(result.Groups, ProxyGroupPreview{
			Name: name,
			Type: strings.ToLower(groupType),
			Now:  now,
			All:  all,
		})
	}

	if len(result.Groups) > 0 && !hasGroupNamed(result.Groups, "GLOBAL") {
		result.Groups = append([]ProxyGroupPreview{buildGlobalGroup(result.Groups)}, result.Groups...)
	}
	return result, nil
}

func hasGroupNamed(groups []ProxyGroupPreview, name string) bool {
	for _, g := range groups {
		if strings.EqualFold(g.Name, name) {
			return true
		}
	}
	return false
}

func buildGlobalGroup(groups []ProxyGroupPreview) ProxyGroupPreview {
	all := make([]string, 0, len(groups)+2)
	for _, g := range groups {
		all = append(all, g.Name)
	}
	all = append(all, "DIRECT", "REJECT")
	now := ""
	if len(all) > 0 {
		now = all[0]
	}
	return ProxyGroupPreview{
		Name: "GLOBAL",
		Type: "select",
		Now:  now,
		All:  all,
	}
}

func toStringList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func ValidateProxySelection(content, group, proxy string) error {
	preview, err := ParseProxyPreview(content)
	if err != nil {
		return err
	}
	for _, g := range preview.Groups {
		if !strings.EqualFold(g.Name, group) {
			continue
		}
		for _, p := range g.All {
			if p == proxy {
				return nil
			}
		}
		return fmt.Errorf("proxy %q not in group %q", proxy, group)
	}
	return fmt.Errorf("group %q not found", group)
}

func ApplyProxySelections(preview ProxyPreviewResult, selections map[string]string) ProxyPreviewResult {
	if len(selections) == 0 {
		return preview
	}
	for i, g := range preview.Groups {
		selected, ok := selections[g.Name]
		if !ok {
			for name, value := range selections {
				if strings.EqualFold(name, g.Name) {
					selected = value
					ok = true
					break
				}
			}
		}
		if !ok {
			continue
		}
		for _, p := range g.All {
			if p == selected {
				preview.Groups[i].Now = selected
				break
			}
		}
	}
	return preview
}
