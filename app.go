package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/metacubex/mihomo/component/updater"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vpn_clash/internal/appmeta"
	"vpn_clash/internal/autostart"
	"vpn_clash/internal/clashapi"
	appconfig "vpn_clash/internal/config"
	"vpn_clash/internal/core"
	"vpn_clash/internal/monitor"
	"vpn_clash/internal/profile"
	"vpn_clash/internal/settings"
	"vpn_clash/internal/subscription"
	"vpn_clash/internal/sysproxy"
	"vpn_clash/internal/tray"
)

type trafficState struct {
	mu            sync.Mutex
	lastUpTotal   int64
	lastDownTotal int64
	lastAt        time.Time
	speedUp       int64
	speedDown     int64
}

type App struct {
	ctx          context.Context
	store        *appconfig.Store
	settings     *settings.Store
	core         *core.Manager
	clash        *clashapi.Client
	subs         *subscription.Manager
	profiles     *profile.Manager
	monitor      *monitor.Hub
	traffic      trafficState
}

type Status struct {
	Running            bool   `json:"running"`
	ConfigPath         string `json:"configPath"`
	Controller         string `json:"controller"`
	MixedPort          int    `json:"mixedPort"`
	SystemProxyEnabled bool   `json:"systemProxyEnabled"`
	SystemProxyServer  string `json:"systemProxyServer"`
	Connected          bool   `json:"connected"`
	Version            string `json:"version"`
}

type TrafficStats struct {
	Upload    int64 `json:"upload"`
	Download  int64 `json:"download"`
	UpTotal   int64 `json:"upTotal"`
	DownTotal int64 `json:"downTotal"`
}

type SubscriptionDTO struct {
	subscription.Item
}

func NewApp() *App {
	return &App{monitor: monitor.NewHub(500)}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	store, err := appconfig.NewStore()
	if err != nil {
		panic(err)
	}
	a.store = store

	settingStore, err := settings.NewStore()
	if err != nil {
		panic(err)
	}
	a.settings = settingStore

	a.core = core.NewManager(store)
	a.refreshAPIClient()

	subs, err := subscription.NewManager()
	if err != nil {
		panic(err)
	}
	a.subs = subs

	profiles, err := profile.NewManager()
	if err != nil {
		panic(err)
	}
	a.profiles = profiles

	if settingStore.Get().AutoStartCore {
		_ = a.Connect()
	}

	tray.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.Disconnect()
	a.monitor.Stop()
}

func (a *App) refreshAPIClient() {
	address, secret := a.store.Controller()
	a.clash = clashapi.NewClient(address, secret)
}

func (a *App) applySettingsToConfig() error {
	s := a.settings.Get()
	content := a.store.Get()

	patched, err := appconfig.PatchYAML(content, map[string]any{
		"mixed-port": s.MixedPort,
		"allow-lan":  s.AllowLan,
		"log-level":  s.LogLevel,
	})
	if err != nil {
		return err
	}

	patched, err = appconfig.PatchNested(patched, "tun", map[string]any{
		"enable": s.TunEnabled,
	})
	if err != nil {
		return err
	}

	return a.store.Save(patched)
}

func (a *App) GetStatus() Status {
	address, _ := a.store.Controller()
	enabled, server, _ := sysproxy.IsEnabled()
	version := ""
	if a.core.IsRunning() {
		if v, err := a.clash.GetVersion(); err == nil {
			if meta, ok := v["meta"].(bool); ok && meta {
				version = "Mihomo Meta"
			}
			if ver, ok := v["version"].(string); ok {
				version = ver
			}
		}
	}

	return Status{
		Running:            a.core.IsRunning(),
		ConfigPath:         a.store.Path(),
		Controller:         address,
		MixedPort:          a.store.MixedPort(),
		SystemProxyEnabled: enabled,
		SystemProxyServer:  server,
		Connected:          a.core.IsRunning() && enabled,
		Version:            version,
	}
}

type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	WebsiteURL  string `json:"websiteUrl"`
	HelpDocsURL string `json:"helpDocsUrl"`
}

type UpdateCheckResult struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Message        string `json:"message"`
}

