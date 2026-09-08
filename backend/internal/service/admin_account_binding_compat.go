package service

import (
	"context"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func (s *adminServiceImpl) ValidateAccountGroupBindings(ctx context.Context, groupIDs []int64) error {
	if len(groupIDs) == 0 || s == nil || s.cfg == nil || s.cfg.RunMode != config.RunModeSimple {
		return nil
	}
	if s.groupRepo == nil {
		return fmt.Errorf("group repository not configured")
	}
	seen := make(map[int64]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		if groupID <= 0 {
			return fmt.Errorf("get group: %w", ErrGroupNotFound)
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		group, err := s.groupRepo.GetByIDLite(ctx, groupID)
		if err != nil {
			return fmt.Errorf("get group: %w", err)
		}
		if !IsGroupBindableInSimpleMode(group) {
			return infraerrors.BadRequest("SIMPLE_MODE_GROUP_NOT_BINDABLE", "composite groups cannot be bound in simple mode")
		}
	}
	return nil
}
