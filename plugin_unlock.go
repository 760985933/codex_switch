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

const codexDebugPort = 9229

// unlockScript — injected via Page.addScriptToEvaluateOnNewDocument so it runs
// BEFORE any Codex JS initializes. Based on Codex++ approach.
const unlockScript = `
(function(){
	// Patch feature flags before any app code reads them
	var origDefineProperty = Object.defineProperty;
	Object.defineProperty = function(obj, prop, desc) {
		if (prop === 'plugins' || prop === 'pluginsEnabled' || prop === 'pluginEnabled') {
			if (desc && (desc.value === false || desc.get)) {
				console.log('[Codex++] Intercepted Object.defineProperty for', prop);
				return;
			}
		}
		return origDefineProperty.call(this, obj, prop, desc);
	};

	// Intercept fetch to patch feature responses
	var origFetch = window.fetch;
	window.fetch = function(url, options) {
		var urlStr = typeof url === 'string' ? url : (url && url.url ? url.url : '');
		return origFetch.apply(this, arguments).then(function(resp) {
			var ct = resp.headers.get('content-type') || '';
			if (ct.indexOf('json') !== -1) {
				var cloned = resp.clone();
				cloned.json().then(function(data) {
					var changed = false;
					function deepPatch(obj) {
						if (!obj || typeof obj !== 'object') return;
						if (Array.isArray(obj)) { obj.forEach(deepPatch); return; }
						Object.keys(obj).forEach(function(k) {
							var lk = k.toLowerCase();
							if (lk === 'plugins' || lk === 'pluginsenabled' || lk === 'pluginenabled') {
								if (typeof obj[k] === 'boolean' || obj[k] === null || obj[k] === undefined) {
									obj[k] = true; changed = true;
								}
							}
							if (lk === 'features' || lk === 'entitlements' || lk === 'featureflags') {
								if (obj[k] && typeof obj[k] === 'object') {
									obj[k].plugins = true;
									obj[k].pluginsEnabled = true;
									obj[k].pluginEnabled = true;
									changed = true;
								}
							}
							if (typeof obj[k] === 'object') deepPatch(obj[k]);
						});
					}
					deepPatch(data);
				}).catch(function(){});
			}
			return resp;
		}).catch(function(e) { return Promise.reject(e); });
	};

	// Intercept WebSocket for real-time feature updates
	var OrigWebSocket = window.WebSocket;
	window.WebSocket = function(url, protocols) {
		var ws = new OrigWebSocket(url, protocols);
		var desc = Object.getOwnPropertyDescriptor(ws.constructor.prototype, 'onmessage');
		if (!desc) {
			desc = Object.getOwnPropertyDescriptor(ws, 'onmessage');
		}
		var realHandler = null;
		try {
			Object.defineProperty(ws, 'onmessage', {
				get: function() { return realHandler; },
				set: function(fn) {
					realHandler = function(event) {
						try {
							if (typeof event.data === 'string') {
								var data = JSON.parse(event.data);
								if (data && typeof data === 'object') {
									if (data.type === 'features' || data.type === 'entitlements') {
										data.plugins = true; data.pluginsEnabled = true;
									}
									event = new MessageEvent('message', {data: JSON.stringify(data), origin: event.origin});
								}
							}
						} catch(e) {}
						return fn.call(this, event);
					};
				},
				configurable: true
			});
		} catch(e) {}
		return ws;
	};

	// Intercept Storage APIs
	var origSetItem = Storage.prototype.setItem;
	Storage.prototype.setItem = function(key, value) {
		try {
			var k = (key||'').toLowerCase();
			if (k.indexOf('feature')!==-1 || k.indexOf('plugin')!==-1 || k.indexOf('entitle')!==-1) {
				var o = JSON.parse(value);
				if (o && typeof o === 'object') { o.plugins=true; o.pluginsEnabled=true; value=JSON.stringify(o); }
			}
		} catch(e) {}
		return origSetItem.call(this, key, value);
	};

	// Intercept IPC calls (Electron)
	try {
		var apiKeys = Object.keys(window).filter(function(k) {
			try { return window[k] && typeof window[k] === 'object' && typeof window[k].invoke === 'function'; } catch(e) {}
		});
		apiKeys.forEach(function(k) {
			var api = window[k];
			var origInvoke = api.invoke.bind(api);
			api.invoke = function() {
				return origInvoke.apply(this, arguments).then(function(r) {
					if (r && typeof r === 'object') {
						function deepPatch(obj) {
							if (!obj || typeof obj !== 'object') return;
							if (Array.isArray(obj)) { obj.forEach(deepPatch); return; }
							Object.keys(obj).forEach(function(k) {
								var lk = k.toLowerCase();
								if (lk === 'plugins' || lk === 'pluginsenabled') { obj[k]=true; }
								if (lk === 'features' || lk === 'entitlements') {
									if (obj[k] && typeof obj[k] === 'object') { obj[k].plugins=true; obj[k].pluginsEnabled=true; }
								}
								if (typeof obj[k] === 'object') deepPatch(obj[k]);
							});
						}
						deepPatch(r);
					}
					return r;
				}).catch(function(e) { return Promise.reject(e); });
			};
		});
	} catch(e) {}

	// Set globals before any app code runs
	Object.defineProperty(window, '__PLUGINS_ENABLED', {value:true, writable:false, configurable:false});
	Object.defineProperty(window, 'hasPluginAccess', {value:function(){return true}, writable:false, configurable:false});

	// Periodic CSS/direct patching for late-loading elements
	function patchOnce() {
		// CSS
		if (!document.getElementById('codexpp-unlock-css')) {
			var style = document.createElement('style');
			style.id = 'codexpp-unlock-css';
			style.textContent = '[class*="plugin"],[class*="Plugin"]{display:flex!important;visibility:visible!important;opacity:1!important;pointer-events:auto!important}[class*="overlay"]:has([class*="plugin"]){display:none!important}[class*="plugin"] [disabled]{pointer-events:auto!important;opacity:1!important}';
			document.head.appendChild(style);
		}
		// Button/click enabler
		document.querySelectorAll('[class*="plugin"] [disabled],[class*="Plugin"] [disabled]').forEach(function(el){
			el.removeAttribute('disabled');
		});
		document.querySelectorAll('[aria-disabled="true"]').forEach(function(el){
			if (el.className && typeof el.className === 'string' && el.className.toLowerCase().indexOf('plugin') !== -1) {
				el.setAttribute('aria-disabled','false');
			}
		});
	}
	document.addEventListener('DOMContentLoaded', function(){ patchOnce(); setTimeout(patchOnce, 1000); setTimeout(patchOnce, 5000); });
	if (document.readyState !== 'loading') { patchOnce(); setTimeout(patchOnce, 1000); setTimeout(patchOnce, 5000); }

	console.log('[Codex++] Plugin unlock injected via addScriptToEvaluateOnNewDocument');
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

	var pages []cdpTarget
	for _, t := range all {
		if t.WebSocketDebuggerURL == "" {
			continue
		}
		if t.Type == "page" || t.Type == "" {
			pages = append(pages, t)
		}
	}

	var codexPages, otherPages []cdpTarget
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.Title), "codex") || strings.HasPrefix(p.URL, "file://") {
			codexPages = append(codexPages, p)
		} else {
			otherPages = append(otherPages, p)
		}
	}
	return append(codexPages, otherPages...)
}

// injectUnlockScript connects via CDP WebSocket and uses
// Page.addScriptToEvaluateOnNewDocument to register the injection
// BEFORE any page JS runs. Then navigates to trigger execution.
func injectUnlockScript(logFn func(level, source, msg, requestID string), wsURL string) error {
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %w", err)
	}
	defer conn.Close()

	// Enable Page domain
	conn.WriteJSON(map[string]interface{}{"id": 1, "method": "Page.enable"})
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var res map[string]interface{}
	for i := 0; i < 3; i++ {
		if err := conn.ReadJSON(&res); err != nil {
			break
		}
		if _, ok := res["result"]; ok {
			break
		}
	}
	logFn("info", "plugin", "CDP Page.enable 完成", "")

	// Register script to evaluate on new document BEFORE any page JS
	msg := map[string]interface{}{
		"id":     2,
		"method": "Page.addScriptToEvaluateOnNewDocument",
		"params": map[string]interface{}{
			"source": unlockScript,
		},
	}
	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("addScriptToEvaluateOnNewDocument 失败: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err := conn.ReadJSON(&res); err != nil {
		return fmt.Errorf("读取结果失败: %w", err)
	}
	if errInfo, ok := res["error"]; ok {
		return fmt.Errorf("addScriptToEvaluateOnNewDocument 异常: %v", errInfo)
	}
	scriptID, _ := res["result"].(map[string]interface{})["identifier"].(string)
	logFn("info", "plugin", fmt.Sprintf("注入脚本已注册 (scriptID=%s)，将在页面加载前执行", scriptID), "")

	// Also evaluate immediately in current page context
	conn.WriteJSON(map[string]interface{}{
		"id":     3,
		"method": "Runtime.evaluate",
		"params": map[string]interface{}{
			"expression":    unlockScript,
			"returnByValue": true,
		},
	})
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.ReadJSON(&res)

	return nil
}

// TryPluginUnlock attempts to inject the plugin unlock script into Codex.
func TryPluginUnlock(logFn func(level, source, msg, requestID string)) error {
	// Step 1: try existing Codex with debug port
	targets := getCodexTargets()
	if len(targets) > 0 {
		logFn("info", "plugin", fmt.Sprintf("检测到 Codex 调试端口（%d 个页面）", len(targets)), "")

		var lastErr error
		successCount := 0
		for _, t := range targets {
			logFn("info", "plugin", fmt.Sprintf("注入目标: %s (%s)", t.Title, t.URL), "")
			if err := injectUnlockScript(logFn, t.WebSocketDebuggerURL); err != nil {
				lastErr = err
				logFn("warn", "plugin", fmt.Sprintf("目标 %s: %v", t.Title, err), "")
			} else {
				successCount++
			}
		}
		if successCount > 0 {
			logFn("info", "plugin", fmt.Sprintf("插件解锁已注入 %d 个页面（addScriptToEvaluateOnNewDocument）", successCount), "")
			return nil
		}
		if lastErr != nil {
			return fmt.Errorf("注入失败: %w", lastErr)
		}
		return fmt.Errorf("未找到 Codex 页面")
	}

	// Step 2: launch Codex with debug flags
	codexPath := findCodexPath()
	if codexPath == "" {
		return fmt.Errorf("未找到 Codex 安装路径")
	}

	logFn("info", "plugin", fmt.Sprintf("启动 Codex: %s", codexPath), "")

	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", "Codex.exe").Run()
	} else {
		exec.Command("pkill", "-f", "Codex").Run()
	}
	time.Sleep(800 * time.Millisecond)

	cmd := exec.Command(codexPath,
		fmt.Sprintf("--remote-debugging-port=%d", codexDebugPort),
		fmt.Sprintf("--remote-allow-origins=http://127.0.0.1:%d", codexDebugPort),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}

	// Step 3: wait and inject
	logFn("info", "plugin", "等待 Codex 调试端口...", "")
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		targets = getCodexTargets()
		if len(targets) > 0 {
			successCount := 0
			for _, t := range targets {
				if err := injectUnlockScript(logFn, t.WebSocketDebuggerURL); err != nil {
					logFn("warn", "plugin", fmt.Sprintf("%s: %v", t.Title, err), "")
				} else {
					successCount++
				}
			}
			if successCount > 0 {
				logFn("info", "plugin", fmt.Sprintf("插件解锁已注入 %d 个页面", successCount), "")
				return nil
			}
		}
	}
	return fmt.Errorf("Codex 启动超时")
}
