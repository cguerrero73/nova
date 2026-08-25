package cache

import (
	"fmt"
	"sync"
	"time"
)

// LayoutCache provides an in-process cache for resolved layout JSON.
// Key format: formbuilder:layout:{role_name}:{version_id}
// TTL: 10 minutes.
type LayoutCache struct {
	data sync.Map
	ttl  time.Duration
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewLayoutCache creates a new LayoutCache with the given TTL.
func NewLayoutCache(ttl time.Duration) *LayoutCache {
	return &LayoutCache{ttl: ttl}
}

// NewDefaultLayoutCache creates a LayoutCache with the default 10-minute TTL.
func NewDefaultLayoutCache() *LayoutCache {
	return NewLayoutCache(10 * time.Minute)
}

// Key builds a cache key from role name and version ID.
func Key(roleName string, versionID int64) string {
	return fmt.Sprintf("formbuilder:layout:%s:%d", roleName, versionID)
}

// Get retrieves a cached layout definition. Returns nil on miss or expiry.
func (c *LayoutCache) Get(key string) []byte {
	val, ok := c.data.Load(key)
	if !ok {
		return nil
	}
	entry := val.(cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.data.Delete(key)
		return nil
	}
	return entry.value
}

// Set stores a layout definition in the cache with the configured TTL.
func (c *LayoutCache) Set(key string, value []byte) {
	c.data.Store(key, cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	})
}

// Invalidate removes a specific key from the cache.
func (c *LayoutCache) Invalidate(key string) {
	c.data.Delete(key)
}

// Clear removes all entries from the cache.
func (c *LayoutCache) Clear() {
	c.data.Range(func(key, value interface{}) bool {
		c.data.Delete(key)
		return true
	})
}
