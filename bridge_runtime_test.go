package main

import (
	"encoding/json"
	"testing"
)

func TestTranslateChatCompletionsUsesMapping(t *testing.T) {
	cfg := defaultConfig()
	cfg.Mappings["gpt-4.1"] = "deepseek-chat"

	body := []byte(`{"model":"gpt-4.1","messages":[{"role":"user","content":"hello"}],"stream":true}`)
	translated, err := translateChatCompletions(body, cfg)
	if err != nil {
		t.Fatalf("translateChatCompletions returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(translated, &payload); err != nil {
		t.Fatalf("unmarshal translated body failed: %v", err)
	}

	if payload["model"] != "deepseek-chat" {
		t.Fatalf("expected mapped model deepseek-chat, got %v", payload["model"])
	}
}

func TestUpstreamResourceURLNormalizesBasePath(t *testing.T) {
	got, err := upstreamResourceURL("https://api.deepseek.com", "chat/completions")
	if err != nil {
		t.Fatalf("upstreamResourceURL returned error: %v", err)
	}

	want := "https://api.deepseek.com/v1/chat/completions"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
