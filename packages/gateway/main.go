package main

import (
	"context"
	"gateway/cache"
	"gateway/router"
	"gateway/utils"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := cache.New()
	r := router.New(c, time.Hour)

	go c.Run(ctx)

	r.Run(utils.PORT)
}
