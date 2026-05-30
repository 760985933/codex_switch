package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const claudeSettingsDir = ".claude"
const claudeSettingsFile = "settings.json"

// ── Claude-3p gateway config types ──

const claude3pConfigDir = "Claude-3p"
const claude3pConfigLibrary = "configLibrary"
const claude3pMetaFile = "_meta.json"

// claude3pGatewayUUID is the stable UUID used for the gateway config file.
const claude3pGatewayUUID = "00000000-0000-4000-8000-000000157211"

// claude3pGatewayConfig mirrors the JSON schema that the Claude-3p desktop
// app uses to declare a gateway inference provider.
type claude3pGatewayConfig struct {
	InferenceProvider       string             `json:"inferenceProvider"`
	InferenceGatewayBaseURL string             `json:"inferenceGatewayBaseUrl"`
	InferenceModels         []claude3pModelDef `json:"inferenceModels"`
}

type claude3pModelDef struct {
	Name          string `json:"name"`
	LabelOverride string `json:"labelOverride,omitempty"`
}

// claude3pMeta represents the _meta.json file that controls profile switching
// in the Claude-3p config library.
type claude3pMeta struct {
	AppliedID string            `json:"appliedId"`
	Entries   map[string]string `json:"entries"` // UUID → label
}

// getClaudeSettingsPath returns the full path to the Claude Code settings file.
func getClaudeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, claudeSettingsDir, claudeSettingsFile), nil
}

// getClaude3pConfigLibPath returns the platform-appropriate Claude-3p configLibrary path.
// macOS:  ~/Library/Application Support/Claude-3p/configLibrary/
// Windows: %AppData%/Claude-3p/configLibrary/
// Linux:  ~/.config/Claude-3p/configLibrary/
func getClaude3pConfigLibPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, claude3pConfigDir, claude3pConfigLibrary), nil
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
// endpoint configuration into ~/.claude/settings.json AND the Claude-3p
// desktop gateway config (~/Library/Application Support/Claude-3p/configLibrary/).
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

	// Gateway model names — default to Anthropic tier models, overridable via Claude*Model.
	claudeHaiku := profile.DefaultModel
	claudeSonnet := profile.DefaultModel
	claudeOpus := profile.DefaultModel

	if prov := GetProvider(ProviderID(profile.Provider)); prov != nil {
		if prov.AnthropicBaseURL != "" {
			baseURL = prov.AnthropicBaseURL
		}
		if prov.AnthropicHaikuModel != "" {
			haikuModel = prov.AnthropicHaikuModel
		}
		if prov.AnthropicSonnetModel != "" {
			sonnetModel = prov.AnthropicSonnetModel
		}
		if prov.AnthropicOpusModel != "" {
			opusModel = prov.AnthropicOpusModel
		}
		if prov.ClaudeHaikuModel != "" {
			claudeHaiku = prov.ClaudeHaikuModel
		} else {
			claudeHaiku = haikuModel
		}
		if prov.ClaudeSonnetModel != "" {
			claudeSonnet = prov.ClaudeSonnetModel
		} else {
			claudeSonnet = sonnetModel
		}
		if prov.ClaudeOpusModel != "" {
			claudeOpus = prov.ClaudeOpusModel
		} else {
			claudeOpus = opusModel
		}
	}

	// 1. Write ~/.claude/settings.json
	settings := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":           profile.APIKey,
			"ANTHROPIC_BASE_URL":             baseURL,
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  haikuModel,
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   opusModel,
			"ANTHROPIC_DEFAULT_SONNET_MODEL": sonnetModel,
			"ANTHROPIC_MODEL":                defaultModel,
		},
		"experimental": map[string]bool{
			"strip_metadata_user_id": true,
		},
	}

	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", err
	}

	path, err := a.WriteClaudeSettings(string(content))
	if err != nil {
		return "", err
	}

	// 2. Write Claude-3p gateway config so the desktop app also works.
	if gwPath, gwErr := a.enableClaude3pGateway(baseURL, claudeHaiku, claudeSonnet, claudeOpus, profile.Name); gwErr != nil {
		a.appendLog("warn", "app", fmt.Sprintf("Claude-3p gateway 配置写入失败: %v", gwErr), "")
	} else {
		a.appendLog("info", "app", "已写入 Claude-3p gateway 配置: "+gwPath, "")
	}

	return path, nil
}

