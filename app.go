package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	store  *ConfigStore
	bridge *BridgeRuntime

	mu     sync.RWMutex
	config AppConfig

	logsMu sync.RWMutex
	logs   []LogEntry
}

func NewApp() *App {
	store, err := NewConfigStore()
	if err != nil {
		panic(err)
	}

	app := &App{
		store: store,
	}
	app.bridge = NewBridgeRuntime(app)
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := a.store.Load()
	if err != nil {
		a.appendLog("error", "app", "读取本地配置失败: "+err.Error(), "")
		cfg = defaultConfig()
	}

	a.mu.Lock()
	a.config = cfg
	a.mu.Unlock()

	a.appendLog("info", "app", "配置文件路径: "+a.store.Path(), "")

	if cfg.EnableAutoStart {
		go func() {
			time.Sleep(800 * time.Millisecond)
			if _, err := a.StartBridge(); err != nil {
				a.appendLog("error", "app", "自动启动失败: "+err.Error(), "")
			}
		}()
	}
}

func (a *App) GetAppConfig() (AppConfig, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config, nil
}

func (a *App) SaveAppConfig(cfg AppConfig) (AppConfig, error) {
	cfg = normalizeConfig(cfg)
	if err := validateConfig(cfg, false); err != nil {
		a.appendLog("warn", "app", "配置校验失败: "+err.Error(), "")
		return AppConfig{}, err
	}
	if err := a.store.Save(cfg); err != nil {
		a.appendLog("error", "app", "保存配置失败: "+err.Error(), "")
		return AppConfig{}, err
	}

	a.mu.Lock()
	a.config = cfg
	a.mu.Unlock()

	a.appendLog("info", "app", "配置已保存", "")
	return cfg, nil
}

func (a *App) ExportConfig() (string, error) {
	cfg, err := a.GetAppConfig()
	if err != nil {
		return "", err
	}
	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (a *App) ImportConfig(payload string) (AppConfig, error) {
	var cfg AppConfig
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		return AppConfig{}, err
	}
	return a.SaveAppConfig(cfg)
}

func (a *App) StartBridge() (BridgeStatusPayload, error) {
	cfg, err := a.GetAppConfig()
	if err != nil {
		return BridgeStatusPayload{}, err
	}
	a.appendLog("info", "app", fmt.Sprintf("收到启动请求: %s:%d -> %s (%s)", cfg.ListenHost, cfg.ListenPort, cfg.DeepseekBaseURL, cfg.DefaultModel), "")
	if err := validateConfig(cfg, true); err != nil {
		a.appendLog("warn", "app", "启动前配置校验失败: "+err.Error(), "")
		return BridgeStatusPayload{}, err
	}
	if err := a.bridge.Start(cfg); err != nil {
		a.appendLog("error", "app", "启动桥接失败: "+err.Error(), "")
		return BridgeStatusPayload{}, err
	}
	status := a.bridge.Status()
	a.appendLog("info", "app", "启动命令已提交: "+status.ListenAddress, "")
	return status, nil
}

func (a *App) StopBridge() (BridgeStatusPayload, error) {
	a.appendLog("info", "app", "收到停止请求", "")
	if err := a.bridge.Stop(); err != nil {
		a.appendLog("error", "app", "停止桥接失败: "+err.Error(), "")
		return BridgeStatusPayload{}, err
	}
	return a.bridge.Status(), nil
}

func (a *App) RestartBridge() (BridgeStatusPayload, error) {
	a.appendLog("info", "app", "收到重启请求", "")
	if err := a.bridge.Stop(); err != nil {
		a.appendLog("error", "app", "重启时停止失败: "+err.Error(), "")
		return BridgeStatusPayload{}, err
	}
	return a.StartBridge()
}

func (a *App) GetBridgeStatus() BridgeStatusPayload {
	return a.bridge.Status()
}