func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:        appmeta.Name,
		Version:     appmeta.Version,
		WebsiteURL:  appmeta.WebsiteURL,
		HelpDocsURL: appmeta.HelpDocsURL,
	}
}

func (a *App) CheckForUpdates() UpdateCheckResult {
	return UpdateCheckResult{
		HasUpdate:      false,
		CurrentVersion: appmeta.Version,
		LatestVersion:  appmeta.Version,
		Message:        "当前已是最新版本",
	}
}

func (a *App) OpenWebsite() {
	runtime.BrowserOpenURL(a.ctx, appmeta.WebsiteURL)
}

func (a *App) OpenHelpDocs() {
	runtime.BrowserOpenURL(a.ctx, appmeta.HelpDocsURL)
}

func (a *App) GetSettings() settings.Settings {
	return a.settings.Get()
}

func (a *App) SaveSettings(data settings.Settings) error {
	if err := a.settings.Save(data); err != nil {
		return err
	}
	if err := a.applySettingsToConfig(); err != nil {
		return err
	}
	if a.core.IsRunning() {
		return a.core.Reload(a.store.Get())
	}
	return nil
}

func (a *App) Connect() error {
	if err := a.applySettingsToConfig(); err != nil {
		return err
	}
	if !a.core.IsRunning() {
		updater.RegisterGeoUpdater()
		if err := a.core.Start(); err != nil {
			return err
		}
		a.refreshAPIClient()
		address, secret := a.store.Controller()
		a.monitor.Start(address, secret)
		a.applySavedProxySelections()
	}

	s := a.settings.Get()
	if s.AutoSystemProxy || s.SystemProxyEnabled {
		port := s.MixedPort
		if port <= 0 {
			port = a.store.MixedPort()
		}
		if err := sysproxy.SetEnabled("127.0.0.1", port); err != nil {
			return fmt.Errorf("set system proxy: %w", err)
		}
		s.SystemProxyEnabled = true
		_ = a.settings.Save(s)
	}

	runtime.EventsEmit(a.ctx, "vpn:connected", a.GetStatus())
	return nil
}

func (a *App) Disconnect() error {
	s := a.settings.Get()
	if s.SystemProxyEnabled {
		_ = sysproxy.SetDisabled()
		s.SystemProxyEnabled = false
		_ = a.settings.Save(s)
	}

	a.monitor.Stop()

	if a.core.IsRunning() {
		if err := a.core.Stop(); err != nil {
			return err
		}
	}

	runtime.EventsEmit(a.ctx, "vpn:disconnected", a.GetStatus())
	return nil
}

func (a *App) SetSystemProxy(enabled bool) error {
	s := a.settings.Get()
	port := s.MixedPort
	if port <= 0 {
		port = a.store.MixedPort()
	}

	if enabled {
		if !a.core.IsRunning() {
			if err := a.Connect(); err != nil {
				return err
			}
			return nil
		}
		if err := sysproxy.SetEnabled("127.0.0.1", port); err != nil {
			return err
		}
	} else {
		if err := sysproxy.SetDisabled(); err != nil {
			return err
		}
	}

	s.SystemProxyEnabled = enabled
	return a.settings.Save(s)
}

func (a *App) GetConfig() string {
	return a.store.Get()
}

func (a *App) SaveConfig(content string) error {
	if err := a.store.Save(content); err != nil {
		return err
	}
	a.refreshAPIClient()
	if a.core.IsRunning() {
		return a.core.Reload(content)
	}
	return nil
}

func (a *App) TestConfig(content string) error {
	return a.core.TestConfig(content)
}

func (a *App) StartCore() error {
	return a.Connect()
}

func (a *App) StopCore() error {
	return a.Disconnect()
}

func (a *App) GetConfigs() (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.GetConfigs()
}

func (a *App) PatchMode(mode string) error {
	if !a.core.IsRunning() {
		return fmt.Errorf("core is not running")
	}
	return a.clash.PatchConfigs(map[string]any{"mode": mode})
}

func (a *App) SetTunEnabled(enabled bool) error {
	s := a.settings.Get()
	s.TunEnabled = enabled
	if err := a.settings.Save(s); err != nil {
		return err
	}
	if err := a.applySettingsToConfig(); err != nil {
		return err
	}
	if a.core.IsRunning() {
		if err := a.core.Reload(a.store.Get()); err != nil {
			return err
		}
		return a.clash.PatchTun(enabled)
	}
	return nil
}

