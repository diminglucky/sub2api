package service

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	openAIModelsCacheBodyLimit = codexModelsManifestCacheBodyLimit
	openAIModelsCacheMaxEntries = codexModelsManifestCacheMaxEntries
	openAIModelsCacheTTL = codexModelsManifestCacheTTL
	openAIModelsCacheStaleTTL = codexModelsManifestCacheStaleTTL
)

// OpenAIModelsResponse is retained for the standard /v1/models handlers while
// CodexModelsManifest is used by the newer manifest APIs.
type OpenAIModelsResponse = CodexModelsManifest

type openAIModelsCacheEntry struct {
	manifest  *OpenAIModelsResponse
	order     uint64
	expiresAt time.Time
	staleUntil time.Time
}

type openAIModelsCacheState uint8

const (
	openAIModelsCacheMiss openAIModelsCacheState = iota
	openAIModelsCacheFresh
	openAIModelsCacheStale
)

type openAIModelsCache struct {
	mu sync.Mutex
	entries map[string]openAIModelsCacheEntry
	nextOrder uint64
	refresh singleflight.Group
}

func (c *openAIModelsCache) get(key string, now time.Time) (*OpenAIModelsResponse, openAIModelsCacheState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, openAIModelsCacheMiss
	}
	if !now.Before(entry.staleUntil) {
		delete(c.entries, key)
		return nil, openAIModelsCacheMiss
	}
	if now.Before(entry.expiresAt) {
		return entry.manifest, openAIModelsCacheFresh
	}
	return entry.manifest, openAIModelsCacheStale
}

func (c *openAIModelsCache) set(key string, manifest *OpenAIModelsResponse, now time.Time) {
	if manifest == nil || len(manifest.Body) > codexModelsManifestCacheBodyLimit {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]openAIModelsCacheEntry)
	}
	c.nextOrder++
	c.entries[key] = openAIModelsCacheEntry{manifest: manifest, order: c.nextOrder, expiresAt: now.Add(codexModelsManifestCacheTTL), staleUntil: now.Add(codexModelsManifestCacheStaleTTL)}
}
