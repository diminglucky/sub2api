package service

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// opsCyberPolicyKey 在 gin context 中携带 cyber_policy 命中标记。
const opsCyberPolicyKey = "ops_cyber_policy"

// errOpenAICyberPolicyForwarded 表示 cyber_policy 已按当前端点格式透传给客户端。
var errOpenAICyberPolicyForwarded = errors.New("openai cyber_policy forwarded to client")

// CyberPolicyMark 记录一次 cyber_policy 硬阻断的上游证据。
type CyberPolicyMark struct {
	Code           string
	Message        string
	Body           string
	UpstreamStatus int
	UpstreamInTok  int
	UpstreamOutTok int
}

func MarkOpsCyberPolicy(c *gin.Context, mark CyberPolicyMark) {
	if c == nil {
		return
	}
	if GetOpsCyberPolicy(c) != nil {
		return
	}
	mark.Code = "cyber_policy"
	mark.Message = strings.TrimSpace(mark.Message)
	mark.Body = strings.TrimSpace(mark.Body)
	c.Set(opsCyberPolicyKey, &mark)
}

func GetOpsCyberPolicy(c *gin.Context) *CyberPolicyMark {
	if c == nil {
		return nil
	}
	if v, ok := c.Get(opsCyberPolicyKey); ok {
		if mark, ok := v.(*CyberPolicyMark); ok && mark != nil {
			return mark
		}
	}
	return nil
}

func ClearOpsCyberPolicy(c *gin.Context) {
	if c == nil {
		return
	}
	c.Set(opsCyberPolicyKey, (*CyberPolicyMark)(nil))
}

func detectOpenAICyberPolicy(payload []byte) (bool, string, string) {
	if len(payload) == 0 {
		return false, "", ""
	}
	code := strings.TrimSpace(gjson.GetBytes(payload, "error.code").String())
	if code == "" {
		code = strings.TrimSpace(gjson.GetBytes(payload, "response.error.code").String())
	}
	if !strings.EqualFold(code, "cyber_policy") {
		return false, "", ""
	}
	message := strings.TrimSpace(gjson.GetBytes(payload, "error.message").String())
	if message == "" {
		message = strings.TrimSpace(gjson.GetBytes(payload, "response.error.message").String())
	}
	return true, "cyber_policy", sanitizeUpstreamErrorMessage(message)
}

func isOpenAIContextWindowError(upstreamMsg string, upstreamBody []byte) bool {
	match := func(text string) bool {
		lower := strings.ToLower(strings.TrimSpace(text))
		if lower == "" {
			return false
		}
		if strings.Contains(lower, "context_too_large") || strings.Contains(lower, "context_length_exceeded") {
			return true
		}
		if strings.Contains(lower, "maximum context length") || strings.Contains(lower, "max context length") {
			return true
		}
		hasExceeded := strings.Contains(lower, "exceed") || strings.Contains(lower, "too large") || strings.Contains(lower, "too long")
		if strings.Contains(lower, "context window") && hasExceeded {
			return true
		}
		if strings.Contains(lower, "context length") && hasExceeded {
			return true
		}
		return strings.Contains(lower, "token limit") &&
			strings.Contains(lower, "context") &&
			hasExceeded
	}

	if match(upstreamMsg) {
		return true
	}
	if len(upstreamBody) == 0 {
		return false
	}
	for _, path := range []string{
		"error.message",
		"response.error.message",
		"message",
		"error.code",
		"response.error.code",
		"code",
	} {
		if match(gjson.GetBytes(upstreamBody, path).String()) {
			return true
		}
	}
	return match(string(upstreamBody))
}
