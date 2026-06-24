package service

import (
	"strings"

	"github.com/tidwall/gjson"
)

func detectOpenAICyberPolicy(payload []byte) (bool, string) {
	if len(payload) == 0 {
		return false, ""
	}
	code := strings.TrimSpace(gjson.GetBytes(payload, "error.code").String())
	if code == "" {
		code = strings.TrimSpace(gjson.GetBytes(payload, "response.error.code").String())
	}
	if !strings.EqualFold(code, "cyber_policy") {
		return false, ""
	}
	message := strings.TrimSpace(gjson.GetBytes(payload, "error.message").String())
	if message == "" {
		message = strings.TrimSpace(gjson.GetBytes(payload, "response.error.message").String())
	}
	return true, sanitizeUpstreamErrorMessage(message)
}