func (a *App) GetOverviewSnapshot() (OverviewSnapshot, error) {
	cfg, err := a.GetAppConfig()
	if err != nil {
		return OverviewSnapshot{}, err
	}

	return OverviewSnapshot{
		Config:     cfg,
		Status:     a.bridge.Status(),
		RecentLogs: a.GetLogHistory(6),
		QuickTips: []string{
			"先填写 DeepSeek Base URL、API Key 和默认模型。",
			"启动后将本地地址填入 Codex Desktop 的服务端点。",
			"请求失败时先查看最近日志，再进入完整诊断页。",
		},
		Defaults: map[string]string{
			"baseURL": "https://api.deepseek.com/v1",
			"model":   "deepseek-chat",
		},
		Features: map[string]bool{
			"streamingProxy":   true,
			"healthCheck":      true,
			"logPush":          true,
			"compactDashboard": cfg.CompactMode,
		},
	}, nil
}

func (a *App) RunHealthCheck() (HealthCheckResult, error) {
	cfg, err := a.GetAppConfig()
	if err != nil {
		return HealthCheckResult{}, err
	}

	result := HealthCheckResult{
		OK:     true,
		Checks: make([]HealthCheckItem, 0, 3),
	}

	if err := validateConfig(cfg, true); err != nil {
		result.OK = false
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "配置完整性",
			OK:      false,
			Message: err.Error(),
		})
	} else {
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "配置完整性",
			OK:      true,
			Message: "核心配置已填写",
		})
	}

	if a.bridge.IsRunning() {
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "本地桥接服务",
			OK:      true,
			Message: "桥接服务正在运行: " + a.bridge.Status().ListenAddress,
		})
	} else {
		result.OK = false
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "本地桥接服务",
			OK:      false,
			Message: "桥接服务未启动",
		})
	}

	upstreamErr := a.bridge.CheckUpstream(cfg)
	if upstreamErr != nil {
		result.OK = false
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "DeepSeek 上游接口",
			OK:      false,
			Message: upstreamErr.Error(),
		})
	} else {
		result.Checks = append(result.Checks, HealthCheckItem{
			Name:    "DeepSeek 上游接口",
			OK:      true,
			Message: "上游接口可访问",
		})
	}

	return result, nil
}

func (a *App) GetLogHistory(limit int) []LogEntry {
	a.logsMu.RLock()
	defer a.logsMu.RUnlock()

	if limit <= 0 || limit >= len(a.logs) {
		return append([]LogEntry(nil), a.logs...)
	}

	start := len(a.logs) - limit
	return append([]LogEntry(nil), a.logs[start:]...)
}

func (a *App) appendLog(level string, source string, message string, requestID string) {
	entry := LogEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    source,
		Message:   message,
		RequestID: requestID,
	}

	a.logsMu.Lock()
	a.logs = append(a.logs, entry)
	if len(a.logs) > 500 {
		a.logs = a.logs[len(a.logs)-500:]
	}
	a.logsMu.Unlock()

	if a.ctx != nil {
		ctx := a.ctx
		go runtime.EventsEmit(ctx, "log:entry", entry)
	}
}

func (a *App) emitStatus() {
	if a.ctx != nil {
		ctx := a.ctx
		payload := a.bridge.Status()
		go runtime.EventsEmit(ctx, "bridge:status", payload)
	}
}

func validateConfig(cfg AppConfig, requireCredentials bool) error {
	if strings.TrimSpace(cfg.ListenHost) == "" {
		return errors.New("监听地址不能为空")
	}
	if cfg.ListenPort <= 0 {
		return errors.New("监听端口必须大于 0")
	}
	if strings.TrimSpace(cfg.DeepseekBaseURL) == "" {
		return errors.New("DeepSeek Base URL 不能为空")
	}
	if parsed, err := url.Parse(strings.TrimSpace(cfg.DeepseekBaseURL)); err == nil && parsed.Host != "" {
		if parsed.Port() != "" {
			if net.JoinHostPort(parsed.Hostname(), parsed.Port()) == net.JoinHostPort(strings.TrimSpace(cfg.ListenHost), fmt.Sprintf("%d", cfg.ListenPort)) {
				return errors.New("DeepSeek Base URL 不能指向本桥接地址（会导致请求循环）")
			}
		} else if parsed.Hostname() == strings.TrimSpace(cfg.ListenHost) {
			return errors.New("DeepSeek Base URL 不能指向本桥接地址（会导致请求循环）")
		}
	}
	if requireCredentials && strings.TrimSpace(cfg.APIKey) == "" {
		return errors.New("API Key 不能为空")
	}
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		return errors.New("默认模型不能为空")
	}
	return nil
}
