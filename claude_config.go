package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const claudeSettingsDir = ".claude"
const claudeSettingsFile = "settings.json"

// getClaudeSettingsPath returns the full path to the Claude Code settings file.
func getClaudeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, claudeSettingsDir, claudeSettingsFile), nil
}

func (a *App) GetClaudeSettingsPath() (string, error) {
	return getClaudeSettingsPath()
}

func (a *App) ReadClaudeSettings() (string, error) {
	path, err := getClaudeSettingsPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) WriteClaudeSettings(content string) (string, error) {
	path, err := getClaudeSettingsPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}
	a.appendLog("info", "app", "已保存 Claude Code 设置: "+path, "")
	return path, nil
}

// EnableClaudeSettings writes the current Claude-source profile's model and
// endpoint configuration into ~/.claude/settings.json so that Claude Code is
// ready to use the configured provider via its Anthropic-compatible API.
func (a *App) EnableClaudeSettings() (string, error) {
	cfg, err := a.GetAppConfig()
	if err != nil {
		return "", err
	}

	instCfg, ok := cfg.Instances[SourceClaude]
	if !ok {
		return "", errors.New("Claude Code 实例配置不存在")
	}

	profile, ok := cfg.Profiles[instCfg.CurrentProfileID]
	if !ok {
		return "", errors.New("当前没有选中的模型配置，请先添加代理配置")
	}

	if strings.TrimSpace(profile.APIKey) == "" {
		return "", errors.New("当前配置未设置 API Key")
	}

	baseURL := profile.BaseURL

	// Resolve tiered models from the provider when available; otherwise fall
	// back to the profile default for every slot.
	haikuModel := profile.DefaultModel
	sonnetModel := profile.DefaultModel
	opusModel := profile.DefaultModel
	defaultModel := profile.DefaultModel

	if prov := GetProvider(ProviderID(profile.Provider)); prov != nil {
		if prov.AnthropicBaseURL != "" {
			baseURL = prov.AnthropicBaseURL
		}
		// Some providers expose distinct models per capability tier.
		if prov.AnthropicHaikuModel != "" {
			haikuModel = prov.AnthropicHaikuModel
		}
		if prov.AnthropicSonnetModel != "" {
			sonnetModel = prov.AnthropicSonnetModel
		}
		if prov.AnthropicOpusModel != "" {
			opusModel = prov.AnthropicOpusModel
		}
	}

	settings := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":            profile.APIKey,
			"ANTHROPIC_BASE_URL":              baseURL,
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":   haikuModel,
			"ANTHROPIC_DEFAULT_OPUS_MODEL":    opusModel,
			"ANTHROPIC_DEFAULT_SONNET_MODEL":  sonnetModel,
			"ANTHROPIC_MODEL":                 defaultModel,
		},
		"experimental": map[string]bool{
			"strip_metadata_user_id": true,
		},
	}

	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}

	return a.WriteClaudeSettings(string(content))
}

// RestoreClaudeSettings removes ~/.claude/settings.json so that Claude Code
// falls back to its built-in defaults.
func (a *App) RestoreClaudeSettings() (string, error) {
	path, err := getClaudeSettingsPath()
	if err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	a.appendLog("info", "app", "已移除 Claude Code 设置: "+path, "")
	return path, nil
}
