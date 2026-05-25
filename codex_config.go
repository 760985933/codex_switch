package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	codexProviderID = "local-bridge"
)

func codexConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".codex", "config.toml"), nil
}

func (a *App) GetCodexConfigPath() (string, error) {
	return codexConfigPath()
}

func (a *App) ReadCodexConfigToml() (string, error) {
	path, err := codexConfigPath()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(raw), nil
}

func (a *App) GenerateCodexConfigToml() (string, error) {
	status := a.bridge.Status()
	if strings.TrimSpace(status.ListenAddress) == "" {
		return "", errors.New("桥接服务未启动，无法生成 base_url")
	}

	cfg, err := a.GetAppConfig()
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(status.ListenAddress, "/") + "/v1"
	merged, err := mergeCodexConfigToml(nil, baseURL, cfg.DefaultModel)
	if err != nil {
		return "", err
	}
	return string(merged), nil
}

func (a *App) WriteCodexConfigTomlRaw(content string) (string, error) {
	path, err := codexConfigPath()
	if err != nil {
		return "", err
	}

	doc := map[string]any{}
	if len(bytes.TrimSpace([]byte(content))) > 0 {
		if err := toml.Unmarshal([]byte(content), &doc); err != nil {
			return "", err
		}
	}

	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return "", mkErr
	}

	existing, readErr := os.ReadFile(path)
	if readErr == nil && len(existing) > 0 {
		backupPath := path + ".bak"
		_ = os.WriteFile(backupPath, existing, 0o600)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}

	a.appendLog("info", "app", "已写入 Codex config.toml: "+path, "")
	return path, nil
}

func (a *App) WriteCodexConfigToml() (string, error) {
	status := a.bridge.Status()
	if strings.TrimSpace(status.ListenAddress) == "" {
		return "", errors.New("桥接服务未启动，无法生成 base_url")
	}

	cfg, err := a.GetAppConfig()
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(status.ListenAddress, "/") + "/v1"

	path, err := codexConfigPath()
	if err != nil {
		return "", err
	}

	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return "", mkErr
	}

	existing, readErr := os.ReadFile(path)
	if readErr == nil && len(existing) > 0 {
		backupPath := path + ".bak"
		_ = os.WriteFile(backupPath, existing, 0o600)
		a.appendLog("info", "app", "已备份原 Codex config.toml: "+backupPath, "")
	}

	merged, err := mergeCodexConfigToml(existing, baseURL, cfg.DefaultModel)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, merged, 0o600); err != nil {
		return "", err
	}

	a.appendLog("info", "app", "已更新 Codex config.toml（保留原配置项）: "+path, "")
	return path, nil
}

func (a *App) RestoreCodexConfigToml() (string, error) {
	path, err := codexConfigPath()
	if err != nil {
		return "", err
	}

	backupPath := path + ".bak"
	if backup, readBackupErr := os.ReadFile(backupPath); readBackupErr == nil && len(backup) > 0 {
		if writeBackupErr := os.WriteFile(path, backup, 0o600); writeBackupErr != nil {
			return "", writeBackupErr
		}
		a.appendLog("info", "app", "已恢复 Codex config.toml: "+path, "")
		return path, nil
	}

	existing, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	updated, changed, err := removeCodexBridgeFromConfig(existing)
	if err != nil {
		return "", err
	}
	if !changed {
		return path, nil
	}

	if err := os.WriteFile(path, updated, 0o600); err != nil {
		return "", err
	}

	a.appendLog("info", "app", "已从 Codex config.toml 移除 local-bridge 配置: "+path, "")
	return path, nil
}

func mergeCodexConfigToml(existing []byte, baseURL string, defaultModel string) ([]byte, error) {
	doc := map[string]any{}

	if len(bytes.TrimSpace(existing)) > 0 {
		if err := toml.Unmarshal(existing, &doc); err != nil {
			return nil, err
		}
	}

	doc["model_provider"] = codexProviderID
	delete(doc, "model_catalog_json")
	doc["profile"] = codexProviderID
	if strings.TrimSpace(defaultModel) != "" {
		doc["model"] = defaultModel
	} else if _, ok := doc["model"]; !ok {
		doc["model"] = "deepseek-chat"
	}

	modelProviders := ensureTomlMap(doc, "model_providers")
	provider := ensureTomlMap(modelProviders, codexProviderID)
	provider["name"] = "Local Bridge (DeepSeek)"
	provider["base_url"] = baseURL
	provider["wire_api"] = "responses"
	delete(provider, "env_key")
	if _, ok := provider["query_params"]; !ok {
		provider["query_params"] = map[string]any{}
	}

	profiles := ensureTomlMap(doc, "profiles")
	profile := ensureTomlMap(profiles, codexProviderID)
	profile["model_provider"] = codexProviderID
	if model, ok := doc["model"].(string); ok && strings.TrimSpace(model) != "" {
		profile["model"] = model
	} else {
		profile["model"] = "deepseek-chat"
	}
	profile["openai_base_url"] = strings.TrimRight(baseURL, "/") + "/"
	delete(profile, "model_catalog_json")

	return toml.Marshal(doc)
}

func removeCodexBridgeFromConfig(existing []byte) ([]byte, bool, error) {
	doc := map[string]any{}
	if len(bytes.TrimSpace(existing)) == 0 {
		return existing, false, nil
	}
	if err := toml.Unmarshal(existing, &doc); err != nil {
		return nil, false, err
	}

	changed := false

	modelProvidersAny, ok := doc["model_providers"]
	if ok {
		if modelProviders, ok := modelProvidersAny.(map[string]any); ok {
			if _, has := modelProviders[codexProviderID]; has {
				delete(modelProviders, codexProviderID)
				changed = true
			}
			if len(modelProviders) == 0 {
				delete(doc, "model_providers")
				changed = true
			} else {
				doc["model_providers"] = modelProviders
			}

			if current, ok := doc["model_provider"].(string); ok && current == codexProviderID {
				for k := range modelProviders {
					doc["model_provider"] = k
					changed = true
					break
				}
				if doc["model_provider"] == codexProviderID {
					delete(doc, "model_provider")
					changed = true
				}
			}
		}
	}

	if !changed {
		return existing, false, nil
	}
	out, err := toml.Marshal(doc)
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func ensureTomlMap(parent map[string]any, key string) map[string]any {
	if value, ok := parent[key]; ok {
		if m, ok := value.(map[string]any); ok {
			return m
		}
	}

	m := map[string]any{}
	parent[key] = m
	return m
}
