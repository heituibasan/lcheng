package clashapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	secret  string
	client  *http.Client
}

func NewClient(address, secret string) *Client {
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	return &Client{
		baseURL: strings.TrimRight(address, "/"),
		secret:  secret,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) request(method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("clash api %s %s: %s", method, path, string(data))
	}
	return data, nil
}

func (c *Client) GetConfigs() (map[string]any, error) {
	data, err := c.request(http.MethodGet, "/configs", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) PatchConfigs(payload map[string]any) error {
	_, err := c.request(http.MethodPatch, "/configs", payload)
	return err
}

func (c *Client) GetProxies() (map[string]any, error) {
	data, err := c.request(http.MethodGet, "/proxies", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) SelectProxy(group, proxy string) error {
	path := fmt.Sprintf("/proxies/%s", group)
	_, err := c.request(http.MethodPut, path, map[string]string{"name": proxy})
	return err
}

func (c *Client) GetConnections() (map[string]any, error) {
	data, err := c.request(http.MethodGet, "/connections", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CloseAllConnections() error {
	_, err := c.request(http.MethodDelete, "/connections", nil)
	return err
}

func (c *Client) CloseConnection(id string) error {
	path := fmt.Sprintf("/connections/%s", id)
	_, err := c.request(http.MethodDelete, path, nil)
	return err
}

func (c *Client) GetRules() (map[string]any, error) {
	data, err := c.request(http.MethodGet, "/rules", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetVersion() (map[string]any, error) {
	data, err := c.request(http.MethodGet, "/version", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ReloadConfig() error {
	_, err := c.request(http.MethodPut, "/configs?force=true", map[string]string{"path": ""})
	return err
}

func (c *Client) PatchTun(enable bool) error {
	return c.PatchConfigs(map[string]any{
		"tun": map[string]any{"enable": enable},
	})
}

func (c *Client) PatchAllowLan(enable bool) error {
	return c.PatchConfigs(map[string]any{"allow-lan": enable})
}

func (c *Client) PatchLogLevel(level string) error {
	return c.PatchConfigs(map[string]any{"log-level": level})
}

func (c *Client) DelayTestGroup(group, testURL string, timeout int) (map[string]any, error) {
	if testURL == "" {
		testURL = "http://www.gstatic.com/generate_204"
	}
	if timeout <= 0 {
		timeout = 5000
	}
	path := fmt.Sprintf("/group/%s/delay?timeout=%d&url=%s", url.PathEscape(group), timeout, url.QueryEscape(testURL))
	data, err := c.request(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) DelayTest(proxy, testURL string, timeout int) (map[string]any, error) {
	if testURL == "" {
		testURL = "http://www.gstatic.com/generate_204"
	}
	if timeout <= 0 {
		timeout = 5000
	}
	path := fmt.Sprintf("/proxies/%s/delay?timeout=%d&url=%s", url.PathEscape(proxy), timeout, url.QueryEscape(testURL))
	data, err := c.request(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
