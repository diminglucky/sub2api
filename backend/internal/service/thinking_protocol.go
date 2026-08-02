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

// ResolveThinkingProtocol 根据「作为 thinking block 处理参考的模型 ID」推断 thinking 协议族。
//
// 传入参数的语义随调用路径不同：
//   - **Anthropic gateway**（转发原始 Anthropic 请求）：传 mappedModel（账号级 model mapping
//     后的上游 model ID）。例：用户配置「claude-sonnet-4-6 → deepseek-v4-pro」后，
//     传 deepseek-v4-pro 才能被正确判为 passback-required。
//   - **Gemini messages compat**（Anthropic body → Gemini upstream）：传 originalModel
//     （客户端 Anthropic 请求的 model ID）。原因：此场景下上游是 Gemini，但被剥
//     离的 body 是 Anthropic 格式，需按客户端请求的 Anthropic 子协议族判定剥离行为。
//
// 匹配规则按厂商前缀硬编码：
//   - anthropic-strict: claude-* / opus-* / sonnet-* / haiku-*
//   - passback-required: deepseek-* / kimi-* / moonshot-* / glm-* /
//     minimax-* / minimax-m* / (qwen-|qwen2-|qwen3-|qwen4-)*-thinking /
//     Kimi Code bare aliases k3 / k3-256k（精确匹配，避免宽泛 k3 前缀）
//   - unknown: 其他模型（保守不剥离）
//
// 已知局限：前缀贪婪匹配（如 `claudette-`、`claude-foreign-relay-` 也会被分类为
// strict）。当遇到伪装命名时改成显式名单匹配，但现实场景几乎不会出现。
//
// 不覆盖的厂商（截至 2026-04）：
//   - Doubao / Seed (ByteDance)：走 Volcano Engine OpenAI 协议，非 Anthropic 路径
//   - Hunyuan T1 (Tencent)：未提供 Anthropic 兼容端点
//   - 若未来出现这些厂商的 Anthropic 兼容代理，需扩展前缀列表
func ResolveThinkingProtocol(modelID string) ThinkingProtocol {
	if modelID == "" {
		return ThinkingProtocolUnknown
	}
	id := strings.ToLower(modelID)

	// Passback-required 优先匹配（特定厂商前缀），避免误判 claude-* 时也命中。
	// kimi-k3* 已由 kimi- 前缀覆盖；Kimi Code endpoint 的 bare model ID（k3 / k3-256k）
	// 无厂商前缀，仅精确匹配，避免 "foo-k3" 等未知型号被宽泛前缀误判。
	switch {
	case strings.HasPrefix(id, "deepseek-"),
		strings.HasPrefix(id, "kimi-"),
		strings.HasPrefix(id, "moonshot-"),
		strings.HasPrefix(id, "glm-"),
		id == "k3",
		id == "k3-256k":
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
