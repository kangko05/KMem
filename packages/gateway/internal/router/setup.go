package router

import (
	"context"
	"fmt"
	"gateway/internal/cache"
	"gateway/internal/middlewares"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type services struct {
	cache *cache.Cache
}

func initServices(ctx context.Context) (*services, error) {
	c := cache.New()

	go c.Run(ctx)
	// go uq.Run(ctx)

	return &services{cache: c}, nil
}

func New(ctx context.Context, dur time.Duration) (*gin.Engine, error) {
	serv, err := initServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init services: %v", err)
	}

	r := gin.Default()

	// middlewares
	r.Use(middlewares.Cache(serv.cache, dur))

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	// connecting with other services here - going to call grpc functions from its handlers
	setupFileApi(r)
	setupAuthApi(r)

	return r, nil
}
