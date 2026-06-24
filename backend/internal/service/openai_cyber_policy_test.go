package service

import "testing"

func TestDetectOpenAICyberPolicy(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantMsg string
	}{
		{
			name:    "plain error envelope",
			payload: `{"error":{"code":"cyber_policy","message":"blocked by policy"}}`,
			wantMsg: "blocked by policy",
		},
		{
			name:    "response failed envelope",
			payload: `{"type":"response.failed","response":{"error":{"code":"cyber_policy","message":"high-risk cyber activity"}}}`,
			wantMsg: "high-risk cyber activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, msg := detectOpenAICyberPolicy([]byte(tt.payload))
			if !got {
				t.Fatal("expected cyber_policy to be detected")
			}
			if msg != tt.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestOpenAIStreamFailedEventShouldNotFailoverCyberPolicy(t *testing.T) {
	payload := []byte(`{"type":"response.failed","response":{"error":{"code":"cyber_policy","message":"blocked"}}}`)
	if openAIStreamFailedEventShouldFailover(payload, "blocked") {
		t.Fatal("cyber_policy must be passed through, not failed over")
	}
}
