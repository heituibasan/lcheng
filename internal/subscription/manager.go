package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"vpn_clash/internal/paths"
)

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	UpdatedAt   int64  `json:"updatedAt"`
	TrafficUsed string `json:"trafficUsed,omitempty"`
	TrafficTotal string `json:"trafficTotal,omitempty"`
	ExpireAt    string `json:"expireAt,omitempty"`
}

type Manager struct {
	mu    sync.RWMutex
	path  string
	items []Item
}

func NewManager() (*Manager, error) {
	dir, err := paths.AppDataDir()
	if err != nil {
		return nil, err
	}
	m := &Manager{path: dir + string(os.PathSeparator) + "subscriptions.json"}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	m.items = []Item{}
	raw, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return m.save()
		}
		return err
	}
	return json.Unmarshal(raw, &m.items)
}

func (m *Manager) save() error {
	raw, err := json.MarshalIndent(m.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, raw, 0o644)
}

func (m *Manager) List() []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Item, len(m.items))
	copy(out, m.items)
	return out
}

func (m *Manager) Add(name, url string) (Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item := Item{
		ID:   fmt.Sprintf("sub_%d", time.Now().UnixNano()),
		Name: name,
		URL:  url,
	}
	m.items = append(m.items, item)
	return item, m.save()
}

func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	next := make([]Item, 0, len(m.items))
	for _, item := range m.items {
		if item.ID != id {
			next = append(next, item)
		}
	}
	m.items = next
	return m.save()
}

func (m *Manager) UpdateMeta(id string, meta Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.items {
		if item.ID == id {
			m.items[i].UpdatedAt = meta.UpdatedAt
			m.items[i].TrafficUsed = meta.TrafficUsed
			m.items[i].TrafficTotal = meta.TrafficTotal
			m.items[i].ExpireAt = meta.ExpireAt
			return m.save()
		}
	}
	return fmt.Errorf("subscription not found")
}

func Fetch(url, userAgent string) (string, map[string]string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	if resp.StatusCode >= 400 {
		return "", nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	meta := parseSubscriptionHeaders(resp.Header)
	content := decodeSubscriptionBody(body)
	if err := validateYAML(content); err != nil {
		return "", nil, fmt.Errorf("invalid subscription content: %w", err)
	}
	return content, meta, nil
}

func parseSubscriptionHeaders(h http.Header) map[string]string {
	meta := map[string]string{}
	keys := []string{
		"Subscription-Userinfo",
		"Content-Disposition",
		"Profile-Update-Interval",
	}
	for _, key := range keys {
		if v := h.Get(key); v != "" {
			meta[key] = v
		}
	}
	return meta
}

func decodeSubscriptionBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if decoded, err := base64.StdEncoding.DecodeString(text); err == nil {
		if validateYAML(string(decoded)) == nil {
			return string(decoded)
		}
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(text); err == nil {
		if validateYAML(string(decoded)) == nil {
			return string(decoded)
		}
	}
	return text
}

func validateYAML(content string) error {
	var raw map[string]any
	return yaml.Unmarshal([]byte(content), &raw)
}

func MergeIntoConfig(baseYAML, subYAML string) (string, error) {
	var base, sub map[string]any
	if err := yaml.Unmarshal([]byte(baseYAML), &base); err != nil {
		return "", err
	}
	if err := yaml.Unmarshal([]byte(subYAML), &sub); err != nil {
		return "", err
	}

	for _, key := range []string{"proxies", "proxy-groups", "rules", "proxy-providers", "rule-providers"} {
		if v, ok := sub[key]; ok {
			base[key] = v
		}
	}

	out, err := yaml.Marshal(base)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ParseUserInfo(header string) (used, total, expire string) {
	// upload=0; download=0; total=107374182400; expire=1735689600
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "upload", "download":
			used = kv[1]
		case "total":
			total = kv[1]
		case "expire":
			expire = kv[1]
		}
	}
	return used, total, expire
}
