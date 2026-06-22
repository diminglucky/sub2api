package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

const passbackThinkingBody = `{
	"model":"deepseek-v4-pro",
	"thinking":{"type":"enabled","budget_tokens":1024},
	"messages":[
		{"role":"user","content":[{"type":"text","text":"Hi"}]},
		{"role":"assistant","content":[
			{"type":"thinking","thinking":"Let me think..."},
			{"type":"text","text":"Answer"}
		]}
	]
}`

func TestFilterThinkingBlocks_SkipsForPassbackRequired(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterThinkingBlocks(in, "deepseek-v4-pro")
	require.True(t, bytes.Equal(in, out), "passback-required upstream must preserve thinking blocks")
}

func TestFilterThinkingBlocksForRetry_SkipsForPassbackRequired(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterThinkingBlocksForRetry(in, "kimi-coding")
	require.True(t, bytes.Equal(in, out), "passback-required retry path must preserve thinking blocks")
}

func TestFilterSignatureSensitiveBlocksForRetry_SkipsForPassbackRequired(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterSignatureSensitiveBlocksForRetry(in, "glm-5.1")
	require.True(t, bytes.Equal(in, out), "passback-required tool retry path must preserve thinking blocks")
}

func TestFilterThinkingBlocks_StripsForAnthropicStrict(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterThinkingBlocks(in, "claude-sonnet-4-5")
	require.False(t, bytes.Equal(in, out), "anthropic-strict upstream should strip invalid thinking blocks")
	require.NotContains(t, string(out), `"type":"thinking"`)
}

func TestFilterThinkingBlocks_SkipsForUnknownModel(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterThinkingBlocks(in, "yi-large")
	require.True(t, bytes.Equal(in, out), "unknown protocol must preserve body")
}
