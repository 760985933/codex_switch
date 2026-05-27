package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

func sandboxConfigPath() (string, error) {
	return codexConfigPath()
}

func (a *App) GetSandboxConfig() (SandboxWorkspaceConfig, error) {
	path, err := sandboxConfigPath()
	if err != nil {
		return SandboxWorkspaceConfig{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SandboxWorkspaceConfig{NetworkAccess: true}, nil
		}
		return SandboxWorkspaceConfig{}, err
	}

	doc := map[string]any{}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := toml.Unmarshal(raw, &doc); err != nil {
			return SandboxWorkspaceConfig{}, err
		}
	}

	cfg := SandboxWorkspaceConfig{NetworkAccess: true}
	if sw, ok := doc["sandbox_workspace_write"]; ok {
		if swMap, ok := sw.(map[string]any); ok {
			if na, ok := swMap["network_access"]; ok {
				if b, ok := na.(bool); ok {
					cfg.NetworkAccess = b
				}
			}
		}
	}

	return cfg, nil
}

func (a *App) SetSandboxConfig(cfg SandboxWorkspaceConfig) (SandboxWorkspaceConfig, error) {
	path, err := sandboxConfigPath()
	if err != nil {
		return SandboxWorkspaceConfig{}, err
	}

	raw := []byte{}
	existing, readErr := os.ReadFile(path)
	if readErr == nil && len(existing) > 0 {
		raw = existing
	}

	doc := map[string]any{}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := toml.Unmarshal(raw, &doc); err != nil {
			return SandboxWorkspaceConfig{}, err
		}
	}

	// Dedup: skip write if the value is already the same
	current := SandboxWorkspaceConfig{NetworkAccess: true}
	if sw, ok := doc["sandbox_workspace_write"]; ok {
		if swMap, ok := sw.(map[string]any); ok {
			if na, ok := swMap["network_access"]; ok {
				if b, ok := na.(bool); ok {
					current.NetworkAccess = b
				}
			}
		}
	}
	if current.NetworkAccess == cfg.NetworkAccess {
		return cfg, nil
	}

	doc["sandbox_workspace_write"] = map[string]any{
		"network_access": cfg.NetworkAccess,
	}

	out, err := toml.Marshal(doc)
	if err != nil {
		return SandboxWorkspaceConfig{}, err
	}

	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return SandboxWorkspaceConfig{}, mkErr
	}

	if err := os.WriteFile(path, out, 0o600); err != nil {
		return SandboxWorkspaceConfig{}, err
	}

	a.appendLog("info", "app", "已更新 sandbox 配置: network_access="+boolStr(cfg.NetworkAccess)+" → "+path, "")
	return cfg, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
