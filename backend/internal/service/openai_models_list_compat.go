package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *OpenAIGatewayService) validateOutboundURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" { return "", fmt.Errorf("invalid outbound URL") }
	return u.String(), nil
}

type openAIModelsRequest = codexModelsManifestRequest

func buildOpenAIAPIKeyModelsRequest(ctx context.Context, account *Account, validateBaseURL func(string) (string, error)) (*http.Request, error) {
	if account == nil || account.Type != AccountTypeAPIKey {
		return nil, fmt.Errorf("unsupported OpenAI account type")
	}
	key := strings.TrimSpace(account.GetOpenAIProtocolAPIKey())
	if key == "" {
		return nil, fmt.Errorf("OpenAI API key is empty")
	}
	base := account.GetOpenAIFormatBaseURL()
	if base == "" { base = "https://api.openai.com" }
	if validateBaseURL != nil {
		var err error
		base, err = validateBaseURL(base)
		if err != nil { return nil, err }
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/v1/models", nil)
	if err != nil { return nil, err }
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	account.ApplyHeaderOverrides(req.Header)
	return req, nil
}

func openAIModelsResponseForClient(response *OpenAIModelsResponse, ifNoneMatch string) *OpenAIModelsResponse {
	if response == nil { return nil }
	if codexModelsManifestETagMatches(ifNoneMatch, response.ETag) {
		return &OpenAIModelsResponse{ETag: response.ETag, NotModified: true}
	}
	copy := *response
	copy.Body = append([]byte(nil), response.Body...)
	return &copy
}

func (s *OpenAIGatewayService) fetchCachedOpenAIModels(ctx context.Context, request openAIModelsRequest, fetch func(context.Context, string) (*OpenAIModelsResponse, error), ifNoneMatch string) (*OpenAIModelsResponse, error) {
	key := fmt.Sprintf("%s|%d|%t", request.url, request.accountID, request.standardModelsList)
	if cached, state := s.openAIModelsCache.get(key, time.Now()); cached != nil {
		if state == openAIModelsCacheFresh || state == openAIModelsCacheStale {
			return openAIModelsResponseForClient(cached, ifNoneMatch), nil
		}
	}
	result := <-s.openAIModelsCache.refresh.DoChan(key, func() (any, error) {
		cached, _ := s.openAIModelsCache.get(key, time.Now())
		etag := ""
		if cached != nil { etag = cached.ETag }
		response, err := fetch(ctx, etag)
		if err != nil { return nil, err }
		if response != nil && !response.NotModified { s.openAIModelsCache.set(key, response, time.Now()) } else if response != nil && cached != nil { response = cached }
		return response, nil
	})
	if result.Err != nil { return nil, result.Err }
	response, ok := result.Val.(*OpenAIModelsResponse)
	if !ok || response == nil { return nil, fmt.Errorf("invalid OpenAI models cache result") }
	return openAIModelsResponseForClient(response, ifNoneMatch), nil
}

func (s *OpenAIGatewayService) fetchOpenAIModelsUpstream(ctx context.Context, request openAIModelsRequest, ifNoneMatch string) (*OpenAIModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, request.url, nil)
	if err != nil { return nil, err }
	for k, values := range request.headers { req.Header[k] = append([]string(nil), values...) }
	if ifNoneMatch != "" { req.Header.Set("If-None-Match", ifNoneMatch) }
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	body := make([]byte, 0)
	if resp.StatusCode != http.StatusNotModified {
		body, err = io.ReadAll(resp.Body)
		if err != nil { return nil, err }
	}
	return &OpenAIModelsResponse{Body: body, ETag: resp.Header.Get("ETag"), NotModified: resp.StatusCode == http.StatusNotModified}, nil
}
