package service

import "testing"

func TestResolveThinkingProtocol(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		want    ThinkingProtocol
	}{
		{"claude sonnet", "claude-sonnet-4-5", ThinkingProtocolAnthropicStrict},
		{"opus short", "opus-4-5", ThinkingProtocolAnthropicStrict},
		{"sonnet short", "sonnet-4-5", ThinkingProtocolAnthropicStrict},
		{"haiku short", "haiku-4-5", ThinkingProtocolAnthropicStrict},
		{"upper case claude", "Claude-Sonnet-4-5", ThinkingProtocolAnthropicStrict},
		{"deepseek", "deepseek-v4-pro", ThinkingProtocolPassbackRequired},
		{"kimi", "kimi-coding-v2", ThinkingProtocolPassbackRequired},
		{"moonshot", "moonshot-v1-32k", ThinkingProtocolPassbackRequired},
		{"glm", "glm-5.1", ThinkingProtocolPassbackRequired},
		{"qwen thinking", "qwen3-235b-a22b-thinking-2507", ThinkingProtocolPassbackRequired},
		{"minimax mixed case", "MiniMax-M2.7", ThinkingProtocolPassbackRequired},
		{"empty", "", ThinkingProtocolUnknown},
		{"gpt", "gpt-5.1", ThinkingProtocolUnknown},
		{"gemini", "gemini-3-pro-preview", ThinkingProtocolUnknown},
		{"qwen non-thinking", "qwen3-32b", ThinkingProtocolUnknown},
		{"minimax non-m", "abab6.5-chat", ThinkingProtocolUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveThinkingProtocol(tt.modelID); got != tt.want {
				t.Fatalf("ResolveThinkingProtocol(%q) = %v, want %v", tt.modelID, got, tt.want)
			}
		})
	}
}

func TestThinkingProtocolGuards(t *testing.T) {
	if !ShouldPreFilterThinkingBlocks("claude-sonnet-4-5") {
		t.Fatal("anthropic strict should pre-filter")
	}
	if ShouldPreFilterThinkingBlocks("deepseek-v4-pro") {
		t.Fatal("passback-required must not pre-filter")
	}
	if ShouldRectifyThinkingSignatureError("kimi-coding") {
		t.Fatal("passback-required must not trigger signature rectifier")
	}
	if ShouldApplyRetryFilters("gpt-5.1") {
		t.Fatal("unknown models must not apply retry filters")
	}
}
