package service

import "testing"

func TestOpenAIModelMappingIncludesGPT56Passthroughs(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"gpt-5.5": "gpt-5.5",
			},
		},
	}

	mapping := account.GetModelMapping()
	for _, model := range []string{"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		if got := mapping[model]; got != model {
			t.Fatalf("GetModelMapping()[%s] = %q, want %s", model, got, model)
		}
		if !account.IsModelSupported(model) {
			t.Fatalf("IsModelSupported(%s) = false, want true", model)
		}
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
