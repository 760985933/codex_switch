#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="$(go env GOPATH)/bin:$PATH"

cd "$ROOT_DIR"
mkdir -p release
rm -rf release/CodexDeepSeekBridge-darwin-arm64.app
rm -rf release/CodexDeepSeekBridge-darwin-amd64.app
rm -f release/CodexDeepSeekBridge-windows-amd64.exe
rm -f release/CodexDeepSeekBridge-windows-arm64.exe

echo "==> Building macOS Apple Silicon package"
wails build -platform darwin/arm64
rm -rf release/nettopo-switch-darwin-arm64.app
cp -R build/bin/codex-deepseek-bridge.app release/nettopo-switch-darwin-arm64.app

echo "==> Building macOS Intel package"
wails build -platform darwin/amd64
rm -rf release/nettopo-switch-darwin-amd64.app
cp -R build/bin/codex-deepseek-bridge.app release/nettopo-switch-darwin-amd64.app

echo "==> Building Windows x64 package"
wails build -platform windows/amd64
rm -f release/nettopo-switch-windows-amd64.exe
cp build/bin/CodexDeepSeekBridge.exe release/nettopo-switch-windows-amd64.exe

echo "==> Building Windows ARM64 package"
wails build -platform windows/arm64
rm -f release/nettopo-switch-windows-arm64.exe
cp build/bin/CodexDeepSeekBridge.exe release/nettopo-switch-windows-arm64.exe

echo "==> Done"
