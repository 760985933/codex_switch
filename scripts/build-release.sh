#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="$(go env GOPATH)/bin:$PATH"

cd "$ROOT_DIR"
mkdir -p release

echo "==> Building macOS Apple Silicon package"
wails build -platform darwin/arm64
rm -rf release/CodexDeepSeekBridge-darwin-arm64.app
cp -R build/bin/codex-deepseek-bridge.app release/CodexDeepSeekBridge-darwin-arm64.app

echo "==> Building Windows x64 package"
wails build -platform windows/amd64
cp build/bin/CodexDeepSeekBridge.exe release/CodexDeepSeekBridge-windows-amd64.exe

echo "==> Building Windows ARM64 package"
wails build -platform windows/arm64
cp build/bin/CodexDeepSeekBridge.exe release/CodexDeepSeekBridge-windows-arm64.exe

echo "==> Done"
