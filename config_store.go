package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ConfigStore struct {
	mu   sync.Mutex
	path string
}

func NewConfigStore() (*ConfigStore, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(baseDir, "codex-deepseek-bridge")
	return &ConfigStore{
		path: filepath.Join(dir, "app-config.json"),
	}, nil
}

func (s *ConfigStore) Path() string {
	return s.path
}

func (s *ConfigStore) Load() (AppConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg := defaultConfig()
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return defaultConfig(), err
	}
	return normalizeConfig(cfg), nil
}

func (s *ConfigStore) Save(cfg AppConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg = normalizeConfig(cfg)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, content, 0o600)
}

func defaultConfig() AppConfig {
	defaultProfile := Profile{
		ID:               "default",
		Name:             "DeepSeek",
		BaseURL:          "https://api.deepseek.com/v1",
		APIKey:           "",
		DefaultModel:     "deepseek-v4-flash",
		RequestTimeoutMs: 60000,
		MaxRetries:       1,
		Mappings: map[string]string{
			"gpt-5.5":       "deepseek-v4-pro",
			"gpt-5.4":       "deepseek-v4-pro",
			"gpt-5.4-mini":  "deepseek-v4-flash",
			"gpt-5.3-codex": "deepseek-v4-pro",
			"gpt-4.1":       "deepseek-v4-flash",
			"gpt-4o":        "deepseek-v4-flash",
			"gpt-4o-mini":   "deepseek-v4-flash",
			"o4-mini":       "deepseek-v4-flash",
		},
		Headers: map[string]string{},
	}

	return AppConfig{
		ListenHost:       "127.0.0.1",
		ListenPort:       17419,
		DeepseekBaseURL:  "https://api.deepseek.com/v1",
		APIKey:           "",
		DefaultModel:     "deepseek-v4-flash",
		RequestTimeoutMs: 60000,
		MaxRetries:       1,
		EnableAutoStart:  false,
		MinimizeToTray:   false,
		LogRetentionDays: 7,
		CompactMode:         true,
		PluginUnlockEnabled: false,
		Mappings: map[string]string{
			"gpt-5.5":       "deepseek-v4-pro",
			"gpt-5.4":       "deepseek-v4-pro",
			"gpt-5.4-mini":  "deepseek-v4-flash",
			"gpt-5.3-codex": "deepseek-v4-pro",
			"gpt-4.1":       "deepseek-v4-flash",
			"gpt-4o":        "deepseek-v4-flash",
			"gpt-4o-mini":   "deepseek-v4-flash",
			"o4-mini":       "deepseek-v4-flash",
		},
		Headers:          map[string]string{},
		Profiles:         map[string]*Profile{"default": &defaultProfile},
		CurrentProfileID: "default",
	}
}

func normalizeConfig(cfg AppConfig) AppConfig {
	defaults := defaultConfig()

	if strings.TrimSpace(cfg.ListenHost) == "" {
		cfg.ListenHost = defaults.ListenHost
	}
	if cfg.ListenPort <= 0 {
		cfg.ListenPort = defaults.ListenPort
	}
	cfg.DeepseekBaseURL = strings.TrimRight(strings.TrimSpace(cfg.DeepseekBaseURL), "/")
	if cfg.DeepseekBaseURL == "" {
		cfg.DeepseekBaseURL = defaults.DeepseekBaseURL
	}
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		cfg.DefaultModel = defaults.DefaultModel
	}
	if cfg.RequestTimeoutMs <= 0 {
		cfg.RequestTimeoutMs = defaults.RequestTimeoutMs
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = defaults.MaxRetries
	}
	if cfg.LogRetentionDays <= 0 {
		cfg.LogRetentionDays = defaults.LogRetentionDays
	}
	if cfg.Mappings == nil {
		cfg.Mappings = map[string]string{}
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}

	for key, value := range defaults.Mappings {
		if _, ok := cfg.Mappings[key]; !ok {
			cfg.Mappings[key] = value
		}
	}

	// --- Multi-profile migration & sync ---

	// Migration: if no profiles exist, create one from old flat fields
	if len(cfg.Profiles) == 0 {
		profile := &Profile{
			ID:               "default",
			Name:             "DeepSeek",
			BaseURL:          cfg.DeepseekBaseURL,
			APIKey:           cfg.APIKey,
			DefaultModel:     cfg.DefaultModel,
			RequestTimeoutMs: cfg.RequestTimeoutMs,
			MaxRetries:       cfg.MaxRetries,
			Mappings:         copyMap(cfg.Mappings),
			Headers:          copyMap(cfg.Headers),
		}
		cfg.Profiles = map[string]*Profile{"default": profile}
		cfg.CurrentProfileID = "default"
	}

	// Ensure current profile ID is valid
	if _, ok := cfg.Profiles[cfg.CurrentProfileID]; !ok {
		for id := range cfg.Profiles {
			cfg.CurrentProfileID = id
			break
		}
	}

	// Normalize each profile and sync current profile → flat fields
	if profile, ok := cfg.Profiles[cfg.CurrentProfileID]; ok {
		normalizeProfile(profile, defaults)
		// Sync current profile back to flat fields for backward compat
		cfg.DeepseekBaseURL = profile.BaseURL
		cfg.APIKey = profile.APIKey
		cfg.DefaultModel = profile.DefaultModel
		cfg.RequestTimeoutMs = profile.RequestTimeoutMs
		cfg.MaxRetries = profile.MaxRetries
		cfg.Mappings = profile.Mappings
		cfg.Headers = profile.Headers
	}

	// Normalize non-current profiles too
	for id, p := range cfg.Profiles {
		if id != cfg.CurrentProfileID {
			normalizeProfile(p, defaults)
		}
	}

	return cfg
}

func normalizeProfile(p *Profile, defaults AppConfig) {
	p.BaseURL = strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if p.BaseURL == "" {
		p.BaseURL = defaults.DeepseekBaseURL
	}
	if strings.TrimSpace(p.DefaultModel) == "" {
		p.DefaultModel = defaults.DefaultModel
	}
	if p.RequestTimeoutMs <= 0 {
		p.RequestTimeoutMs = defaults.RequestTimeoutMs
	}
	if p.MaxRetries < 0 {
		p.MaxRetries = defaults.MaxRetries
	}
	if p.Mappings == nil {
		p.Mappings = map[string]string{}
	}
	if p.Headers == nil {
		p.Headers = map[string]string{}
	}
	for key, value := range defaults.Mappings {
		if _, ok := p.Mappings[key]; !ok {
			p.Mappings[key] = value
		}
	}
}

func copyMap[K comparable, V any](src map[K]V) map[K]V {
	if src == nil {
		return nil
	}
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
