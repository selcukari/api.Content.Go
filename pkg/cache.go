package pkg

import (
	"sync"
	"time"
)

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

type Cache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

// Global cache instance'ı
var GlobalCache = NewCache()

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]cacheItem),
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var expiresAt time.Time

	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: expiresAt,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.data[key]
	if !ok {
		return nil, false
	}
	// Süresi dolmuş mu kontrol et
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.mu.RUnlock()
		// Süresi dolduğu için silme işlemini yazma kilidi (Lock) ile yapalım
		c.Delete(key)
		return nil, false
	}

	c.mu.RUnlock()
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
}
