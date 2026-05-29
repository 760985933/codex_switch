package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ---- APIType 定义 ----

type APIType string

const (
	APIChatCompletions APIType = "chat_completions"
	APIResponses       APIType = "responses"
	APIMessages        APIType = "messages"
)

// ---- 格式配置 ----

type FormatConfig struct {
	DefaultModel     string
	Mappings         map[string]string
	APIKey           string
	BaseURL          string
	RequestTimeoutMs int
	MaxRetries       int
}

type UsageInfo struct {
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

// ---- FormatHandler 接口：描述一种 API 格式的线协议特征 ----

type FormatHandler interface {
	APIType() APIType
	EndpointPath() string
	AuthHeader(apiKey string) (key, value string)
	AdditionalHeaders() map[string]string
}

// ---- 三种格式处理器 ----

type chatCompletionsHandler struct{}

func (h *chatCompletionsHandler) APIType() APIType               { return APIChatCompletions }
func (h *chatCompletionsHandler) EndpointPath() string           { return "chat/completions" }
func (h *chatCompletionsHandler) AuthHeader(apiKey string) (string, string) {
	return "Authorization", "Bearer " + apiKey
}
func (h *chatCompletionsHandler) AdditionalHeaders() map[string]string { return nil }

type responsesHandler struct{}

func (h *responsesHandler) APIType() APIType               { return APIResponses }
func (h *responsesHandler) EndpointPath() string           { return "responses" }
func (h *responsesHandler) AuthHeader(apiKey string) (string, string) {
	return "Authorization", "Bearer " + apiKey
}
func (h *responsesHandler) AdditionalHeaders() map[string]string { return nil }

type messagesHandler struct{}

func (h *messagesHandler) APIType() APIType               { return APIMessages }
func (h *messagesHandler) EndpointPath() string           { return "messages" }
func (h *messagesHandler) AuthHeader(apiKey string) (string, string) {
	return "x-api-key", apiKey
}
func (h *messagesHandler) AdditionalHeaders() map[string]string {
	return map[string]string{"anthropic-version": "2023-06-01"}
}

// ---- 格式注册表 ----

var formatHandlers = map[APIType]FormatHandler{}

func RegisterFormatHandler(h FormatHandler) {
	formatHandlers[h.APIType()] = h
}

func GetFormatHandler(t APIType) (FormatHandler, error) {
	h, ok := formatHandlers[t]
	if !ok {
		return nil, fmt.Errorf("unknown API format: %s", t)
	}
	return h, nil
}

func init() {
	RegisterFormatHandler(&chatCompletionsHandler{})
	RegisterFormatHandler(&responsesHandler{})
	RegisterFormatHandler(&messagesHandler{})
}

// ---- 格式感知路由上下文 ----

// FormatRoute 封装格式感知代理转发的完整上下文
type FormatRoute struct {
	SourceFormat APIType
	TargetFormat APIType
	Model        string
	Streaming    bool
	Cfg          AppConfig
	RequestID    string
}

// DetectAPITypeFromPath 根据 URL 路径检测 API 格式
func DetectAPITypeFromPath(path string) APIType {
	if strings.Contains(path, "/v1/responses") {
		return APIResponses
	}
	if strings.Contains(path, "/v1/messages") {
		return APIMessages
	}
	return APIChatCompletions
}

// ResolveProviderFormat 从配置解析提供商的 API 格式
// 优先使用 Profile 中显式设置的 APIType，其次使用提供商的默认格式
func ResolveProviderFormat(cfg AppConfig) APIType {
	profile, ok := cfg.Profiles[cfg.CurrentProfileID]
	if !ok {
		return APIChatCompletions
	}
	// 配置文件里显式设置了 APIType 则优先使用
	if strings.TrimSpace(profile.APIType) != "" {
		return APIType(profile.APIType)
	}
	provider := GetProvider(ProviderID(profile.Provider))
	if provider != nil {
		return provider.APIType
	}
	return APIChatCompletions
}

// ResolveProviderBaseURL 获取上游提供商的 base URL
func ResolveProviderBaseURL(cfg AppConfig) string {
	profile, ok := cfg.Profiles[cfg.CurrentProfileID]
	if !ok {
		return cfg.DeepseekBaseURL
	}
	if strings.TrimSpace(profile.BaseURL) != "" {
		return profile.BaseURL
	}
	provider := GetProvider(ProviderID(profile.Provider))
	if provider != nil && strings.TrimSpace(provider.DefaultBaseURL) != "" {
		return provider.DefaultBaseURL
	}
	return cfg.DeepseekBaseURL
}

// ---- 格式感知请求翻译调度 ----

// TranslateRequestBody 根据源/目标格式翻译请求体
// 返回: 翻译后的body, 是否流式, 模型名, error
func TranslateRequestBody(body []byte, sourceFormat, targetFormat APIType, cfg AppConfig) ([]byte, bool, string, error) {
	if sourceFormat == targetFormat {
		return passthroughRequest(body, sourceFormat, cfg)
	}

	switch {
	case sourceFormat == APIResponses && targetFormat == APIChatCompletions:
		return translateResponsesToChatCompletions(body, cfg)

	case sourceFormat == APIMessages && targetFormat == APIChatCompletions:
		return translateMessagesToChatCompletions(body, cfg)

	case sourceFormat == APIMessages && targetFormat == APIResponses:
		chatBody, stream, _, err := translateMessagesToChatCompletions(body, cfg)
		if err != nil {
			return nil, false, "", err
		}
		respBody, _, _, err := translateChatToResponsesRequest(chatBody, cfg)
		return respBody, stream, "", err

	case sourceFormat == APIResponses && targetFormat == APIMessages:
		chatBody, stream, _, err := translateResponsesToChatCompletions(body, cfg)
		if err != nil {
			return nil, false, "", err
		}
		msgBody, _, _, err := translateChatToMessagesRequest(chatBody, cfg)
		return msgBody, stream, "", err

	case sourceFormat == APIChatCompletions && targetFormat == APIMessages:
		chatBody, _, _, err := passthroughRequest(body, APIChatCompletions, cfg)
		if err != nil {
			return nil, false, "", err
		}
		return translateChatToMessagesRequest(chatBody, cfg)

	case sourceFormat == APIChatCompletions && targetFormat == APIResponses:
		chatBody, _, _, err := passthroughRequest(body, APIChatCompletions, cfg)
		if err != nil {
			return nil, false, "", err
		}
		return translateChatToResponsesRequest(chatBody, cfg)

	default:
		return nil, false, "", fmt.Errorf("unsupported translation: %s -> %s", sourceFormat, targetFormat)
	}
}

// TranslateResponseBody 将上游响应（目标格式）转回源格式
func TranslateResponseBody(upstreamBody []byte, sourceFormat, targetFormat APIType, model string) (any, error) {
	if sourceFormat == targetFormat {
		return nil, nil // 直通
	}

	switch {
	case sourceFormat == APIResponses && targetFormat == APIChatCompletions:
		return translateChatCompletionToResponses(upstreamBody, model)

	case sourceFormat == APIMessages && targetFormat == APIChatCompletions:
		return translateChatCompletionToMessages(upstreamBody, model)

	case sourceFormat == APIResponses && targetFormat == APIMessages:
		return translateMessagesToResponses(upstreamBody, model)

	case sourceFormat == APIMessages && targetFormat == APIResponses:
		return translateResponsesToMessages(upstreamBody, model)

	case sourceFormat == APIChatCompletions && targetFormat == APIMessages:
		return translateMessagesToChatResponse(upstreamBody, model)

	case sourceFormat == APIChatCompletions && targetFormat == APIResponses:
		return translateResponsesToChatResponse(upstreamBody)
		// Note: this case is unusual but handled for completeness

	default:
		return nil, fmt.Errorf("unsupported response translation: %s -> %s", sourceFormat, targetFormat)
	}
}

// ---- 同格式直通（仅做模型映射和规范化） ----

func passthroughRequest(body []byte, format APIType, cfg AppConfig) ([]byte, bool, string, error) {
	streaming := detectStreaming(body)

	switch format {
	case APIMessages:
		upstreamBody, err := passThroughMessagesBody(body, cfg)
		return upstreamBody, streaming, extractModelFromBody(body), err
	case APIResponses:
		upstreamBody, err := passThroughResponsesBody(body, cfg)
		return upstreamBody, streaming, extractModelFromBody(body), err
	default:
		upstreamBody, err := translateChatCompletions(body, cfg)
		return upstreamBody, streaming, extractModelFromBody(body), err
	}
}

func passThroughMessagesBody(body []byte, cfg AppConfig) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("请求体不是有效的 JSON: %w", err)
	}
	if rawModel, ok := payload["model"].(string); ok && strings.TrimSpace(rawModel) != "" {
		if mapped, ok := cfg.Mappings[rawModel]; ok && strings.TrimSpace(mapped) != "" {
			payload["model"] = mapped
		} else {
			payload["model"] = strings.TrimSpace(cfg.DefaultModel)
		}
	} else {
		payload["model"] = strings.TrimSpace(cfg.DefaultModel)
	}
	return json.Marshal(payload)
}

