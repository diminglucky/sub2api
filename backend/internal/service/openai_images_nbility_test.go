package service

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeNbilityDocumentedTaskResults(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"documented URLs", `{"code":"success","data":{"status":"succeeded","result_url":"https://cdn.example/a.gif","image_urls":["https://cdn.example/a.gif","https://cdn.example/b.png"]}}`, `{"data":[{"url":"https://cdn.example/a.gif","output_format":"gif"},{"url":"https://cdn.example/b.png","output_format":"png"}]}`},
		{"result URL only", `{"data":{"status":"succeeded","result_url":"https://cdn.example/a.gif"}}`, `{"data":[{"url":"https://cdn.example/a.gif","output_format":"gif"}]}`},
		{"unwrapped URLs", `{"status":"succeeded","image_urls":["https://cdn.example/a.gif"]}`, `{"data":[{"url":"https://cdn.example/a.gif","output_format":"gif"}]}`},
		{"structured result preferred", `{"data":{"status":"succeeded","result":{"data":[{"url":"https://cdn.example/a.gif","output_format":"gif","width":512,"height":512}],"usage":{"total_tokens":12}},"image_urls":["https://cdn.example/sheet.png"]}}`, `{"data":[{"url":"https://cdn.example/a.gif","output_format":"gif","width":512,"height":512}],"usage":{"total_tokens":12}}`},
		{"data URL result", `{"data":{"status":"succeeded","result":"data:image/gif;base64,R0lGODlh"}}`, `{"data":[{"b64_json":"R0lGODlh","output_format":"gif"}]}`},
		{"single PNG contact sheet", `{"data":{"result":{"url":"https://cdn.example/sheet.png","output_format":"png"}}}`, `{"data":[{"url":"https://cdn.example/sheet.png","output_format":"png"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.JSONEq(t, tc.want, string(normalizeNbilityImageTaskResponse([]byte(tc.body), "gif")))
			require.True(t, nbilityImageTaskBodyHasResult([]byte(tc.body)))
		})
	}
}

func TestNormalizeNbilityImageBinaryResponse(t *testing.T) {
	gifBytes, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
	require.NoError(t, err)
	body := normalizeNbilityImageBinaryResponse(gifBytes, "image/gif", "gif")
	require.JSONEq(t, `{"data":[{"b64_json":"R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7","output_format":"gif"}]}`, string(body))
	require.Empty(t, normalizeNbilityImageBinaryResponse([]byte("upstream error"), "text/html", "gif"))
}

func TestShouldUseNbilityImageAsync(t *testing.T) {
	for _, tc := range []struct {
		host, model, format string
		stream, want        bool
	}{
		{"api.nbility.ai", "gpt-image-2-6", "gif", false, true},
		{"nbility.ai", "gpt-image-2-6-flare", "gif", false, true},
		{"api.nbility.ai", "gpt-image-2-6-sunburst", "GIF", false, true},
		{"api.nbility.ai", "gpt-image-2.6-flare", "gif", false, true},
		{"api.openai.com", "gpt-image-2-6", "gif", false, false},
		{"api.nbility.ai.example.org", "gpt-image-2-6", "gif", false, false},
		{"api.nbility.ai", "gpt-image-2-5", "gif", false, false},
		{"api.nbility.ai", "gpt-image-2-6", "png", false, false},
		{"api.nbility.ai", "gpt-image-2-6", "gif", true, false},
	} {
		t.Run(tc.host+"/"+tc.model+"/"+tc.format, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "https://"+tc.host+"/v1/images/generations", nil)
			require.NoError(t, err)
			require.Equal(t, tc.want, shouldUseNbilityImageAsync(newNbilityImagesAPIKeyAccount(), req, tc.model, &OpenAIImagesRequest{OutputFormat: tc.format, Stream: tc.stream}))
		})
	}
}

func TestForwardOpenAIImagesAPIKey_NbilityDocumentedGIFResponse(t *testing.T) {
	for _, model := range []string{"gpt-image-2-6", "gpt-image-2-6-flare", "gpt-image-2-6-sunburst"} {
		t.Run(model, func(t *testing.T) {
			body := []byte(`{"model":"` + model + `","prompt":"animated cat","output_format":"gif"}`)
			c, rec := newOpenAIImagesTestContext(t, body)
			upstream := &httpUpstreamRecorder{responses: []*http.Response{
				{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"code":"success","data":{"task_id":"task_gif_1","status":"submitted"}}`))},
				{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"code":"success","data":{"status":"succeeded","result_url":"https://cdn.example/a.gif","image_urls":["https://cdn.example/a.gif"]}}`))},
			}}
			svc := newOpenAIImagesTestService(upstream)
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			result, err := svc.ForwardImages(context.Background(), c, newNbilityImagesAPIKeyAccount(), body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, 1, result.ImageCount)
			require.Len(t, upstream.requests, 2)
			require.Equal(t, "true", upstream.requests[0].Header.Get("X-New-Api-Async-Task"))
			require.Equal(t, "https://cdn.example/a.gif", gjson.GetBytes(rec.Body.Bytes(), "data.0.url").String())
		})
	}
}

func TestForwardOpenAIImagesAPIKey_NbilityBinaryGIFResponse(t *testing.T) {
	gifBytes, err := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
	require.NoError(t, err)
	body := []byte(`{"model":"gpt-image-2-6","prompt":"animated cat","output_format":"gif"}`)
	c, rec := newOpenAIImagesTestContext(t, body)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"task_id":"task_gif_binary"}}`))},
		{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"image/gif"}}, Body: io.NopCloser(strings.NewReader(string(gifBytes)))},
	}}
	parsed, err := newOpenAIImagesTestService(upstream).ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	result, err := newOpenAIImagesTestService(upstream).ForwardImages(context.Background(), c, newNbilityImagesAPIKeyAccount(), body, parsed, "")
	require.NoError(t, err)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7", gjson.GetBytes(rec.Body.Bytes(), "data.0.b64_json").String())
}

