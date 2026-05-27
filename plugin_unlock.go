package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const codexDebugPort = 9222

const unlockScript = `
(function(){
	console.log("[Nettopo] Injecting plugin unlock...");

	// 1. Force-show all plugin-related UI elements
	var css = [
		'[class*="plugin"]',
		'[class*="Plugin"]',
		'[class*="plugin" i]',
		'.plugin-entry',
		'.plugins-menu',
		'[class*="sidebar"] [class*="item"]',
		'[role="button"][class*="plug"]',
		'[data-testid*="plugin"]',
		'[id*="plugin"]',
		'[id*="Plugin"]',
		'button:has([class*="plugin"])',
		'a:has([class*="plugin"])',
	].join(',');

	var style = document.createElement('style');
	style.id = 'nettopo-plugin-unlock';
	style.textContent = css + '{display:flex!important;visibility:visible!important;opacity:1!important;pointer-events:auto!important;height:auto!important;width:auto!important;overflow:visible!important;clip:auto!important;position:static!important;transform:none!important;max-height:none!important}';
	document.head.appendChild(style);

	// 2. Hijack feature flags
	var originalSetItem = localStorage.setItem;
	localStorage.setItem = function(key, value) {
		if (key && key.toLowerCase().indexOf('feature') !== -1) {
			try {
				var obj = JSON.parse(value);
				if (obj && typeof obj === 'object') {
					obj.plugins = true;
					obj.pluginsEnabled = true;
					value = JSON.stringify(obj);
				}
			} catch(e) {}
		}
		originalSetItem.apply(this, arguments);
	};

	// 3. Set global flags
	window.__PLUGINS_ENABLED = true;
	window.__PLUGIN_UNLOCKED = true;
	window.hasPluginAccess = function(){return true};
	if (window.APP_CONFIG && window.APP_CONFIG.features) {
		window.APP_CONFIG.features.plugins = true;
		window.APP_CONFIG.features.pluginEnabled = true;
	}
	if (window.userFeatures) {
		window.userFeatures.plugins = true;
		window.userFeatures.pluginEnabled = true;
	}

	// 4. Watch for dynamically added plugin elements and make them visible
	var observer = new MutationObserver(function(mutations) {
		mutations.forEach(function(mutation) {
			mutation.addedNodes.forEach(function(node) {
				if (node.nodeType === 1) {
					if (node.className && (node.className.toLowerCase().indexOf('plugin') !== -1)) {
						node.style.setProperty('display', 'flex', 'important');
						node.style.setProperty('visibility', 'visible', 'important');
						node.style.setProperty('opacity', '1', 'important');
						node.style.setProperty('pointer-events', 'auto', 'important');
					}
					if (node.querySelectorAll) {
						var els = node.querySelectorAll('[class*="plugin"], [class*="Plugin"]');
						els.forEach(function(el) {
							el.style.setProperty('display', 'flex', 'important');
							el.style.setProperty('visibility', 'visible', 'important');
							el.style.setProperty('opacity', '1', 'important');
							el.style.setProperty('pointer-events', 'auto', 'important');
						});
					}
				}
			});
		});
	});
	observer.observe(document.documentElement, {childList: true, subtree: true});

	// 5. Patch fetch/API responses that check plugin access
	var originalFetch = window.fetch;
	window.fetch = function() {
		return originalFetch.apply(this, arguments).then(function(response) {
			var cloned = response.clone();
			var url = arguments[0] && typeof arguments[0] === 'string' ? arguments[0] : (arguments[0] && arguments[0].url || '');
			if (url.toLowerCase().indexOf('feature') !== -1 || url.toLowerCase().indexOf('entitlement') !== -1) {
				cloned.json().then(function(data) {
					if (data && typeof data === 'object') {
						data.plugins = true;
						data.pluginsEnabled = true;
					}
				}).catch(function(){});
			}
			return response;
		}).catch(function() {
			return Promise.reject(arguments[0]);
		});
	};

	console.log("[Nettopo] Plugin unlock injected successfully");
})();
`