func (a *App) SetAllowLan(enabled bool) error {
	s := a.settings.Get()
	s.AllowLan = enabled
	if err := a.settings.Save(s); err != nil {
		return err
	}
	if err := a.applySettingsToConfig(); err != nil {
		return err
	}
	if a.core.IsRunning() {
		return a.clash.PatchAllowLan(enabled)
	}
	return nil
}

func (a *App) SetLaunchAtLogin(enabled bool) error {
	if err := autostart.SetEnabled(enabled); err != nil {
		return err
	}
	s := a.settings.Get()
	s.LaunchAtLogin = enabled
	return a.settings.Save(s)
}

func (a *App) GetProxies() (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.GetProxies()
}

func (a *App) SelectProxy(group, proxy string) error {
	configID := a.settings.Get().ActiveProfileID
	content, err := a.GetConfigContent(configID)
	if err != nil {
		return err
	}
	if err := appconfig.ValidateProxySelection(content, group, proxy); err != nil {
		if a.core.IsRunning() {
			if liveErr := a.validateLiveProxySelection(group, proxy); liveErr != nil {
				return err
			}
		} else {
			return err
		}
	}

	s := a.settings.Get()
	if s.ProxySelections == nil {
		s.ProxySelections = make(map[string]map[string]string)
	}
	if s.ProxySelections[configID] == nil {
		s.ProxySelections[configID] = make(map[string]string)
	}
	s.ProxySelections[configID][group] = proxy
	if err := a.settings.Save(s); err != nil {
		return err
	}

	if a.core.IsRunning() {
		if err := a.clash.SelectProxy(group, proxy); err != nil {
			return err
		}
	}

	runtime.EventsEmit(a.ctx, "proxy:selected", map[string]string{
		"group": group,
		"proxy": proxy,
	})
	return nil
}

func (a *App) validateLiveProxySelection(group, proxy string) error {
	data, err := a.clash.GetProxies()
	if err != nil {
		return err
	}
	root, ok := data["proxies"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid proxies response")
	}
	item, ok := root[group].(map[string]any)
	if !ok {
		return fmt.Errorf("group %q not found", group)
	}
	all, _ := item["all"].([]any)
	for _, entry := range all {
		if name, ok := entry.(string); ok && name == proxy {
			return nil
		}
	}
	return fmt.Errorf("proxy %q not in group %q", proxy, group)
}

func (a *App) GetConnections() (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.GetConnections()
}

func (a *App) CloseAllConnections() error {
	if !a.core.IsRunning() {
		return fmt.Errorf("core is not running")
	}
	return a.clash.CloseAllConnections()
}

func (a *App) CloseConnection(id string) error {
	if !a.core.IsRunning() {
		return fmt.Errorf("core is not running")
	}
	return a.clash.CloseConnection(id)
}

func (a *App) DelayTest(proxy string) (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.DelayTest(proxy, "", 5000)
}

func (a *App) DelayTestGroup(group string) (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.DelayTestGroup(group, "", 5000)
}

func (a *App) GetRules() (map[string]any, error) {
	if !a.core.IsRunning() {
		return nil, fmt.Errorf("core is not running")
	}
	return a.clash.GetRules()
}

