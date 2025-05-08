package tests

import (
	"fmt"
	"gateway/internal/cache"
	"gateway/internal/middlewares"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("Set & Get items", func(t *testing.T) {
		assert := assert.New(t)

		cache := cache.New()
		go cache.Run(t.Context())

		keys := []string{"first", "second", "third"}

		for i, key := range keys {
			cache.Set(key, i+1, 30*time.Second)
		}

		for i, key := range keys {
			cit, ok := cache.Get(key)

			assert.True(ok)
			assert.Equal(cit.Value, i+1)
		}
	})

	t.Run("Delete itmes", func(t *testing.T) {
		assert := assert.New(t)

		var empty cache.CacheItem
		cache := cache.New()
		go cache.Run(t.Context())

		keys := []string{"first", "second", "third"}

		for i, key := range keys {
			cache.Set(key, i+1, 30*time.Second)
		}

		for _, key := range keys {
			cache.Delete(key)
		}

		for _, key := range keys {
			cit, ok := cache.Get(key)
			assert.False(ok)
			assert.Equal(cit, empty)
		}
	})

	t.Run("test on multiple goroutines", func(t *testing.T) {
		assert := assert.New(t)

		nGoroutines := 100

		ch := make(chan int, nGoroutines)
		for i := range nGoroutines {
			ch <- i
		}
		close(ch)

		c := cache.New()
		go c.Run(t.Context())

		var wg sync.WaitGroup
		wg.Add(nGoroutines)

		for i := range nGoroutines {
			go func() {
				defer wg.Done()
				c.Set(fmt.Sprint(i), i, time.Second*30)
				time.Sleep(time.Second)
			}()
		}

		wg.Wait()

		for i := range nGoroutines {
			v, ok := c.Get(fmt.Sprint(i))
			assert.True(ok)
			assert.Equal(v.Value, i)
		}
	})

	t.Run("test expiration", func(t *testing.T) {
		assert := assert.New(t)
		c := cache.New()
		go c.Run(t.Context())

		exp := "expiring-item"

		c.Set(exp, 123, time.Second*3)

		time.Sleep(time.Second)

		v, ok := c.Get(exp)
		assert.True(ok)
		assert.Equal(v.Value, 123)

		// time doesn't have to be exact
		time.Sleep(3 * time.Second)
		v, ok = c.Get(exp)
		assert.False(ok)
		assert.Equal(v, cache.CacheItem{})
	})
}

func TestCacheResponse(t *testing.T) {
	assert := assert.New(t)

	c := cache.New()
	go c.Run(t.Context())

	r := gin.New()
	r.Use(middlewares.Cache(c, time.Hour))

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	req1, _ := http.NewRequest("GET", "/ping", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(http.StatusOK, w1.Code)

	found := false
	c.Items().Range(func(key, value any) bool {
		k := key.(string)
		v := value.(cache.CacheItem)
		fmt.Println("key:", k)
		fmt.Printf("value: %+v\n", v.Value)
		found = true
		return true
	})

	assert.True(found, "item must be stored in cache")

	req2, _ := http.NewRequest("GET", "/ping", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(http.StatusOK, w2.Code)
}
