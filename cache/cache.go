package cache

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type PageState int

const (
	HIT PageState = iota
	STALE
	MISS
)

type Cache struct {
	cache          *cache.Cache
	staleInterval  time.Duration
	expiryDuration time.Duration
}

func NewCache(cfg CacheConfig) *Cache {
	if cfg.DisableCache {
		return &Cache{}
	}
	return &Cache{
		staleInterval:  time.Duration(cfg.StaleInterval),
		expiryDuration: time.Duration(cfg.DefaultExpiry),
		cache:          cache.New(time.Duration(cfg.DefaultExpiry), time.Duration(cfg.CleanupInterval)),
	}
}

func (c *Cache) GetPage(host, path string) ([]byte, PageState) {
	page, expiry, found := c.cache.GetWithExpiration(fmt.Sprintf("%s%s", host, path))
	if found {
		page := page.([]byte)
		if c.isFresh(expiry, c.staleInterval) {
			return page, HIT
		}
		return page, STALE
	}
	return nil, MISS
}

func (c *Cache) CachePage(host, path string, page []byte) {
	c.cache.SetDefault(fmt.Sprintf("%s%s", host, path), page)
}

func (c *Cache) UpdatePage(host, path string, page []byte) {
	c.cache.Replace(fmt.Sprintf("%s%s", host, path), page, cache.DefaultExpiration)
}

func (c *Cache) ClearPage(host, path string) {
	c.cache.Delete(fmt.Sprintf("%s%s", host, path))
}

func (c *Cache) isFresh(expiry time.Time, staleInterval time.Duration) bool {
	timeUntilExpiry := expiry.Unix() - time.Now().Unix()
	timeElapsed := int64(c.expiryDuration.Seconds()) - timeUntilExpiry
	return timeElapsed < int64(staleInterval.Seconds())
}
