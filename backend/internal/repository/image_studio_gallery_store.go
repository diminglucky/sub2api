package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const imageStudioGalleryKeyPrefix = "image_studio_gallery:"

type imageStudioGalleryStore struct {
	rdb *redis.Client
}

func NewImageStudioGalleryStore(rdb *redis.Client) service.ImageStudioGalleryStore {
	return &imageStudioGalleryStore{rdb: rdb}
}

func (s *imageStudioGalleryStore) Save(ctx context.Context, entry *service.ImageStudioGalleryEntry, ttl time.Duration) error {
	if s == nil || s.rdb == nil || entry == nil {
		return service.ErrImageStudioGalleryUnavailable
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	pipe := s.rdb.TxPipeline()
	pipe.ZRemRangeByScore(
		ctx,
		imageStudioGalleryIndexKey(entry.UserID),
		"-inf",
		fmt.Sprintf("%d", entry.CreatedAt-int64(ttl/time.Second)),
	)
	pipe.Set(ctx, imageStudioGalleryEntryKey(entry.UserID, entry.ID), data, ttl)
	pipe.ZAdd(ctx, imageStudioGalleryIndexKey(entry.UserID), redis.Z{
		Score:  float64(entry.CreatedAt),
		Member: entry.ID,
	})
	pipe.Expire(ctx, imageStudioGalleryIndexKey(entry.UserID), ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *imageStudioGalleryStore) List(ctx context.Context, userID int64, limit int) ([]*service.ImageStudioGalleryEntry, error) {
	if s == nil || s.rdb == nil || userID <= 0 || limit <= 0 {
		return []*service.ImageStudioGalleryEntry{}, nil
	}
	indexKey := imageStudioGalleryIndexKey(userID)
	ids, err := s.rdb.ZRevRange(ctx, indexKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*service.ImageStudioGalleryEntry{}, nil
	}

	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, imageStudioGalleryEntryKey(userID, id))
	}
	values, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Unix()
	out := make([]*service.ImageStudioGalleryEntry, 0, len(values))
	staleIDs := make([]any, 0)
	for i, value := range values {
		if value == nil {
			staleIDs = append(staleIDs, ids[i])
			continue
		}
		raw, ok := value.(string)
		if !ok {
			staleIDs = append(staleIDs, ids[i])
			continue
		}
		var entry service.ImageStudioGalleryEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			staleIDs = append(staleIDs, ids[i])
			continue
		}
		if entry.ExpiresAt <= now {
			staleIDs = append(staleIDs, ids[i])
			continue
		}
		out = append(out, &entry)
	}
	if len(staleIDs) > 0 {
		_ = s.rdb.ZRem(ctx, indexKey, staleIDs...).Err()
	}
	return out, nil
}

func (s *imageStudioGalleryStore) Count(ctx context.Context, userID int64) (int64, error) {
	if s == nil || s.rdb == nil || userID <= 0 {
		return 0, nil
	}
	return s.rdb.ZCard(ctx, imageStudioGalleryIndexKey(userID)).Result()
}

func imageStudioGalleryEntryKey(userID int64, id string) string {
	return fmt.Sprintf("%s%d:%s", imageStudioGalleryKeyPrefix, userID, id)
}

func imageStudioGalleryIndexKey(userID int64) string {
	return fmt.Sprintf("%s%d:index", imageStudioGalleryKeyPrefix, userID)
}
