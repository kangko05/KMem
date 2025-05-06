package middlewares

import (
	"bytes"
	"encoding/json"
	"gateway/cache"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func Cache(c *cache.Cache, duration time.Duration) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// only cache GET request
		if ctx.Request.Method != "GET" {
			ctx.Next()
			return
		}

		cacheKey := c.GenKey(ctx)

		cit, exists := c.Get(cacheKey)
		if exists {
			mapVal, ok := cit.Value.(map[string]any)
			if ok {
				ctx.JSON(http.StatusOK, mapVal)
			} else {
				ctx.JSON(http.StatusOK, gin.H{"data": cit})
			}

			ctx.Abort()
			return
		}

		writer := &responseBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}

		ctx.Writer = writer

		ctx.Next()

		if ctx.Writer.Status() == http.StatusOK {
			var response map[string]any
			if err := json.Unmarshal(writer.body.Bytes(), &response); err == nil {
				c.Set(cacheKey, response, duration)
			}
		}
	}
}
