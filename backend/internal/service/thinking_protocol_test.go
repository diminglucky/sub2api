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
		{"upper case Claude", "Claude-Sonnet-4-5", ThinkingProtocolAnthropicStrict},

		// 第三方兼容上游
		{"deepseek-v4-pro", "deepseek-v4-pro", ThinkingProtocolPassbackRequired},
		{"deepseek-r2-thinking", "deepseek-r2-thinking", ThinkingProtocolPassbackRequired},
		{"kimi-coding", "kimi-coding-v2", ThinkingProtocolPassbackRequired},
		{"kimi-k2-thinking", "kimi-k2-thinking", ThinkingProtocolPassbackRequired},
		{"kimi-k3 platform", "kimi-k3", ThinkingProtocolPassbackRequired},
		{"kimi code bare k3", "k3", ThinkingProtocolPassbackRequired},
		{"kimi code bare k3-256k", "k3-256k", ThinkingProtocolPassbackRequired},
		{"moonshot-v1", "moonshot-v1-32k", ThinkingProtocolPassbackRequired},
		{"glm-5.1", "glm-5.1", ThinkingProtocolPassbackRequired},
		{"qwen-2 thinking variant", "qwen-2-72b-thinking", ThinkingProtocolPassbackRequired},
		{"qwen3 thinking (real Alibaba naming)", "qwen3-235b-a22b-thinking-2507", ThinkingProtocolPassbackRequired},
		{"qwen3-next thinking", "qwen3-next-80b-a3b-thinking", ThinkingProtocolPassbackRequired},
		{"upper case Deepseek", "DeepSeek-V4-Pro", ThinkingProtocolPassbackRequired},

		// MiniMax M 系列（Anthropic 兼容端点要求 thinking round-trip）
		{"MiniMax-M2 (case-sensitive original)", "MiniMax-M2", ThinkingProtocolPassbackRequired},
		{"MiniMax-M2.1", "MiniMax-M2.1", ThinkingProtocolPassbackRequired},
		{"MiniMax-M2.5", "MiniMax-M2.5", ThinkingProtocolPassbackRequired},
		{"MiniMax-M2.7", "MiniMax-M2.7", ThinkingProtocolPassbackRequired},
		{"MiniMax-M2.7-highspeed", "MiniMax-M2.7-highspeed", ThinkingProtocolPassbackRequired},
		{"minimax-m2 lowercase", "minimax-m2", ThinkingProtocolPassbackRequired},

		// 未知 / 保守
		{"empty", "", ThinkingProtocolUnknown},
		{"gpt", "gpt-5.1", ThinkingProtocolUnknown},
		{"gemini", "gemini-3-pro-preview", ThinkingProtocolUnknown},
		{"qwen3 non-thinking", "qwen3-32b", ThinkingProtocolUnknown},
		{"qwen2 non-thinking", "qwen-2-72b", ThinkingProtocolUnknown},
		{"random vendor", "yi-large", ThinkingProtocolUnknown},
		// 相似但未知的 k3 型号：不得因含 k3 被宽泛匹配为 passback-required
		{"k3-like unknown", "foo-k3-bar", ThinkingProtocolUnknown},
		// MiniMax 非 M 系列（如 abab、speech 等其他产品线）—— unknown
		{"minimax abab non-M", "abab6.5-chat", ThinkingProtocolUnknown},
		// Doubao 走 OpenAI 协议，不属于本网关 Anthropic 路径——归 unknown
		{"doubao goes via openai", "doubao-1-5-thinking-vision-pro-250428", ThinkingProtocolUnknown},
		// Hunyuan T1 未暴露 Anthropic 端点——归 unknown
		{"hunyuan t1 no anthropic endpoint", "hunyuan-t1", ThinkingProtocolUnknown},
		{"hy-t1 short alias", "hy-t1", ThinkingProtocolUnknown},
		// claude-something 但不是 anthropic 官方命名风格——也归 strict（前缀匹配优先）
		{"weird claude prefix", "claude-experimental-fork", ThinkingProtocolAnthropicStrict},
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
