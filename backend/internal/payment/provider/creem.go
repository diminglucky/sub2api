package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const (
	creemProdAPIBase       = "https://api.creem.io/v1"
	creemTestAPIBase       = "https://test-api.creem.io/v1"
	creemHTTPTimeout       = 15 * time.Second
	creemMaxResponseSize   = 1 << 20
	creemMaxErrorSummary   = 512
	creemSignatureHeader   = "creem-signature"
	creemEventCheckoutPaid = "checkout.completed"
)

type Creem struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

func NewCreem(instanceID string, config map[string]string) (*Creem, error) {
	for _, k := range []string{"apiKey", "apiBase", "productId", "webhookSecret"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("creem config missing required key: %s", k)
		}
	}
	cfg := cloneStringMap(config)
	apiBase, err := normalizeCreemAPIBase(cfg["apiBase"])
	if err != nil {
		return nil, err
	}
	cfg["apiBase"] = apiBase
	currency, err := payment.NormalizePaymentCurrency(cfg["currency"])
	if err != nil {
		return nil, fmt.Errorf("creem config currency: %w", err)
	}
	cfg["currency"] = currency
	return &Creem{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: creemHTTPTimeout},
	}, nil
}

func normalizeCreemAPIBase(raw string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return "", fmt.Errorf("creem apiBase is required")
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("creem apiBase must be an HTTPS URL")
	}
	host := strings.ToLower(parsed.Host)
	if host != "api.creem.io" && host != "test-api.creem.io" {
		return "", fmt.Errorf("creem apiBase host must be api.creem.io or test-api.creem.io")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "" {
		parsed.Path = "/v1"
	}
	if parsed.Path != "/v1" {
		return "", fmt.Errorf("creem apiBase path must be /v1")
	}
	return parsed.String(), nil
}

func (c *Creem) Name() string        { return "Creem" }
func (c *Creem) ProviderKey() string { return payment.TypeCreem }
func (c *Creem) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeCreem}
}

func (c *Creem) MerchantIdentityMetadata() map[string]string {
	if c == nil {
		return nil
	}
	return map[string]string{
		"currency":   c.currency(),
		"product_id": strings.TrimSpace(c.config["productId"]),
		"api_base":   strings.TrimSpace(c.config["apiBase"]),
	}
}

func (c *Creem) currency() string {
	if c == nil {
		return payment.DefaultPaymentCurrency
	}
	currency, err := payment.NormalizePaymentCurrency(c.config["currency"])
	if err != nil {
		return payment.DefaultPaymentCurrency
	}
	return currency
}

func (c *Creem) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("creem create checkout: invalid amount %s", req.Amount)
	}
	currency := c.currency()
	amountInMinorUnit, err := payment.AmountToMinorUnit(req.Amount, currency)
	if err != nil {
		return nil, fmt.Errorf("creem create checkout: %w", err)
	}
	if amountInMinorUnit <= 0 {
		return nil, fmt.Errorf("creem create checkout: amount must be positive")
	}

	payload := creemCreateCheckoutRequest{
		ProductID:  strings.TrimSpace(c.config["productId"]),
		RequestID:  req.OrderID,
		Units:      amountInMinorUnit,
		SuccessURL: req.ReturnURL,
		Metadata: map[string]string{
			"order_id": req.OrderID,
			"amount":   req.Amount,
		},
	}
	if req.Subject != "" {
		payload.Metadata["subject"] = req.Subject
	}
	if req.NotifyURL != "" {
		payload.Metadata["notify_url"] = req.NotifyURL
	}

	var checkout creemCheckout
	if err := c.doJSON(ctx, http.MethodPost, "/checkouts", payload, &checkout); err != nil {
		return nil, fmt.Errorf("creem create checkout: %w", err)
	}
	checkoutID := strings.TrimSpace(checkout.ID)
	checkoutURL := strings.TrimSpace(checkout.CheckoutURL)
	if checkoutURL == "" {
		checkoutURL = strings.TrimSpace(checkout.URL)
	}
	if checkoutID == "" || checkoutURL == "" {
		return nil, fmt.Errorf("creem create checkout: missing checkout id or url")
	}
	return &payment.CreatePaymentResponse{
		TradeNo:    checkoutID,
		PayURL:     checkoutURL,
		Currency:   currency,
		PaymentEnv: c.paymentEnv(),
	}, nil
}

