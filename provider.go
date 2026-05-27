package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProviderID 定义支持的提供商标识
type ProviderID string

const (
	ProviderDeepSeek ProviderID = "deepseek"
	ProviderAlibaba  ProviderID = "alibaba"
	ProviderXiaomi   ProviderID = "xiaomi"
	ProviderZhipu    ProviderID = "zhipu"
	ProviderCustom   ProviderID = "custom"
)

// ProviderInfo 描述一个 LLM 提供商的元信息
type ProviderInfo struct {
	ID              ProviderID        // 内部标识
	Name            string            // 显示名称
	DefaultBaseURL  string            // 默认 API 地址
	DefaultModel    string            // 默认模型
	DocsURL         string            // API 文档地址
	DefaultMappings map[string]string // Codex 模型 → 提供商模型映射
	HasBalanceAPI   bool              // 是否有公开余额查询接口
	BalanceCheckFn  func(apiKey, baseURL string) (*UsageBalance, error) // 余额查询函数（nil 表示不支持）
}

// GetProvider 根据 ID 获取提供商信息；未知 ID 返回 nil
func GetProvider(id ProviderID) *ProviderInfo {
	p, _ := registeredProviders[string(id)]
	return p
}

// GetDefaultProvider 返回默认的 DeepSeek 提供商
func GetDefaultProvider() *ProviderInfo {
	return GetProvider(ProviderDeepSeek)
}

// AllProviders 返回所有预置提供商列表（不含 "custom"）
func AllProviders() []ProviderInfo {
	list := make([]ProviderInfo, 0, len(registeredProviders))
	for _, p := range registeredProviders {
		if p.ID != ProviderCustom {
			list = append(list, *p)
		}
	}
	return list
}

// deepseekBalanceCheck 查询 DeepSeek 余额
func deepseekBalanceCheck(apiKey, baseURL string) (*UsageBalance, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	balanceURL := fmt.Sprintf("%s://%s/user/balance", parsed.Scheme, parsed.Host)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, balanceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var balanceResp struct {
		IsAvailable  bool `json:"is_available"`
		BalanceInfos []struct {
			Currency        string `json:"currency"`
			TotalBalance    string `json:"total_balance"`
			GrantedBalance  string `json:"granted_balance"`
			ToppedUpBalance string `json:"topped_up_balance"`
		} `json:"balance_infos"`
	}
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	result := &UsageBalance{
		AvailableBalance: "",
		TotalBalance:     "",
		Currency:         "",
		IsDepleted:       !balanceResp.IsAvailable,
	}
	if len(balanceResp.BalanceInfos) > 0 {
		info := balanceResp.BalanceInfos[0]
		result.AvailableBalance = info.ToppedUpBalance
		result.TotalBalance = info.TotalBalance
		result.Currency = info.Currency
	}
	return result, nil
}

// registeredProviders 是全局提供商注册表
var registeredProviders = map[string]*ProviderInfo{
	string(ProviderDeepSeek): {
		ID:              ProviderDeepSeek,
		Name:            "DeepSeek",
		DefaultBaseURL:  "https://api.deepseek.com/v1",
		DefaultModel:    "deepseek-v4-flash",
		DocsURL:         "https://api-docs.deepseek.com/",
		HasBalanceAPI:   true,
		BalanceCheckFn:  deepseekBalanceCheck,
		DefaultMappings: deepseekDefaultMappings(),
	},
	string(ProviderAlibaba): {
		ID:              ProviderAlibaba,
		Name:            "阿里通义千问",
		DefaultBaseURL:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultModel:    "qwen3.6-plus",
		DocsURL:         "https://help.aliyun.com/zh/model-studio/models",
		HasBalanceAPI:   false,
		BalanceCheckFn:  nil,
		DefaultMappings: alibabaDefaultMappings(),
	},
	string(ProviderXiaomi): {
		ID:              ProviderXiaomi,
		Name:            "小米 MiMo",
		DefaultBaseURL:  "https://api.xiaomimimo.com/v1",
		DefaultModel:    "mimo-v2.5-pro",
		DocsURL:         "https://platform.xiaomimimo.com/#/docs/welcome",
		HasBalanceAPI:   false,
		BalanceCheckFn:  nil,
		DefaultMappings: xiaomiDefaultMappings(),
	},
	string(ProviderZhipu): {
		ID:              ProviderZhipu,
		Name:            "智谱 GLM",
		DefaultBaseURL:  "https://open.bigmodel.cn/api/paas/v4",
		DefaultModel:    "glm-4.7-flash",
		DocsURL:         "https://docs.bigmodel.cn/",
		HasBalanceAPI:   false,
		BalanceCheckFn:  nil,
		DefaultMappings: zhipuDefaultMappings(),
	},
}

func deepseekDefaultMappings() map[string]string {
	return map[string]string{
		"gpt-5.5":       "deepseek-v4-pro",
		"gpt-5.4":       "deepseek-v4-pro",
		"gpt-5.4-mini":  "deepseek-v4-flash",
		"gpt-5.3-codex": "deepseek-v4-pro",
		"gpt-4.1":       "deepseek-v4-flash",
		"gpt-4o":        "deepseek-v4-flash",
		"gpt-4o-mini":   "deepseek-v4-flash",
		"o4-mini":       "deepseek-v4-flash",
	}
}

func alibabaDefaultMappings() map[string]string {
	return map[string]string{
		"gpt-5.5":       "qwen3.6-max-preview",
		"gpt-5.4":       "qwen3.6-max-preview",
		"gpt-5.4-mini":  "qwen3.6-flash",
		"gpt-5.3-codex": "qwen3.6-max-preview",
		"gpt-4.1":       "qwen3.6-flash",
		"gpt-4o":        "qwen3.6-flash",
		"gpt-4o-mini":   "qwen3.6-flash",
		"o4-mini":       "qwq-plus",
	}
}

func xiaomiDefaultMappings() map[string]string {
	return map[string]string{
		"gpt-5.5":       "mimo-v2.5-pro",
		"gpt-5.4":       "mimo-v2.5-pro",
		"gpt-5.4-mini":  "mimo-v2-flash",
		"gpt-5.3-codex": "mimo-v2.5-pro",
		"gpt-4.1":       "mimo-v2-flash",
		"gpt-4o":        "mimo-v2-flash",
		"gpt-4o-mini":   "mimo-v2-flash",
		"o4-mini":       "mimo-v2-flash",
	}
}

func zhipuDefaultMappings() map[string]string {
	return map[string]string{
		"gpt-5.5":       "glm-5",
		"gpt-5.4":       "glm-5",
		"gpt-5.4-mini":  "glm-4.7-flash",
		"gpt-5.3-codex": "glm-5",
		"gpt-4.1":       "glm-4.7-flash",
		"gpt-4o":        "glm-4.7-flash",
		"gpt-4o-mini":   "glm-4.7-flash",
		"o4-mini":       "glm-4.7-flash",
	}
}

// RegisterProvider 允许外部动态注册新提供商（用于扩展）
func RegisterProvider(info *ProviderInfo) {
	if info != nil && info.ID != "" {
		registeredProviders[string(info.ID)] = info
	}
}