func TestForwardOpenAIImagesAPIKey_NbilityLargeGIFResult(t *testing.T) {
	encoded := "R0lGODlh" + strings.Repeat("A", 2<<20)
	body := []byte(`{"model":"gpt-image-2-6","prompt":"animated cat","output_format":"gif"}`)
	c, rec := newOpenAIImagesTestContext(t, body)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"task_id":"task_large"}}`))},
		{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"status":"succeeded","result":{"data":[{"b64_json":"` + encoded + `","output_format":"gif"}]}}}`))},
	}}
	svc := newOpenAIImagesTestService(upstream)
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	result, err := svc.ForwardImages(context.Background(), c, newNbilityImagesAPIKeyAccount(), body, parsed, "")
	require.NoError(t, err)
	require.Equal(t, 1, result.ImageCount)
	require.Equal(t, encoded, gjson.GetBytes(rec.Body.Bytes(), "data.0.b64_json").String())
}

func TestForwardOpenAIImagesAPIKey_NbilityTaskResponseLimit(t *testing.T) {
	body := []byte(`{"model":"gpt-image-2-6","prompt":"animated cat","output_format":"gif"}`)
	c, _ := newOpenAIImagesTestContext(t, body)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"task_id":"task_large"}}`))},
		{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":{"status":"succeeded","result_url":"https://cdn.example/` + strings.Repeat("a", 2048) + `.gif"}}`))},
	}}
	svc := newOpenAIImagesTestService(upstream)
	svc.cfg.Gateway.UpstreamResponseReadMaxBytes = 1024
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	_, err = svc.ForwardImages(context.Background(), c, newNbilityImagesAPIKeyAccount(), body, parsed, "")
	require.ErrorContains(t, err, "configured response size limit")
	require.Len(t, upstream.requests, 2, "an oversized result must fail explicitly without repeatedly polling it")
}