// enableClaude3pGateway writes a gateway config JSON file and updates
// _meta.json under the Claude-3p configLibrary directory.
func (a *App) enableClaude3pGateway(baseURL, haikuModel, sonnetModel, opusModel, profileName string) (string, error) {
	libPath, err := getClaude3pConfigLibPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(libPath, 0o755); err != nil {
		return "", fmt.Errorf("创建 configLibrary 目录失败: %w", err)
	}

	// Ensure baseURL has no trailing slash
	cleanBaseURL := strings.TrimRight(baseURL, "/")

	gw := claude3pGatewayConfig{
		InferenceProvider:       "gateway",
		InferenceGatewayBaseURL: cleanBaseURL,
		InferenceModels: []claude3pModelDef{
			{Name: opusModel, LabelOverride: profileName + " Opus"},
			{Name: sonnetModel, LabelOverride: profileName + " Sonnet"},
			{Name: haikuModel, LabelOverride: profileName + " Haiku"},
		},
	}

	gwPath := filepath.Join(libPath, claude3pGatewayUUID+".json")
	gwData, err := json.MarshalIndent(gw, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化 gateway 配置失败: %w", err)
	}
	if err := os.WriteFile(gwPath, gwData, 0o600); err != nil {
		return "", fmt.Errorf("写入 gateway 配置失败: %w", err)
	}

	// Update _meta.json to activate this gateway config.
	if err := a.updateClaude3pMeta(claude3pGatewayUUID, profileName); err != nil {
		return "", err
	}

	return gwPath, nil
}

// updateClaude3pMeta sets the active config entry in _meta.json.
func (a *App) updateClaude3pMeta(uuid, label string) error {
	libPath, err := getClaude3pConfigLibPath()
	if err != nil {
		return err
	}
	metaPath := filepath.Join(libPath, claude3pMetaFile)

	meta := claude3pMeta{
		AppliedID: uuid,
	}
	meta.Entries = map[string]string{uuid: label}

	// If _meta.json already exists, merge to preserve other entries.
	if data, err := os.ReadFile(metaPath); err == nil {
		var existing claude3pMeta
		if err := json.Unmarshal(data, &existing); err == nil {
			meta.Entries = existing.Entries
			meta.Entries[uuid] = label
			meta.AppliedID = uuid
		}
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 _meta.json 失败: %w", err)
	}
	if err := os.WriteFile(metaPath, []byte(data), 0o600); err != nil {
		return fmt.Errorf("写入 _meta.json 失败: %w", err)
	}
	return nil
}

// RestoreClaudeSettings writes an empty settings object to ~/.claude/settings.json
// and removes the Claude-3p gateway config, so Claude Code falls back to its
// built-in defaults.
func (a *App) RestoreClaudeSettings() (string, error) {
	// 1. Overwrite ~/.claude/settings.json with empty object (not delete).
	path, err := getClaudeSettingsPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		return "", err
	}
	a.appendLog("info", "app", "已重置 Claude Code 设置: "+path, "")

	// 2. Clean up Claude-3p gateway config.
	a.restoreClaude3pGateway()

	return path, nil
}

// restoreClaude3pGateway removes the gateway config JSON and updates _meta.json.
func (a *App) restoreClaude3pGateway() {
	libPath, err := getClaude3pConfigLibPath()
	if err != nil {
		return
	}

	// Remove the gateway config file.
	gwPath := filepath.Join(libPath, claude3pGatewayUUID+".json")
	if err := os.Remove(gwPath); err != nil && !os.IsNotExist(err) {
		a.appendLog("warn", "app", "移除 Claude-3p gateway 配置失败: "+err.Error(), "")
	}

	// Remove the entry from _meta.json.
	metaPath := filepath.Join(libPath, claude3pMetaFile)
	if data, err := os.ReadFile(metaPath); err == nil {
		var meta claude3pMeta
		if err := json.Unmarshal(data, &meta); err == nil {
			delete(meta.Entries, claude3pGatewayUUID)
			if meta.AppliedID == claude3pGatewayUUID {
				meta.AppliedID = ""
				// If there are other entries, pick the first one.
				for id := range meta.Entries {
					meta.AppliedID = id
					break
				}
			}
			if len(meta.Entries) == 0 {
				_ = os.Remove(metaPath)
			} else {
				cleaned, _ := json.MarshalIndent(meta, "", "  ")
				_ = os.WriteFile(metaPath, cleaned, 0o600)
			}
		}
	}
}