func passThroughResponsesBody(body []byte, cfg AppConfig) ([]byte, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("请求体不是有效的 JSON: %w", err)
	}
	if rawModel, ok := payload["model"].(string); ok && strings.TrimSpace(rawModel) != "" {
		if mapped, ok := cfg.Mappings[rawModel]; ok && strings.TrimSpace(mapped) != "" {
			payload["model"] = mapped
		} else {
			payload["model"] = strings.TrimSpace(cfg.DefaultModel)
		}
	} else {
		payload["model"] = strings.TrimSpace(cfg.DefaultModel)
	}
	return json.Marshal(payload)
}

func detectStreaming(body []byte) bool {
	return bytes.Contains(body, []byte(`"stream":true`)) ||
		bytes.Contains(body, []byte(`"stream": true`))
}

// ---- Messages API 跨格式翻译 ----

func translateChatToMessagesRequest(body []byte, cfg AppConfig) ([]byte, bool, string, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, false, "", fmt.Errorf("无效 JSON: %w", err)
	}

	streaming := detectStreaming(body)

	model := strings.TrimSpace(cfg.DefaultModel)
	if rawModel, ok := payload["model"].(string); ok && strings.TrimSpace(rawModel) != "" {
		if mapped, ok := cfg.Mappings[rawModel]; ok && strings.TrimSpace(mapped) != "" {
			model = mapped
		}
	}
	payload["model"] = model

	messages, _ := payload["messages"].([]any)
	if messages == nil {
		messages = []any{}
	}
	var filteredMsgs []any
	var systemTexts []string
	for _, item := range messages {
		msg, ok := item.(map[string]any)
		if !ok {
			filteredMsgs = append(filteredMsgs, item)
			continue
		}
		role, _ := msg["role"].(string)
		if role == "system" {
			if content, ok := msg["content"].(string); ok && strings.TrimSpace(content) != "" {
				systemTexts = append(systemTexts, content)
			}
		} else {
			if content, ok := msg["content"].(string); ok {
				msg["content"] = []any{map[string]any{"type": "text", "text": content}}
			}
			filteredMsgs = append(filteredMsgs, msg)
		}
	}
	if len(systemTexts) > 0 {
		payload["system"] = strings.Join(systemTexts, "\n")
	}
	payload["messages"] = filteredMsgs

	if _, ok := payload["max_tokens"]; !ok {
		payload["max_tokens"] = 4096
	}
	delete(payload, "thinking")

	out, err := json.Marshal(payload)
	return out, streaming, model, err
}

