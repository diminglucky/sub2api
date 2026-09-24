package service

import (
	"context"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

const (
	ImageStudioGalleryTTL        = 7 * 24 * time.Hour
	imageStudioMaxItems          = 100
	imageStudioMaxEntriesPerUser = 100
)

var ErrImageStudioGalleryUnavailable = infraerrors.New(
	http.StatusServiceUnavailable,
	"IMAGE_STUDIO_GALLERY_UNAVAILABLE",
	"image studio gallery storage is unavailable",
)

var ErrImageStudioGalleryQuotaExceeded = infraerrors.New(
	http.StatusTooManyRequests,
	"IMAGE_STUDIO_GALLERY_QUOTA_EXCEEDED",
	"image studio gallery item limit reached",
)

type ImageStudioGalleryEntry struct {
	ID        string `json:"id"`
	UserID    int64  `json:"-"`
	URL       string `json:"url"`
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	Size      string `json:"size"`
	Format    string `json:"format"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

type ImageStudioGalleryStore interface {
	Save(ctx context.Context, entry *ImageStudioGalleryEntry, ttl time.Duration) error
	List(ctx context.Context, userID int64, limit int) ([]*ImageStudioGalleryEntry, error)
	Count(ctx context.Context, userID int64) (int64, error)
}

type ImageStudioImageSaver interface {
	SaveImage(ctx context.Context, key, contentType string, data []byte) (string, error)
	GalleryRetention() time.Duration
}

type ImageStudioGalleryService struct {
	store   ImageStudioGalleryStore
	storage ImageStudioImageSaver
}

func NewImageStudioGalleryService(store ImageStudioGalleryStore, storage ImageStudioImageSaver) *ImageStudioGalleryService {
	return &ImageStudioGalleryService{store: store, storage: storage}
}

func (s *ImageStudioGalleryService) Save(
	ctx context.Context,
	userID int64,
	prompt string,
	model string,
	size string,
	format string,
	contentType string,
	data []byte,
) (*ImageStudioGalleryEntry, error) {
	if s == nil || s.store == nil || s.storage == nil || userID <= 0 || len(data) == 0 {
		return nil, ErrImageStudioGalleryUnavailable
	}

	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if !strings.HasPrefix(contentType, "image/") {
		contentType = detectImageContentType(data)
	}
	retention := s.storage.GalleryRetention()
	if retention <= 0 {
		retention = ImageStudioGalleryTTL
	}
	count, err := s.store.Count(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= imageStudioMaxEntriesPerUser {
		return nil, ErrImageStudioGalleryQuotaExceeded
	}

	key := path.Join(strings.TrimSpace(formatUserID(userID)), id+extensionForContentType(contentType))
	url, err := s.storage.SaveImage(ctx, key, contentType, data)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	entry := &ImageStudioGalleryEntry{
		ID:        id,
		UserID:    userID,
		URL:       url,
		Prompt:    strings.TrimSpace(prompt),
		Model:     strings.TrimSpace(model),
		Size:      strings.TrimSpace(size),
		Format:    strings.TrimSpace(format),
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(retention).Unix(),
	}
	if err := s.store.Save(ctx, entry, retention); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *ImageStudioGalleryService) List(ctx context.Context, userID int64, limit int) ([]*ImageStudioGalleryEntry, error) {
	if s == nil || s.store == nil || userID <= 0 {
		return []*ImageStudioGalleryEntry{}, nil
	}
	if limit <= 0 || limit > imageStudioMaxItems {
		limit = imageStudioMaxItems
	}
	entries, err := s.store.List(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []*ImageStudioGalleryEntry{}
	}
	return entries, nil
}

func formatUserID(userID int64) string {
	return strconv.FormatInt(userID, 10)
}
