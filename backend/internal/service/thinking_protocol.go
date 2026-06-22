package service

import "strings"

// ThinkingProtocol describes how an upstream handles Anthropic thinking blocks.
// Anthropic official upstreams require valid signatures; several third-party
// Anthropic-compatible upstreams require historical thinking blocks to be passed
// back verbatim.
type ThinkingProtocol int

const (
	ThinkingProtocolUnknown ThinkingProtocol = iota
	ThinkingProtocolAnthropicStrict
	ThinkingProtocolPassbackRequired
)

// ResolveThinkingProtocol classifies the mapped upstream model ID.
// Callers must pass the model after account model mapping, not the client model.
func ResolveThinkingProtocol(modelID string) ThinkingProtocol {
	if modelID == "" {
		return ThinkingProtocolUnknown
	}
	id := strings.ToLower(modelID)

	switch {
	case strings.HasPrefix(id, "deepseek-"),
		strings.HasPrefix(id, "kimi-"),
		strings.HasPrefix(id, "moonshot-"),
		strings.HasPrefix(id, "glm-"):
		return ThinkingProtocolPassbackRequired
	}
	if strings.HasPrefix(id, "minimax-m") {
		return ThinkingProtocolPassbackRequired
	}
	if (strings.HasPrefix(id, "qwen-") ||
		strings.HasPrefix(id, "qwen2-") ||
		strings.HasPrefix(id, "qwen3-") ||
		strings.HasPrefix(id, "qwen4-")) && strings.Contains(id, "-thinking") {
		return ThinkingProtocolPassbackRequired
	}

	switch {
	case strings.HasPrefix(id, "claude-"),
		strings.HasPrefix(id, "opus-"),
		strings.HasPrefix(id, "sonnet-"),
		strings.HasPrefix(id, "haiku-"):
		return ThinkingProtocolAnthropicStrict
	}

	return ThinkingProtocolUnknown
}

func ShouldPreFilterThinkingBlocks(modelID string) bool {
	return ResolveThinkingProtocol(modelID) == ThinkingProtocolAnthropicStrict
}

func ShouldRectifyThinkingSignatureError(modelID string) bool {
	return ResolveThinkingProtocol(modelID) == ThinkingProtocolAnthropicStrict
}

func ShouldApplyRetryFilters(modelID string) bool {
	return ResolveThinkingProtocol(modelID) == ThinkingProtocolAnthropicStrict
}
