//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiRelayHTTPUpstream struct {
	client *http.Client
}

func (u geminiRelayHTTPUpstream) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return u.client.Do(req)
}

func (u geminiRelayHTTPUpstream) DoWithTLS(
	req *http.Request,
	proxyURL string,
	accountID int64,
	accountConcurrency int,
	profile *tlsfingerprint.Profile,
) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func TestGeminiForwardNative_AcceptsSub2APIUpstreamBaseURLAndKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotPath string
	var gotAPIKey string
	var gotBody map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-goog-api-key")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotBody))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":4}}`)
	}))
	defer upstream.Close()

	cfg := &config.Config{}
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	svc := &GeminiMessagesCompatService{
		httpUpstream: geminiRelayHTTPUpstream{client: upstream.Client()},
		cfg:          cfg,
	}
	account := &Account{
		ID:       42,
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "upstream-sub2api-key",
			"base_url": upstream.URL,
		},
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1beta/models/gemini-2.5-flash-image:generateContent",
		strings.NewReader("{}"),
	)
	requestBody := []byte(`{
		"contents":[{
			"role":"user",
			"parts":[
				{"text":"turn this into a poster"},
				{"inlineData":{"mimeType":"image/png","data":"cmVmZXJlbmNl"}}
			]
		}],
		"generationConfig":{"responseModalities":["TEXT","IMAGE"],"imageConfig":{"aspectRatio":"1:1"}}
	}`)

	result, err := svc.ForwardNative(
		context.Background(),
		c,
		account,
		"gemini-2.5-flash-image",
		"generateContent",
		false,
		requestBody,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "/v1beta/models/gemini-2.5-flash-image:generateContent", gotPath)
	require.Equal(t, "upstream-sub2api-key", gotAPIKey)

	contents := gotBody["contents"].([]any)
	parts := contents[0].(map[string]any)["parts"].([]any)
	reference := parts[1].(map[string]any)["inlineData"].(map[string]any)
	require.Equal(t, "image/png", reference["mimeType"])
	require.Equal(t, "cmVmZXJlbmNl", reference["data"])

	require.Contains(t, rec.Body.String(), `"inlineData"`)
	require.Contains(t, rec.Body.String(), `"data":"aGVsbG8="`)
}