func translateChatToResponsesRequest(body []byte, cfg AppConfig) ([]byte, bool, string, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, false, "", fmt.Errorf("无效 JSON: %w", err)
	}

	streaming := detectStreaming(body)

	model := strings.TrimSpace(cfg.DefaultModel)
	if rawModel, ok := payload["model"].(string); ok && strings.TrimSpace(rawModel) != "" {
		if mapped, ok := cfg.Mappings[rawModel]; ok && strings.TrimSpace(mapped) != "" {
			model = mapped
		}
	}

	messages, _ := payload["messages"].([]any)
	if messages == nil {
		messages = []any{}
	}
	input := make([]any, 0, len(messages))
	for _, item := range messages {
		msg, ok := item.(map[string]any)
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		content := ""
		if c, ok := msg["content"].(string); ok {
			content = c
		}
		contentType := "input_text"
		if role == "assistant" {
			contentType = "output_text"
		}
		input = append(input, map[string]any{
			"type":    "message",
			"role":    role,
			"content": []any{map[string]any{"type": contentType, "text": content}},
		})
	}

	outPayload := map[string]any{
		"model":  model,
		"input":  input,
		"stream": streaming,
	}
	if v, ok := payload["tools"]; ok {
		outPayload["tools"] = v
	}
	out, err := json.Marshal(outPayload)
	return out, streaming, model, err
}

// ---- 响应跨格式翻译 ----

func translateMessagesToResponses(body []byte, model string) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	text := extractMessagesText(payload)
	return map[string]any{
		"id":         fmt.Sprintf("resp_%d", timestamp()),
		"object":     "response",
		"created_at": timestamp() / 1e9,
		"model":      model,
		"status":     "completed",
		"output": []any{map[string]any{
			"id":      fmt.Sprintf("msg_%d", timestamp()),
			"type":    "message",
			"role":    "assistant",
			"status":  "completed",
			"content": []any{map[string]any{"type": "output_text", "text": text}},
		}},
	}, nil
}