func (a *App) GetTraffic() TrafficStats {
	ws := a.monitor.GetTraffic()
	if ws.Upload > 0 || ws.Download > 0 {
		return TrafficStats{
			Upload:   ws.Upload,
			Download: ws.Download,
		}
	}

	if !a.core.IsRunning() {
		return TrafficStats{}
	}

	data, err := a.clash.GetConnections()
	if err != nil {
		return TrafficStats{}
	}

	upTotal := toInt64(data["uploadTotal"])
	downTotal := toInt64(data["downloadTotal"])

	a.traffic.mu.Lock()
	defer a.traffic.mu.Unlock()

	now := time.Now()
	if !a.traffic.lastAt.IsZero() {
		sec := now.Sub(a.traffic.lastAt).Seconds()
		if sec > 0 {
			a.traffic.speedUp = int64(float64(upTotal-a.traffic.lastUpTotal) / sec)
			a.traffic.speedDown = int64(float64(downTotal-a.traffic.lastDownTotal) / sec)
			if a.traffic.speedUp < 0 {
				a.traffic.speedUp = 0
			}
			if a.traffic.speedDown < 0 {
				a.traffic.speedDown = 0
			}
		}
	}
	a.traffic.lastUpTotal = upTotal
	a.traffic.lastDownTotal = downTotal
	a.traffic.lastAt = now

	return TrafficStats{
		Upload:    a.traffic.speedUp,
		Download:  a.traffic.speedDown,
		UpTotal:   upTotal,
		DownTotal: downTotal,
	}
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func (a *App) GetLogs() []string {
	return a.monitor.GetLogs()
}

func (a *App) ImportConfigFromFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入 Clash 配置",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML", Pattern: "*.yaml;*.yml"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	data, err := readFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	if err := appconfig.Validate(content); err != nil {
		return "", fmt.Errorf("invalid yaml: %w", err)
	}
	if err := a.SaveConfig(content); err != nil {
		return "", err
	}
	return content, nil
}

func (a *App) ExportConfigToFile() error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出 Clash 配置",
		DefaultFilename: "config.yaml",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML", Pattern: "*.yaml"},
		},
	})
	if err != nil || path == "" {
		return err
	}
	return writeFile(path, []byte(a.store.Get()))
}

func (a *App) ListSubscriptions() []subscription.Item {
	return a.subs.List()
}

func (a *App) AddSubscription(name, url string) (subscription.Item, error) {
	return a.subs.Add(name, url)
}

func (a *App) RemoveSubscription(id string) error {
	return a.subs.Remove(id)
}

func (a *App) UpdateSubscription(id string) (string, error) {
	items := a.subs.List()
	var target *subscription.Item
	for i := range items {
		if items[i].ID == id {
			target = &items[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("subscription not found")
	}

	s := a.settings.Get()
	content, meta, err := subscription.Fetch(target.URL, s.SubscriptionUserAgent)
	if err != nil {
		return "", err
	}

	merged, err := subscription.MergeIntoConfig(a.store.Get(), content)
	if err != nil {
		return "", err
	}
	if err := a.SaveConfig(merged); err != nil {
		return "", err
	}

	updateMeta := subscription.Item{UpdatedAt: time.Now().Unix()}
	if info := meta["Subscription-Userinfo"]; info != "" {
		_, total, expire := subscription.ParseUserInfo(info)
		updateMeta.TrafficTotal = total
		updateMeta.ExpireAt = expire
		_ = updateMeta.TrafficUsed
	}
	_ = a.subs.UpdateMeta(id, updateMeta)

	runtime.EventsEmit(a.ctx, "subscription:updated", id)
	return merged, nil
}

func (a *App) UpdateAllSubscriptions() error {
	for _, item := range a.subs.List() {
		if _, err := a.UpdateSubscription(item.ID); err != nil {
			return fmt.Errorf("%s: %w", item.Name, err)
		}
	}
	return nil
}

func (a *App) ListProfiles() []profile.Profile {
	return a.profiles.List()
}

func (a *App) CreateProfile(name string) (profile.Profile, error) {
	content := a.store.Get()
	return a.profiles.Create(name, content)
}

func (a *App) DeleteProfile(id string) error {
	return a.profiles.Delete(id)
}

func (a *App) ActivateProfile(id string) error {
	content, err := a.profiles.Read(id)
	if err != nil {
		return err
	}
	if content == "" {
		content = appconfig.DefaultYAML
	}
	if err := a.SaveConfig(content); err != nil {
		return err
	}
	s := a.settings.Get()
	s.ActiveProfileID = id
	return a.settings.Save(s)
}

func (a *App) SaveProfileContent(id, content string) error {
	if err := appconfig.Validate(content); err != nil {
		return err
	}
	if err := a.profiles.Write(id, content); err != nil {
		return err
	}
	s := a.settings.Get()
	if s.ActiveProfileID == id {
		return a.SaveConfig(content)
	}
	return nil
}

func (a *App) RenameProfile(id, name string) error {
	return a.profiles.Rename(id, name)
}

type ConfigEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SourceURL  string `json:"sourceUrl"`
	UpdatedAt  int64  `json:"updatedAt"`
	IsActive   bool   `json:"isActive"`
	ProxyCount int    `json:"proxyCount"`
}

type RulesPageResult struct {
	Total  int                 `json:"total"`
	Rules  []appconfig.RuleEntry `json:"rules"`
	Offset int                 `json:"offset"`
	Limit  int                 `json:"limit"`
}

func (a *App) ListConfigs() []ConfigEntry {
	activeID := a.settings.Get().ActiveProfileID
	items := a.profiles.List()
	out := make([]ConfigEntry, 0, len(items))
	for _, item := range items {
		entry := ConfigEntry{
			ID:        item.ID,
			Name:      item.Name,
			SourceURL: item.SourceURL,
			UpdatedAt: item.UpdatedAt,
			IsActive:  item.ID == activeID,
		}
		if content, err := a.profiles.Read(item.ID); err == nil {
			if preview, err := appconfig.ParseProxyPreview(content); err == nil {
				entry.ProxyCount = preview.ProxyCount
			}
		}
		out = append(out, entry)
	}
	return out
}

func (a *App) DownloadConfigFromURL(url string) (ConfigEntry, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return ConfigEntry{}, fmt.Errorf("url is empty")
	}

	s := a.settings.Get()
	content, _, err := subscription.Fetch(url, s.SubscriptionUserAgent)
	if err != nil {
		return ConfigEntry{}, err
	}

	name := configNameFromURL(url)
	prof, err := a.profiles.CreateWithSource(name, url, content)
	if err != nil {
		return ConfigEntry{}, err
	}
	if err := a.activateProfileContent(prof.ID, content); err != nil {
		return ConfigEntry{}, err
	}
	runtime.EventsEmit(a.ctx, "config:selected", prof.ID)

	return ConfigEntry{
		ID:         prof.ID,
		Name:       prof.Name,
		SourceURL:  prof.SourceURL,
		UpdatedAt:  prof.UpdatedAt,
		IsActive:   true,
		ProxyCount: countProxies(content),
	}, nil
}

