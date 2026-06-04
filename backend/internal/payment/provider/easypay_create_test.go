package provider

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestEasyPayCreateAPIPaymentParsesAliasFields(t *testing.T) {
	t.Parallel()

	var gotForm url.Values
	provider := newTestEasyPay(t, "https://easypay.example")
	provider.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/mapi.php" {
			t.Errorf("path = %q, want /mapi.php", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		gotForm = r.PostForm
		return jsonResponse(`{"code":1,"msg":"ok","trade_no":"trade-1","urlscheme":"weixin://dl/business/?ticket=abc","qr_code":"weixin://wxpay/bizpayurl?pr=xyz"}`), nil
	})}

	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "order-1",
		Amount:      "5.00",
		PaymentType: payment.TypeWxpay,
		Subject:     "SuperAI 5.00 CNY",
		ClientIP:    "127.0.0.1",
		IsMobile:    true,
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.TradeNo != "trade-1" {
		t.Fatalf("trade_no = %q, want trade-1", resp.TradeNo)
	}
	if resp.PayURL != "weixin://dl/business/?ticket=abc" {
		t.Fatalf("pay_url = %q, want urlscheme", resp.PayURL)
	}
	if resp.QRCode != "weixin://wxpay/bizpayurl?pr=xyz" {
		t.Fatalf("qr_code = %q, want qr_code alias", resp.QRCode)
	}
	if got := gotForm.Get("sign_type"); got != signTypeMD5 {
		t.Fatalf("sign_type = %q, want MD5 (form=%v)", got, gotForm)
	}
	unsigned := map[string]string{}
	for key, values := range gotForm {
		if len(values) > 0 {
			unsigned[key] = values[0]
		}
	}
	if got, want := gotForm.Get("sign"), easyPaySign(unsigned, "pkey-1"); got != want {
		t.Fatalf("sign = %q, want %q (form=%v)", got, want, gotForm)
	}
}

func TestEasyPayCreateAPIPaymentFallsBackToSubmitWhenNoPaymentEntry(t *testing.T) {
	t.Parallel()

	provider := newTestEasyPay(t, "https://easypay.example")
	provider.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/mapi.php" {
			t.Errorf("path = %q, want /mapi.php", r.URL.Path)
		}
		return jsonResponse(`{"code":1,"msg":"ok","trade_no":"trade-1"}`), nil
	})}

	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "order-2",
		Amount:      "5.00",
		PaymentType: payment.TypeAlipay,
		Subject:     "SuperAI 5.00 CNY",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.PayURL == "" {
		t.Fatal("pay_url is empty, want submit.php fallback URL")
	}
	payURL, err := url.Parse(resp.PayURL)
	if err != nil {
		t.Fatalf("parse pay_url: %v", err)
	}
	if payURL.Path != "/submit.php" {
		t.Fatalf("fallback path = %q, want /submit.php", payURL.Path)
	}
	query := payURL.Query()
	for key, want := range map[string]string{
		"pid":          "pid-1",
		"type":         payment.TypeAlipay,
		"out_trade_no": "order-2",
		"money":        "5.00",
		"name":         "SuperAI 5.00 CNY",
		"sign_type":    signTypeMD5,
	} {
		if got := query.Get(key); got != want {
			t.Fatalf("query[%s] = %q, want %q (url=%s)", key, got, want, resp.PayURL)
		}
	}
	unsigned := map[string]string{}
	for key, values := range query {
		if len(values) > 0 {
			unsigned[key] = values[0]
		}
	}
	if got, want := query.Get("sign"), easyPaySign(unsigned, "pkey-1"); got != want {
		t.Fatalf("sign = %q, want %q (url=%s)", got, want, resp.PayURL)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