func (c *Creem) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	checkoutID := strings.TrimSpace(tradeNo)
	if checkoutID == "" {
		return nil, fmt.Errorf("creem query checkout: missing checkout id")
	}
	var checkout creemCheckout
	if err := c.doJSON(ctx, http.MethodGet, "/checkouts?checkout_id="+url.QueryEscape(checkoutID), nil, &checkout); err != nil {
		return nil, fmt.Errorf("creem query checkout: %w", err)
	}
	status := creemProviderStatus(checkout.Status)
	if status == payment.ProviderStatusPending && checkout.Order.Status != "" {
		status = creemProviderStatus(checkout.Order.Status)
	}
	currency := strings.ToUpper(strings.TrimSpace(checkout.Order.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(checkout.Currency))
	}
	if currency == "" {
		currency = c.currency()
	}
	return &payment.QueryOrderResponse{
		TradeNo: checkoutID,
		Status:  status,
		Amount:  payment.MinorUnitToAmount(checkout.Order.Amount, currency),
		Metadata: map[string]string{
			"currency": currency,
			"status":   strings.TrimSpace(checkout.Order.Status),
		},
	}, nil
}

func (c *Creem) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	if err := verifyCreemWebhookSignature(rawBody, headers, c.config["webhookSecret"]); err != nil {
		return nil, err
	}
	var event creemWebhookEvent
	if err := json.Unmarshal([]byte(rawBody), &event); err != nil {
		return nil, fmt.Errorf("creem parse webhook: %w", err)
	}
	if event.EventType != creemEventCheckoutPaid {
		return nil, nil
	}
	checkout := event.Object
	orderID := firstCreemString(checkout.Metadata, "order_id", "orderId")
	if strings.TrimSpace(checkout.ID) == "" || strings.TrimSpace(orderID) == "" {
		return nil, fmt.Errorf("creem webhook missing checkout id or order_id")
	}
	currency := strings.ToUpper(strings.TrimSpace(checkout.Order.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(checkout.Currency))
	}
	if currency == "" {
		currency = c.currency()
	}
	return &payment.PaymentNotification{
		TradeNo: checkout.ID,
		OrderID: orderID,
		Amount:  payment.MinorUnitToAmount(checkout.Order.Amount, currency),
		Status:  payment.NotificationStatusSuccess,
		RawData: rawBody,
		Metadata: map[string]string{
			"currency": currency,
			"event":    event.EventType,
		},
	}, nil
}

func (c *Creem) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("creem refund is not supported by this integration")
}

func (c *Creem) doJSON(ctx context.Context, method, path string, payload any, out any) error {
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.config["apiBase"]+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config["apiKey"])

	body, status, err := c.do(req)
	if err != nil {
		return err
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP %d: %s", status, summarizeCreemResponse(body))
	}
	if out == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	return nil
}

func (c *Creem) do(req *http.Request) ([]byte, int, error) {
	client := c.httpClient
	if client == nil {
		client = &http.Client{Timeout: creemHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, creemMaxResponseSize))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (c *Creem) paymentEnv() string {
	if strings.EqualFold(c.config["apiBase"], creemTestAPIBase) {
		return "test"
	}
	return "prod"
}

func verifyCreemWebhookSignature(rawBody string, headers map[string]string, secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Errorf("creem webhookSecret not configured")
	}
	signature := strings.ToLower(strings.TrimSpace(headers[creemSignatureHeader]))
	if signature == "" {
		return fmt.Errorf("creem notification missing creem-signature header")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(rawBody))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("creem invalid signature")
	}
	return nil
}

func creemProviderStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "paid", "succeeded", "success":
		return payment.ProviderStatusPaid
	case "failed", "expired", "cancelled", "canceled":
		return payment.ProviderStatusFailed
	default:
		return payment.ProviderStatusPending
	}
}

func firstCreemString(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			return value
		}
	}
	return ""
}

func summarizeCreemResponse(body []byte) string {
	summary := strings.Join(strings.Fields(string(body)), " ")
	if summary == "" {
		return "<empty>"
	}
	if len(summary) > creemMaxErrorSummary {
		return summary[:creemMaxErrorSummary] + "..."
	}
	return summary
}

type creemCreateCheckoutRequest struct {
	ProductID  string            `json:"product_id"`
	RequestID  string            `json:"request_id"`
	Units      int64             `json:"units,omitempty"`
	SuccessURL string            `json:"success_url,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type creemCheckout struct {
	ID          string            `json:"id"`
	Object      string            `json:"object"`
	Mode        string            `json:"mode"`
	CheckoutURL string            `json:"checkout_url"`
	URL         string            `json:"url"`
	Status      string            `json:"status"`
	Amount      int64             `json:"amount"`
	Currency    string            `json:"currency"`
	Metadata    map[string]string `json:"metadata"`
	Order       struct {
		ID       string `json:"id"`
		Object   string `json:"object"`
		Product  string `json:"product"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
		Status   string `json:"status"`
	} `json:"order"`
}

type creemWebhookEvent struct {
	ID        string        `json:"id"`
	EventType string        `json:"eventType"`
	Object    creemCheckout `json:"object"`
}

func (e *creemWebhookEvent) UnmarshalJSON(data []byte) error {
	type alias creemWebhookEvent
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*e = creemWebhookEvent(raw)
	return nil
}

var (
	_ payment.Provider                 = (*Creem)(nil)
	_ payment.MerchantIdentityProvider = (*Creem)(nil)
)
