package cache

import (
	"context"
	"gateway/utils"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CacheItem struct {
	Value      any
	Expiration int64
}

type Cache struct {
	items sync.Map
}

func New() *Cache {
	return &Cache{
		items: sync.Map{},
	}
}

func (c *Cache) GenKey(ctx *gin.Context) string {
	return utils.MD5(ctx.Request.URL.RawPath)
}

// key is hashed with md5
func (c *Cache) Set(key string, value any, duration time.Duration) {
	key = utils.MD5(key)

	c.items.Store(key, CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration).UnixNano(),
	})
}

func (c *Cache) Get(key string) (CacheItem, bool) {
	v, exists := c.items.Load(utils.MD5(key))
	if exists {
		cit := v.(CacheItem) // this should be ok because only thing that goes into cache will be cache item

		if cit.Expiration < time.Now().UnixNano() {
			return CacheItem{}, false
		}

		return cit, exists
	}

	return CacheItem{}, false
}

func (c *Cache) Delete(key string) {
	c.items.Delete(utils.MD5(key))
}

// run this as a goroutine
func (c *Cache) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping cache process")
			return
		case <-ticker.C:
			c.cleanUp()
		}
	}
}

func (c *Cache) cleanUp() {
	now := time.Now().UnixNano()

	c.items.Range(func(key, value any) bool {
		it := value.(CacheItem)

		if it.Expiration < now {
			c.items.Delete(key)
		}

		return true
	})
}

func (c *Cache) Items() *sync.Map {
	return &c.items
}
