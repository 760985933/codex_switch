# Nettopo Switch (Codex ↔ LLM Local Proxy)

[中文](README.zh.md)

## Overview

Nettopo Switch is a local proxy that adapts Codex-compatible `POST /v1/responses` (including SSE streaming) and forwards requests to various LLM providers' `POST /v1/chat/completions`. It provides a desktop UI to manage endpoints, API keys, ports, and model mappings across multiple providers, plus tools to manage Codex `config.toml` with safe backups.

This is designed to work with Codex Desktop by pointing Codex's Base URL to the local proxy, so Codex can use different LLM models without changing its workflow.

## Features

- **Codex Responses adapter**: supports `POST /v1/responses` (SSE) and forwards to provider `POST /v1/chat/completions`
- **Multi-provider support**: switch between 9+ built-in LLM providers, or add custom ones
- **Model mapping**: maps Codex-side model names (e.g. `gpt-5.4-mini`) to provider-specific models
- **Menu bar balance display**: shows real-time available usage balance in the system tray (for supported providers like DeepSeek)
- **Visual configuration**: configure Base URL, API key, port, and mappings in the desktop app
- **Codex config.toml management**: merge-write, edit raw content, create/restore/delete/clean backup history
- **Health check & logs**: one-click upstream connectivity check; logs for each request
- **UI i18n**: `zh-CN` (简体中文), `en-US` (English), `ja-JP` (日本語), `ko-KR` (한국어), `fr-FR` (Français), `de-DE` (Deutsch), `es-ES` (Español)
- **Cross-platform builds**: macOS arm64 / Windows amd64 / Windows arm64

## Supported Providers

| Provider | Default Base URL | Default Model |
|---|---|---|
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-v4-flash` |
| 阿里通义千问 (Alibaba) | `https://dashscope.aliyuncs.com/compatible-mode/v1` | `qwen3.6-plus` |
| 小米 MiMo (Xiaomi) | `https://api.xiaomimimo.com/v1` | `mimo-v2.5-pro` |
| 智谱 GLM (Zhipu) | `https://open.bigmodel.cn/api/paas/v4` | `glm-4.7-flash` |
| 百度千帆 (Baidu) | `https://qianfan.baidubce.com/v2` | `ernie-5.1` |
| 火山引擎豆包 (Volcano) | `https://ark.cn-beijing.volces.com/api/v3` | `doubao-seed-2-0-lite-260215` |
| 腾讯混元 (Tencent) | `https://api.hunyuan.cloud.tencent.com/v1` | `hunyuan-2.0-thinking-20251109` |
| 硅基流动 (SiliconFlow) | `https://api.siliconflow.cn/v1` | `deepseek-ai/DeepSeek-V4-Flash` |
| Kimi (Moonshot) | `https://api.moonshot.cn/v1` | `kimi-k2.6` |
| Custom | User-defined | User-defined |

## Endpoints

- `GET /`: service info
- `GET /health`: health status
- `GET /v1/models`: model list (for Codex UI model selection)
- `POST /v1/responses`: Codex entrypoint (recommended)
- `POST /v1/chat/completions`: compatibility endpoint

## Quick Start

1) In the desktop app, select a provider and configure:
   - Base URL & API Key for the chosen provider
   - Default model

2) Start the proxy service (default listen: `http://127.0.0.1:11434`)

3) Verify:

```bash
curl http://127.0.0.1:11434/health
```

## Configure Codex

In the app: **Preferences → Codex config.toml**

- **Merge write**: updates `~/.codex/config.toml` while preserving other existing settings
- **Backups**: each write creates a non-overwriting backup under `~/.codex/backups/` which can be restored or deleted

Codex Base URL should point to:

```
http://127.0.0.1:11434/v1
```

## Menu Bar Balance

On macOS, the system tray icon displays the available usage balance (e.g. `12.34 CNY`) for providers that support balance checking (currently DeepSeek). The balance refreshes every 60 seconds.

## FAQ

### Codex shows 502 / Reconnecting

- Check “Recent Logs” in the app for `upstream 4xx/5xx` or forwarding failures
- If the provider does not recognize a model name, update “Model Mapping” to map Codex models to the correct provider-specific model names

### How do I add a custom provider?

Use the “Custom” provider type in the app and fill in your own Base URL, API key, and model mappings.

## Development & Build

### Local development

```bash
npm -C frontend install
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
wails dev
```

### Build (example)

```bash
export PATH=”$(go env GOPATH)/bin:$PATH”
wails build -platform darwin/arm64
```
