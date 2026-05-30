package main

import (
	"os"
	"path/filepath"
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
