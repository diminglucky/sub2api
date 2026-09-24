//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type galleryStoreStub struct {
	entries []*ImageStudioGalleryEntry
	ttl     time.Duration
}

func (s *galleryStoreStub) Save(_ context.Context, entry *ImageStudioGalleryEntry, ttl time.Duration) error {
	s.entries = append(s.entries, entry)
	s.ttl = ttl
	return nil
}

func (s *galleryStoreStub) List(context.Context, int64, int) ([]*ImageStudioGalleryEntry, error) {
	return s.entries, nil
}

func (s *galleryStoreStub) Count(_ context.Context, _ int64) (int64, error) {
	return int64(len(s.entries)), nil
}

type gallerySaverStub struct {
	key         string
	contentType string
	data        []byte
	retention   time.Duration
}

func (s *gallerySaverStub) SaveImage(_ context.Context, key, contentType string, data []byte) (string, error) {
	s.key = key
	s.contentType = contentType
	s.data = append([]byte(nil), data...)
	return "https://cdn.example.com/" + key, nil
}

func (s *gallerySaverStub) GalleryRetention() time.Duration {
	if s.retention <= 0 {
		return ImageStudioGalleryTTL
	}
	return s.retention
}

func TestImageStudioGallerySaveUsesSevenDayRetention(t *testing.T) {
	store := &galleryStoreStub{}
	saver := &gallerySaverStub{}
	svc := NewImageStudioGalleryService(store, saver)

	entry, err := svc.Save(
		context.Background(),
		42,
		"a cat",
		"gpt-image-1",
		"1024x1024",
		"png",
		"image/png",
		smallPNG(t),
	)
	require.NoError(t, err)
	require.Equal(t, int64(42), entry.UserID)
	require.Equal(t, "42/"+entry.ID+".png", saver.key)
	require.Equal(t, "image/png", saver.contentType)
	require.Len(t, saver.data, len(smallPNG(t)))
	require.Equal(t, ImageStudioGalleryTTL, store.ttl)
	require.Equal(t, entry.ExpiresAt-entry.CreatedAt, int64(ImageStudioGalleryTTL/time.Second))
}

func TestImageStudioGalleryUsesConfiguredRetention(t *testing.T) {
	store := &galleryStoreStub{}
	saver := &gallerySaverStub{retention: 48 * time.Hour}
	svc := NewImageStudioGalleryService(store, saver)

	entry, err := svc.Save(
		context.Background(),
		42,
		"a cat",
		"gpt-image-1",
		"1024x1024",
		"png",
		"image/png",
		smallPNG(t),
	)
	require.NoError(t, err)
	require.Equal(t, 48*time.Hour, store.ttl)
	require.Equal(t, entry.ExpiresAt-entry.CreatedAt, int64((48*time.Hour)/time.Second))
}

func TestImageStudioGalleryRejectsWhenUserLimitReached(t *testing.T) {
	entries := make([]*ImageStudioGalleryEntry, imageStudioMaxEntriesPerUser)
	for i := range entries {
		entries[i] = &ImageStudioGalleryEntry{}
	}
	store := &galleryStoreStub{entries: entries}
	saver := &gallerySaverStub{}
	svc := NewImageStudioGalleryService(store, saver)

	_, err := svc.Save(
		context.Background(),
		42,
		"a cat",
		"gpt-image-1",
		"1024x1024",
		"png",
		"image/png",
		smallPNG(t),
	)
	require.ErrorIs(t, err, ErrImageStudioGalleryQuotaExceeded)
	require.Empty(t, saver.key)
}

func smallPNG(t *testing.T) []byte {
	t.Helper()
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0xf0,
		0x1f, 0x00, 0x05, 0x00, 0x01, 0xff, 0x89, 0x99,
		0x3d, 0x1d, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
		0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