func (a *App) ImportConfigAsNew() (ConfigEntry, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入 YAML 配置",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML", Pattern: "*.yaml;*.yml"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ConfigEntry{}, err
	}
	if path == "" {
		return ConfigEntry{}, nil
	}

	data, err := readFile(path)
	if err != nil {
		return ConfigEntry{}, err
	}
	content := string(data)
	if err := appconfig.Validate(content); err != nil {
		return ConfigEntry{}, fmt.Errorf("invalid yaml: %w", err)
	}

	name := configNameFromPath(path)
	prof, err := a.profiles.CreateWithSource(name, "", content)
	if err != nil {
		return ConfigEntry{}, err
	}
	if err := a.activateProfileContent(prof.ID, content); err != nil {
		return ConfigEntry{}, err
	}
	runtime.EventsEmit(a.ctx, "config:selected", prof.ID)

	return ConfigEntry{
		ID:         prof.ID,
		Name:       prof.Name,
		UpdatedAt:  prof.UpdatedAt,
		IsActive:   true,
		ProxyCount: countProxies(content),
	}, nil
}

func countProxies(content string) int {
	preview, err := appconfig.ParseProxyPreview(content)
	if err != nil {
		return 0
	}
	return preview.ProxyCount
}

func (a *App) RefreshConfigFromURL(id string) error {
	content, err := a.profiles.Read(id)
	if err != nil {
		return err
	}
	_ = content

	items := a.profiles.List()
	var sourceURL string
	for _, item := range items {
		if item.ID == id {
			sourceURL = item.SourceURL
			break
		}
	}
	if sourceURL == "" {
		return fmt.Errorf("this config has no download url")
	}

	s := a.settings.Get()
	fetched, _, err := subscription.Fetch(sourceURL, s.SubscriptionUserAgent)
	if err != nil {
		return err
	}
	if err := a.profiles.Write(id, fetched); err != nil {
		return err
	}
	if a.settings.Get().ActiveProfileID == id {
		if err := a.SaveConfig(fetched); err != nil {
			return err
		}
		runtime.EventsEmit(a.ctx, "config:selected", id)
	}
	return nil
}

