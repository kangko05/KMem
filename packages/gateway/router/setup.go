package router

import (
	"gateway/cache"
	"gateway/middlewares"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func New(c *cache.Cache, dur time.Duration) *gin.Engine {
	r := gin.Default()

	r.Use(middlewares.Cache(c, dur))

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	return r
}
