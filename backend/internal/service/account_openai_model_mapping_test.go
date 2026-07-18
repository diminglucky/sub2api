package service

import "testing"

func TestOpenAIModelMappingIncludesGPT56SolPassthrough(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.5": "gpt-5.5",
			},
		},
	}

	mapping := account.GetModelMapping()
	if got := mapping["gpt-5.6-sol"]; got != "gpt-5.6-sol" {
		t.Fatalf("GetModelMapping()[gpt-5.6-sol] = %q, want gpt-5.6-sol", got)
	}
	if !account.IsModelSupported("gpt-5.6-sol") {
		t.Fatalf("IsModelSupported(gpt-5.6-sol) = false, want true")
	}
}

func TestOpenAIModelMappingDoesNotOverrideExplicitGPT56Sol(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.6-sol": "upstream-sol",
			},
		},
	}

	if got := account.GetModelMapping()["gpt-5.6-sol"]; got != "upstream-sol" {
		t.Fatalf("GetModelMapping()[gpt-5.6-sol] = %q, want upstream-sol", got)
	}
}