func (a *App) ActivateConfig(id string) error {
	content, err := a.profiles.Read(id)
	if err != nil {
		return err
	}
	return a.activateProfileContent(id, content)
}

func (a *App) SelectConfig(id string) error {
	if err := a.ActivateConfig(id); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, "config:selected", id)
	return nil
}

func (a *App) applySavedProxySelections() {
	if !a.core.IsRunning() {
		return
	}
	s := a.settings.Get()
	selections := s.ProxySelections[s.ActiveProfileID]
	for group, proxy := range selections {
		_ = a.clash.SelectProxy(group, proxy)
	}
}

func (a *App) GetConfigProxies(configID string) (appconfig.ProxyPreviewResult, error) {
	if configID == "" {
		configID = a.settings.Get().ActiveProfileID
	}
	content, err := a.GetConfigContent(configID)
	if err != nil {
		return appconfig.ProxyPreviewResult{}, err
	}
	preview, err := appconfig.ParseProxyPreview(content)
	if err != nil {
		return appconfig.ProxyPreviewResult{}, err
	}
	s := a.settings.Get()
	if s.ProxySelections != nil {
		preview = appconfig.ApplyProxySelections(preview, s.ProxySelections[configID])
	}
	return preview, nil
}

func (a *App) RefreshAllConfigs() error {
	for _, item := range a.ListConfigs() {
		if item.SourceURL == "" {
			continue
		}
		if err := a.RefreshConfigFromURL(item.ID); err != nil {
			return fmt.Errorf("%s: %w", item.Name, err)
		}
	}
	runtime.EventsEmit(a.ctx, "config:updated", "")
	return nil
}

func (a *App) GetActiveConfigID() string {
	return a.settings.Get().ActiveProfileID
}

func (a *App) DeleteConfig(id string) error {
	activeID := a.settings.Get().ActiveProfileID
	if err := a.profiles.Delete(id); err != nil {
		return err
	}
	if activeID == id {
		items := a.profiles.List()
		if len(items) > 0 {
			return a.ActivateConfig(items[0].ID)
		}
	}
	return nil
}

func (a *App) GetConfigContent(id string) (string, error) {
	content, err := a.profiles.Read(id)
	if err != nil {
		return "", err
	}
	if content == "" {
		return appconfig.DefaultYAML, nil
	}
	return content, nil
}

func (a *App) SaveConfigContent(id, content string) error {
	if err := appconfig.Validate(content); err != nil {
		return err
	}
	if err := a.profiles.Write(id, content); err != nil {
		return err
	}
	if a.settings.Get().ActiveProfileID == id {
		if err := a.SaveConfig(content); err != nil {
			return err
		}
		runtime.EventsEmit(a.ctx, "config:selected", id)
	}
	return nil
}

func (a *App) RenameConfig(id, name string) error {
	return a.profiles.Rename(id, name)
}

func (a *App) GetConfigRulesPage(offset, limit int, keyword string) (RulesPageResult, error) {
	content := a.store.Get()
	rules, err := appconfig.ParseRulesFromYAML(content)
	if err != nil {
		return RulesPageResult{}, err
	}
	filtered := appconfig.FilterRules(rules, keyword)
	page := appconfig.PageRules(filtered, offset, limit)
	return RulesPageResult{
		Total:  page.Total,
		Rules:  page.Rules,
		Offset: offset,
		Limit:  limit,
	}, nil
}

func (a *App) activateProfileContent(id, content string) error {
	if content == "" {
		content = appconfig.DefaultYAML
	}
	if err := a.SaveConfig(content); err != nil {
		return err
	}
	s := a.settings.Get()
	s.ActiveProfileID = id
	return a.settings.Save(s)
}

func configNameFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if idx := strings.Index(raw, "/"); idx > 0 {
		raw = raw[:idx]
	}
	if raw == "" {
		return fmt.Sprintf("config_%d", time.Now().Unix())
	}
	return raw
}

func configNameFromPath(path string) string {
	base := path
	if idx := strings.LastIndexAny(path, `/\`); idx >= 0 {
		base = path[idx+1:]
	}
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	if base == "" {
		return fmt.Sprintf("import_%d", time.Now().Unix())
	}
	return base
}
