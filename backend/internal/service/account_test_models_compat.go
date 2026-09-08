package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

// FetchOpenAIAccountModels preserves the admin picker API for older callers.
func (s *AccountTestService) FetchOpenAIAccountModels(ctx context.Context, account *Account) ([]openai.Model, error) {
	return append([]openai.Model(nil), openai.DefaultModels...), nil
}