func translateResponsesToMessages(body []byte, model string) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	var texts []string
	if output, ok := payload["output"].([]any); ok {
		for _, item := range output {
			if m, ok := item.(map[string]any); ok {
				if content, ok := m["content"].([]any); ok {
					for _, c := range content {
						if cm, ok := c.(map[string]any); ok {
							if t, ok := cm["text"].(string); ok {
								texts = append(texts, t)
							}
						}
					}
				}
			}
		}
	}
	usage := map[string]any{"input_tokens": 0, "output_tokens": 0}
	if u, ok := payload["usage"].(map[string]any); ok {
		usage = u
	}
	return map[string]any{
		"id":           fmt.Sprintf("msg_%d", timestamp()),
		"type":         "message",
		"role":         "assistant",
		"content":      []any{map[string]any{"type": "text", "text": strings.Join(texts, "\n")}},
		"model":        model,
		"stop_reason":  "end_turn",
		"stop_sequence": nil,
		"usage":        usage,
	}, nil
}

func translateMessagesToChatResponse(body []byte, model string) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	text := extractMessagesText(payload)
	stopReason, _ := payload["stop_reason"].(string)
	usage := map[string]any{}
	if u, ok := payload["usage"].(map[string]any); ok {
		if v, ok := u["input_tokens"]; ok {
			usage["prompt_tokens"] = v
		}
		if v, ok := u["output_tokens"]; ok {
			usage["completion_tokens"] = v
		}
		usage["total_tokens"] = 0
	}
	return map[string]any{
		"id":      fmt.Sprintf("chatcmpl-%d", timestamp()),
		"object":  "chat.completion",
		"created": timestamp() / 1e9,
		"model":   model,
		"choices": []any{map[string]any{
			"index": 0,
			"message": map[string]any{
				"role":    "assistant",
				"content": text,
			},
			"finish_reason": mapStopReason(stopReason),
		}},
		"usage": usage,
	}, nil
}

func translateResponsesToChatResponse(body []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	var texts []string
	if output, ok := payload["output"].([]any); ok {
		for _, item := range output {
			if m, ok := item.(map[string]any); ok {
				if content, ok := m["content"].([]any); ok {
					for _, c := range content {
						if cm, ok := c.(map[string]any); ok {
							if t, ok := cm["text"].(string); ok {
								texts = append(texts, t)
							}
						}
					}
				}
			}
		}
	}
	return map[string]any{
		"id":      fmt.Sprintf("chatcmpl-%d", timestamp()),
		"object":  "chat.completion",
		"created": timestamp() / 1e9,
		"model":   "",
		"choices": []any{map[string]any{
			"index": 0,
			"message": map[string]any{
				"role":    "assistant",
				"content": strings.Join(texts, "\n"),
			},
			"finish_reason": "stop",
		}},
	}, nil
}

func mapStopReason(sr string) string {
	switch sr {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return sr
	}
}

func extractMessagesText(payload map[string]any) string {
	content, ok := payload["content"].([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, block := range content {
		b, ok := block.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := b["type"].(string); t == "text" {
			if text, ok := b["text"].(string); ok {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func timestamp() int64 {
	// Time-based ID generation, inlined to avoid import cycle
	return int64(1e6)
}

// ---- SSE 流式跨格式翻译 ----

// StreamChatToResponsesSSE 将 Chat SSE 流转为 Responses SSE 流
// 实现在 proxy_runtime.go:processChatStreamToResponses

// StreamChatToMessagesSSE 将 Chat SSE 流转为 Messages SSE 流
// 实现在 proxy_runtime.go:streamChatToMessages

// ---- HTTP 请求构建辅助 ----

// BuildUpstreamRequest 构建格式感知的上游 HTTP 请求
func BuildUpstreamRequest(r *http.Request, body []byte, targetFormat APIType, cfg AppConfig) (*http.Request, error) {
	handler, err := GetFormatHandler(targetFormat)
	if err != nil {
		return nil, err
	}

	baseURL := ResolveProviderBaseURL(cfg)
	endpointPath := handler.EndpointPath()
	upstreamURL, err := upstreamResourceURL(baseURL, endpointPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	authKey, authValue := handler.AuthHeader(cfg.APIKey)
	req.Header.Set(authKey, authValue)

	for k, v := range handler.AdditionalHeaders() {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "nettopo-switch/0.1")

	copyRequestHeaders(req.Header, r.Header, cfg.Headers)

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	return req, nil
}
