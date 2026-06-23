package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TrafficStats struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

type Hub struct {
	mu      sync.RWMutex
	address string
	secret  string
	traffic TrafficStats
	logs    []string
	maxLogs int
	cancel  context.CancelFunc
}

func NewHub(maxLogs int) *Hub {
	if maxLogs <= 0 {
		maxLogs = 500
	}
	return &Hub{maxLogs: maxLogs}
}

func (h *Hub) Start(address, secret string) {
	h.Stop()
	h.mu.Lock()
	h.address = address
	h.secret = secret
	h.traffic = TrafficStats{}
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.mu.Unlock()

	go h.runTrafficPoll(ctx, address, secret)
	go h.runLogsPoll(ctx, address, secret)
}

func (h *Hub) Stop() {
	h.mu.Lock()
	if h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}
	h.mu.Unlock()
}

func (h *Hub) GetTraffic() TrafficStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.traffic
}

func (h *Hub) GetLogs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, len(h.logs))
	copy(out, h.logs)
	return out
}

func (h *Hub) appendLog(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = append(h.logs, line)
	if len(h.logs) > h.maxLogs {
		h.logs = h.logs[len(h.logs)-h.maxLogs:]
	}
}

func (h *Hub) baseURL(address string) string {
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	return strings.TrimRight(address, "/")
}

func (h *Hub) runTrafficPoll(ctx context.Context, address, secret string) {
	client := &http.Client{Timeout: 3 * time.Second}
	url := h.baseURL(address) + "/traffic"

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if secret != "" {
			req.Header.Set("Authorization", "Bearer "+secret)
		}

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		reader := bufio.NewReader(resp.Body)
		line, err := reader.ReadBytes('\n')
		resp.Body.Close()
		if err != nil && err != io.EOF {
			time.Sleep(2 * time.Second)
			continue
		}

		var payload struct {
			Up   int64 `json:"up"`
			Down int64 `json:"down"`
		}
		if json.Unmarshal(bytesTrim(line), &payload) == nil {
			h.mu.Lock()
			h.traffic = TrafficStats{Upload: payload.Up, Download: payload.Down}
			h.mu.Unlock()
		}

		time.Sleep(time.Second)
	}
}

func (h *Hub) runLogsPoll(ctx context.Context, address, secret string) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := h.baseURL(address) + "/logs"

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if secret != "" {
			req.Header.Set("Authorization", "Bearer "+secret)
		}

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			h.appendLog(scanner.Text())
			select {
			case <-ctx.Done():
				resp.Body.Close()
				return
			default:
			}
		}
		resp.Body.Close()
		time.Sleep(2 * time.Second)
	}
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
