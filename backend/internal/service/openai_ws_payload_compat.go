package service

import "encoding/json"

func cloneOpenAIWSRawMessages(items []json.RawMessage) []json.RawMessage {
	if items == nil { return nil }
	out := make([]json.RawMessage, len(items))
	for i, item := range items { out[i] = append(json.RawMessage(nil), item...) }
	return out
}

func cloneOpenAIWSPayloadBytes(payload []byte) []byte { return append([]byte(nil), payload...) }
