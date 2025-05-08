package main

import (
	"context"
	"gateway/internal/router"
	"gateway/internal/utils"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := router.New(ctx, time.Hour)
	if err != nil {
		log.Fatal(err)
	}

	if err := r.Run(utils.PORT); err != nil {
		log.Fatal(err)
	}
}