func findCodexPath() string {
	if runtime.GOOS == "darwin" {
		paths := []string{
			"/Applications/Codex.app/Contents/MacOS/Codex",
			os.Getenv("HOME") + "/Applications/Codex.app/Contents/MacOS/Codex",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return ""
	}

	if runtime.GOOS == "windows" {
		paths := []string{
			os.Getenv("LOCALAPPDATA") + "\\Programs\\Codex\\Codex.exe",
			os.Getenv("USERPROFILE") + "\\AppData\\Local\\Programs\\Codex\\Codex.exe",
			"C:\\Program Files\\Codex\\Codex.exe",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return ""
	}

	return ""
}

type cdpTarget struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	Type                 string `json:"type"`
}

// getCodexTargets fetches all CDP targets and returns page-type targets,
// preferring the Codex main window.
func getCodexTargets() []cdpTarget {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json/list", codexDebugPort))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var all []cdpTarget
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil
	}

	// Filter to page targets (skip service workers, extensions, etc.)
	var pages []cdpTarget
	for _, t := range all {
		if t.WebSocketDebuggerURL == "" {
			continue
		}
		// Accept page type or any target that isn't clearly background
		if t.Type == "page" || t.Type == "" {
			pages = append(pages, t)
		}
	}

	// Sort: prefer targets whose title contains "Codex" or that have a file:// URL
	var codexPages, otherPages []cdpTarget
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.Title), "codex") ||
			strings.HasPrefix(p.URL, "file://") {
			codexPages = append(codexPages, p)
		} else {
			otherPages = append(otherPages, p)
		}
	}

	result := append(codexPages, otherPages...)
	return result
}

func injectViaWebSocket(wsURL string) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %w", err)
	}
	defer conn.Close()

	msg := map[string]interface{}{
		"id":     1,
		"method": "Runtime.evaluate",
		"params": map[string]interface{}{
			"expression":    unlockScript,
			"returnByValue": true,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("发送注入命令失败: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var result map[string]interface{}
	if err := conn.ReadJSON(&result); err != nil {
		return fmt.Errorf("读取注入结果失败: %w", err)
	}

	if errInfo, ok := result["error"]; ok {
		return fmt.Errorf("注入执行异常: %v", errInfo)
	}

	return nil
}

// TryPluginUnlock attempts to inject the plugin unlock script into Codex.
// If Codex is not running with a debug port, it tries to launch it.
func TryPluginUnlock(logFn func(level, source, msg, requestID string)) error {
	// Step 1: try existing Codex with debug port
	targets := getCodexTargets()
	if len(targets) > 0 {
		logFn("info", "plugin", fmt.Sprintf("检测到 Codex 调试端口（%d 个页面），注入解锁脚本...", len(targets)), "")

		var lastErr error
		successCount := 0
		for _, t := range targets {
			logFn("info", "plugin", fmt.Sprintf("注入目标: %s (%s)", t.Title, t.URL), "")
			if err := injectViaWebSocket(t.WebSocketDebuggerURL); err != nil {
				lastErr = err
				logFn("warn", "plugin", fmt.Sprintf("目标 %s 注入失败: %v", t.Title, err), "")
			} else {
				successCount++
			}
		}

		if successCount > 0 {
			logFn("info", "plugin", fmt.Sprintf("插件解锁脚本已注入 %d 个页面", successCount), "")
			return nil
		}
		if lastErr != nil {
			return fmt.Errorf("所有目标注入失败，最后错误: %w", lastErr)
		}
		return fmt.Errorf("未找到可注入的 Codex 页面")
	}

	// Step 2: try to launch Codex with debug flag
	codexPath := findCodexPath()
	if codexPath == "" {
		return fmt.Errorf("未找到 Codex 安装路径，请确认已安装 Codex Desktop")
	}

	logFn("info", "plugin", fmt.Sprintf("正在启动 Codex（路径: %s）...", codexPath), "")

	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", "Codex.exe").Run()
	} else {
		exec.Command("pkill", "-f", "Codex").Run()
	}
	time.Sleep(800 * time.Millisecond)

	cmd := exec.Command(codexPath, fmt.Sprintf("--remote-debugging-port=%d", codexDebugPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Codex 失败: %w", err)
	}

	// Step 3: wait for Codex to start and expose debug port
	logFn("info", "plugin", "等待 Codex 启动并连接调试端口...", "")
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		targets = getCodexTargets()
		if len(targets) > 0 {
			logFn("info", "plugin", fmt.Sprintf("已连接 Codex（%d 个页面），注入解锁脚本...", len(targets)), "")
			successCount := 0
			for _, t := range targets {
				if err := injectViaWebSocket(t.WebSocketDebuggerURL); err != nil {
					logFn("warn", "plugin", fmt.Sprintf("目标 %s 注入失败: %v", t.Title, err), "")
				} else {
					successCount++
				}
			}
			if successCount > 0 {
				logFn("info", "plugin", fmt.Sprintf("插件解锁脚本已注入 %d 个页面", successCount), "")
				return nil
			}
		}
	}

	return fmt.Errorf("等待 Codex 启动超时，请手动启动 Codex 并添加 --remote-debugging-port=%d 参数", codexDebugPort)
}
