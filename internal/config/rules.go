package config

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type RuleEntry struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
}

type RulesPage struct {
	Total int         `json:"total"`
	Rules []RuleEntry `json:"rules"`
}

func ParseRulesFromYAML(content string) ([]RuleEntry, error) {
	var raw struct {
		Rules []any `yaml:"rules"`
	}
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return nil, err
	}

	out := make([]RuleEntry, 0, len(raw.Rules))
	for _, item := range raw.Rules {
		entry := parseRuleItem(item)
		if entry != nil {
			out = append(out, *entry)
		}
	}
	return out, nil
}

func parseRuleItem(item any) *RuleEntry {
	switch v := item.(type) {
	case string:
		parts := splitRule(v)
		if len(parts) == 0 {
			return nil
		}
		if len(parts) == 1 {
			return &RuleEntry{Type: "MATCH", Payload: "", Proxy: parts[0]}
		}
		if len(parts) == 2 && strings.EqualFold(parts[0], "MATCH") {
			return &RuleEntry{Type: "MATCH", Payload: "", Proxy: parts[1]}
		}
		return &RuleEntry{
			Type:    parts[0],
			Payload: strings.Join(parts[1:len(parts)-1], ","),
			Proxy:   parts[len(parts)-1],
		}
	default:
		return nil
	}
}

func splitRule(rule string) []string {
	var parts []string
	var current strings.Builder
	escaped := false
	for _, ch := range rule {
		if ch == '\\' && !escaped {
			escaped = true
			continue
		}
		if ch == ',' && !escaped {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(ch)
		escaped = false
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

func FilterRules(rules []RuleEntry, keyword string) []RuleEntry {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return rules
	}
	filtered := make([]RuleEntry, 0, len(rules))
	for _, rule := range rules {
		if strings.Contains(strings.ToLower(rule.Type), keyword) ||
			strings.Contains(strings.ToLower(rule.Payload), keyword) ||
			strings.Contains(strings.ToLower(rule.Proxy), keyword) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func PageRules(rules []RuleEntry, offset, limit int) RulesPage {
	total := len(rules)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}
	if offset >= total {
		return RulesPage{Total: total, Rules: []RuleEntry{}}
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return RulesPage{
		Total: total,
		Rules: rules[offset:end],
	}
}
